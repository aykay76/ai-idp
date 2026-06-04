# Context Engineering SDK (Go)

> **Status:** Early adoption  
> **Audience:** Engineers building AI-enabled services and agents  
> **Maintained by:** AI Adoption – Context Engineering Workstream

---

## 1. Purpose

This SDK provides **reusable Go primitives for assembling, validating, and rendering context** before invoking an LLM.

It exists to:
- Reduce variance across AI-enabled services
- Make AI behaviour predictable and reviewable
- Enable safe reuse of skills and agents
- Turn context from “prompt text” into an engineered input

This SDK is intended for **production systems**, not one-off experiments or demos.

---

## 2. Scope, Responsibilities, and Boundaries

This SDK focuses exclusively on **context handling** for AI-enabled systems.

### What This SDK Owns

The SDK is responsible for:
- Representing context as typed, structured inputs
- Assembling context from multiple sources
- Applying explicit ordering, precedence, and composition rules
- Validating context against declared contracts
- Managing token budgets before model invocation
- Rendering context into model-specific request formats
- Preserving provenance and metadata for observability

### What This SDK Does Not Own

The SDK intentionally does **not** provide or replace:
- Model hosting, routing, or selection
- LLM gateways or providers
- Agent orchestration or workflows
- RAG pipelines or vector stores
- Data ingestion or ETL
- Tool execution frameworks
- Guardrail policy definition or enforcement
- Long-running state management

These concerns are owned elsewhere and integrate *around* this SDK.

### Why These Boundaries Exist

Clear ownership boundaries:
- Prevent duplication with AI platform capabilities
- Avoid hidden coupling between systems
- Keep the SDK stable, composable, and reusable
- Allow teams to adopt incrementally without architectural lock-in

---

## 3. Design Principles

- **Explicit over implicit**  
  Context composition is visible and deterministic.

- **Strongly typed inputs**  
  Context is structured data, not concatenated strings.

- **Composable, not prescriptive**  
  Teams assemble what they need; nothing is forced.

- **Model-agnostic by default**  
  Rendering adapts to the target interface, not the other way around.

- **Predictability over cleverness**  
  Prefer boring correctness to smart-but-fragile abstractions.

---

## 4. Conceptual Model (SDK View)

At the SDK level, an AI request is treated as:

```

Context Sources
↓
Context Assembly
↓
Contract Validation
↓
Token Budgeting
↓
Rendering
↓
LLM Invocation (outside this SDK)

```

This separation keeps responsibilities clear and testable.

---

## 5. Core SDK Capabilities

### 5.1 Context Blocks

- Typed, named units of context
- Explicit ownership and intent
- Optional mutability and lifetime rules
- Associated metadata (source, freshness, risk)

### 5.2 Context Sources

- Pluggable providers (code, files, snapshots, APIs)
- Clear separation between raw data and curated context
- No assumptions about storage or retrieval mechanisms

### 5.3 Context Assembly & Composition

- Deterministic ordering
- Priority and precedence rules
- Explicit inclusion / exclusion
- Conflict visibility rather than silent override

### 5.4 Context Contracts

Contracts define:
- Required context blocks
- Optional blocks
- Expected semantics (not just shape)
- Degradation behaviour when context is missing
- Version compatibility rules

Contracts are enforced before model calls.

### 5.5 Token Budgeting

- Estimate token usage before rendering
- Enforce hard and soft limits
- Enable predictable cost and latency profiles
- Avoid model-side truncation as a failure mode

### 5.6 Rendering

- Convert assembled context into model-compatible request formats
- Support multiple rendering strategies without changing upstream logic
- Preserve traceability from rendered output back to context blocks

---

## 6. Context Patterns Support

The SDK supports reusable **context composition patterns**, including (but not limited to):
- Policy sandwich patterns
- Read-only context fences
- Thin prompt / fat context
- Execution vs explanation context
- Preflight validation patterns

Patterns are guidance and examples, not mandatory frameworks.

---

## 7. Reference Implementations

The repository includes **intentionally boring reference services** demonstrating:
- Correct context assembly
- Contract enforcement
- Pattern usage
- Predictable token behaviour
- Observable provenance

These exist to answer:
> “What does a correct implementation look like?”

They are not intended as production templates.

---

## 8. Adoption: What It Means

### Adoption Criteria

A team or service is considered to have **adopted the Context Engineering SDK** when:

1. ✅ Context is assembled using SDK primitives (not raw prompt strings)
2. ✅ Context is explicitly typed and named
3. ✅ At least one AI interaction enforces a context contract
4. ✅ Token budgeting occurs before making the model call
5. ✅ Context provenance is observable (logging or tracing)

### What Adoption Does *Not* Mean

- ❌ Rewriting existing agents end-to-end
- ❌ Using every SDK feature
- ❌ Central approval or gating
- ❌ Loss of autonomy over models or architecture choices

Adoption is incremental and opt-in.

---

## 9. When You Should Use This SDK

Use this SDK when you:
- Build an AI-enabled service or agent
- Share skills across teams
- Need stable, explainable AI behaviour
- Expect solutions to evolve over time
- Care about operability and reviewability

You likely do *not* need it for:
- Throwaway experiments
- One-off prompts
- Demos with no production intent

---

## 10. Non-Goals

To keep the SDK sustainable and focused, we explicitly avoid:
- Becoming an agent framework
- Encoding business logic in abstractions
- Competing with AI platform capabilities
- Introducing hidden behaviour or magic defaults
- Centralising control over how teams use AI

---

## 11. Getting Started

- Start with `/examples` for minimal integrations
- Review `/reference-agents` for end-to-end usage
- Introduce contracts first, then refine patterns
- Optimise for clarity and debuggability over cleverness
