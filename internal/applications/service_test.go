package applications

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aykay76/ai-idp/internal/database"
	"github.com/aykay76/ai-idp/internal/testutils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationService_CreateApplication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database
	ctx := context.Background()
	pool, cleanup := testutils.SetupTestDB(t, ctx)
	defer cleanup()

	// Create service
	service := NewService(pool)

	// Create test tenant
	tenantManager := database.NewTenantManager(pool)
	tenant, err := tenantManager.CreateTenant(ctx, &database.CreateTenantRequest{
		Name:        "test-tenant",
		DisplayName: "Test Tenant",
		Description: stringPtr("Test tenant for application tests"),
	})
	require.NoError(t, err)

	// Test data
	req := &CreateApplicationRequest{
		Name:        "test-app",
		DisplayName: "Test Application",
		Description: stringPtr("Test application for unit tests"),
		TeamName:    "platform-team",
		OwnerEmail:  "test@example.com",
		Lifecycle:   "development",
	}

	// Test successful creation
	app, err := service.CreateApplication(ctx, tenant.ID, req, "test-user")
	require.NoError(t, err)
	assert.NotNil(t, app)
	assert.NotEqual(t, uuid.Nil, app.ID)
	assert.Equal(t, tenant.ID, app.TenantID)
	assert.Equal(t, req.Name, app.Name)
	assert.Equal(t, req.DisplayName, app.DisplayName)
	assert.Equal(t, req.TeamName, app.TeamName)
	assert.Equal(t, req.OwnerEmail, app.OwnerEmail)
	assert.Equal(t, req.Lifecycle, app.Lifecycle)
	assert.Equal(t, "pending", app.Status)
	assert.Equal(t, "test-user", app.CreatedBy)
	assert.WithinDuration(t, time.Now(), app.CreatedAt, 5*time.Second)
	assert.NotNil(t, app.ResourceQuota)
	assert.NotNil(t, app.ComplianceSettings)
	assert.NotNil(t, app.Dependencies)
	assert.NotNil(t, app.ObservabilityConfig)
}

