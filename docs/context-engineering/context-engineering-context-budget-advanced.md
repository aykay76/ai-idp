# Context Engineering Part 4: Advanced Context Budgeting

This page answers the questions teams usually hit after they adopt a fixed starter budget: how to tune budgets by risk and task type, how to compress safely, and how to observe when the budget is hurting quality or wasting capacity.

For a concrete build flow and starter numbers, see [Part 3 - practical Go example](context-engineering-go-example.md).

---

## Why Budget Management Matters

Context windows and gateway limits are finite. Poor budget management shows up as:

- Truncated constraints or acceptance criteria (quality and safety risk)
- Bloated history and duplicate sources (cost and latency)
- Silent summarisation without traceability (trust risk)
- No telemetry, so teams cannot tell whether failures are model, prompt, or context

Budget management is an engineering discipline: allocate, measure, trim, escalate.

---

## Core Concepts

### Total envelope

Split any call into explicit buckets:

1. **System and instructions** - fixed overhead (tool definitions, policy text, formatting)
2. **Context package** - your curated Jira, Confluence, GitHub material
3. **User input** - follow-up questions and clarifications
4. **Model output reserve** - room for the answer you need

Never allocate 100% of the window to context. Always reserve for output and user follow-up.

### Section budgets

Assign each logical section (Jira snapshot, Confluence excerpt, file bundle) a **min**, **target**, and **max** token budget. Trimming moves from max toward min in priority order.

### Estimation vs measurement

- **Estimation:** fast, good enough for trimming heuristics (character-based ratios, tokeniser where available).
- **Measurement:** use gateway or provider token counts when the API returns them; use these to calibrate estimators over time.

Treat estimation as a product feature with confidence bounds, not a single magic number.

---

## Risk-Tier Budget Profiles

Align budget with how wrong an answer can be. These are example profiles; calibrate to your gateway limits and typical prompt overhead.

### High risk (production, incidents, security-sensitive)

Goals: maximise correctness and traceability; accept higher latency and cost.

- Larger **context** slice; smaller compromise on constraints and conflicts
- Prefer **extractive** snippets (quoted excerpts, file ranges) over aggressive summarisation
- Stricter **freshness** requirements; refetch if stale
- **Escalate** early if budget cannot fit acceptance criteria and implementation truth

Illustrative split on a `32k` usable envelope (adjust for your model and system overhead):

- System/instructions: `4k`
- Context package: `18k`
- User input reserve: `4k`
- Output reserve: `6k`

### Medium risk (delivery planning, design reviews, implementation plans)

Goals: balance completeness with cost.

- Standard context slice; summarise long Confluence history
- GitHub: prefer diff summaries and key files over full trees

Illustrative split on `32k`:

- System/instructions: `3k`
- Context package: `14k`
- User input reserve: `5k`
- Output reserve: `10k`

### Low risk (drafts, onboarding, exploration)

Goals: speed and cheap iteration; clearly label draft outputs.

- Smaller context slice; more summarisation
- Allow older snapshots with explicit caveats

Illustrative split on `32k`:

- System/instructions: `3k`
- Context package: `10k`
- User input reserve: `6k`
- Output reserve: `13k`

---

## Dynamic Reallocation by Task Type

Static per-section budgets work until they do not. Use **task type** to shift weight between sources.

| Task type | Favor | De-prioritise |
|-----------|--------|----------------|
| Incident response | Runbooks, recent commits, active PRs, Jira blocker state | Long Confluence background |
| Feature implementation plan | Jira AC, GitHub default branch, ADR pages | Full git history |
| PR review assist | Diff, tests, CI, linked Jira | Broad wiki search |
| Architecture refresh | Confluence ADRs, diagrams, service map | Raw large logs |

Implementation sketch:

1. Classify `task_type` and `risk_tier`.
2. Load a **budget profile** (high/medium/low).
3. Apply **task weights** to section targets (e.g. +30% GitHub for PR review).
4. Run **normalisation** then **conflict detection** before trimming (never trim away disclosed conflicts without escalation).

---

## Compression Strategy Matrix

Choose a strategy per section based on content type and sensitivity.

