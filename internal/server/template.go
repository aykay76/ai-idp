package server

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// ServiceTemplate provides a reusable template for creating microservices
type ServiceTemplate struct {
	*Server
	name    string
	version string
}

// NewServiceTemplate creates a new service template
func NewServiceTemplate(name, version string, config *Config) *ServiceTemplate {
	config.ServiceName = name
	server := NewServer(config)

	template := &ServiceTemplate{
		Server:  server,
		name:    name,
		version: version,
	}

	// Add service-specific endpoints
	template.HandleFunc("GET /version", template.versionHandler)
	template.HandleFunc("GET /metrics", template.metricsHandler)

	return template
}

// RegisterCRUDRoutes registers standard CRUD routes for a resource
func (st *ServiceTemplate) RegisterCRUDRoutes(basePath string, handlers *CRUDHandlers) {
	// Ensure basePath starts and ends correctly
	if basePath[0] != '/' {
		basePath = "/" + basePath
	}
	if basePath[len(basePath)-1] == '/' {
		basePath = basePath[:len(basePath)-1]
	}

	// Register CRUD routes
	st.HandleFunc(fmt.Sprintf("POST %s", basePath), handlers.Create)
	st.HandleFunc(fmt.Sprintf("GET %s", basePath), handlers.List)
	st.HandleFunc(fmt.Sprintf("GET %s/{id}", basePath), handlers.Get)
	st.HandleFunc(fmt.Sprintf("PUT %s/{id}", basePath), handlers.Update)
	st.HandleFunc(fmt.Sprintf("DELETE %s/{id}", basePath), handlers.Delete)

	// Optional additional routes
	if handlers.Patch != nil {
		st.HandleFunc(fmt.Sprintf("PATCH %s/{id}", basePath), handlers.Patch)
	}
}

// CRUDHandlers contains handlers for standard CRUD operations
type CRUDHandlers struct {
	Create http.HandlerFunc
	List   http.HandlerFunc
	Get    http.HandlerFunc
	Update http.HandlerFunc
	Delete http.HandlerFunc
	Patch  http.HandlerFunc // Optional
}

// TenantContext provides tenant information extracted from request
type TenantContext struct {
	TenantID uuid.UUID
	UserID   string
}

// ExtractTenantContext extracts tenant context from HTTP request headers
func ExtractTenantContext(r *http.Request) (*TenantContext, error) {
	// Extract tenant ID from header
	tenantIDStr, err := GetHeaderValue(r, "X-Tenant-ID", true)
	if err != nil {
		return nil, fmt.Errorf("missing tenant context: %w", err)
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	// Extract user ID (optional in development)
	userID, _ := GetHeaderValue(r, "X-User-Email", false)
	if userID == "" {
		userID = "system" // Default for development
	}

	return &TenantContext{
		TenantID: tenantID,
		UserID:   userID,
	}, nil
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Limit  int
	Offset int
}

// ParsePaginationParams parses limit and offset from query parameters
func ParsePaginationParams(r *http.Request) (*PaginationParams, error) {
	query := r.URL.Query()

	params := &PaginationParams{
		Limit:  50, // default limit
		Offset: 0,  // default offset
	}

	if limitStr := query.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, fmt.Errorf("invalid limit parameter: %w", err)
		}
		if limit < 1 || limit > 1000 {
			return nil, fmt.Errorf("limit must be between 1 and 1000")
		}
		params.Limit = limit
	}

	if offsetStr := query.Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return nil, fmt.Errorf("invalid offset parameter: %w", err)
		}
		if offset < 0 {
			return nil, fmt.Errorf("offset must be non-negative")
		}
		params.Offset = offset
	}

	return params, nil
}

// FilterParams represents common filtering parameters
type FilterParams struct {
	Search    string
	Status    string
	CreatedBy string
	TeamName  string
	// Add more common filters as needed
}

// ParseFilterParams parses common filter parameters from query string
func ParseFilterParams(r *http.Request) *FilterParams {
	query := r.URL.Query()

	return &FilterParams{
		Search:    query.Get("search"),
		Status:    query.Get("status"),
		CreatedBy: query.Get("created_by"),
		TeamName:  query.Get("team_name"),
	}
}

// PathParam extracts a path parameter from the request URL
func PathParam(r *http.Request, paramName string) string {
	return r.PathValue(paramName)
}

// QueryParam gets a query parameter value
func QueryParam(r *http.Request, paramName string) string {
	return r.URL.Query().Get(paramName)
}

