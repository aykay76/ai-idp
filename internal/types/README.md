# Types Package - AIIDP-20: Core Domain Types ✅

The `types` package provides comprehensive core domain types, constants, and data structures used across the AI-IDP platform. This package establishes the foundational type system for the entire platform.

## Architecture

This package follows **Kubernetes-style API design** with Group/Version/Kind patterns and provides complete domain modeling for:

- **Applications**: Full lifecycle management with AI integration
- **Teams**: Multi-tenant team management with RBAC
- **Users**: User management with authentication providers
- **Tenants**: Multi-tenant isolation and resource quotas
- **Templates**: AI-enhanced application templates
- **Policies**: Governance and policy enforcement
- **Audit**: Complete audit trail and event tracking

## Core Domain Types

### 1. Application Types
```go
// Kubernetes-style application resource
type Application struct {
    TypeMeta   `json:",inline"`
    Metadata   ObjectMeta        `json:"metadata"`
    Spec       ApplicationSpec   `json:"spec"`
    Status     ApplicationStatus `json:"status,omitempty"`
}
```

**Features:**
- ✅ Kubernetes-style Group/Version/Kind API design
- ✅ Full lifecycle management (development → production)
- ✅ AI integration with generation context and templates
- ✅ Repository and deployment specifications
- ✅ Resource dependencies and configuration management
- ✅ Contact management and ownership tracking

### 2. Team Management
```go
type Team struct {
    TypeMeta `json:",inline"`
    Metadata ObjectMeta `json:"metadata"`
    Spec     TeamSpec   `json:"spec"`
    Status   TeamStatus `json:"status,omitempty"`
}
```

**Features:**
- ✅ Multi-tenant team isolation
- ✅ Role-based access control (Owner, Maintainer, Developer, Viewer)
- ✅ Resource quotas and namespace management
- ✅ Permission systems and notification settings

### 3. User Management
```go
type User struct {
    TypeMeta `json:",inline"`
    Metadata ObjectMeta `json:"metadata"`
    Spec     UserSpec   `json:"spec"`
    Status   UserStatus `json:"status,omitempty"`
}
```

**Features:**
- ✅ Multi-provider authentication (GitHub, Google, SAML)
- ✅ User preferences and avatar support
- ✅ Team membership tracking
- ✅ Activity and usage analytics

### 4. Multi-Tenant Support
```go
type Tenant struct {
    TypeMeta `json:",inline"`
    Metadata ObjectMeta  `json:"metadata"`
    Spec     TenantSpec  `json:"spec"`
    Status   TenantStatus `json:"status,omitempty"`
}
```

**Features:**
- ✅ Complete tenant isolation with resource quotas
- ✅ Provider configuration per tenant
- ✅ Domain-based tenant identification
- ✅ Billing plan integration ready

### 5. AI-Enhanced Templates
```go
type Template struct {
    TypeMeta `json:",inline"`
    Metadata ObjectMeta     `json:"metadata"`
    Spec     TemplateSpec   `json:"spec"`
    Status   TemplateStatus `json:"status,omitempty"`
}
```

**Features:**
- ✅ Dynamic AI-powered file generation
- ✅ Parameterized templates with validation
- ✅ Post-creation actions and hooks
- ✅ Category-based organization (microservice, frontend, ML, etc.)

### 6. Policy & Governance
```go
type Policy struct {
    TypeMeta `json:",inline"`
    Metadata ObjectMeta   `json:"metadata"`
    Spec     PolicySpec   `json:"spec"`
    Status   PolicyStatus `json:"status,omitempty"`
}
```

**Features:**
- ✅ Multi-scope policies (global, tenant, team, namespace)
- ✅ Multiple enforcement modes (warn, block, monitor)
- ✅ Resource, security, cost, and compliance policies
- ✅ Rule-based conditions with flexible operators

### 7. Audit & Traceability
```go
type AuditEvent struct {
    TypeMeta `json:",inline"`
    Metadata ObjectMeta     `json:"metadata"`
    Spec     AuditEventSpec `json:"spec"`
}
```

