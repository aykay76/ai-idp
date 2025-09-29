package types

import (
	"time"

	"github.com/google/uuid"
)

// ContextKey is used for context keys to avoid collisions
type ContextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey ContextKey = "request_id"
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
	// TenantIDKey is the context key for tenant ID
	TenantIDKey ContextKey = "tenant_id"
	// TeamIDKey is the context key for team ID
	TeamIDKey ContextKey = "team_id"
	// OrganizationIDKey is the context key for organization ID
	OrganizationIDKey ContextKey = "organization_id"
)

// =============================================================================
// KUBERNETES-STYLE RESOURCE DEFINITIONS
// =============================================================================

// ObjectMeta contains metadata for all resources (Kubernetes-style)
type ObjectMeta struct {
	Name        string            `json:"name" validate:"required,dns1123"`
	Namespace   string            `json:"namespace,omitempty"`
	UID         uuid.UUID         `json:"uid,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	DeletedAt   *time.Time        `json:"deleted_at,omitempty"`
}

// TypeMeta contains API version and kind (Kubernetes-style)
type TypeMeta struct {
	APIVersion string `json:"apiVersion" validate:"required"`
	Kind       string `json:"kind" validate:"required"`
}

// =============================================================================
// APPLICATION DOMAIN TYPES
// =============================================================================

// Application represents an application in the platform
type Application struct {
	TypeMeta `json:",inline"`
	Metadata ObjectMeta        `json:"metadata"`
	Spec     ApplicationSpec   `json:"spec"`
	Status   ApplicationStatus `json:"status,omitempty"`
}

// ApplicationSpec contains the desired state of an application
type ApplicationSpec struct {
	// Basic application information
	DisplayName string `json:"display_name,omitempty"`
	Description string `json:"description,omitempty"`

	// Ownership and team information
	Team     string    `json:"team" validate:"required"`
	Owner    string    `json:"owner" validate:"required,email"`
	Contacts []Contact `json:"contacts,omitempty"`

	// Application lifecycle
	Lifecycle   ApplicationLifecycle `json:"lifecycle" validate:"required"`
	Environment string               `json:"environment,omitempty"`

	// Repository and deployment information
	Repository RepositorySpec `json:"repository,omitempty"`
	Deployment DeploymentSpec `json:"deployment,omitempty"`

	// Configuration and resources
	Resources     []ResourceSpec    `json:"resources,omitempty"`
	Dependencies  []DependencySpec  `json:"dependencies,omitempty"`
	Configuration map[string]string `json:"configuration,omitempty"`

	// AI and template information
	Template    *TemplateReference `json:"template,omitempty"`
	AIGenerated bool               `json:"ai_generated,omitempty"`
	AIContext   *AIContext         `json:"ai_context,omitempty"`
}

// ApplicationStatus represents the observed state of an application
type ApplicationStatus struct {
	Phase              ApplicationPhase `json:"phase"`
	Message            string           `json:"message,omitempty"`
	Conditions         []Condition      `json:"conditions,omitempty"`
	Resources          []ResourceStatus `json:"resources,omitempty"`
	LastUpdate         time.Time        `json:"last_update"`
	ObservedGeneration int64            `json:"observed_generation,omitempty"`
}

// ApplicationPhase represents the phase of an application
type ApplicationPhase string

const (
	ApplicationPhasePending      ApplicationPhase = "pending"
	ApplicationPhaseInitializing ApplicationPhase = "initializing"
	ApplicationPhaseDeploying    ApplicationPhase = "deploying"
	ApplicationPhaseRunning      ApplicationPhase = "running"
	ApplicationPhaseFailed       ApplicationPhase = "failed"
	ApplicationPhaseTerminating  ApplicationPhase = "terminating"
	ApplicationPhaseTerminated   ApplicationPhase = "terminated"
)

// ApplicationLifecycle represents the lifecycle stage
type ApplicationLifecycle string

const (
	LifecycleDevelopment ApplicationLifecycle = "development"
	LifecycleTesting     ApplicationLifecycle = "testing"
	LifecycleStaging     ApplicationLifecycle = "staging"
	LifecycleProduction  ApplicationLifecycle = "production"
	LifecycleRetired     ApplicationLifecycle = "retired"
)

// =============================================================================
// SUPPORTING APPLICATION TYPES
// =============================================================================

// Contact represents a contact person for an application
type Contact struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role,omitempty"`
	Type  string `json:"type,omitempty"` // primary, secondary, escalation
}

