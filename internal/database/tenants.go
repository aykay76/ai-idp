package database

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// TenantManager handles tenant database operations
type TenantManager struct {
	pool *Pool
}

// NewTenantManager creates a new tenant manager
func NewTenantManager(pool *Pool) *TenantManager {
	return &TenantManager{pool: pool}
}

// Tenant represents a tenant record
type Tenant struct {
	ID             uuid.UUID              `json:"id" db:"id"`
	Name           string                 `json:"name" db:"name"`
	DisplayName    string                 `json:"display_name" db:"display_name"`
	Description    *string                `json:"description,omitempty" db:"description"`
	DatabaseName   string                 `json:"database_name" db:"database_name"`
	Status         string                 `json:"status" db:"status"`
	Settings       map[string]interface{} `json:"settings" db:"settings"`
	ResourceLimits map[string]interface{} `json:"resource_limits" db:"resource_limits"`
	CreatedAt      string                 `json:"created_at" db:"created_at"`
	UpdatedAt      string                 `json:"updated_at" db:"updated_at"`
}

// CreateTenantRequest represents a request to create a new tenant
type CreateTenantRequest struct {
	Name           string                 `json:"name" validate:"required,min=2,max=63,alphanum"`
	DisplayName    string                 `json:"display_name" validate:"required,min=2,max=255"`
	Description    *string                `json:"description,omitempty" validate:"omitempty,max=1000"`
	Settings       map[string]interface{} `json:"settings,omitempty"`
	ResourceLimits map[string]interface{} `json:"resource_limits,omitempty"`
}

