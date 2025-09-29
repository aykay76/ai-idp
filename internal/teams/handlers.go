package teams

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/aykay76/ai-idp/internal/logger"
	"github.com/google/uuid"
)

// Handlers provides HTTP handlers for team resources using native Go HTTP
type Handlers struct {
	service TeamService
	logger  *logger.Logger
}

// NewHandlers creates new team handlers
func NewHandlers(service TeamService, appLogger *logger.Logger) *Handlers {
	return &Handlers{
		service: service,
		logger:  appLogger,
	}
}

// ListTeamsResponse represents the response for listing teams
type ListTeamsResponse struct {
	Teams      []Team         `json:"teams"`
	Pagination PaginationMeta `json:"pagination"`
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

// CreateTeam handles POST /api/v1/teams
func (h *Handlers) CreateTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	h.logger.WithFields(logger.LogFields{
		logger.FieldHTTPMethod: r.Method,
		logger.FieldHTTPPath:   r.URL.Path,
	}).Debug("Creating new team")

	// Parse request body
	var teamReq Team
	if err := json.NewDecoder(r.Body).Decode(&teamReq); err != nil {
		h.logger.WithFields(logger.LogFields{
			logger.FieldError: err.Error(),
		}).Error("Failed to decode team request")

		h.writeError(w, "Invalid JSON in request body", http.StatusBadRequest, "INVALID_JSON")
		return
	}

	// Create team using service
	team, err := h.service.CreateTeam(ctx, teamReq)
	if err != nil {
		h.logger.WithFields(logger.LogFields{
			logger.FieldError: err.Error(),
		}).Error("Failed to create team")

		h.writeError(w, "Failed to create team", http.StatusInternalServerError, "CREATE_FAILED")
		return
	}

	h.logger.WithFields(logger.LogFields{
		"team_id":   team.ID,
		"team_name": team.Name,
	}).Info("Team created successfully")

	// Return created team
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(team); err != nil {
		h.logger.WithFields(logger.LogFields{
			logger.FieldError: err.Error(),
		}).Error("Failed to encode team response")
	}
}

// GetTeam handles GET /api/v1/teams/{id}
func (h *Handlers) GetTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract team ID from path
	teamID := r.PathValue("id")
	if teamID == "" {
		h.writeError(w, "Team ID is required", http.StatusBadRequest, "MISSING_TEAM_ID")
		return
	}

	// Parse team ID as UUID
	id, err := uuid.Parse(teamID)
	if err != nil {
		h.logger.WithFields(logger.LogFields{
			logger.FieldError: err.Error(),
			"team_id":         teamID,
		}).Error("Invalid team ID format")

		h.writeError(w, "Invalid team ID format", http.StatusBadRequest, "INVALID_TEAM_ID")
		return
	}

	h.logger.WithFields(logger.LogFields{
		logger.FieldHTTPMethod: r.Method,
		logger.FieldHTTPPath:   r.URL.Path,
		"team_id":              id.String(),
	}).Debug("Getting team")

	// Get team using service
	team, err := h.service.GetTeam(ctx, id)
	if err != nil {
		if err == ErrTeamNotFound {
			h.writeError(w, "Team not found", http.StatusNotFound, "TEAM_NOT_FOUND")
			return
		}

		h.logger.WithFields(logger.LogFields{
			logger.FieldError: err.Error(),
			"team_id":         id.String(),
		}).Error("Failed to get team")

		h.writeError(w, "Failed to get team", http.StatusInternalServerError, "GET_FAILED")
		return
	}

	// Return team
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(team); err != nil {
		h.logger.WithFields(logger.LogFields{
			logger.FieldError: err.Error(),
		}).Error("Failed to encode team response")
	}
}

// ListTeams handles GET /api/v1/teams
func (h *Handlers) ListTeams(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	h.logger.WithFields(logger.LogFields{
		logger.FieldHTTPMethod: r.Method,
		logger.FieldHTTPPath:   r.URL.Path,
	}).Debug("Listing teams")

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Set default values
	limit := 50
	offset := 0

	// Parse limit if provided
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Parse offset if provided
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// List teams using service
	teams, total, err := h.service.ListTeams(ctx, limit, offset)
	if err != nil {
		h.logger.WithFields(logger.LogFields{
			logger.FieldError: err.Error(),
		}).Error("Failed to list teams")

		h.writeError(w, "Failed to list teams", http.StatusInternalServerError, "LIST_FAILED")
		return
	}

	// Create response
	response := ListTeamsResponse{
		Teams: teams,
		Pagination: PaginationMeta{
			Limit:  limit,
			Offset: offset,
			Total:  total,
		},
	}

	// Return teams list
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.WithFields(logger.LogFields{
			logger.FieldError: err.Error(),
		}).Error("Failed to encode teams response")
	}
}

