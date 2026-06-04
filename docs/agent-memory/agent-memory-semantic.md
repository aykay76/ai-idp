# Agent Memory Part 4: Semantic Memory

Semantic memory is the agent's long-term knowledge store. Unlike episodic memory — which records what *happened* — semantic memory holds what is *known*: facts, decisions, policies, and domain knowledge that persist indefinitely and apply across many sessions and many agents.

It is the closest analogue to how a skilled engineer accumulates expertise: not by replaying every project from memory, but by recalling relevant facts on demand.

---

## What Belongs in Semantic Memory

The write discipline for semantic memory is strict. Write too little and agents rediscover the same knowledge repeatedly. Write too much and the store fills with noise that degrades retrieval quality.

**Write:**

- Factual conclusions — "Service X owns table Y"
- Architectural decisions — "We chose pattern Z for reason R"
- User or team preferences — "Team A always deploys to staging before production"
- Reference knowledge — runbook excerpts, API behaviour notes, data contracts
- Resolved ambiguities — "Field X means Y in this context, not Z"

**Do not write:**

- Intermediate reasoning steps
- Raw tool outputs (store these in working memory; distil them before writing to semantic)
- Opinions without supporting evidence
- Information that can be freshly fetched cheaply on demand

---

## Schema

```sql
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE semantic_memory (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    namespace    TEXT        NOT NULL,           -- scoping: agent, user, team (see Part 5)
    key          TEXT,                           -- optional stable key for exact lookup
    content      TEXT        NOT NULL,
    kind         TEXT        NOT NULL,           -- fact | decision | preference | reference
    source       TEXT,                           -- where this came from
    confidence   REAL        NOT NULL DEFAULT 1.0, -- 0.0–1.0; lower = uncertain
    stored_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at   TIMESTAMPTZ,                    -- NULL = never expires
    superseded   BOOLEAN     NOT NULL DEFAULT FALSE,
    embedding    vector(1536) NOT NULL
);

CREATE INDEX ON semantic_memory USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);
CREATE INDEX ON semantic_memory (namespace, kind);
CREATE INDEX ON semantic_memory (key) WHERE key IS NOT NULL;
CREATE INDEX ON semantic_memory (expires_at) WHERE expires_at IS NOT NULL;
```

---

## Go Implementation

### Types

```go
import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/pgvector/pgvector-go"
)

type MemoryRecord struct {
    ID         uuid.UUID
    Namespace  string
    Key        string
    Content    string
    Kind       string     // fact | decision | preference | reference
    Source     string
    Confidence float32
    StoredAt   time.Time
    ExpiresAt  *time.Time
}

type SemanticMemory struct {
    db       *pgxpool.Pool
    embedder EmbeddingClient
}
```

### Writing Knowledge

```go
func (s *SemanticMemory) Write(ctx context.Context, rec MemoryRecord) error {
    embedding, err := s.embedder.Embed(ctx, rec.Content)
    if err != nil {
        return fmt.Errorf("embedding content: %w", err)
    }

    // When a stable key is provided, supersede any existing record for that key
    // before inserting so that Write is idempotent by key.
    if rec.Key != "" {
        if err := s.SupersedeByKey(ctx, rec.Namespace, rec.Key); err != nil {
            return fmt.Errorf("superseding previous record: %w", err)
        }
    }

    _, err = s.db.Exec(ctx, `
        INSERT INTO semantic_memory
            (namespace, key, content, kind, source, confidence, expires_at, embedding)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `,
        rec.Namespace,
        nullableString(rec.Key),
        rec.Content,
        rec.Kind,
        nullableString(rec.Source),
        rec.Confidence,
        rec.ExpiresAt,
        pgvector.NewVector(embedding),
    )
    return err
}

func nullableString(s string) any {
    if s == "" {
        return nil
    }
    return s
}
```

### Semantic Search