// CreateTenant creates a new tenant and its isolated database
func (tm *TenantManager) CreateTenant(ctx context.Context, req *CreateTenantRequest) (*Tenant, error) {
	// Validate and sanitize the tenant name
	dbName, err := tm.generateDatabaseName(req.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant name: %w", err)
	}

	// Start a transaction for atomic tenant creation
	var tenant *Tenant
	err = tm.pool.WithTransaction(ctx, func(tx *Transaction) error {
		// Insert tenant record
		tenant = &Tenant{
			ID:             uuid.New(),
			Name:           req.Name,
			DisplayName:    req.DisplayName,
			Description:    req.Description,
			DatabaseName:   dbName,
			Status:         "active",
			Settings:       req.Settings,
			ResourceLimits: req.ResourceLimits,
		}

		if tenant.Settings == nil {
			tenant.Settings = make(map[string]interface{})
		}
		if tenant.ResourceLimits == nil {
			tenant.ResourceLimits = make(map[string]interface{})
		}

		query := `
			INSERT INTO control_plane.tenants (
				id, name, display_name, description, database_name, status, settings, resource_limits
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING created_at, updated_at
		`

		err := tx.QueryRow(ctx, query,
			tenant.ID, tenant.Name, tenant.DisplayName, tenant.Description,
			tenant.DatabaseName, tenant.Status, tenant.Settings, tenant.ResourceLimits,
		).Scan(&tenant.CreatedAt, &tenant.UpdatedAt)

		if err != nil {
			return fmt.Errorf("failed to insert tenant record: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Create the tenant database (outside transaction to avoid nested transactions)
	if err := tm.createTenantDatabase(ctx, dbName); err != nil {
		// Cleanup: delete the tenant record if database creation fails
		_ = tm.deleteTenantRecord(ctx, tenant.ID)
		return nil, fmt.Errorf("failed to create tenant database: %w", err)
	}

	// Initialize tenant database schema
	if err := tm.initializeTenantDatabase(ctx, dbName); err != nil {
		// Cleanup: delete both database and record
		_ = tm.dropTenantDatabase(ctx, dbName)
		_ = tm.deleteTenantRecord(ctx, tenant.ID)
		return nil, fmt.Errorf("failed to initialize tenant database: %w", err)
	}

	return tenant, nil
}

// GetTenant retrieves a tenant by ID
func (tm *TenantManager) GetTenant(ctx context.Context, tenantID uuid.UUID) (*Tenant, error) {
	query := `
		SELECT id, name, display_name, description, database_name, status, 
			   settings, resource_limits, created_at, updated_at
		FROM control_plane.tenants
		WHERE id = $1
	`

	var tenant Tenant
	err := tm.pool.QueryRow(ctx, query, tenantID).Scan(
		&tenant.ID, &tenant.Name, &tenant.DisplayName, &tenant.Description,
		&tenant.DatabaseName, &tenant.Status, &tenant.Settings, &tenant.ResourceLimits,
		&tenant.CreatedAt, &tenant.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tenant not found: %s", tenantID)
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return &tenant, nil
}

// GetTenantByName retrieves a tenant by name
func (tm *TenantManager) GetTenantByName(ctx context.Context, name string) (*Tenant, error) {
	query := `
		SELECT id, name, display_name, description, database_name, status,
			   settings, resource_limits, created_at, updated_at
		FROM control_plane.tenants
		WHERE name = $1
	`

	var tenant Tenant
	err := tm.pool.QueryRow(ctx, query, name).Scan(
		&tenant.ID, &tenant.Name, &tenant.DisplayName, &tenant.Description,
		&tenant.DatabaseName, &tenant.Status, &tenant.Settings, &tenant.ResourceLimits,
		&tenant.CreatedAt, &tenant.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tenant not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get tenant by name: %w", err)
	}

	return &tenant, nil
}

// ListTenants retrieves all tenants with optional filtering
func (tm *TenantManager) ListTenants(ctx context.Context, status string, limit, offset int) ([]*Tenant, error) {
	qb := NewQueryBuilder(`
		SELECT id, name, display_name, description, database_name, status,
			   settings, resource_limits, created_at, updated_at
		FROM control_plane.tenants
	`)

	qb.AddOptionalCondition("status = $"+fmt.Sprint(qb.ArgIndex), status)
	qb.AddOrderBy("created_at DESC")
	qb.AddLimit(limit)
	qb.AddOffset(offset)

	rows, err := qb.BuildAndQuery(ctx, tm.pool)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*Tenant
	for rows.Next() {
		var tenant Tenant
		err := rows.Scan(
			&tenant.ID, &tenant.Name, &tenant.DisplayName, &tenant.Description,
			&tenant.DatabaseName, &tenant.Status, &tenant.Settings, &tenant.ResourceLimits,
			&tenant.CreatedAt, &tenant.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant row: %w", err)
		}
		tenants = append(tenants, &tenant)
	}

	return tenants, nil
}

// UpdateTenant updates a tenant's metadata
func (tm *TenantManager) UpdateTenant(ctx context.Context, tenantID uuid.UUID, updates map[string]interface{}) (*Tenant, error) {
	// Build dynamic update query
	setParts := make([]string, 0)
	args := make([]interface{}, 0)
	argIndex := 1

	for field, value := range updates {
		switch field {
		case "display_name", "description", "status", "settings", "resource_limits":
			setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
			args = append(args, value)
			argIndex++
		default:
			return nil, fmt.Errorf("field '%s' is not updatable", field)
		}
	}

	if len(setParts) == 0 {
		return nil, fmt.Errorf("no valid fields to update")
	}

	// Add tenant ID as last argument
	args = append(args, tenantID)

	query := fmt.Sprintf(`
		UPDATE control_plane.tenants
		SET %s, updated_at = NOW()
		WHERE id = $%d
		RETURNING id, name, display_name, description, database_name, status,
				  settings, resource_limits, created_at, updated_at
	`, strings.Join(setParts, ", "), argIndex)

	var tenant Tenant
	err := tm.pool.QueryRow(ctx, query, args...).Scan(
		&tenant.ID, &tenant.Name, &tenant.DisplayName, &tenant.Description,
		&tenant.DatabaseName, &tenant.Status, &tenant.Settings, &tenant.ResourceLimits,
		&tenant.CreatedAt, &tenant.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tenant not found: %s", tenantID)
		}
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return &tenant, nil
}

// DeleteTenant soft-deletes a tenant (marks as terminated)
func (tm *TenantManager) DeleteTenant(ctx context.Context, tenantID uuid.UUID) error {
	// First, mark tenant as terminating
	_, err := tm.UpdateTenant(ctx, tenantID, map[string]interface{}{
		"status": "terminating",
	})
	if err != nil {
		return fmt.Errorf("failed to mark tenant as terminating: %w", err)
	}

	// Get tenant details for cleanup
	tenant, err := tm.GetTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant for deletion: %w", err)
	}

	// Drop tenant database
	if err := tm.dropTenantDatabase(ctx, tenant.DatabaseName); err != nil {
		// Continue with soft delete even if database drop fails
		fmt.Printf("Warning: failed to drop tenant database %s: %v\n", tenant.DatabaseName, err)
	}

	// Mark tenant as terminated
	_, err = tm.UpdateTenant(ctx, tenantID, map[string]interface{}{
		"status": "terminated",
	})
	if err != nil {
		return fmt.Errorf("failed to mark tenant as terminated: %w", err)
	}

	return nil
}

// generateDatabaseName creates a valid PostgreSQL database name from tenant name
func (tm *TenantManager) generateDatabaseName(tenantName string) (string, error) {
	// PostgreSQL database naming rules:
	// - Must start with letter or underscore
	// - Can contain letters, numbers, underscores
	// - Max 63 characters
	// - Case insensitive

	// Convert to lowercase and replace invalid characters
	dbName := strings.ToLower(tenantName)
	dbName = regexp.MustCompile(`[^a-z0-9_]`).ReplaceAllString(dbName, "_")

	// Ensure it starts with a letter
	if !regexp.MustCompile(`^[a-z]`).MatchString(dbName) {
		dbName = "tenant_" + dbName
	}

	// Truncate if too long (leaving room for potential suffixes)
	if len(dbName) > 50 {
		dbName = dbName[:50]
	}

	// Add tenant prefix for clarity
	if !strings.HasPrefix(dbName, "tenant_") {
		dbName = "tenant_" + dbName
	}

	// Validate final name
	if !regexp.MustCompile(`^[a-z][a-z0-9_]*$`).MatchString(dbName) {
		return "", fmt.Errorf("cannot generate valid database name from '%s'", tenantName)
	}

	return dbName, nil
}

// createTenantDatabase creates a new database for the tenant
func (tm *TenantManager) createTenantDatabase(ctx context.Context, dbName string) error {
	// Note: Database names cannot be parameterized in PostgreSQL
	// We've already validated the name format, so this should be safe
	query := fmt.Sprintf("CREATE DATABASE %s", pgx.Identifier{dbName}.Sanitize())

	_, err := tm.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create database %s: %w", dbName, err)
	}

	return nil
}

// initializeTenantDatabase sets up the initial schema for a tenant database
func (tm *TenantManager) initializeTenantDatabase(ctx context.Context, dbName string) error {
	// Connect to the tenant database
	tenantConfig := *tm.pool.config
	tenantConfig.URL = strings.Replace(tenantConfig.URL, "/platform", "/"+dbName, 1)

	tenantPool, err := NewPool(ctx, &tenantConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to tenant database: %w", err)
	}
	defer tenantPool.Close()

	// Create tenant-specific tables
	initQueries := []string{
		`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`,

		// Tenant metadata table
		`CREATE TABLE IF NOT EXISTS tenant_metadata (
			key VARCHAR(255) PRIMARY KEY,
			value JSONB NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Insert tenant info
		fmt.Sprintf(`INSERT INTO tenant_metadata (key, value) VALUES 
			('tenant_id', '"%s"'::jsonb),
			('schema_version', '1'::jsonb),
			('created_at', 'NOW()'::jsonb)
		`, dbName),
	}

	for _, query := range initQueries {
		if _, err := tenantPool.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to initialize tenant database schema: %w", err)
		}
	}

	return nil
}

// dropTenantDatabase drops the tenant database
func (tm *TenantManager) dropTenantDatabase(ctx context.Context, dbName string) error {
	// Terminate existing connections to the database
	terminateQuery := `
		SELECT pg_terminate_backend(pg_stat_activity.pid)
		FROM pg_stat_activity
		WHERE pg_stat_activity.datname = $1
		  AND pid <> pg_backend_pid()
	`
	_, _ = tm.pool.Exec(ctx, terminateQuery, dbName)

	// Drop the database
	query := fmt.Sprintf("DROP DATABASE IF EXISTS %s", pgx.Identifier{dbName}.Sanitize())
	_, err := tm.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to drop database %s: %w", dbName, err)
	}

	return nil
}

// deleteTenantRecord removes the tenant record (used for cleanup on errors)
func (tm *TenantManager) deleteTenantRecord(ctx context.Context, tenantID uuid.UUID) error {
	query := "DELETE FROM control_plane.tenants WHERE id = $1"
	_, err := tm.pool.Exec(ctx, query, tenantID)
	return err
}

// GetTenantPool returns a connection pool for a specific tenant's database
func (tm *TenantManager) GetTenantPool(ctx context.Context, tenantID uuid.UUID) (*Pool, error) {
	tenant, err := tm.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	if tenant.Status != "active" {
		return nil, fmt.Errorf("tenant is not active: %s", tenant.Status)
	}

	// Create connection config for tenant database
	tenantConfig := *tm.pool.config
	tenantConfig.URL = strings.Replace(tenantConfig.URL, "/platform", "/"+tenant.DatabaseName, 1)

	return NewPool(ctx, &tenantConfig)
}
