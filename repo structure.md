```
idp-manifests/
├── .github/                      # CI/CD workflows
│   └── workflows/
│       ├── validate-score.yaml   # Lint Score manifests
│       └── opa-verify.yaml       # OPA policy checks
├── schemas/                      # Score.dev schema definitions
│   └── score.dev/                # Follows Score's GVK model
│       ├── v1/                   # API Version
│       │   ├── workload.yaml     # e.g., `kind: Workload`
│       │   └── resource.yaml     # e.g., `kind: Resource`
│       └── v1beta1/              # Experimental version
├── policies/                     # OPA policies + governance docs
│   ├── security/                 # Cybersecurity team's policies
│   │   ├── encryption.rego       # e.g., "All databases encrypted at rest"
│   │   └── network.rego          # e.g., "Isolate PCI workloads"
│   ├── cost/                     # FinOps/Cloud Governance
│   │   └── budget-alerts.rego    # "Alert if monthly spend exceeds $10k"
│   ├── compliance/               # Enterprise Architecture
│   │   ├── gdpr.rego             # GDPR compliance rules
│   │   └── hipaa.rego            # HIPAA compliance rules
│   └── examples/                 # Test cases for policies
│       ├── allowed-manifest.yaml
│       └── denied-manifest.yaml
├── manifests/                    # Actual application manifests
│   ├── team-frontend/
│   │   └── app.score.yaml
│   └── team-data/
│       └── analytics.score.yaml
├── docs/
│   ├── CONTRIBUTING.md           # How to add policies/schemas
│   ├── ADR/                      # Architecture Decision Records
│   │   └── 001-use-score.md      # Why Score was chosen
│   └── policy-guide.md           # For cyber/cloud governance teams
└── Makefile                      # Local validation/test commands
```
