package applications

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/aykay76/ai-idp/internal/server"
	"github.com/go-playground/validator/v10"
)

// ApplicationHandlers provides HTTP handlers for application resources
type ApplicationHandlers struct {
	service  *Service
	validate *validator.Validate
}

// NewApplicationHandlers creates new application handlers
func NewApplicationHandlers(service *Service) *ApplicationHandlers {
	return &ApplicationHandlers{
		service:  service,
		validate: validator.New(),
	}
}

// GetCRUDHandlers returns the CRUD handlers for applications
func (h *ApplicationHandlers) GetCRUDHandlers() *server.CRUDHandlers {
	return &server.CRUDHandlers{
		Create: server.WithTenantValidation(h.CreateApplication),
		List:   server.WithTenantValidation(h.ListApplications),
		Get:    server.WithTenantValidation(h.GetApplication),
		Update: server.WithTenantValidation(h.UpdateApplication),
		Delete: server.WithTenantValidation(h.DeleteApplication),
	}
}

// CreateApplication handles POST /applications
func (h *ApplicationHandlers) CreateApplication(w http.ResponseWriter, r *http.Request) {
	// Get tenant context
	tenantCtx, err := server.GetTenantFromContext(r)
	if err != nil {
		server.RespondWithError(w, http.StatusBadRequest, err, "Invalid tenant context")
		return
	}

	// Parse request body
	var req CreateApplicationRequest
	if err := server.ParseJSONBody(r, &req); err != nil {
		server.RespondWithValidationError(w, err)
		return
	}

	// Validate request
	if err := h.validate.Struct(&req); err != nil {
		server.RespondWithValidationError(w, err)
		return
	}

	// Create application
	app, err := h.service.CreateApplication(r.Context(), tenantCtx.TenantID, &req, tenantCtx.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			server.RespondWithError(w, http.StatusConflict, err, "Application name already exists")
			return
		}
		server.RespondWithInternalError(w, err)
		return
	}

	server.RespondCreated(w, app)
}

// ListApplications handles GET /applications
func (h *ApplicationHandlers) ListApplications(w http.ResponseWriter, r *http.Request) {
	// Get tenant context
	tenantCtx, err := server.GetTenantFromContext(r)
	if err != nil {
		server.RespondWithError(w, http.StatusBadRequest, err, "Invalid tenant context")
		return
	}

	// Parse pagination parameters
	pagination, err := server.ParsePaginationParams(r)
	if err != nil {
		server.RespondWithError(w, http.StatusBadRequest, err, "Invalid pagination parameters")
		return
	}

	// Parse filter parameters
	filters := server.ParseFilterParams(r)

	// Build list request
	req := &ListApplicationsRequest{
		TenantID:  tenantCtx.TenantID,
		TeamName:  filters.TeamName,
		Lifecycle: server.QueryParam(r, "lifecycle"),
		Status:    filters.Status,
		Limit:     pagination.Limit,
		Offset:    pagination.Offset,
	}

	// List applications
	applications, err := h.service.ListApplications(r.Context(), req)
	if err != nil {
		server.RespondWithInternalError(w, err)
		return
	}

	server.RespondWithData(w, map[string]interface{}{
		"applications": applications,
		"count":        len(applications),
		"filters": map[string]interface{}{
			"team_name": req.TeamName,
			"lifecycle": req.Lifecycle,
			"status":    req.Status,
			"limit":     req.Limit,
			"offset":    req.Offset,
		},
	})
}

// GetApplication handles GET /applications/{id}
func (h *ApplicationHandlers) GetApplication(w http.ResponseWriter, r *http.Request) {
	// Get tenant context
	tenantCtx, err := server.GetTenantFromContext(r)
	if err != nil {
		server.RespondWithError(w, http.StatusBadRequest, err, "Invalid tenant context")
		return
	}

	// Extract application ID from path
	applicationIDStr := server.PathParam(r, "id")
	applicationID, err := server.ParseUUIDParam(applicationIDStr, "application ID")
	if err != nil {
		server.RespondWithError(w, http.StatusBadRequest, err, "Invalid application ID")
		return
	}

	// Get application
	app, err := h.service.GetApplication(r.Context(), tenantCtx.TenantID, applicationID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			server.RespondWithNotFound(w, "Application")
			return
		}
		server.RespondWithInternalError(w, err)
		return
	}

	server.RespondWithData(w, app)
}

