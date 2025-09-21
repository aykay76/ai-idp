package database

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq" // Required for golang-migrate postgres driver
)

// MigrationManager handles database migrations
type MigrationManager struct {
	pool         *Pool
	migrateUp    *migrate.Migrate
	migrateDown  *migrate.Migrate
	migrationDir string
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(pool *Pool, migrationDir string) (*MigrationManager, error) {
	// Create a standard database/sql connection for golang-migrate
	// golang-migrate doesn't support pgx directly, so we need database/sql
	db, err := sql.Open("postgres", pool.config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database for migrations: %w", err)
	}

	// Create postgres driver instance
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Get absolute path to migrations
	absPath, err := filepath.Abs(migrationDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for migrations: %w", err)
	}

	sourceURL := fmt.Sprintf("file://%s", absPath)

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return &MigrationManager{
		pool:         pool,
		migrateUp:    m,
		migrationDir: migrationDir,
	}, nil
}

// Up runs all available migrations
func (mm *MigrationManager) Up(ctx context.Context) error {
	if err := mm.migrateUp.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations up: %w", err)
	}
	return nil
}

// Down runs one migration down
func (mm *MigrationManager) Down(ctx context.Context) error {
	if err := mm.migrateUp.Steps(-1); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migration down: %w", err)
	}
	return nil
}

// Steps runs a specific number of migrations (positive = up, negative = down)
func (mm *MigrationManager) Steps(ctx context.Context, steps int) error {
	if err := mm.migrateUp.Steps(steps); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run %d migration steps: %w", steps, err)
	}
	return nil
}

// Version returns the current migration version
func (mm *MigrationManager) Version() (uint, bool, error) {
	version, dirty, err := mm.migrateUp.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}
	return version, dirty, nil
}

// Drop drops the entire database (use with caution!)
func (mm *MigrationManager) Drop(ctx context.Context) error {
	if err := mm.migrateUp.Drop(); err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}
	return nil
}

// Reset drops and recreates the database with all migrations
func (mm *MigrationManager) Reset(ctx context.Context) error {
	// Drop all tables and data
	if err := mm.Drop(ctx); err != nil {
		return fmt.Errorf("failed to drop database during reset: %w", err)
	}

	// Run all migrations
	if err := mm.Up(ctx); err != nil {
		return fmt.Errorf("failed to run migrations during reset: %w", err)
	}

	return nil
}

// Status returns detailed migration status
func (mm *MigrationManager) Status() (*MigrationStatus, error) {
	version, dirty, err := mm.Version()
	if err != nil {
		return nil, err
	}

	status := &MigrationStatus{
		Version: version,
		Dirty:   dirty,
	}

	if version == 0 {
		status.Status = "No migrations applied"
	} else if dirty {
		status.Status = fmt.Sprintf("Version %d (dirty)", version)
	} else {
		status.Status = fmt.Sprintf("Version %d (clean)", version)
	}

	return status, nil
}

// MigrationStatus represents the current migration status
type MigrationStatus struct {
	Version uint   `json:"version"`
	Dirty   bool   `json:"dirty"`
	Status  string `json:"status"`
}

// Close closes the migration manager
func (mm *MigrationManager) Close() error {
	sourceErr, dbErr := mm.migrateUp.Close()
	if sourceErr != nil {
		return fmt.Errorf("failed to close migration source: %w", sourceErr)
	}
	if dbErr != nil {
		return fmt.Errorf("failed to close migration database: %w", dbErr)
	}
	return nil
}

// WaitForMigrations waits for database migrations to be applied
// This is useful in containerized environments where services might start
// before migrations are complete
func WaitForMigrations(ctx context.Context, pool *Pool, expectedVersion uint) error {
	mm, err := NewMigrationManager(pool, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration manager: %w", err)
	}
	defer mm.Close()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for migrations: %w", ctx.Err())
		default:
			version, dirty, err := mm.Version()
			if err != nil && err != migrate.ErrNilVersion {
				return fmt.Errorf("failed to check migration version: %w", err)
			}

			if !dirty && version >= expectedVersion {
				return nil
			}

			// Wait a bit before checking again
			select {
			case <-ctx.Done():
				return fmt.Errorf("timeout waiting for migrations: %w", ctx.Err())
			case <-time.After(1 * time.Second):
			}
		}
	}
}
