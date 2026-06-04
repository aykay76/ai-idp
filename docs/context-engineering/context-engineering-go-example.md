# Context Engineering Part 3: Practical Go Example

This page shows a practical way to build a context package for a real engineering task.

It uses Go-flavored pseudocode so teams can move from concept to implementation quickly.

---

## Example Scenario

Task: "Create an implementation plan for a Jira epic, grounded in current code and architecture decisions."

Sources:

- Jira for delivery intent and acceptance criteria
- Confluence for architecture/process decisions
- GitHub for implementation truth and change history

Output:

- A structured context package passed to an LLM through the gateway
- Traceable references to each source
- A confidence and conflict summary

---

## Step-by-Step Build Flow

### Step 1: Define the context contract

```go
type BuildContextRequest struct {
	TaskID          string   // Internal request id
	JiraIssueKey    string   // e.g., "IE-1234"
	ConfluencePages []string // relevant page IDs or URLs
	Repo            string   // e.g., "maersk/service-a"
	Branch          string   // usually default branch for truth
	RiskTier        string   // high | medium | low
}

type ContextPackage struct {
	TaskBrief        string
	Sources          []SourceRef
	Sections         []Section
	Conflicts        []Conflict
	ValidationReport ValidationReport
	BudgetReport     BudgetReport
}
```

### Step 2: Fetch source snapshots

```go
func FetchSnapshots(ctx context.Context, req BuildContextRequest) (Snapshots, error) {
	jira, err := jiraClient.GetIssue(ctx, req.JiraIssueKey)
	if err != nil {
		return Snapshots{}, err
	}

	pages, err := confluenceClient.GetPages(ctx, req.ConfluencePages)
	if err != nil {
		return Snapshots{}, err
	}

	repoState, err := githubClient.GetRepoState(ctx, req.Repo, req.Branch)
	if err != nil {
		return Snapshots{}, err
	}

	return Snapshots{
		Jira:       jira,
		Confluence: pages,
		GitHub:     repoState,
	}, nil
}
```

### Step 3: Normalise into common sections

```go
func Normalise(s Snapshots) []Section {
	return []Section{
		{Kind: "task_brief", Content: FromJiraIntent(s.Jira)},
		{Kind: "acceptance_criteria", Content: FromJiraAC(s.Jira)},
		{Kind: "decisions", Content: FromConfluenceDecisions(s.Confluence)},
		{Kind: "runbook_constraints", Content: FromConfluenceRunbooks(s.Confluence)},
		{Kind: "implementation_truth", Content: FromGitHubState(s.GitHub)},
		{Kind: "test_and_ci_signals", Content: FromGitHubChecks(s.GitHub)},
	}
}
```

### Step 4: Apply precedence and detect conflicts

```go
func ResolveConflicts(sections []Section) ([]Section, []Conflict) {
	// Precedence:
	// 1) GitHub for implemented behavior
	// 2) Jira for planned delivery intent
	// 3) Confluence for decision rationale
	//
	// Any material mismatch is not silently overwritten.
	// It is carried forward as an explicit conflict.

	var conflicts []Conflict
	resolved := make([]Section, 0, len(sections))

	// Pseudocode: detect contradictions and annotate chosen source.
	// ...

	return resolved, conflicts
}
```

### Step 5: Enforce a context budget

```go
type BudgetPolicy struct {
	MaxTokensTotal int
	MaxPerSection  map[string]int
	ReserveForUser int // keep room for user's additional input
	ReserveForLLM  int // keep room for model response
}

type BudgetReport struct {
	UsedTotal      int
	UsedPerSection map[string]int
	Trimmed        []string // section kinds that were summarised
}

func ApplyBudget(sections []Section, p BudgetPolicy) ([]Section, BudgetReport) {
	// Strategy:
	// 1) Keep high-priority sections as-is
	// 2) Summarise medium-priority sections
	// 3) Drop low-value duplicates
	// 4) Never drop constraints/conflicts/acceptance criteria
	//
	// Pseudocode implementation omitted for brevity.
	return sections, BudgetReport{}
}
```

### Step 6: Validate before send

```go
func Validate(pkg ContextPackage) ValidationReport {
	return ValidationReport{
		HasTraceableSources: CheckSourceLinks(pkg),
		HasFreshnessMeta:    CheckFreshness(pkg),
		HasConflictFlags:    CheckConflictDisclosure(pkg),
		HasUnknowns:         CheckUnknowns(pkg),
		Pass:                true, // set from checks above
	}
}
```

### Step 7: Emit package to gateway

```go
func BuildAndSend(ctx context.Context, req BuildContextRequest) (LLMResponse, error) {
	snapshots, err := FetchSnapshots(ctx, req)
	if err != nil {
		return LLMResponse{}, err
	}

	sections := Normalise(snapshots)
	sections, conflicts := ResolveConflicts(sections)

	sections, budget := ApplyBudget(sections, DefaultBudgetFor(req.RiskTier))

	pkg := ContextPackage{
		TaskBrief:        "Implementation plan for " + req.JiraIssueKey,
		Sources:          CollectSources(snapshots),
		Sections:         sections,
		Conflicts:        conflicts,
		ValidationReport: Validate(ContextPackage{Sections: sections, Conflicts: conflicts}),
		BudgetReport:     budget,
	}

	if !pkg.ValidationReport.Pass {
		return LLMResponse{}, fmt.Errorf("context package validation failed")
	}

	return gatewayClient.Generate(ctx, pkg)
}
```

---

## Starter Context Budget (Practical Default)

Start with a simple budget that is easy to explain and tune.

Example total budget: `12,000` tokens

- `2,000` reserved for user follow-up input
- `3,000` reserved for model output
- `7,000` available for context package

Suggested section budget for the `7,000` context tokens:

- Task brief + acceptance criteria: `1,200`
- Implementation truth (GitHub): `2,200`
- Decisions/runbooks (Confluence): `1,600`
- Delivery context (Jira details): `1,200`
- Constraints, conflicts, unknowns: `800`

This is intentionally conservative. It prevents overfilling context windows and makes failures easier to diagnose.

---

## Budget Management Rules of Thumb

- Do not add raw documents if summaries with references are enough.
- Prefer newest and most authoritative artifacts over long history.
- Keep constraints and acceptance criteria in full fidelity.
- Summarise duplicates across sources into one canonical statement.
- Always include a budget report in debugging logs.

If output quality drops, inspect budget trimming decisions first.

---

## Advanced budgeting

For dynamic profiles, compression choices, observability, and escalation rules, see [Part 4 - advanced context budgeting](context-engineering-context-budget-advanced.md).

For most teams, start with the fixed budget above and evolve only after collecting real usage data.