// UpdateApplication handles PUT /applications/{id}
func (h *ApplicationHandlers) UpdateApplication(w http.ResponseWriter, r *http.Request) {
	// Get tenant context
	tenantCtx, err := server.GetTenantFromContext(r)
	if err != nil {
		server.RespondWithError(w, http.StatusBadRequest, err, "Invalid tenant context")
		return
	}

	// Extract application ID from path
	applicationIDStr := server.PathParam(r, "id")
	applicationID, err := server.ParseUUIDParam(applicationIDStr, "application ID")
	if err != nil {
		server.RespondWithError(w, http.StatusBadRequest, err, "Invalid application ID")
		return
	}

	// Parse request body
	var req UpdateApplicationRequest
	if err := server.ParseJSONBody(r, &req); err != nil {
		server.RespondWithValidationError(w, err)
		return
	}

	// Validate request
	if err := h.validate.Struct(&req); err != nil {
		server.RespondWithValidationError(w, err)
		return
	}

	// Update application
	app, err := h.service.UpdateApplication(r.Context(), tenantCtx.TenantID, applicationID, &req, tenantCtx.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			server.RespondWithNotFound(w, "Application")
			return
		}
		if strings.Contains(err.Error(), "no fields to update") {
			server.RespondWithError(w, http.StatusBadRequest, err, "No fields provided for update")
			return
		}
		server.RespondWithInternalError(w, err)
		return
	}

	server.RespondWithData(w, app)
}

// DeleteApplication handles DELETE /applications/{id}
func (h *ApplicationHandlers) DeleteApplication(w http.ResponseWriter, r *http.Request) {
	// Get tenant context
	tenantCtx, err := server.GetTenantFromContext(r)
	if err != nil {
		server.RespondWithError(w, http.StatusBadRequest, err, "Invalid tenant context")
		return
	}

	// Extract application ID from path
	applicationIDStr := server.PathParam(r, "id")
	applicationID, err := server.ParseUUIDParam(applicationIDStr, "application ID")
	if err != nil {
		server.RespondWithError(w, http.StatusBadRequest, err, "Invalid application ID")
		return
	}

	// Delete application
	err = h.service.DeleteApplication(r.Context(), tenantCtx.TenantID, applicationID, tenantCtx.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			server.RespondWithNotFound(w, "Application")
			return
		}
		server.RespondWithInternalError(w, err)
		return
	}

	server.RespondWithMessage(w, "Application deleted successfully")
}

// Additional helper handlers

// GetApplicationsByTeam handles GET /applications/by-team/{teamName}
func (h *ApplicationHandlers) GetApplicationsByTeam(w http.ResponseWriter, r *http.Request) {
	// Get tenant context
	tenantCtx, err := server.GetTenantFromContext(r)
	if err != nil {
		server.RespondWithError(w, http.StatusBadRequest, err, "Invalid tenant context")
		return
	}

	// Extract team name from path
	teamName := server.PathParam(r, "teamName")
	if teamName == "" {
		server.RespondWithError(w, http.StatusBadRequest,
			fmt.Errorf("missing team name"), "Team name is required")
		return
	}

	// Parse pagination parameters
	pagination, err := server.ParsePaginationParams(r)
	if err != nil {
		server.RespondWithError(w, http.StatusBadRequest, err, "Invalid pagination parameters")
		return
	}

	// Build list request
	req := &ListApplicationsRequest{
		TenantID: tenantCtx.TenantID,
		TeamName: teamName,
		Limit:    pagination.Limit,
		Offset:   pagination.Offset,
	}

	// List applications for team
	applications, err := h.service.ListApplications(r.Context(), req)
	if err != nil {
		server.RespondWithInternalError(w, err)
		return
	}

	server.RespondWithData(w, map[string]interface{}{
		"applications": applications,
		"team_name":    teamName,
		"count":        len(applications),
	})
}

// GetApplicationStats handles GET /applications/stats
func (h *ApplicationHandlers) GetApplicationStats(w http.ResponseWriter, r *http.Request) {
	// Get tenant context
	tenantCtx, err := server.GetTenantFromContext(r)
	if err != nil {
		server.RespondWithError(w, http.StatusBadRequest, err, "Invalid tenant context")
		return
	}

	// Get all applications for stats
	req := &ListApplicationsRequest{
		TenantID: tenantCtx.TenantID,
		Limit:    1000, // Large limit to get all applications
		Offset:   0,
	}

	applications, err := h.service.ListApplications(r.Context(), req)
	if err != nil {
		server.RespondWithInternalError(w, err)
		return
	}

	// Calculate statistics
	stats := calculateApplicationStats(applications)

	server.RespondWithData(w, stats)
}

// calculateApplicationStats calculates statistics for applications
func calculateApplicationStats(applications []*Application) map[string]interface{} {
	stats := map[string]interface{}{
		"total":        len(applications),
		"by_lifecycle": make(map[string]int),
		"by_status":    make(map[string]int),
		"by_team":      make(map[string]int),
	}

	lifecycleStats := stats["by_lifecycle"].(map[string]int)
	statusStats := stats["by_status"].(map[string]int)
	teamStats := stats["by_team"].(map[string]int)

	for _, app := range applications {
		lifecycleStats[app.Lifecycle]++
		statusStats[app.Status]++
		teamStats[app.TeamName]++
	}

	return stats
}