// RequiredQueryParam gets a required query parameter
func RequiredQueryParam(r *http.Request, paramName string) (string, error) {
	value := r.URL.Query().Get(paramName)
	if value == "" {
		return "", fmt.Errorf("missing required query parameter: %s", paramName)
	}
	return value, nil
}

// ParseUUIDParam parses a UUID from path or query parameters
func ParseUUIDParam(value string, paramName string) (uuid.UUID, error) {
	if value == "" {
		return uuid.Nil, fmt.Errorf("missing %s parameter", paramName)
	}

	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid %s format: %w", paramName, err)
	}

	return id, nil
}

// WithTenantValidation wraps a handler with tenant context validation
func WithTenantValidation(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantCtx, err := ExtractTenantContext(r)
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, err, "Invalid tenant context")
			return
		}

		// Add tenant context to request context
		ctx := context.WithValue(r.Context(), "tenant", tenantCtx)
		r = r.WithContext(ctx)

		handler(w, r)
	}
}

// GetTenantFromContext extracts tenant context from request context
func GetTenantFromContext(r *http.Request) (*TenantContext, error) {
	tenantCtx, ok := r.Context().Value("tenant").(*TenantContext)
	if !ok {
		return nil, fmt.Errorf("tenant context not found")
	}
	return tenantCtx, nil
}

// ValidationError represents a validation error with field details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "validation failed"
	}
	return fmt.Sprintf("validation failed: %s", ve[0].Message)
}

// RespondWithValidationErrors sends validation errors response
func RespondWithValidationErrors(w http.ResponseWriter, errors ValidationErrors) {
	response := &ErrorResponse{
		Error:   errors.Error(),
		Message: "Validation failed",
		Code:    http.StatusBadRequest,
		Details: map[string]interface{}{
			"validation_errors": errors,
		},
		Time: time.Now().UTC().Format(time.RFC3339),
	}
	RespondWithJSON(w, http.StatusBadRequest, response)
}

// Service template handlers

// versionHandler returns service version information
func (st *ServiceTemplate) versionHandler(w http.ResponseWriter, r *http.Request) {
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"service": st.name,
		"version": st.version,
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

// metricsHandler returns basic service metrics
func (st *ServiceTemplate) metricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := map[string]interface{}{
		"service": st.name,
		"uptime":  time.Since(startTime).String(),
		"time":    time.Now().UTC().Format(time.RFC3339),
	}

	// Add database metrics if available
	if st.database != nil {
		stats := st.database.Stats()
		metrics["database"] = map[string]interface{}{
			"total_connections": stats.TotalConnections,
			"idle_connections":  stats.IdleConnections,
			"used_connections":  stats.UsedConnections,
		}
	}

	RespondWithJSON(w, http.StatusOK, metrics)
}

// ResourceTemplate provides a template for resource-based services
type ResourceTemplate struct {
	*ServiceTemplate
	resourceName string
	resourcePath string
}

// NewResourceTemplate creates a template for managing a specific resource type
func NewResourceTemplate(serviceName, resourceName string, config *Config) *ResourceTemplate {
	template := NewServiceTemplate(serviceName, "1.0.0", config)

	return &ResourceTemplate{
		ServiceTemplate: template,
		resourceName:    resourceName,
		resourcePath:    fmt.Sprintf("/api/v1/%s", resourceName),
	}
}

// RegisterResourceHandlers registers CRUD handlers for the resource
func (rt *ResourceTemplate) RegisterResourceHandlers(handlers *CRUDHandlers) {
	rt.RegisterCRUDRoutes(rt.resourcePath, handlers)
}

// Helper functions

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv(serviceName, defaultPort string) *Config {
	return &Config{
		Port:            getEnv("PORT", defaultPort),
		ServiceName:     getEnv("SERVICE_NAME", serviceName),
		Environment:     getEnv("ENVIRONMENT", "development"),
		Debug:           getEnv("DEBUG", "false") == "true",
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://platform:platform_dev_password@localhost:5432/platform?sslmode=disable"),
		RedisURL:        getEnv("REDIS_URL", "redis://:redis_dev_password@localhost:6379/0"),
		JWTSecret:       getEnv("JWT_SECRET", "dev_jwt_secret_change_in_production"),
		ShutdownTimeout: 30 * time.Second,
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// BuildURL builds a URL with query parameters
func BuildURL(base string, params url.Values) string {
	if len(params) == 0 {
		return base
	}
	return base + "?" + params.Encode()
}

// Package-level variables
var startTime = time.Now()
