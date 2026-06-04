# Context Engineering Part 2: Maersk Lens

This guide applies Context Engineering 101 to how engineering teams work at Maersk.

The goal is simple: make AI outputs reliable by grounding them in trusted internal sources of truth, not just generic model knowledge.

---

## What Changes With the Maersk Lens

In the 101 guide, we focused on the general process.  
In this guide, we define:

- Which internal systems provide trusted context
- What each system is authoritative for
- How to resolve conflicts between sources
- How fresh context should be for different task types
- How teams can start small and scale safely

---

## Core Source-of-Truth Model

Use a source only for the domain where it is strongest.

### Jira is authoritative for work state and delivery intent

Use Jira for:

- Current status (To Do, In Progress, Blocked, Done)
- Ownership and accountability
- Priority and target delivery windows
- Acceptance criteria and business intent
- Linked dependencies and execution risks

Do not use Jira as the final truth for code behavior.

### Confluence is authoritative for decisions and operating guidance

Use Confluence for:

- Architecture decisions and rationale
- Team process and operating runbooks
- Product/domain background and definitions
- Policy and governance guidance

Do not use Confluence alone to infer current implementation details.

### GitHub is authoritative for implementation truth

Use GitHub for:

- Actual code and current behavior
- Pull request intent and change history
- Test coverage and validation evidence
- CI/CD signals for build and quality state

Do not use GitHub alone to infer delivery priority or business commitment.

---

## Precedence Rules (When Sources Conflict)

When AI sees contradictory information, apply these rules in order:

1. **Implementation truth:** GitHub wins for "what the system does now."
2. **Delivery truth:** Jira wins for "what we intend to deliver next."
3. **Decision truth:** Confluence wins for "why we chose this approach."
4. **Freshness tie-breaker:** Prefer newer, traceable artifacts.
5. **Escalation rule:** If conflict is material and unresolved, flag for human review.

This avoids silent hallucination by making conflict handling explicit.

---

## Freshness Expectations

Set context freshness by task criticality.

### High-risk tasks (production-impacting)

Examples: incident actions, rollback plans, access changes, critical fixes.

- Pull latest Jira issue state
- Pull latest relevant Confluence pages
- Pull latest default branch code and open PR status
- Require explicit "last refreshed" metadata in AI output

### Medium-risk tasks (delivery and planning)

Examples: sprint planning support, implementation proposals, handoffs.

- Use current sprint Jira data
- Use most recent decision/runbook pages
- Use current repository/PR state
- Warn when context age exceeds agreed threshold

### Low-risk tasks (learning and drafting)

Examples: onboarding drafts, summaries, first-pass docs.

- Allow slightly older context
- Require clear caveat that output is draft-level

---

## Standard Context Pack for Maersk Engineering Tasks

For most tasks, package context in this order:

1. **Task brief** (objective, audience, definition of done)
2. **Jira snapshot** (issue keys, status, owners, acceptance criteria)
3. **Confluence snapshot** (decision links, runbooks, policy excerpts)
4. **GitHub snapshot** (repo, branch/PR state, relevant files/tests)
5. **Constraints** (security, compliance, architecture boundaries)
6. **Output contract** (exact required format)
7. **Validation checks** (what must be true before accepting output)

Keep context intentionally small. Relevance beats volume.

---

## Starter Workflow for Teams

Use this four-step rollout so teams can adopt quickly.

### Step 1: Pick one recurring engineering task

Good examples:

- "Draft implementation plan from Jira epic + linked PRs"
- "Summarise incident context and propose next safe actions"
- "Prepare PR review brief with policy and test checks"

### Step 2: Define the minimum source bundle

Specify exactly:

- Which Jira fields are required
- Which Confluence spaces/pages are trusted
- Which GitHub repos/branches are in scope

### Step 3: Define acceptance checks

At minimum, verify:

- Output references source links
- Claims are traceable to source systems
- Conflicts are surfaced, not hidden
- Unknowns are listed explicitly

### Step 4: Run, review, improve

- Run 5-10 real examples
- Record failure modes
- Improve source selection and constraints first
- Only then tune prompts or model settings

---

## Governance and Guardrails

At a minimum, every team should define:

- Approved source systems per workflow
- Data handling boundaries (what cannot be copied into prompts)
- Human approval points for high-impact actions
- Auditability expectations (input references and output traceability)

Treat context packages as controlled engineering assets, not ad-hoc prompt text.

---

## Common Failure Modes in Enterprise Context Engineering

- Mixing "planned" and "implemented" states without labeling the difference
- Relying on stale Confluence pages without freshness checks
- Treating Jira comments as policy-level truth
- Asking AI to infer missing data instead of flagging missing context
- No explicit conflict resolution logic between systems

---

## Deliverables to Create Next

To operationalise this at Maersk, create:

- A reusable "Maersk context pack" template
- A source precedence policy per workflow
- A freshness policy by risk level
- A validation checklist for AI outputs
- First Go package examples to assemble and validate context packs

These deliverables become the bridge from guidance to production usage.

See Part 3: [Context engineering Part 3 - practical Go example](context-engineering-go-example.md).

For advanced budgeting (profiles, compression, observability): [Part 4 - advanced context budgeting](context-engineering-context-budget-advanced.md).