func TestApplicationService_GetApplication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database
	ctx := context.Background()
	pool, cleanup := testutils.SetupTestDB(t, ctx)
	defer cleanup()

	service := NewService(pool)

	// Create test tenant and application
	tenant := createTestTenant(t, ctx, pool)
	app := createTestApplication(t, ctx, service, tenant.ID)

	// Test successful retrieval
	retrievedApp, err := service.GetApplication(ctx, tenant.ID, app.ID)
	require.NoError(t, err)
	assert.Equal(t, app.ID, retrievedApp.ID)
	assert.Equal(t, app.Name, retrievedApp.Name)
	assert.Equal(t, app.DisplayName, retrievedApp.DisplayName)

	// Test not found
	nonExistentID := uuid.New()
	_, err = service.GetApplication(ctx, tenant.ID, nonExistentID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test wrong tenant
	otherTenant := createTestTenant(t, ctx, pool)
	_, err = service.GetApplication(ctx, otherTenant.ID, app.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestApplicationService_ListApplications(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database
	ctx := context.Background()
	pool, cleanup := testutils.SetupTestDB(t, ctx)
	defer cleanup()

	service := NewService(pool)
	tenant := createTestTenant(t, ctx, pool)

	// Create multiple test applications
	apps := make([]*Application, 3)
	for i := 0; i < 3; i++ {
		apps[i] = createTestApplicationWithName(t, ctx, service, tenant.ID, fmt.Sprintf("test-app-%d", i+1))
	}

	// Test listing all applications
	req := &ListApplicationsRequest{
		TenantID: tenant.ID,
		Limit:    10,
		Offset:   0,
	}

	result, err := service.ListApplications(ctx, req)
	require.NoError(t, err)
	assert.Len(t, result, 3)

	// Test filtering by team name (all our test apps have the same team)
	req.TeamName = "platform-team"
	result, err = service.ListApplications(ctx, req)
	require.NoError(t, err)
	assert.Len(t, result, 3)

	// Test filtering by non-existent team
	req.TeamName = "non-existent-team"
	result, err = service.ListApplications(ctx, req)
	require.NoError(t, err)
	assert.Len(t, result, 0)

	// Test pagination
	req.TeamName = ""
	req.Limit = 2
	req.Offset = 0
	result, err = service.ListApplications(ctx, req)
	require.NoError(t, err)
	assert.Len(t, result, 2)

	// Test offset
	req.Offset = 2
	result, err = service.ListApplications(ctx, req)
	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestApplicationService_UpdateApplication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database
	ctx := context.Background()
	pool, cleanup := testutils.SetupTestDB(t, ctx)
	defer cleanup()

	service := NewService(pool)
	tenant := createTestTenant(t, ctx, pool)
	app := createTestApplication(t, ctx, service, tenant.ID)

	// Test update
	newDisplayName := "Updated Test Application"
	newDescription := "Updated description"
	newLifecycle := "staging"

	req := &UpdateApplicationRequest{
		DisplayName: &newDisplayName,
		Description: &newDescription,
		Lifecycle:   &newLifecycle,
	}

	updatedApp, err := service.UpdateApplication(ctx, tenant.ID, app.ID, req, "test-updater")
	require.NoError(t, err)
	assert.Equal(t, newDisplayName, updatedApp.DisplayName)
	assert.Equal(t, newDescription, *updatedApp.Description)
	assert.Equal(t, newLifecycle, updatedApp.Lifecycle)
	assert.Equal(t, "test-updater", *updatedApp.UpdatedBy)
	assert.True(t, updatedApp.UpdatedAt.After(updatedApp.CreatedAt))

	// Test update with no fields
	emptyReq := &UpdateApplicationRequest{}
	_, err = service.UpdateApplication(ctx, tenant.ID, app.ID, emptyReq, "test-updater")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no fields to update")

	// Test update non-existent application
	nonExistentID := uuid.New()
	_, err = service.UpdateApplication(ctx, tenant.ID, nonExistentID, req, "test-updater")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestApplicationService_DeleteApplication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database
	ctx := context.Background()
	pool, cleanup := testutils.SetupTestDB(t, ctx)
	defer cleanup()

	service := NewService(pool)
	tenant := createTestTenant(t, ctx, pool)
	app := createTestApplication(t, ctx, service, tenant.ID)

	// Test successful deletion
	err := service.DeleteApplication(ctx, tenant.ID, app.ID, "test-deleter")
	require.NoError(t, err)

	// Verify application is marked as terminated
	deletedApp, err := service.GetApplication(ctx, tenant.ID, app.ID)
	require.NoError(t, err)
	assert.Equal(t, "terminated", deletedApp.Status)
	assert.Equal(t, "test-deleter", *deletedApp.UpdatedBy)

	// Test delete non-existent application
	nonExistentID := uuid.New()
	err = service.DeleteApplication(ctx, tenant.ID, nonExistentID, "test-deleter")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// Test helper functions

func createTestTenant(t *testing.T, ctx context.Context, pool *database.Pool) *database.Tenant {
	tenantManager := database.NewTenantManager(pool)
	tenant, err := tenantManager.CreateTenant(ctx, &database.CreateTenantRequest{
		Name:        fmt.Sprintf("test-tenant-%d", time.Now().UnixNano()),
		DisplayName: "Test Tenant",
		Description: stringPtr("Test tenant for application tests"),
	})
	require.NoError(t, err)
	return tenant
}

func createTestApplication(t *testing.T, ctx context.Context, service *Service, tenantID uuid.UUID) *Application {
	return createTestApplicationWithName(t, ctx, service, tenantID, "test-app")
}

func createTestApplicationWithName(t *testing.T, ctx context.Context, service *Service, tenantID uuid.UUID, name string) *Application {
	req := &CreateApplicationRequest{
		Name:        name,
		DisplayName: "Test Application",
		Description: stringPtr("Test application for unit tests"),
		TeamName:    "platform-team",
		OwnerEmail:  "test@example.com",
		Lifecycle:   "development",
	}

	app, err := service.CreateApplication(ctx, tenantID, req, "test-user")
	require.NoError(t, err)
	return app
}

func stringPtr(s string) *string {
	return &s
}