// RepositorySpec contains repository information
type RepositorySpec struct {
	URL      string `json:"url,omitempty"`
	Branch   string `json:"branch,omitempty"`
	Provider string `json:"provider,omitempty"` // github, gitlab, etc.
	Path     string `json:"path,omitempty"`     // subdirectory in repo
}

// DeploymentSpec contains deployment configuration
type DeploymentSpec struct {
	Strategy    string            `json:"strategy,omitempty"` // rolling, blue-green, canary
	Replicas    int               `json:"replicas,omitempty"`
	Resources   ResourceLimits    `json:"resources,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Secrets     []SecretRef       `json:"secrets,omitempty"`
}

// ResourceSpec represents a resource requirement
type ResourceSpec struct {
	Type         string            `json:"type" validate:"required"` // database, cache, queue, etc.
	Name         string            `json:"name" validate:"required"`
	Provider     string            `json:"provider,omitempty"`
	Config       map[string]string `json:"config,omitempty"`
	Dependencies []string          `json:"dependencies,omitempty"`
}

// ResourceStatus represents the status of a resource
type ResourceStatus struct {
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	Message    string    `json:"message,omitempty"`
	Ready      bool      `json:"ready"`
	LastUpdate time.Time `json:"last_update"`
}

// DependencySpec represents a dependency on another service or resource
type DependencySpec struct {
	Name     string            `json:"name" validate:"required"`
	Type     string            `json:"type" validate:"required"` // service, database, external-api
	Version  string            `json:"version,omitempty"`
	Required bool              `json:"required,omitempty"`
	Config   map[string]string `json:"config,omitempty"`
}

// TemplateReference references an application template
type TemplateReference struct {
	Name    string `json:"name" validate:"required"`
	Version string `json:"version,omitempty"`
	Source  string `json:"source,omitempty"` // builtin, git-url, etc.
}

// AIContext contains AI-related metadata
type AIContext struct {
	Prompt        string            `json:"prompt,omitempty"`
	Model         string            `json:"model,omitempty"`
	GeneratedAt   time.Time         `json:"generated_at,omitempty"`
	Confidence    float64           `json:"confidence,omitempty"`
	Parameters    map[string]string `json:"parameters,omitempty"`
	Modifications []string          `json:"modifications,omitempty"`
}

// ResourceLimits defines resource limits and requests
type ResourceLimits struct {
	CPU      string           `json:"cpu,omitempty"`
	Memory   string           `json:"memory,omitempty"`
	Storage  string           `json:"storage,omitempty"`
	Requests ResourceRequests `json:"requests,omitempty"`
}

// ResourceRequests defines resource requests
type ResourceRequests struct {
	CPU     string `json:"cpu,omitempty"`
	Memory  string `json:"memory,omitempty"`
	Storage string `json:"storage,omitempty"`
}

// SecretRef references a secret
type SecretRef struct {
	Name string `json:"name" validate:"required"`
	Key  string `json:"key,omitempty"`
}

// =============================================================================
// TEAM DOMAIN TYPES
// =============================================================================

// Team represents a team in the platform
type Team struct {
	TypeMeta `json:",inline"`
	Metadata ObjectMeta `json:"metadata"`
	Spec     TeamSpec   `json:"spec"`
	Status   TeamStatus `json:"status,omitempty"`
}

// TeamSpec contains the desired state of a team
type TeamSpec struct {
	DisplayName string       `json:"display_name,omitempty"`
	Description string       `json:"description,omitempty"`
	Members     []Member     `json:"members,omitempty"`
	Permissions []Permission `json:"permissions,omitempty"`
	Settings    TeamSettings `json:"settings,omitempty"`
}

// TeamStatus represents the observed state of a team
type TeamStatus struct {
	MemberCount  int       `json:"member_count"`
	Applications int       `json:"applications"`
	LastActivity time.Time `json:"last_activity,omitempty"`
}

// Member represents a team member
type Member struct {
	UserID   string       `json:"user_id" validate:"required"`
	Email    string       `json:"email" validate:"required,email"`
	Role     TeamRole     `json:"role" validate:"required"`
	JoinedAt time.Time    `json:"joined_at"`
	Status   MemberStatus `json:"status"`
}

// TeamRole represents a member's role in a team
type TeamRole string

const (
	TeamRoleOwner      TeamRole = "owner"
	TeamRoleMaintainer TeamRole = "maintainer"
	TeamRoleDeveloper  TeamRole = "developer"
	TeamRoleViewer     TeamRole = "viewer"
)

// MemberStatus represents the status of a team member
type MemberStatus string

const (
	MemberStatusActive   MemberStatus = "active"
	MemberStatusInactive MemberStatus = "inactive"
	MemberStatusPending  MemberStatus = "pending"
)

// Permission represents a permission for resources
type Permission struct {
	Resource string   `json:"resource" validate:"required"`
	Actions  []string `json:"actions" validate:"required"`
	Scope    string   `json:"scope,omitempty"` // team, namespace, cluster
}

// TeamSettings contains team-specific configuration
type TeamSettings struct {
	AutoApproval      bool                 `json:"auto_approval,omitempty"`
	AllowedNamespaces []string             `json:"allowed_namespaces,omitempty"`
	ResourceQuotas    ResourceQuotas       `json:"resource_quotas,omitempty"`
	Notifications     NotificationSettings `json:"notifications,omitempty"`
}

// ResourceQuotas defines resource quotas for teams or tenants
type ResourceQuotas struct {
	CPU          string `json:"cpu,omitempty"`
	Memory       string `json:"memory,omitempty"`
	Storage      string `json:"storage,omitempty"`
	Applications int    `json:"applications,omitempty"`
	Namespaces   int    `json:"namespaces,omitempty"`
	Users        int    `json:"users,omitempty"`
}

// ResourceUsage tracks current resource usage
type ResourceUsage struct {
	CPU          string `json:"cpu,omitempty"`
	Memory       string `json:"memory,omitempty"`
	Storage      string `json:"storage,omitempty"`
	Applications int    `json:"applications"`
	Namespaces   int    `json:"namespaces"`
	Users        int    `json:"users"`
}

// NotificationSettings contains notification preferences
type NotificationSettings struct {
	Email    bool     `json:"email,omitempty"`
	Slack    bool     `json:"slack,omitempty"`
	Webhook  bool     `json:"webhook,omitempty"`
	Channels []string `json:"channels,omitempty"`
}

// =============================================================================
// TENANT AND ORGANIZATION TYPES
// =============================================================================

// Tenant represents a tenant in the multi-tenant platform
type Tenant struct {
	TypeMeta `json:",inline"`
	Metadata ObjectMeta   `json:"metadata"`
	Spec     TenantSpec   `json:"spec"`
	Status   TenantStatus `json:"status,omitempty"`
}

// TenantSpec contains the desired state of a tenant
type TenantSpec struct {
	DisplayName string           `json:"display_name,omitempty"`
	Description string           `json:"description,omitempty"`
	Domain      string           `json:"domain,omitempty"`
	Plan        string           `json:"plan,omitempty"`
	Settings    TenantSettings   `json:"settings,omitempty"`
	Quotas      ResourceQuotas   `json:"quotas,omitempty"`
	Providers   []ProviderConfig `json:"providers,omitempty"`
}

// TenantStatus represents the observed state of a tenant
type TenantStatus struct {
	Phase         TenantPhase   `json:"phase"`
	UserCount     int           `json:"user_count"`
	TeamCount     int           `json:"team_count"`
	AppCount      int           `json:"app_count"`
	ResourceUsage ResourceUsage `json:"resource_usage,omitempty"`
}

// TenantPhase represents the phase of a tenant
type TenantPhase string

const (
	TenantPhaseProvisioning TenantPhase = "provisioning"
	TenantPhaseActive       TenantPhase = "active"
	TenantPhaseSuspended    TenantPhase = "suspended"
	TenantPhaseTerminating  TenantPhase = "terminating"
)

// TenantSettings contains tenant-specific configuration
type TenantSettings struct {
	AllowSelfRegistration bool     `json:"allow_self_registration,omitempty"`
	AllowedDomains        []string `json:"allowed_domains,omitempty"`
	RequireMFA            bool     `json:"require_mfa,omitempty"`
	SessionTimeout        int      `json:"session_timeout,omitempty"` // in minutes
	MaxTeams              int      `json:"max_teams,omitempty"`
	MaxUsers              int      `json:"max_users,omitempty"`
}

// ProviderConfig represents configuration for a resource provider
type ProviderConfig struct {
	Name    string            `json:"name" validate:"required"`
	Type    string            `json:"type" validate:"required"` // github, aws, gcp, azure, k8s
	Enabled bool              `json:"enabled"`
	Config  map[string]string `json:"config,omitempty"`
	Secrets []SecretRef       `json:"secrets,omitempty"`
}

// TenantContext represents tenant information in requests
type TenantContext struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	DisplayName string `json:"display_name,omitempty"`
}

// =============================================================================
// USER DOMAIN TYPES
// =============================================================================

// User represents a user in the platform
type User struct {
	TypeMeta `json:",inline"`
	Metadata ObjectMeta `json:"metadata"`
	Spec     UserSpec   `json:"spec"`
	Status   UserStatus `json:"status,omitempty"`
}

// UserSpec contains the desired state of a user
type UserSpec struct {
	Email       string            `json:"email" validate:"required,email"`
	DisplayName string            `json:"display_name,omitempty"`
	AvatarURL   string            `json:"avatar_url,omitempty"`
	Roles       []string          `json:"roles,omitempty"`
	Teams       []string          `json:"teams,omitempty"`
	Preferences map[string]string `json:"preferences,omitempty"`
	Provider    AuthProvider      `json:"provider,omitempty"`
}

// UserStatus represents the observed state of a user
type UserStatus struct {
	Active     bool      `json:"active"`
	LastLogin  time.Time `json:"last_login,omitempty"`
	LoginCount int       `json:"login_count"`
	TeamCount  int       `json:"team_count"`
	AppCount   int       `json:"app_count"`
}

// AuthProvider represents authentication provider information
type AuthProvider struct {
	Type       string `json:"type"` // github, google, saml, etc.
	ProviderID string `json:"provider_id"`
	Username   string `json:"username,omitempty"`
}

// =============================================================================
// TEMPLATE DOMAIN TYPES
// =============================================================================

// Template represents an application template
type Template struct {
	TypeMeta `json:",inline"`
	Metadata ObjectMeta     `json:"metadata"`
	Spec     TemplateSpec   `json:"spec"`
	Status   TemplateStatus `json:"status,omitempty"`
}

// TemplateSpec contains the desired state of a template
type TemplateSpec struct {
	DisplayName string              `json:"display_name,omitempty"`
	Description string              `json:"description,omitempty"`
	Category    string              `json:"category,omitempty"` // microservice, frontend, data-pipeline, ml-service
	Tags        []string            `json:"tags,omitempty"`
	Parameters  []TemplateParameter `json:"parameters,omitempty"`
	Files       []TemplateFile      `json:"files,omitempty"`
	PostActions []TemplateAction    `json:"post_actions,omitempty"`

	// AI enhancement capabilities
	AIEnhanced   bool              `json:"ai_enhanced,omitempty"`
	AIPrompts    []AIPrompt        `json:"ai_prompts,omitempty"`
	DynamicFiles []DynamicFileRule `json:"dynamic_files,omitempty"`
}

// TemplateStatus represents the observed state of a template
type TemplateStatus struct {
	UsageCount   int       `json:"usage_count"`
	LastUsed     time.Time `json:"last_used,omitempty"`
	Validated    bool      `json:"validated"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// TemplateParameter represents a template parameter
type TemplateParameter struct {
	Name        string      `json:"name" validate:"required"`
	DisplayName string      `json:"display_name,omitempty"`
	Description string      `json:"description,omitempty"`
	Type        string      `json:"type" validate:"required"` // string, int, bool, select, multiselect
	Required    bool        `json:"required,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Options     []string    `json:"options,omitempty"`    // for select/multiselect
	Validation  string      `json:"validation,omitempty"` // regex pattern
}

// TemplateFile represents a file in a template
type TemplateFile struct {
	Path     string `json:"path" validate:"required"`
	Content  string `json:"content,omitempty"`
	Source   string `json:"source,omitempty"`   // url, base64, inline
	Template bool   `json:"template,omitempty"` // whether to process as template
	Mode     string `json:"mode,omitempty"`     // file permissions
}

// TemplateAction represents a post-creation action
type TemplateAction struct {
	Name        string            `json:"name" validate:"required"`
	Type        string            `json:"type" validate:"required"` // command, api-call, webhook
	Command     string            `json:"command,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	OnFailure   string            `json:"on_failure,omitempty"` // continue, abort
}

// AIPrompt represents an AI prompt for dynamic generation
type AIPrompt struct {
	Trigger    string `json:"trigger"`     // when to use this prompt
	Prompt     string `json:"prompt"`      // AI prompt template
	OutputFile string `json:"output_file"` // where to put AI output
	OutputType string `json:"output_type"` // code, config, documentation
}

// DynamicFileRule represents a rule for dynamic file generation
type DynamicFileRule struct {
	Condition string `json:"condition"` // when to generate this file
	Generator string `json:"generator"` // ai-prompt, template, script
	Template  string `json:"template"`  // base template if needed
}

// =============================================================================
// POLICY AND GOVERNANCE TYPES
// =============================================================================

// Policy represents a governance policy
type Policy struct {
	TypeMeta `json:",inline"`
	Metadata ObjectMeta   `json:"metadata"`
	Spec     PolicySpec   `json:"spec"`
	Status   PolicyStatus `json:"status,omitempty"`
}

// PolicySpec contains the desired state of a policy
type PolicySpec struct {
	DisplayName string            `json:"display_name,omitempty"`
	Description string            `json:"description,omitempty"`
	Type        PolicyType        `json:"type" validate:"required"`
	Scope       PolicyScope       `json:"scope" validate:"required"`
	Rules       []PolicyRule      `json:"rules" validate:"required"`
	Enforcement PolicyEnforcement `json:"enforcement" validate:"required"`
	Parameters  map[string]string `json:"parameters,omitempty"`
}

// PolicyStatus represents the observed state of a policy
type PolicyStatus struct {
	Active        bool      `json:"active"`
	Violations    int       `json:"violations"`
	LastEvaluated time.Time `json:"last_evaluated,omitempty"`
	ErrorMessage  string    `json:"error_message,omitempty"`
}

// PolicyType represents the type of policy
type PolicyType string

const (
	PolicyTypeResource   PolicyType = "resource"
	PolicyTypeSecurity   PolicyType = "security"
	PolicyTypeCost       PolicyType = "cost"
	PolicyTypeCompliance PolicyType = "compliance"
	PolicyTypeQuality    PolicyType = "quality"
)

// PolicyScope represents the scope of a policy
type PolicyScope string

const (
	PolicyScopeGlobal    PolicyScope = "global"
	PolicyScopeTenant    PolicyScope = "tenant"
	PolicyScopeTeam      PolicyScope = "team"
	PolicyScopeNamespace PolicyScope = "namespace"
)

// PolicyEnforcement represents policy enforcement mode
type PolicyEnforcement string

const (
	PolicyEnforcementWarn    PolicyEnforcement = "warn"
	PolicyEnforcementBlock   PolicyEnforcement = "block"
	PolicyEnforcementMonitor PolicyEnforcement = "monitor"
)

// PolicyRule represents a single policy rule
type PolicyRule struct {
	Name        string            `json:"name" validate:"required"`
	Description string            `json:"description,omitempty"`
	Resource    string            `json:"resource" validate:"required"`
	Action      string            `json:"action" validate:"required"`
	Conditions  []PolicyCondition `json:"conditions,omitempty"`
	Effect      PolicyEffect      `json:"effect" validate:"required"`
}

// PolicyCondition represents a condition in a policy rule
type PolicyCondition struct {
	Field    string      `json:"field" validate:"required"`
	Operator string      `json:"operator" validate:"required"` // eq, ne, gt, lt, in, not_in, regex
	Value    interface{} `json:"value" validate:"required"`
}

// PolicyEffect represents the effect of a policy rule
type PolicyEffect string

const (
	PolicyEffectAllow PolicyEffect = "allow"
	PolicyEffectDeny  PolicyEffect = "deny"
	PolicyEffectAudit PolicyEffect = "audit"
)

// =============================================================================
// AUDIT AND EVENT TYPES
// =============================================================================

// AuditEvent represents an audit event
type AuditEvent struct {
	TypeMeta `json:",inline"`
	Metadata ObjectMeta     `json:"metadata"`
	Spec     AuditEventSpec `json:"spec"`
}

// AuditEventSpec contains the audit event data
type AuditEventSpec struct {
	Timestamp time.Time         `json:"timestamp" validate:"required"`
	Actor     Actor             `json:"actor" validate:"required"`
	Action    string            `json:"action" validate:"required"`
	Resource  ResourceReference `json:"resource" validate:"required"`
	Result    AuditResult       `json:"result" validate:"required"`
	Details   map[string]string `json:"details,omitempty"`
	UserAgent string            `json:"user_agent,omitempty"`
	IPAddress string            `json:"ip_address,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}

// Actor represents the entity that performed an action
type Actor struct {
	Type           ActorType `json:"type" validate:"required"`
	ID             string    `json:"id" validate:"required"`
	Name           string    `json:"name,omitempty"`
	Email          string    `json:"email,omitempty"`
	ServiceAccount string    `json:"service_account,omitempty"`
}

// ActorType represents the type of actor
type ActorType string

const (
	ActorTypeUser           ActorType = "user"
	ActorTypeServiceAccount ActorType = "service_account"
	ActorTypeSystem         ActorType = "system"
	ActorTypeBot            ActorType = "bot"
)

// ResourceReference represents a reference to a resource
type ResourceReference struct {
	APIVersion string `json:"api_version" validate:"required"`
	Kind       string `json:"kind" validate:"required"`
	Name       string `json:"name" validate:"required"`
	Namespace  string `json:"namespace,omitempty"`
	UID        string `json:"uid,omitempty"`
}

// AuditResult represents the result of an audited action
type AuditResult string

const (
	AuditResultSuccess AuditResult = "success"
	AuditResultFailure AuditResult = "failure"
	AuditResultDenied  AuditResult = "denied"
)

// =============================================================================
// LEGACY COMPATIBILITY TYPES
// =============================================================================

// DesiredState represents the desired state of a resource (legacy compatibility)
type DesiredState struct {
	APIVersion string                 `json:"apiVersion" validate:"required"`
	Kind       string                 `json:"kind" validate:"required"`
	Metadata   map[string]interface{} `json:"metadata"`
	Spec       map[string]interface{} `json:"spec"`
}

// ReconcileStatus represents the reconciliation status
type ReconcileStatus struct {
	Phase      ReconcilePhase `json:"phase"`
	Message    string         `json:"message,omitempty"`
	LastUpdate time.Time      `json:"last_update"`
	Conditions []Condition    `json:"conditions,omitempty"`
}

// ReconcilePhase represents the phase of reconciliation
type ReconcilePhase string

const (
	ReconcilePhasePending     ReconcilePhase = "pending"
	ReconcilePhaseReconciling ReconcilePhase = "reconciling"
	ReconcilePhaseCompleted   ReconcilePhase = "completed"
	ReconcilePhaseFailed      ReconcilePhase = "failed"
)

// Condition represents a condition in the reconciliation process
type Condition struct {
	Type               string    `json:"type"`
	Status             string    `json:"status"`
	LastTransitionTime time.Time `json:"last_transition_time"`
	Reason             string    `json:"reason,omitempty"`
	Message            string    `json:"message,omitempty"`
}

// APIResponse represents a standardized API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// APIError represents an API error
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Meta represents metadata in API responses
type Meta struct {
	RequestID  string      `json:"request_id,omitempty"`
	Timestamp  time.Time   `json:"timestamp"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}
