package testutils

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aykay76/ai-idp/internal/database"
)

// SetupTestDB creates a test database connection and ensures migrations are run
func SetupTestDB(t *testing.T, ctx context.Context) (*database.Pool, func()) {
	t.Helper()

	// Get test database URL from environment
	dbURL := getTestDatabaseURL()

	// Create database connection
	config := database.DefaultConfig(dbURL)
	config.MaxConnections = 5 // Smaller pool for tests
	config.MinConnections = 1
	config.ConnectTimeout = 10 * time.Second

	pool, err := database.NewPool(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create test database pool: %v", err)
	}

	// Verify connection
	if err := pool.HealthCheck(ctx); err != nil {
		pool.Close()
		t.Fatalf("Test database health check failed: %v", err)
	}

	// Run migrations
	migrationManager, err := database.NewMigrationManager(pool, "../../migrations")
	if err != nil {
		pool.Close()
		t.Fatalf("Failed to create migration manager: %v", err)
	}

	if err := migrationManager.Up(ctx); err != nil {
		migrationManager.Close()
		pool.Close()
		t.Fatalf("Failed to run test migrations: %v", err)
	}

	// Cleanup function
	cleanup := func() {
		migrationManager.Close()
		pool.Close()
	}

	return pool, cleanup
}

// SetupIsolatedTestDB creates a completely isolated test database for tests that need it
func SetupIsolatedTestDB(t *testing.T, ctx context.Context) (*database.Pool, func()) {
	t.Helper()

	// Create unique database name for this test
	testDBName := fmt.Sprintf("test_%s_%d", sanitizeTestName(t.Name()), time.Now().UnixNano())

	// Connect to postgres database to create test database
	adminURL := getAdminDatabaseURL()
	adminConfig := database.DefaultConfig(adminURL)
	adminPool, err := database.NewPool(ctx, adminConfig)
	if err != nil {
		t.Fatalf("Failed to create admin database connection: %v", err)
	}

	// Create test database
	createQuery := fmt.Sprintf("CREATE DATABASE %s", testDBName)
	if _, err := adminPool.Exec(ctx, createQuery); err != nil {
		adminPool.Close()
		t.Fatalf("Failed to create test database %s: %v", testDBName, err)
	}

	adminPool.Close()

	// Connect to the new test database
	testDBURL := getTestDatabaseURLForDB(testDBName)
	testConfig := database.DefaultConfig(testDBURL)
	testConfig.MaxConnections = 5
	testConfig.MinConnections = 1

	testPool, err := database.NewPool(ctx, testConfig)
	if err != nil {
		t.Fatalf("Failed to connect to test database %s: %v", testDBName, err)
	}

	// Run migrations on test database
	migrationManager, err := database.NewMigrationManager(testPool, "../../migrations")
	if err != nil {
		testPool.Close()
		t.Fatalf("Failed to create migration manager for test database: %v", err)
	}

	if err := migrationManager.Up(ctx); err != nil {
		migrationManager.Close()
		testPool.Close()
		t.Fatalf("Failed to run migrations on test database: %v", err)
	}

	// Cleanup function
	cleanup := func() {
		migrationManager.Close()
		testPool.Close()

		// Drop the test database
		adminPool, err := database.NewPool(ctx, adminConfig)
		if err == nil {
			// Terminate connections to test database
			terminateQuery := `
				SELECT pg_terminate_backend(pg_stat_activity.pid)
				FROM pg_stat_activity
				WHERE pg_stat_activity.datname = $1
				  AND pid <> pg_backend_pid()
			`
			adminPool.Exec(ctx, terminateQuery, testDBName)

			// Drop database
			dropQuery := fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName)
			adminPool.Exec(ctx, dropQuery)
			adminPool.Close()
		}
	}

	return testPool, cleanup
}

// SetupTestTenant creates a test tenant in the provided database
func SetupTestTenant(t *testing.T, ctx context.Context, pool *database.Pool) *database.Tenant {
	t.Helper()

	tenantManager := database.NewTenantManager(pool)

	// Create unique tenant name
	tenantName := fmt.Sprintf("test-tenant-%s-%d", sanitizeTestName(t.Name()), time.Now().UnixNano())

	req := &database.CreateTenantRequest{
		Name:        tenantName,
		DisplayName: fmt.Sprintf("Test Tenant for %s", t.Name()),
		Description: stringPtr("Auto-created test tenant"),
		Settings:    map[string]interface{}{"test": true},
	}

	tenant, err := tenantManager.CreateTenant(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	return tenant
}

// Helper functions

func getTestDatabaseURL() string {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		// Default to development database for tests
		dbURL = "postgres://platform:platform_dev_password@localhost:5432/platform?sslmode=disable"
	}
	return dbURL
}

func getAdminDatabaseURL() string {
	// Connect to postgres database for administrative operations
	adminURL := os.Getenv("TEST_ADMIN_DATABASE_URL")
	if adminURL == "" {
		adminURL = "postgres://platform:platform_dev_password@localhost:5432/postgres?sslmode=disable"
	}
	return adminURL
}

func getTestDatabaseURLForDB(dbName string) string {
	// Use hardcoded connection string for test databases
	return fmt.Sprintf("postgres://platform:platform_dev_password@localhost:5432/%s?sslmode=disable", dbName)
}

func sanitizeTestName(name string) string {
	// Replace invalid characters for database names
	result := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			result += string(r)
		} else if r >= 'A' && r <= 'Z' {
			result += string(r + 32) // Convert to lowercase
		} else {
			result += "_"
		}
	}

	// Ensure it starts with a letter
	if len(result) > 0 && result[0] >= '0' && result[0] <= '9' {
		result = "test_" + result
	}

	// Truncate if too long
	if len(result) > 30 {
		result = result[:30]
	}

	return result
}

func stringPtr(s string) *string {
	return &s
}

// SkipIfShort skips the test if running in short mode (for integration tests)
func SkipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
}
