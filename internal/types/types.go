package types

import (
	"time"
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
)

// Application represents an application in the platform
type Application struct {
	ID           string                 `json:"id" validate:"required"`
	Name         string                 `json:"name" validate:"required"`
	Description  string                 `json:"description"`
	TenantID     string                 `json:"tenant_id" validate:"required"`
	DesiredState map[string]interface{} `json:"desired_state"`
	ActualState  map[string]interface{} `json:"actual_state"`
	Status       ApplicationStatus      `json:"status"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// ApplicationStatus represents the status of an application
type ApplicationStatus string

const (
	ApplicationStatusPending     ApplicationStatus = "pending"
	ApplicationStatusDeploying   ApplicationStatus = "deploying"
	ApplicationStatusRunning     ApplicationStatus = "running"
	ApplicationStatusFailed      ApplicationStatus = "failed"
	ApplicationStatusTerminating ApplicationStatus = "terminating"
)

// TenantContext represents tenant information in requests
type TenantContext struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// DesiredState represents the desired state of a resource
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