```go
type SearchQuery struct {
    Namespace string
    Query     string
    Kind      string         // optional — filter by knowledge kind
    Limit     int
    MaxAge    *time.Duration // optional — exclude records older than this
}

func (s *SemanticMemory) Search(ctx context.Context, q SearchQuery) ([]MemoryRecord, error) {
    embedding, err := s.embedder.Embed(ctx, q.Query)
    if err != nil {
        return nil, fmt.Errorf("embedding query: %w", err)
    }

    if q.Limit == 0 {
        q.Limit = 5
    }

    // Build dynamic filter clauses to exclude expired and superseded records.
    args := []any{q.Namespace, pgvector.NewVector(embedding), q.Limit}
    filter := `
        AND superseded = FALSE
        AND (expires_at IS NULL OR expires_at > now())`

    if q.Kind != "" {
        args = append(args, q.Kind)
        filter += fmt.Sprintf(" AND kind = $%d", len(args))
    }
    if q.MaxAge != nil {
        cutoff := time.Now().Add(-*q.MaxAge)
        args = append(args, cutoff)
        filter += fmt.Sprintf(" AND stored_at > $%d", len(args))
    }

    rows, err := s.db.Query(ctx, fmt.Sprintf(`
        SELECT id, namespace, key, content, kind, source, confidence, stored_at, expires_at
        FROM semantic_memory
        WHERE namespace = $1 %s
        ORDER BY embedding <=> $2
        LIMIT $3
    `, filter), args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    return scanRecords(rows)
}
```

### Exact Key Lookup

```go
func (s *SemanticMemory) GetByKey(ctx context.Context, namespace, key string) (*MemoryRecord, error) {
    rows, err := s.db.Query(ctx, `
        SELECT id, namespace, key, content, kind, source, confidence, stored_at, expires_at
        FROM semantic_memory
        WHERE namespace = $1
          AND key = $2
          AND superseded = FALSE
          AND (expires_at IS NULL OR expires_at > now())
        ORDER BY stored_at DESC
        LIMIT 1
    `, namespace, key)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    records, err := scanRecords(rows)
    if err != nil {
        return nil, err
    }
    if len(records) == 0 {
        return nil, nil
    }
    return &records[0], nil
}
```

---

## Retrieval as a Tool

For an agent to search semantic memory, expose it as a named tool. The description shapes write discipline — an LLM calling this tool is exercising the decision of what is worth persisting.

```go
func SemanticRetrieveTool(mem *SemanticMemory, namespace string) ToolDefinition {
    return ToolDefinition{
        Name: "retrieve_knowledge",
        Description: `Search long-term memory for facts, decisions, preferences, or reference knowledge.
Use this before attempting to discover information that may already be known.
Retrieval is cheap — prefer it over re-investigating the same question.`,
        Parameters: map[string]ParameterDef{
            "query": {
                Type:        "string",
                Description: "What you want to know, in natural language.",
            },
            "kind": {
                Type:        "string",
                Enum:        []string{"fact", "decision", "preference", "reference"},
                Description: "Optional. Filter by knowledge kind.",
            },
            "key": {
                Type:        "string",
                Description: "Optional. If you know the exact stable key, use this for a direct lookup.",
            },
        },
        Handler: func(ctx context.Context, params map[string]any) (string, error) {
            if key, ok := params["key"].(string); ok && key != "" {
                rec, err := mem.GetByKey(ctx, namespace, key)
                if err != nil {
                    return "", err
                }
                if rec == nil {
                    return "No record found for that key.", nil
                }
                return formatRecord(*rec), nil
            }

            q := SearchQuery{
                Namespace: namespace,
                Query:     params["query"].(string),
                Limit:     5,
            }
            if kind, ok := params["kind"].(string); ok {
                q.Kind = kind
            }
            records, err := mem.Search(ctx, q)
            if err != nil {
                return "", err
            }
            if len(records) == 0 {
                return "No relevant knowledge found.", nil
            }
            return formatRecords(records), nil
        },
    }
}
```

## Write-Back as a Tool

```go
func SemanticWriteTool(mem *SemanticMemory, namespace string) ToolDefinition {
    return ToolDefinition{
        Name: "store_knowledge",
        Description: `Persist a fact, decision, or conclusion to long-term memory.
Call this only for information that is:
- Factually confirmed (not speculated)
- Not cheaply derivable by fetching live data again
- Useful to future sessions or other agents

