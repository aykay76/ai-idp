package applications

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aykay76/ai-idp/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Service provides clean application management operations using native Go HTTP
type Service struct {
	db *database.Pool
}

// NewService creates a new clean application service
func NewService(db *database.Pool) *Service {
	return &Service{db: db}
}

// Application represents an application in the simplified model
type Application struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	TenantID    uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	Name        string                 `json:"name" db:"name"`
	DisplayName string                 `json:"display_name" db:"display_name"`
	Description *string                `json:"description,omitempty" db:"description"`
	TeamName    string                 `json:"team_name" db:"team_name"`
	OwnerEmail  string                 `json:"owner_email" db:"owner_email"`
	Lifecycle   string                 `json:"lifecycle" db:"lifecycle"` // development, testing, staging, production
	Status      string                 `json:"status" db:"status"`       // pending, running, failed, stopped
	Config      map[string]interface{} `json:"config" db:"config"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	CreatedBy   string                 `json:"created_by" db:"created_by"`
	UpdatedBy   *string                `json:"updated_by,omitempty" db:"updated_by"`
}

// ListApplicationsRequest represents a request to list applications
type ListApplicationsRequest struct {
	TenantID  uuid.UUID
	TeamName  string
	Lifecycle string
	Status    string
	Limit     int
	Offset    int
}

// CreateApplicationRequest represents a request to create a new application
type CreateApplicationRequest struct {
	Name        string                 `json:"name" validate:"required,min=1,max=63"`
	DisplayName string                 `json:"display_name" validate:"required,min=1,max=255"`
	Description *string                `json:"description,omitempty"`
	TeamName    string                 `json:"team_name" validate:"required"`
	OwnerEmail  string                 `json:"owner_email" validate:"required,email"`
	Lifecycle   string                 `json:"lifecycle" validate:"required,oneof=development staging production deprecated"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// UpdateApplicationRequest represents a request to update an application
type UpdateApplicationRequest struct {
	DisplayName *string                 `json:"display_name,omitempty"`
	Description *string                 `json:"description,omitempty"`
	TeamName    *string                 `json:"team_name,omitempty"`
	OwnerEmail  *string                 `json:"owner_email,omitempty"`
	Lifecycle   *string                 `json:"lifecycle,omitempty"`
	Config      *map[string]interface{} `json:"config,omitempty"`
}

// CreateApplication creates a new application
func (s *Service) CreateApplication(ctx context.Context, tenantID uuid.UUID, req *CreateApplicationRequest) (*Application, error) {
	app := &Application{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		TeamName:    req.TeamName,
		OwnerEmail:  req.OwnerEmail,
		Lifecycle:   req.Lifecycle,
		Status:      "pending",
		Config:      req.Config,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		CreatedBy:   "system", // TODO: Get from context when auth is implemented
	}

	if app.Config == nil {
		app.Config = make(map[string]interface{})
	}

	configJSON, err := json.Marshal(app.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	query := `
		INSERT INTO resource_management.applications (
			id, tenant_id, name, display_name, description, team_name, 
			owner_email, lifecycle, status, observability_config, created_at, updated_at, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`

	_, err = s.db.Exec(ctx, query,
		app.ID, app.TenantID, app.Name, app.DisplayName, app.Description,
		app.TeamName, app.OwnerEmail, app.Lifecycle, app.Status,
		configJSON, app.CreatedAt, app.UpdatedAt, app.CreatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	return app, nil
}

// ListApplications lists applications for a tenant
func (s *Service) ListApplications(ctx context.Context, req *ListApplicationsRequest) ([]Application, int, error) {
	whereClause := "WHERE tenant_id = $1"
	args := []interface{}{req.TenantID}
	argCount := 1

	if req.TeamName != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND team_name = $%d", argCount)
		args = append(args, req.TeamName)
	}

	if req.Lifecycle != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND lifecycle = $%d", argCount)
		args = append(args, req.Lifecycle)
	}

	if req.Status != "" {
		argCount++
		whereClause += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, req.Status)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM resource_management.applications " + whereClause
	var total int
	err := s.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count applications: %w", err)
	}

	// Get applications with pagination
	query := `
		SELECT id, tenant_id, name, display_name, description, team_name, 
		       owner_email, lifecycle, status, observability_config, created_at, updated_at, 
		       created_by, updated_by
		FROM resource_management.applications 
	` + whereClause + `
		ORDER BY created_at DESC 
		LIMIT $` + fmt.Sprintf("%d", argCount+1) + ` OFFSET $` + fmt.Sprintf("%d", argCount+2)

	args = append(args, req.Limit, req.Offset)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query applications: %w", err)
	}
	defer rows.Close()

	var applications []Application
	for rows.Next() {
		var app Application
		var configJSON []byte

		err := rows.Scan(
			&app.ID, &app.TenantID, &app.Name, &app.DisplayName, &app.Description,
			&app.TeamName, &app.OwnerEmail, &app.Lifecycle, &app.Status,
			&configJSON, &app.CreatedAt, &app.UpdatedAt, &app.CreatedBy, &app.UpdatedBy,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan application row: %w", err)
		}

		if err := json.Unmarshal(configJSON, &app.Config); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal config: %w", err)
		}

		applications = append(applications, app)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating application rows: %w", err)
	}

	return applications, total, nil
}

