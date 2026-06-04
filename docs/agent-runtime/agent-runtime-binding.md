# Agent Runtime Binding (Draft Design)

## Purpose

Runtime binding resolves a portable agent contract into executable runtime configuration.

It is platform-owned and environment-aware.

## Responsibilities

- Resolve model routing from logical tier/profile to concrete provider/model.
- Bind abstract tool permissions to concrete MCP servers and credentials.
- Attach memory providers and retention behavior.
- Apply retry/timeouts and execution safeguards.
- Bind and enforce cost policy (predicted and actual spend).
- Enforce policy references and organization defaults.
- Produce an execution plan consumable by the runtime engine.

## Inputs

- Agent contract.
- Runtime profile (for example: `mva-default`, `mva-prod`).
- Environment context (local, dev, int, prod).
- Policy catalog and secrets references.

## Outputs

- Fully bound runtime configuration.
- Validation report and policy decision trace.
- Deterministic execution metadata for audit.

## Suggested Binding Steps

1. Validate contract schema.
2. Resolve capability and constraint references from policy catalogs.
3. Apply organization defaults.
4. Resolve runtime profile and model mapping.
5. Bind tools and credentials by permission scope.
6. Apply guardrail ceilings and policy checks (including cost policy).
7. Emit final runtime plan and provenance metadata.

## Cost Binding and Enforcement

Cost behavior should be deterministic at runtime, not prompt-dependent.

Recommended controls:

- **Pre-call estimate gate:** estimate token cost before each LLM call using provider pricing metadata.
- **In-flight accounting:** track actual usage after each response and tool turn.
- **Dual thresholds:** support warning threshold (soft) and stop threshold (hard).
- **Scoped budgets:** enforce per-step, per-run, and per-tenant/day budgets as configured.
- **Fallback ladder:** define ordered degradations before escalation.

Recommended precedence:

1. Organization-level budget policy
2. Environment/runtime profile ceilings
3. Contract-local limits (if stricter)

If thresholds conflict, choose the stricter effective limit.

### Fallback semantics

On soft breach, runtime may continue with controlled degradation:

- Reduce context payload
- Reduce loop depth (iterations/reflection)
- Downgrade model tier for eligible steps
- Force tool-first mode where possible

On hard breach, runtime should fail closed for autonomous continuation and require escalation or explicit operator override.

### Traceability requirements

Execution metadata should include:

- Resolved budget policy ID/version
- Predicted vs actual spend per step
- Triggered thresholds and chosen fallback path
- Final termination reason (completed, budget_stop, escalated)

## Capability and Constraint Resolution

Runtime binding is the layer that turns capability/constraint declarations into enforceable behavior.

- Capability refs: resolve `ref` IDs against the selected capability catalog version.
- Inline capabilities: normalize `define` blocks into internal capability objects, run policy validation, then merge with catalog-backed capabilities.
- Constraint refs: resolve to executable policy checks and runtime interceptors.
- Conflict handling: fail closed if a capability and constraint cannot both be satisfied.
- Traceability: include catalog version and resolution decisions in the policy decision trace.

This model keeps the contract declarative while giving the runtime deterministic behavior.

## Example Binding Fragment (Illustrative)

```yaml
runtimeBinding:
  profile: mva-default

  modelResolution:
    default: anthropic/claude-sonnet-4-6
    quick: anthropic/claude-haiku

  toolBindings:
    jira.read:
      server: atlassian
      allowedProjects:
        - HWS

  memory:
    sessionStore: redis://session-cache
    episodicStore: postgres://agent-memory

  limits:
    maxIterations: 6
    maxDuration: 10m

  costPolicy:
    ref: default_cost_guardrails
    thresholds:
      stepUsdWarn: 0.03
      stepUsdStop: 0.08
      runUsdWarn: 1.00
      runUsdStop: 2.00
    fallbackOrder:
      - trim_context
      - reduce_reflection
      - downgrade_model
      - force_tool_first
      - escalate_human
    version: 2026-05-01
```

## Design Rule

A contract should not need edits when moving from local execution to Kubernetes execution. Only runtime binding and deployment config should vary.
