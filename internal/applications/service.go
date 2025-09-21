package applications

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aykay76/ai-idp/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Service provides application management operations
type Service struct {
	db *database.Pool
}

// NewService creates a new application service
func NewService(db *database.Pool) *Service {
	return &Service{db: db}
}

// Application represents an application resource
type Application struct {
	ID                  uuid.UUID              `json:"id" db:"id"`
	TenantID            uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	Name                string                 `json:"name" db:"name"`
	DisplayName         string                 `json:"display_name" db:"display_name"`
	Description         *string                `json:"description,omitempty" db:"description"`
	TeamName            string                 `json:"team_name" db:"team_name"`
	OwnerEmail          string                 `json:"owner_email" db:"owner_email"`
	Lifecycle           string                 `json:"lifecycle" db:"lifecycle"`
	EnvironmentName     *string                `json:"environment_name,omitempty" db:"environment_name"`
	EnvironmentRegion   *string                `json:"environment_region,omitempty" db:"environment_region"`
	ResourceQuota       map[string]interface{} `json:"resource_quota" db:"resource_quota"`
	ComplianceSettings  map[string]interface{} `json:"compliance_settings" db:"compliance_settings"`
	Dependencies        []interface{}          `json:"dependencies" db:"dependencies"`
	ObservabilityConfig map[string]interface{} `json:"observability_config" db:"observability_config"`
	Status              string                 `json:"status" db:"status"`
	Conditions          []interface{}          `json:"conditions" db:"conditions"`
	CurrentResources    map[string]interface{} `json:"current_resources" db:"current_resources"`
	CreatedAt           time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at" db:"updated_at"`
	CreatedBy           string                 `json:"created_by" db:"created_by"`
	UpdatedBy           *string                `json:"updated_by,omitempty" db:"updated_by"`
}

// CreateApplicationRequest represents a request to create a new application
type CreateApplicationRequest struct {
	Name                string                 `json:"name" validate:"required,min=2,max=255"`
	DisplayName         string                 `json:"display_name" validate:"required,min=2,max=255"`
	Description         *string                `json:"description,omitempty" validate:"omitempty,max=1000"`
	TeamName            string                 `json:"team_name" validate:"required"`
	OwnerEmail          string                 `json:"owner_email" validate:"required,email"`
	Lifecycle           string                 `json:"lifecycle" validate:"required,oneof=development staging production deprecated"`
	EnvironmentName     *string                `json:"environment_name,omitempty"`
	EnvironmentRegion   *string                `json:"environment_region,omitempty"`
	ResourceQuota       map[string]interface{} `json:"resource_quota,omitempty"`
	ComplianceSettings  map[string]interface{} `json:"compliance_settings,omitempty"`
	Dependencies        []interface{}          `json:"dependencies,omitempty"`
	ObservabilityConfig map[string]interface{} `json:"observability_config,omitempty"`
}

// UpdateApplicationRequest represents a request to update an application
type UpdateApplicationRequest struct {
	DisplayName         *string                `json:"display_name,omitempty" validate:"omitempty,min=2,max=255"`
	Description         *string                `json:"description,omitempty" validate:"omitempty,max=1000"`
	TeamName            *string                `json:"team_name,omitempty"`
	OwnerEmail          *string                `json:"owner_email,omitempty" validate:"omitempty,email"`
	Lifecycle           *string                `json:"lifecycle,omitempty" validate:"omitempty,oneof=development staging production deprecated"`
	EnvironmentName     *string                `json:"environment_name,omitempty"`
	EnvironmentRegion   *string                `json:"environment_region,omitempty"`
	ResourceQuota       map[string]interface{} `json:"resource_quota,omitempty"`
	ComplianceSettings  map[string]interface{} `json:"compliance_settings,omitempty"`
	Dependencies        []interface{}          `json:"dependencies,omitempty"`
	ObservabilityConfig map[string]interface{} `json:"observability_config,omitempty"`
}

// ListApplicationsRequest represents filtering options for listing applications
type ListApplicationsRequest struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	TeamName  string    `json:"team_name,omitempty"`
	Lifecycle string    `json:"lifecycle,omitempty"`
	Status    string    `json:"status,omitempty"`
	Limit     int       `json:"limit,omitempty"`
	Offset    int       `json:"offset,omitempty"`
}

