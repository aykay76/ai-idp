# Agent Runtime Design (Draft)

This section defines a simple, practical design for making agents declarative while keeping onboarding lightweight.

## Executive Summary

To reduce confusion and improve governance, split agent design into three concerns:

1. **Agent Contract**: What the agent is and what it is allowed to do.
2. **Runtime Binding**: How the platform resolves and executes that contract.
3. **Packaging and Deployment**: How the runtime is containerized and operated.

The authoring experience can still stay simple:

- Teams primarily write one agent contract file.
- Platform teams maintain runtime binding profiles and deployment templates.
- The runtime composes these layers at execution time.

## Documents

- [Agent Contract](./agent-contract.md)
- [Agent Runtime Binding](./agent-runtime-binding.md)
- [Agent Packaging and Deployment](./agent-packaging-deployment.md)
- [Capability Catalog](./capability-catalog.md)
- [Constraint Catalog](./constraint-catalog.md)

## Design Goals

- Keep agent definitions declarative and portable.
- Enforce guardrails outside prompts.
- Minimize user-authored YAML while preserving strong governance.
- Support local development and Kubernetes deployment without changing contract semantics.

## Cost Governance Placement

Cost management should be split across two layers:

- **Design layer (agentic patterns):** guidance for choosing cheapest sufficient strategies (tool-first paths, model tiering, bounded reflection).
- **Runtime layer (this section):** enforceable limits, fallback policy, and auditable decisions.

Runtime is the source of truth for hard controls. Agents may reason about cost, but runtime must enforce ceilings deterministically.

Minimum runtime cost controls:

- Per-step USD ceiling
- Per-run USD ceiling
- Retry and iteration ceilings
- Fallback ladder when thresholds are breached
- Escalation policy for high-risk unresolved actions

The binding output should include the resolved budget policy ID and all budget-related decisions in trace metadata.