// UpdateTeam handles PUT /api/v1/teams/{id}
func (h *Handlers) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract team ID from path
	teamID := r.PathValue("id")
	if teamID == "" {
		h.writeError(w, "Team ID is required", http.StatusBadRequest, "MISSING_TEAM_ID")
		return
	}

	// Parse team ID as UUID
	id, err := uuid.Parse(teamID)
	if err != nil {
		h.logger.WithFields(logger.LogFields{
			logger.FieldError: err.Error(),
			"team_id":         teamID,
		}).Error("Invalid team ID format")

		h.writeError(w, "Invalid team ID format", http.StatusBadRequest, "INVALID_TEAM_ID")
		return
	}

	h.logger.WithFields(logger.LogFields{
		logger.FieldHTTPMethod: r.Method,
		logger.FieldHTTPPath:   r.URL.Path,
		"team_id":              id.String(),
	}).Debug("Updating team")

	// Parse request body
	var teamReq Team
	if err := json.NewDecoder(r.Body).Decode(&teamReq); err != nil {
		h.logger.WithFields(logger.LogFields{
			logger.FieldError: err.Error(),
		}).Error("Failed to decode team update request")

		h.writeError(w, "Invalid JSON in request body", http.StatusBadRequest, "INVALID_JSON")
		return
	}

	// Set the ID from path
	teamReq.ID = id

	// Update team using service
	team, err := h.service.UpdateTeam(ctx, teamReq)
	if err != nil {
		if err == ErrTeamNotFound {
			h.writeError(w, "Team not found", http.StatusNotFound, "TEAM_NOT_FOUND")
			return
		}

		h.logger.WithFields(logger.LogFields{
			logger.FieldError: err.Error(),
			"team_id":         id.String(),
		}).Error("Failed to update team")

		h.writeError(w, "Failed to update team", http.StatusInternalServerError, "UPDATE_FAILED")
		return
	}

	h.logger.WithFields(logger.LogFields{
		"team_id":   team.ID,
		"team_name": team.Name,
	}).Info("Team updated successfully")

	// Return updated team
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(team); err != nil {
		h.logger.WithFields(logger.LogFields{
			logger.FieldError: err.Error(),
		}).Error("Failed to encode team response")
	}
}

// DeleteTeam handles DELETE /api/v1/teams/{id}
func (h *Handlers) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract team ID from path
	teamID := r.PathValue("id")
	if teamID == "" {
		h.writeError(w, "Team ID is required", http.StatusBadRequest, "MISSING_TEAM_ID")
		return
	}

	// Parse team ID as UUID
	id, err := uuid.Parse(teamID)
	if err != nil {
		h.logger.WithFields(logger.LogFields{
			logger.FieldError: err.Error(),
			"team_id":         teamID,
		}).Error("Invalid team ID format")

		h.writeError(w, "Invalid team ID format", http.StatusBadRequest, "INVALID_TEAM_ID")
		return
	}

	h.logger.WithFields(logger.LogFields{
		logger.FieldHTTPMethod: r.Method,
		logger.FieldHTTPPath:   r.URL.Path,
		"team_id":              id.String(),
	}).Debug("Deleting team")

	// Delete team using service
	err = h.service.DeleteTeam(ctx, id)
	if err != nil {
		if err == ErrTeamNotFound {
			h.writeError(w, "Team not found", http.StatusNotFound, "TEAM_NOT_FOUND")
			return
		}

		h.logger.WithFields(logger.LogFields{
			logger.FieldError: err.Error(),
			"team_id":         id.String(),
		}).Error("Failed to delete team")

		h.writeError(w, "Failed to delete team", http.StatusInternalServerError, "DELETE_FAILED")
		return
	}

	h.logger.WithFields(logger.LogFields{
		"team_id": id.String(),
	}).Info("Team deleted successfully")

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// writeError writes an error response
func (h *Handlers) writeError(w http.ResponseWriter, message string, statusCode int, code string) {
	response := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
		Code:    code,
		Time:    time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.WithFields(logger.LogFields{
			logger.FieldError: err.Error(),
		}).Error("Failed to encode error response")
	}
}