// CreateApplication creates a new application
func (s *Service) CreateApplication(ctx context.Context, tenantID uuid.UUID, req *CreateApplicationRequest, createdBy string) (*Application, error) {
	app := &Application{
		ID:                  uuid.New(),
		TenantID:            tenantID,
		Name:                req.Name,
		DisplayName:         req.DisplayName,
		Description:         req.Description,
		TeamName:            req.TeamName,
		OwnerEmail:          req.OwnerEmail,
		Lifecycle:           req.Lifecycle,
		EnvironmentName:     req.EnvironmentName,
		EnvironmentRegion:   req.EnvironmentRegion,
		ResourceQuota:       req.ResourceQuota,
		ComplianceSettings:  req.ComplianceSettings,
		Dependencies:        req.Dependencies,
		ObservabilityConfig: req.ObservabilityConfig,
		Status:              "pending",
		Conditions:          []interface{}{},
		CurrentResources:    map[string]interface{}{},
		CreatedBy:           createdBy,
	}

	// Set defaults for nil maps/slices
	if app.ResourceQuota == nil {
		app.ResourceQuota = make(map[string]interface{})
	}
	if app.ComplianceSettings == nil {
		app.ComplianceSettings = make(map[string]interface{})
	}
	if app.Dependencies == nil {
		app.Dependencies = []interface{}{}
	}
	if app.ObservabilityConfig == nil {
		app.ObservabilityConfig = make(map[string]interface{})
	}

	query := `
		INSERT INTO resource_management.applications (
			id, tenant_id, name, display_name, description, team_name, owner_email,
			lifecycle, environment_name, environment_region, resource_quota,
			compliance_settings, dependencies, observability_config, status,
			conditions, current_resources, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		) RETURNING created_at, updated_at
	`

	err := s.db.QueryRow(ctx, query,
		app.ID, app.TenantID, app.Name, app.DisplayName, app.Description,
		app.TeamName, app.OwnerEmail, app.Lifecycle, app.EnvironmentName,
		app.EnvironmentRegion, app.ResourceQuota, app.ComplianceSettings,
		app.Dependencies, app.ObservabilityConfig, app.Status,
		app.Conditions, app.CurrentResources, app.CreatedBy,
	).Scan(&app.CreatedAt, &app.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	return app, nil
}

// GetApplication retrieves an application by ID
func (s *Service) GetApplication(ctx context.Context, tenantID, applicationID uuid.UUID) (*Application, error) {
	query := `
		SELECT id, tenant_id, name, display_name, description, team_name, owner_email,
			   lifecycle, environment_name, environment_region, resource_quota,
			   compliance_settings, dependencies, observability_config, status,
			   conditions, current_resources, created_at, updated_at, created_by, updated_by
		FROM resource_management.applications
		WHERE id = $1 AND tenant_id = $2
	`

	var app Application
	err := s.db.QueryRow(ctx, query, applicationID, tenantID).Scan(
		&app.ID, &app.TenantID, &app.Name, &app.DisplayName, &app.Description,
		&app.TeamName, &app.OwnerEmail, &app.Lifecycle, &app.EnvironmentName,
		&app.EnvironmentRegion, &app.ResourceQuota, &app.ComplianceSettings,
		&app.Dependencies, &app.ObservabilityConfig, &app.Status,
		&app.Conditions, &app.CurrentResources, &app.CreatedAt, &app.UpdatedAt,
		&app.CreatedBy, &app.UpdatedBy,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("application not found: %s", applicationID)
		}
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	return &app, nil
}

// ListApplications retrieves applications with filtering
func (s *Service) ListApplications(ctx context.Context, req *ListApplicationsRequest) ([]*Application, error) {
	qb := database.NewQueryBuilder(`
		SELECT id, tenant_id, name, display_name, description, team_name, owner_email,
			   lifecycle, environment_name, environment_region, resource_quota,
			   compliance_settings, dependencies, observability_config, status,
			   conditions, current_resources, created_at, updated_at, created_by, updated_by
		FROM resource_management.applications
	`)

	// Always filter by tenant
	qb.AddCondition("tenant_id = $"+fmt.Sprint(qb.ArgIndex), req.TenantID)

	// Optional filters
	qb.AddOptionalCondition("team_name = $"+fmt.Sprint(qb.ArgIndex), req.TeamName)
	qb.AddOptionalCondition("lifecycle = $"+fmt.Sprint(qb.ArgIndex), req.Lifecycle)
	qb.AddOptionalCondition("status = $"+fmt.Sprint(qb.ArgIndex), req.Status)

	qb.AddOrderBy("created_at DESC")
	qb.AddLimit(req.Limit)
	qb.AddOffset(req.Offset)

	rows, err := qb.BuildAndQuery(ctx, s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to list applications: %w", err)
	}
	defer rows.Close()

	var applications []*Application
	for rows.Next() {
		var app Application
		err := rows.Scan(
			&app.ID, &app.TenantID, &app.Name, &app.DisplayName, &app.Description,
			&app.TeamName, &app.OwnerEmail, &app.Lifecycle, &app.EnvironmentName,
			&app.EnvironmentRegion, &app.ResourceQuota, &app.ComplianceSettings,
			&app.Dependencies, &app.ObservabilityConfig, &app.Status,
			&app.Conditions, &app.CurrentResources, &app.CreatedAt, &app.UpdatedAt,
			&app.CreatedBy, &app.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan application row: %w", err)
		}
		applications = append(applications, &app)
	}

	return applications, nil
}

// UpdateApplication updates an application's fields
func (s *Service) UpdateApplication(ctx context.Context, tenantID, applicationID uuid.UUID, req *UpdateApplicationRequest, updatedBy string) (*Application, error) {
	// Build dynamic update query
	updates := make(map[string]interface{})

	if req.DisplayName != nil {
		updates["display_name"] = *req.DisplayName
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.TeamName != nil {
		updates["team_name"] = *req.TeamName
	}
	if req.OwnerEmail != nil {
		updates["owner_email"] = *req.OwnerEmail
	}
	if req.Lifecycle != nil {
		updates["lifecycle"] = *req.Lifecycle
	}
	if req.EnvironmentName != nil {
		updates["environment_name"] = *req.EnvironmentName
	}
	if req.EnvironmentRegion != nil {
		updates["environment_region"] = *req.EnvironmentRegion
	}
	if req.ResourceQuota != nil {
		updates["resource_quota"] = req.ResourceQuota
	}
	if req.ComplianceSettings != nil {
		updates["compliance_settings"] = req.ComplianceSettings
	}
	if req.Dependencies != nil {
		updates["dependencies"] = req.Dependencies
	}
	if req.ObservabilityConfig != nil {
		updates["observability_config"] = req.ObservabilityConfig
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	// Add updated_by
	updates["updated_by"] = updatedBy

	return s.updateApplicationFields(ctx, tenantID, applicationID, updates)
}

// updateApplicationFields performs the actual database update
func (s *Service) updateApplicationFields(ctx context.Context, tenantID, applicationID uuid.UUID, updates map[string]interface{}) (*Application, error) {
	// Build SET clause dynamically
	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+2)
	argIndex := 1

	for field, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	// Add WHERE clause arguments
	args = append(args, applicationID, tenantID)

	query := fmt.Sprintf(`
		UPDATE resource_management.applications
		SET %s, updated_at = NOW()
		WHERE id = $%d AND tenant_id = $%d
		RETURNING id, tenant_id, name, display_name, description, team_name, owner_email,
				  lifecycle, environment_name, environment_region, resource_quota,
				  compliance_settings, dependencies, observability_config, status,
				  conditions, current_resources, created_at, updated_at, created_by, updated_by
	`, strings.Join(setParts, ", "), argIndex, argIndex+1)

	var app Application
	err := s.db.QueryRow(ctx, query, args...).Scan(
		&app.ID, &app.TenantID, &app.Name, &app.DisplayName, &app.Description,
		&app.TeamName, &app.OwnerEmail, &app.Lifecycle, &app.EnvironmentName,
		&app.EnvironmentRegion, &app.ResourceQuota, &app.ComplianceSettings,
		&app.Dependencies, &app.ObservabilityConfig, &app.Status,
		&app.Conditions, &app.CurrentResources, &app.CreatedAt, &app.UpdatedAt,
		&app.CreatedBy, &app.UpdatedBy,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("application not found: %s", applicationID)
		}
		return nil, fmt.Errorf("failed to update application: %w", err)
	}

	return &app, nil
}

// DeleteApplication soft-deletes an application (marks as terminated)
func (s *Service) DeleteApplication(ctx context.Context, tenantID, applicationID uuid.UUID, deletedBy string) error {
	// First check if application exists
	_, err := s.GetApplication(ctx, tenantID, applicationID)
	if err != nil {
		return err
	}

	// Update status to terminating, then to terminated
	updates := map[string]interface{}{
		"status":     "terminating",
		"updated_by": deletedBy,
	}

	_, err = s.updateApplicationFields(ctx, tenantID, applicationID, updates)
	if err != nil {
		return fmt.Errorf("failed to mark application as terminating: %w", err)
	}

	// Here you would trigger cleanup of associated resources
	// For now, we'll just mark as terminated
	updates["status"] = "terminated"
	_, err = s.updateApplicationFields(ctx, tenantID, applicationID, updates)
	if err != nil {
		return fmt.Errorf("failed to mark application as terminated: %w", err)
	}

	return nil
}