**Features:**
- ✅ Complete audit trail for all platform actions
- ✅ Actor tracking (users, service accounts, system, bots)
- ✅ Resource references with full lineage
- ✅ IP address and user agent tracking

## Context Keys

Type-safe context keys for request processing:

```go
const (
    RequestIDKey      ContextKey = "request_id"
    UserIDKey         ContextKey = "user_id"
    TenantIDKey       ContextKey = "tenant_id"
    TeamIDKey         ContextKey = "team_id"
    OrganizationIDKey ContextKey = "organization_id"
)
```

## Usage Examples

### Creating Applications
```go
app := &types.Application{
    TypeMeta: types.TypeMeta{
        APIVersion: "platform.company.com/v1",
        Kind:       "Application",
    },
    Metadata: types.ObjectMeta{
        Name:      "ecommerce-api",
        Namespace: "retail-team",
        Labels: map[string]string{
            "domain": "retail",
            "criticality": "high",
        },
        Annotations: map[string]string{
            "platform.company.com/ai-generated": "true",
        },
    },
    Spec: types.ApplicationSpec{
        DisplayName: "E-Commerce API",
        Team:        "retail-backend-team",
        Owner:       "platform-team@company.com",
        Lifecycle:   types.LifecycleProduction,
        AIGenerated: true,
        Repository: types.RepositorySpec{
            URL:      "https://github.com/company/ecommerce-api",
            Branch:   "main",
            Provider: "github",
        },
    },
}
```

### Working with Teams
```go
team := &types.Team{
    TypeMeta: types.TypeMeta{
        APIVersion: "platform.company.com/v1",
        Kind:       "Team",
    },
    Metadata: types.ObjectMeta{
        Name: "retail-backend-team",
    },
    Spec: types.TeamSpec{
        DisplayName: "Retail Backend Team",
        Members: []types.Member{
            {
                Email:  "john.doe@company.com",
                Role:   types.TeamRoleOwner,
                Status: types.MemberStatusActive,
            },
        },
        Settings: types.TeamSettings{
            ResourceQuotas: types.ResourceQuotas{
                Applications: 50,
                CPU:         "100",
                Memory:      "200Gi",
            },
        },
    },
}
```

### Policy Definition
```go
policy := &types.Policy{
    TypeMeta: types.TypeMeta{
        APIVersion: "platform.company.com/v1",
        Kind:       "Policy",
    },
    Metadata: types.ObjectMeta{
        Name: "cost-limit-policy",
    },
    Spec: types.PolicySpec{
        Type:        types.PolicyTypeCost,
        Scope:       types.PolicyScopeTenant,
        Enforcement: types.PolicyEnforcementBlock,
        Rules: []types.PolicyRule{
            {
                Name:     "database-cost-limit",
                Resource: "database",
                Action:   "create",
                Conditions: []types.PolicyCondition{
                    {
                        Field:    "estimated_cost",
                        Operator: "gt",
                        Value:    1000,
                    },
                },
                Effect: types.PolicyEffectDeny,
            },
        },
    },
}
```

## Benefits

1. **Type Safety**: Comprehensive Go types prevent runtime errors
2. **Kubernetes Compatibility**: Standard Group/Version/Kind API patterns
3. **Multi-Tenancy**: Built-in tenant isolation and resource management
4. **AI Integration**: Native support for AI-generated resources and templates
5. **Audit Ready**: Complete audit trail and governance support
6. **Extensible**: Easy to extend with new resource types
7. **Validation**: Built-in validation tags for data integrity

## Implementation Status

- ✅ **AIIDP-20: Define Core Domain Types** - **COMPLETED**
- ✅ Kubernetes-style resource definitions
- ✅ Multi-tenant support with isolation
- ✅ AI integration types
- ✅ Policy and governance framework
- ✅ Complete audit trail support
- ✅ User and team management
- ✅ Template system with AI enhancement

This provides the foundational type system for the entire AI-IDP platform! 🚀