# Capability Catalog (Draft Design)

## Purpose

The capability catalog defines the shared semantics of capability IDs used in agent contracts.

This enables:

- Portable contract references (`ref`) across runtimes.
- Consistent behavior and validation for platform governance.
- Safe promotion of inline contract capabilities into reusable catalog entries.

## Ownership and Lifecycle

- Owned by platform/runtime maintainers.
- Versioned and immutable once published.
- Promoted through environments (dev -> int -> prod) with release controls.

Recommended lifecycle:

1. Team introduces an inline capability in a contract (`define`).
2. Runtime validates and runs it under policy constraints.
3. Platform review approves the definition for reuse.
4. Capability is added to catalog as a stable `id`.
5. Contracts migrate from `define` to `ref`.

## Minimal Shape (Illustrative)

```yaml
apiVersion: agent.maersk/v1
kind: CapabilityCatalog

metadata:
  name: default-capabilities
  version: 2026-05-01

spec:
  capabilities:
    - id: sprint_health_summary
      description: Summarize sprint progress and delivery confidence.
      riskTier: low
      requiredPermissions:
        - action: jira.read
          scope: HWS
      defaultConstraints:
        - read_only
      output:
        format: markdown

    - id: blocker_analysis
      description: Identify blockers and likely impact on near-term delivery.
      riskTier: medium
      requiredPermissions:
        - action: jira.read
          scope: HWS
      output:
        format: markdown
```

## Field Guidance

- `id`: globally unique, stable identifier used by contract `capabilities[].ref`.
- `description`: human-readable semantic intent.
- `riskTier`: policy classification (`low`, `medium`, `high`).
- `requiredPermissions`: minimum action/scope envelope required to execute.
- `defaultConstraints`: constraints that should always apply unless policy disallows override.
- `output`: default output contract for consistency and downstream parsing.

## Contract Mapping

From contract:

```yaml
spec:
  capabilities:
    - ref: sprint_health_summary
```

Runtime binding resolves `ref` IDs against the selected catalog version.

Inline capability in contract:

```yaml
spec:
  capabilities:
    - define:
        id: release_risk_forecast
        description: Estimate release risk from blockers and dependency churn.
        riskTier: medium
        requiredPermissions:
          - action: jira.read
            scope: HWS
```

Runtime treats this as an ephemeral capability object, validates it, and may emit a promotion recommendation.

## Validation Expectations

Validation should fail on:

- Duplicate capability IDs.
- Missing required fields (`id`, `description`, `riskTier`, `requiredPermissions`).
- Unknown enum values (risk tier, action verb).
- Policy-prohibited combinations (for example: high risk tier without required escalation rules).

## Compatibility

- Catalog versions should be backward compatible within the same major schema version.
- Runtime binding should include catalog name/version in policy decision trace.
- Contract execution should fail closed if referenced IDs are missing in the selected catalog.
