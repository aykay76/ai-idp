# Constraint Catalog (Draft Design)

## Purpose

The constraint catalog defines reusable policy constraints that can be referenced from agent contracts.

Constraints express what an agent must not do and how runtime must enforce those limits.

## Ownership and Lifecycle

- Owned by platform policy/governance maintainers.
- Versioned and immutable once published.
- Reviewed with legal, privacy, and security policy owners when relevant.

## Minimal Shape (Illustrative)

```yaml
apiVersion: agent.maersk/v1
kind: ConstraintCatalog

metadata:
  name: default-constraints
  version: 2026-05-01

spec:
  constraints:
    - id: read_only
      description: Disallow all write/update/delete tool actions.
      enforcement:
        type: action_filter
        deniedActions:
          - jira.write
          - github.write

    - id: no_individual_performance_data
      description: Prevent responses that expose person-level performance analysis.
      enforcement:
        type: response_policy
        rules:
          - redact_individual_metrics
          - aggregate_min_group_size: 5

    - id: default_cost_guardrails
      description: Enforce spend ceilings and deterministic fallback behavior.
      enforcement:
        type: budget_policy
        currency: USD
        ceilings:
          stepWarn: 0.03
          stepStop: 0.08
          runWarn: 1.00
          runStop: 2.00
        behavior:
          onWarn:
            - trim_context
            - reduce_reflection
            - downgrade_model
          onStop:
            - halt_autonomous_loop
            - escalate_human
```

## Field Guidance

- `id`: stable identifier used by contract `constraints[].ref`.
- `description`: clear policy intent and boundary.
- `enforcement`: runtime mechanism for applying policy.
- `type`: enforcement strategy class (for example: `action_filter`, `input_policy`, `response_policy`, `budget_policy`).

## Contract Mapping

From contract:

```yaml
spec:
  constraints:
    - ref: read_only
    - ref: no_individual_performance_data
```

Runtime binding resolves each `ref` into executable interceptors/checks.

Budget constraints should usually be catalog-managed (not duplicated inline) so teams share consistent thresholds and fallback semantics.

## Enforcement Semantics

Recommended execution order:

1. Pre-execution input checks.
2. Tool action authorization and filtering.
3. In-loop budget/time/iteration controls, including pre-call estimate checks.
4. Post-generation response policy checks.

Recommended behavior:

- Fail closed when enforcement cannot be evaluated.
- Return explicit policy violation reasons in validation/reporting.
- Record policy decisions in execution trace metadata.

## Conflict Rules

Constraint conflicts or unsatisfiable contract states should fail binding/execution.

Examples:

- Capability requires write permissions while `read_only` is active.
- Capability requires person-level output while `no_individual_performance_data` is active.
- Budget policy constraint is stricter than contract guardrails in a non-mergeable way.

Resolution policy:

- Constraints take precedence over capability defaults.
- Organization-level constraints take precedence over contract-local declarations.
- No implicit downgrades; require explicit contract or policy changes.

For budget policy specifically:

- Use the strictest effective limit across organization, environment, and contract levels.
- If a fallback action would violate another active constraint, stop and escalate.
- Never silently ignore a budget stop condition.

## Validation Expectations

Validation should fail on:

- Unknown constraint references for selected catalog version.
- Duplicate constraint IDs.
- Missing or malformed enforcement blocks.
- Incompatible combinations detected during runtime binding.

Budget-policy validation should additionally fail on:

- Non-monotonic thresholds (warn greater than stop).
- Missing currency or unsupported currency.
- Unknown fallback action IDs.
- Stop behavior that allows autonomous continuation without escalation/termination semantics.