// GetApplication gets an application by ID
func (s *Service) GetApplication(ctx context.Context, tenantID, id uuid.UUID) (*Application, error) {
	query := `
		SELECT id, tenant_id, name, display_name, description, team_name, 
		       owner_email, lifecycle, status, observability_config, created_at, updated_at, 
		       created_by, updated_by
		FROM resource_management.applications 
		WHERE tenant_id = $1 AND id = $2
	`

	var app Application
	var configJSON []byte

	err := s.db.QueryRow(ctx, query, tenantID, id).Scan(
		&app.ID, &app.TenantID, &app.Name, &app.DisplayName, &app.Description,
		&app.TeamName, &app.OwnerEmail, &app.Lifecycle, &app.Status,
		&configJSON, &app.CreatedAt, &app.UpdatedAt, &app.CreatedBy, &app.UpdatedBy,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("application not found")
		}
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	if err := json.Unmarshal(configJSON, &app.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &app, nil
}

// UpdateApplication updates an application
func (s *Service) UpdateApplication(ctx context.Context, tenantID, id uuid.UUID, req *UpdateApplicationRequest) (*Application, error) {
	// First get the existing application
	app, err := s.GetApplication(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.DisplayName != nil {
		app.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		app.Description = req.Description
	}
	if req.TeamName != nil {
		app.TeamName = *req.TeamName
	}
	if req.OwnerEmail != nil {
		app.OwnerEmail = *req.OwnerEmail
	}
	if req.Lifecycle != nil {
		app.Lifecycle = *req.Lifecycle
	}
	if req.Config != nil {
		app.Config = *req.Config
	}

	app.UpdatedAt = time.Now().UTC()
	app.UpdatedBy = &[]string{"system"}[0] // TODO: Get from context when auth is implemented

	configJSON, err := json.Marshal(app.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	query := `
		UPDATE resource_management.applications 
		SET display_name = $3, description = $4, team_name = $5, owner_email = $6, 
		    lifecycle = $7, observability_config = $8, updated_at = $9, updated_by = $10
		WHERE tenant_id = $1 AND id = $2
	`

	_, err = s.db.Exec(ctx, query,
		tenantID, id, app.DisplayName, app.Description, app.TeamName,
		app.OwnerEmail, app.Lifecycle, configJSON, app.UpdatedAt, app.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update application: %w", err)
	}

	return app, nil
}

// DeleteApplication deletes an application
func (s *Service) DeleteApplication(ctx context.Context, tenantID, id uuid.UUID) error {
	query := "DELETE FROM resource_management.applications WHERE tenant_id = $1 AND id = $2"

	result, err := s.db.Exec(ctx, query, tenantID, id)
	if err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("application not found")
	}

	return nil
}