Do NOT store: intermediate reasoning, raw tool outputs, or uncertain guesses.`,
        Parameters: map[string]ParameterDef{
            "key": {
                Type:        "string",
                Description: `A stable, descriptive key. For example: "service/payment-api/db-owner" or "incident/PAY-2024-11-root-cause". Omit if no stable key applies.`,
            },
            "content": {
                Type:        "string",
                Description: "The fact or decision to store. Be specific and concise.",
            },
            "kind": {
                Type:        "string",
                Enum:        []string{"fact", "decision", "preference", "reference"},
                Description: "Classification of the knowledge.",
            },
            "source": {
                Type:        "string",
                Description: `Where this knowledge came from. For example: "jira:PAY-1234" or "incident-investigation-2026-05-28".`,
            },
            "expires_in_days": {
                Type:        "integer",
                Description: "Optional. Set for volatile knowledge (deployments, rosters). Omit for permanent facts.",
            },
        },
        Handler: func(ctx context.Context, params map[string]any) (string, error) {
            rec := MemoryRecord{
                Namespace:  namespace,
                Content:    params["content"].(string),
                Kind:       params["kind"].(string),
                Confidence: 1.0,
            }
            if k, ok := params["key"].(string); ok {
                rec.Key = k
            }
            if src, ok := params["source"].(string); ok {
                rec.Source = src
            }
            if d, ok := params["expires_in_days"].(float64); ok && d > 0 {
                exp := time.Now().Add(time.Duration(d) * 24 * time.Hour)
                rec.ExpiresAt = &exp
            }
            if err := mem.Write(ctx, rec); err != nil {
                return "", err
            }
            return "Knowledge stored.", nil
        },
    }
}
```

---

## Staleness and Invalidation

All stored knowledge carries a `stored_at` timestamp. Agents should factor age into how much they trust retrieved knowledge.

| Age | Knowledge class | Agent behaviour |
|---|---|---|
| < 1 day | Operational data (versions, on-call rosters) | Use directly |
| 1–30 days | Architectural decisions | Use, but verify before acting |
| > 90 days | Stable policies and long-lived facts | Valid; low staleness risk |
| Any age | Record with `superseded = TRUE` | Excluded automatically from all retrieval |

**Explicit invalidation** is preferred over silent overwrites. When an agent determines that a prior record is wrong, it marks it superseded and writes the correction as a new record. This preserves the history of what the agent believed and when the belief changed.

```go
func (s *SemanticMemory) Supersede(ctx context.Context, id uuid.UUID) error {
    _, err := s.db.Exec(ctx, `
        UPDATE semantic_memory SET superseded = TRUE WHERE id = $1
    `, id)
    return err
}

func (s *SemanticMemory) SupersedeByKey(ctx context.Context, namespace, key string) error {
    _, err := s.db.Exec(ctx, `
        UPDATE semantic_memory SET superseded = TRUE
        WHERE namespace = $1 AND key = $2
    `, namespace, key)
    return err
}
```

---

## Chunk Design

Retrieval quality depends heavily on how knowledge is chunked. A chunk that mixes several topics either matches too broadly or too narrowly.

| Principle | Poor | Good |
|---|---|---|
| One idea per chunk | "Service X owns table Y and team A deploys on Fridays" | Two separate records |
| Specific keys | `"note-about-payment"` | `"service/payment-api/database-owner"` |
| Concrete content | "The payment service has some issues" | "Payment service returns 409 on duplicate idempotency keys and does not retry" |
| Source attribution | No source field | `source: "incident/PAY-2024-11-23"` |
| Appropriate TTL | No expiry on deployment version | `expires_in_days: 7` for version records |

---

## Putting It Together

The agent uses `retrieve_knowledge` before calling expensive downstream tools. For a well-populated semantic store, a significant proportion of tool calls can be short-circuited.

```go
func (a *Agent) buildTools() []Tool {
    return []Tool{
        // Domain tools (API calls, database queries, etc.)...
        SemanticRetrieveTool(a.semanticMem, a.namespace),
        SemanticWriteTool(a.semanticMem, a.namespace),
    }
}

// An agent ReAct trace using semantic memory:
//
// Thought: I need to know who owns the payments database.
// Action: retrieve_knowledge(query="payments database owner", kind="fact")
// Observation: [fact] Service payment-api owns schema payments.
//              Source: decision/2025-Q3-data-ownership. Stored: 2025-09-01.
// Thought: I have the answer. No need to query the schema registry.
// Action: <next step>
```

---

## What Comes Next

| Part | Topic |
|---|---|
| [Part 5: Memory Scoping](./agent-memory-scoping.md) | Agent vs user vs team, namespacing, multi-agent sharing |
