package applications

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/aykay76/ai-idp/internal/logger"
	"github.com/google/uuid"
)

// Handlers provides HTTP handlers for application resources using native Go HTTP
type Handlers struct {
	service *Service
	logger  *logger.Logger
}

// NewHandlers creates new application handlers
func NewHandlers(service *Service, appLogger *logger.Logger) *Handlers {
	return &Handlers{
		service: service,
		logger:  appLogger,
	}
}

// ListApplicationsResponse represents the response for listing applications
type ListApplicationsResponse struct {
	Applications []Application  `json:"applications"`
	Pagination   PaginationMeta `json:"pagination"`
}

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string    `json:"error"`
	Message string    `json:"message"`
	Code    string    `json:"code,omitempty"`
	Time    time.Time `json:"timestamp"`
}

// CreateApplication handles POST /api/v1/applications
func (h *Handlers) CreateApplication(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// For now, we'll use a hardcoded tenant ID until we implement proper tenant middleware
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001") // Default tenant

	// Parse request body
	var req CreateApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	// Basic validation
	if req.Name == "" {
		h.respondWithError(w, http.StatusBadRequest, "Name is required", nil)
		return
	}
	if req.DisplayName == "" {
		h.respondWithError(w, http.StatusBadRequest, "Display name is required", nil)
		return
	}

	// Create application
	app, err := h.service.CreateApplication(ctx, tenantID, &req)
	if err != nil {
		h.logger.WithFields(logger.LogFields{
			logger.FieldError: err.Error(),
			"name":            req.Name,
		}).Error("Failed to create application")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create application", err)
		return
	}

	h.logger.WithFields(logger.LogFields{
		"application_id": app.ID.String(),
		"name":           app.Name,
	}).Info("Application created successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(app)
}

// ListApplications handles GET /api/v1/applications
func (h *Handlers) ListApplications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// For now, we'll use a hardcoded tenant ID until we implement proper tenant middleware
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001") // Default tenant

	// Parse pagination parameters
	limit := h.parseQueryInt(r, "limit", 20)
	if limit > 100 {
		limit = 100 // Max limit
	}
	offset := h.parseQueryInt(r, "offset", 0)

	// Parse filter parameters
	teamName := r.URL.Query().Get("team_name")
	lifecycle := r.URL.Query().Get("lifecycle")
	status := r.URL.Query().Get("status")

	// Create list request
	listReq := &ListApplicationsRequest{
		TenantID:  tenantID,
		TeamName:  teamName,
		Lifecycle: lifecycle,
		Status:    status,
		Limit:     limit,
		Offset:    offset,
	}

	// Get applications
	apps, total, err := h.service.ListApplications(ctx, listReq)
	if err != nil {
		h.logger.WithFields(logger.LogFields{
			logger.FieldError: err.Error(),
		}).Error("Failed to list applications")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to list applications", err)
		return
	}

	response := ListApplicationsResponse{
		Applications: apps,
		Pagination: PaginationMeta{
			Limit:  limit,
			Offset: offset,
			Total:  total,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetApplication handles GET /api/v1/applications/{id}
func (h *Handlers) GetApplication(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract ID from path
	idStr := r.PathValue("id")
	if idStr == "" {
		h.respondWithError(w, http.StatusBadRequest, "Application ID is required", nil)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid application ID format", err)
		return
	}

	// For now, we'll use a hardcoded tenant ID until we implement proper tenant middleware
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001") // Default tenant

	// Get application
	app, err := h.service.GetApplication(ctx, tenantID, id)
	if err != nil {
		if err.Error() == "application not found" {
			h.respondWithError(w, http.StatusNotFound, "Application not found", err)
			return
		}
		h.logger.WithFields(logger.LogFields{
			logger.FieldError: err.Error(),
			"application_id":  id.String(),
		}).Error("Failed to get application")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get application", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(app)
}

// UpdateApplication handles PUT /api/v1/applications/{id}
func (h *Handlers) UpdateApplication(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract ID from path
	idStr := r.PathValue("id")
	if idStr == "" {
		h.respondWithError(w, http.StatusBadRequest, "Application ID is required", nil)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid application ID format", err)
		return
	}

	// Parse request body
	var req UpdateApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	// For now, we'll use a hardcoded tenant ID until we implement proper tenant middleware
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001") // Default tenant

	// Update application
	app, err := h.service.UpdateApplication(ctx, tenantID, id, &req)
	if err != nil {
		if err.Error() == "application not found" {
			h.respondWithError(w, http.StatusNotFound, "Application not found", err)
			return
		}
		h.logger.WithFields(logger.LogFields{
			logger.FieldError: err.Error(),
			"application_id":  id.String(),
		}).Error("Failed to update application")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update application", err)
		return
	}

	h.logger.WithFields(logger.LogFields{
		"application_id": app.ID.String(),
		"name":           app.Name,
	}).Info("Application updated successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(app)
}

// DeleteApplication handles DELETE /api/v1/applications/{id}
func (h *Handlers) DeleteApplication(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract ID from path
	idStr := r.PathValue("id")
	if idStr == "" {
		h.respondWithError(w, http.StatusBadRequest, "Application ID is required", nil)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid application ID format", err)
		return
	}

	// For now, we'll use a hardcoded tenant ID until we implement proper tenant middleware
	tenantID := uuid.MustParse("00000000-0000-0000-0000-000000000001") // Default tenant

	// Delete application
	err = h.service.DeleteApplication(ctx, tenantID, id)
	if err != nil {
		if err.Error() == "application not found" {
			h.respondWithError(w, http.StatusNotFound, "Application not found", err)
			return
		}
		h.logger.WithFields(logger.LogFields{
			logger.FieldError: err.Error(),
			"application_id":  id.String(),
		}).Error("Failed to delete application")
		h.respondWithError(w, http.StatusInternalServerError, "Failed to delete application", err)
		return
	}

	h.logger.WithFields(logger.LogFields{
		"application_id": id.String(),
	}).Info("Application deleted successfully")

	w.WriteHeader(http.StatusNoContent)
}

// Helper methods

func (h *Handlers) parseQueryInt(r *http.Request, key string, defaultValue int) int {
	if value := r.URL.Query().Get(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func (h *Handlers) respondWithError(w http.ResponseWriter, status int, message string, err error) {
	response := ErrorResponse{
		Error:   message,
		Message: message,
		Time:    time.Now().UTC(),
	}

	if err != nil {
		response.Message = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}
