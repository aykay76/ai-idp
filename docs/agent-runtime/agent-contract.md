# Agent Contract (Draft Design)

## Purpose

The agent contract is the portable, declarative source of truth for agent behavior and policy intent.

It should be stable across runtimes and deployment targets.

## What Belongs Here

- Agent identity and ownership.
- Intended role, goals, and output expectations.
- Declared interaction pattern (for example: `react`, `plan-execute`, `reflection`).
- Capabilities (what the agent can attempt to do).
- Constraints and policy intent (what it must not do).
- Human-in-the-loop requirements.
- Logical guardrails (budget, max iterations, escalation triggers).
- Abstract tool permissions (verbs and scopes), not concrete server wiring.

## What Does Not Belong Here

- Concrete model provider or model IDs.
- MCP server connection details.
- Memory backend technologies.
- Container image, Kubernetes resources, or environment wiring.

## Contract Principles

- **Declarative**: describe intent, not implementation details.
- **Portable**: same contract works across different runtimes.
- **Governable**: policy-critical controls are explicit and enforceable.
- **Inspectable**: easy to validate, audit, and diff.

## Capability and Constraint Definitions

Capabilities and constraints are declared in the contract, but their semantics are defined by versioned catalogs.

- Capability catalog: platform-governed definitions of what a capability means, expected inputs/outputs, risk tier, and default permission envelope.
- Constraint catalog: platform-governed restrictions and policy semantics (for example: `read_only`, `no_individual_performance_data`).
- Contract authoring: teams normally reference catalog entries by ID. For net-new needs, teams may declare an inline capability proposal in the contract.
- Promotion model: inline capabilities are valid for controlled rollout, then promoted into the shared capability catalog once reviewed.

This keeps contracts portable while still allowing product teams to move quickly on new use cases.

## Recommended Shape (Reference + Inline)

```yaml
spec:
  capabilities:
    - ref: sprint_health_summary

    - define:
        id: release_risk_forecast
        description: Estimate release risk from blockers and dependency churn.
        riskTier: medium
        requiredPermissions:
          - action: jira.read
            scope: HWS
        output:
          format: markdown

  constraints:
    - ref: read_only
    - ref: no_individual_performance_data
```

Notes:

- `ref` resolves to a catalog entry.
- `define` is an inline capability declaration owned by the contract team.
- Constraints should be references only; inline constraints make policy harder to govern consistently.

## Minimal Shape (Illustrative)

```yaml
apiVersion: agent.maersk/v1
kind: AgentContract

metadata:
  name: engineering-manager
  owner: team-engx

spec:
  role: Engineering Manager
  goal: Summarize sprint health and release risk
  pattern: react

  capabilities:
    - sprint_health_summary
    - blocker_analysis

  constraints:
    - read_only
    - no_individual_performance_data

  permissions:
    - action: jira.read
      scope: HWS

  guardrails:
    maxIterations: 6
    budgetUsd: 2

  humanInTheLoop:
    escalationOn:
      - release_p0_risk

  output:
    format: markdown
```

## Validation Expectations

Contract validation should fail on:

- Missing required fields.
- Unknown enum values (pattern, risk tier, action verb).
- Unknown capability/constraint references for the selected catalog version.
- Inline capability definitions missing required metadata (id, description, risk tier, permission envelope).
- Invalid combinations (for example: write action with read-only constraint).
- Out-of-policy budget or risk settings.