| Content | Prefer | Avoid when |
|---------|--------|------------|
| Acceptance criteria, legal/security constraints | No summarisation; full text or escalate | Never truncate for budget alone |
| ADR / policy excerpt | Extractive: heading + key paragraphs + link | Abstractive summary without citations |
| Long Confluence page | Section extract + TOC-guided chunks | Dumping entire page |
| Large file / module | Symbol outline + changed hunks + tests | Whole-file paste unless tiny |
| Git history | Last N commits + PR description | Full blame for entire repo |
| Duplicate facts across Jira/GitHub/Confluence | Single canonical bullet + three links | Repeating the same paragraph thrice |

**Extractive** = selected spans from source (high fidelity, traceable).  
**Abstractive** = model-generated summary (efficient, higher drift risk). Use only when extractive budget is exhausted and you still have references.

---

## Decision Rules: Summarise, Truncate, or Escalate

Use an explicit ordered policy so behavior is predictable.

1. **Escalate** if any of the following cannot fit in **min** budget: acceptance criteria, explicit conflicts, blockers, or named security constraints.
2. **Extractive trim:** drop lowest-priority subsections (old comments, tangential pages) while keeping links.
3. **Summarise** only for sections marked `summarisation_allowed` in your policy.
4. **Truncate** raw text only with visible markers (`[... truncated ...]`) and a link to source; never truncate mid-sentence on legal or numeric values.
5. **Refetch or split request:** if two passes still exceed budget, split into two LLM calls (e.g. "gather evidence" then "synthesise") instead of one overloaded call.

---

## Cost, Latency, and Quality Trade-offs

| Lever | Cost/latency | Quality impact |
|-------|----------------|----------------|
| Larger context | Higher input tokens, slower | Better recall, more contradictions to resolve |
| More summarisation | Lower tokens, faster | Higher risk of subtle errors |
| Two-step pipeline | Two round trips | Often better quality per token |
| Smaller model for summarisation | Cheaper | Summaries may lose nuance; validate |
| Caching snapshots | Cheaper repeat calls | Stale risk; tie to TTL |

A common pattern: **cheap summarisation pass** (internal or small model) to produce a structured evidence pack, then **main model** for reasoning on the smaller pack. Document where summarisation happened so reviewers can audit.

---

## Observability and Budget Drift

Instrument context building like any other service.

### Metrics (examples)

- `context_tokens_total` - estimated or actual input context size
- `context_tokens_by_section` - breakdown (jira, confluence, github, user)
- `context_trim_events_total` - count of trim/summarise steps
- `context_escalations_total` - could not fit min policy
- `llm_latency_seconds` and `llm_cost_estimate` - correlated with context size

### Logs (examples)

Each request should log a compact **budget report**: profile name, per-section before/after tokens, trim actions, and links to sources.

### Alerts (examples)

- Spike in `context_escalations_total` for a workflow (policy too tight or sources too noisy)
- High trim rate on sections that should rarely trim (e.g. acceptance criteria)
- Growing gap between estimated and measured tokens (estimator drift)

---

## FAQ-Style Questions Teams Ask

**Should we maximise context every time?**  
No. Unused budget is cheaper than wrong budget. Reserve output space and prefer a second targeted call over one giant prompt.

**Is character count divided by four a good token estimate?**  
Rough starting point for English prose; poor for code, JSON, and non-Latin text. Calibrate per workflow with measured samples.

**Who owns the budget policy?**  
Platform or enablement team owns defaults; product teams own overrides per workflow with review.

**How do we avoid "summary hallucination"?**  
Prefer extractive snippets with links; if you summarise, require a "sources consulted" list and forbid introducing facts not present in sources.

---

## Deliverables Checklist for Advanced Rollout

- [ ] Documented budget profiles per `risk_tier`
- [ ] Task-type weight table and code/config mapping
- [ ] Section min/target/max table per workflow
- [ ] Compression matrix and allowed summarisation sections
- [ ] Escalation rules when min budget cannot be met
- [ ] Metrics, logs, and one dashboard or query for budget health

---

## Related Pages

- [Context engineering 101](context-engineering.md)
- [Part 2 - Maersk lens](context-engineering-maersk-lens.md)
- [Part 3 - practical Go example](context-engineering-go-example.md)
- [Part 5 - progressive disclosure](context-engineering-progressive-disclosure.md)
