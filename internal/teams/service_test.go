package teams

import (
	"context"
	"testing"
	"time"

	"github.com/aykay76/ai-idp/internal/database"
	"github.com/aykay76/ai-idp/internal/testutils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamService_CreateTeam(t *testing.T) {
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
		Description: stringPtr("Test tenant for team tests"),
	})
	require.NoError(t, err)

	t.Run("successful creation", func(t *testing.T) {
		team := Team{
			TenantID:     tenant.ID,
			Name:         "platform-team",
			DisplayName:  "Platform Team",
			Description:  stringPtr("Core platform development team"),
			LeadEmail:    "platform-lead@company.com",
			Department:   stringPtr("Engineering"),
			Organization: stringPtr("Platform Division"),
			CreatedBy:    "test-user",
		}

		result, err := service.CreateTeam(ctx, team)
		require.NoError(t, err)

		assert.NotEqual(t, uuid.Nil, result.ID)
		assert.Equal(t, tenant.ID, result.TenantID)
		assert.Equal(t, "platform-team", result.Name)
		assert.Equal(t, "Platform Team", result.DisplayName)
		assert.Equal(t, "Core platform development team", *result.Description)
		assert.Equal(t, "platform-lead@company.com", result.LeadEmail)
		assert.Equal(t, "Engineering", *result.Department)
		assert.Equal(t, "Platform Division", *result.Organization)
		assert.Equal(t, "test-user", result.CreatedBy)
		assert.WithinDuration(t, time.Now(), result.CreatedAt, time.Minute)
		assert.WithinDuration(t, time.Now(), result.UpdatedAt, time.Minute)
		assert.Equal(t, 0, result.MemberCount)
		assert.Equal(t, 0, result.ActiveApplications)
		assert.Empty(t, result.Members)
		assert.NotNil(t, result.Contacts)
		assert.NotNil(t, result.OwnedApplications)
		assert.NotNil(t, result.OwnedDomains)
		assert.NotNil(t, result.OwnedRepositories)
		assert.NotNil(t, result.Policies)
		assert.NotNil(t, result.BudgetConfig)
	})

	t.Run("missing required fields", func(t *testing.T) {
		// Test missing name
		team := Team{
			TenantID:    tenant.ID,
			DisplayName: "Test Team",
			LeadEmail:   "lead@company.com",
		}

		_, err := service.CreateTeam(ctx, team)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")

		// Test missing lead email
		team = Team{
			TenantID:    tenant.ID,
			Name:        "test-team",
			DisplayName: "Test Team",
		}

		_, err = service.CreateTeam(ctx, team)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "lead_email is required")
	})

	t.Run("auto-generated fields", func(t *testing.T) {
		team := Team{
			Name:      "auto-test-team",
			LeadEmail: "auto-lead@company.com",
		}

		result, err := service.CreateTeam(ctx, team)
		require.NoError(t, err)

		// Should auto-generate ID
		assert.NotEqual(t, uuid.Nil, result.ID)
		
		// Should set default tenant ID
		assert.NotEqual(t, uuid.Nil, result.TenantID)
		
		// Should set display name to name if not provided
		assert.Equal(t, "auto-test-team", result.DisplayName)
		
		// Should set default created by
		assert.Equal(t, "system", result.CreatedBy)
	})

	t.Run("with members", func(t *testing.T) {
		members := []Member{
			{
				UserID:   "user1",
				Email:    "user1@company.com",
				Role:     "owner",
				JoinedAt: time.Now().UTC(),
				Status:   "active",
			},
			{
				UserID:   "user2",
				Email:    "user2@company.com",
				Role:     "developer",
				JoinedAt: time.Now().UTC(),
				Status:   "active",
			},
		}

		team := Team{
			TenantID:    tenant.ID,
			Name:        "team-with-members",
			DisplayName: "Team With Members",
			LeadEmail:   "lead@company.com",
			Members:     members,
			CreatedBy:   "test-user",
		}

		result, err := service.CreateTeam(ctx, team)
		require.NoError(t, err)

		assert.Len(t, result.Members, 2)
		assert.Equal(t, "user1", result.Members[0].UserID)
		assert.Equal(t, "user1@company.com", result.Members[0].Email)
		assert.Equal(t, "owner", result.Members[0].Role)
		assert.Equal(t, "active", result.Members[0].Status)
	})
}

func TestTeamService_GetTeam(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	pool, cleanup := testutils.SetupTestDB(t, ctx)
	defer cleanup()

	service := NewService(pool)

	// Create test tenant
	tenantManager := database.NewTenantManager(pool)
	tenant, err := tenantManager.CreateTenant(ctx, &database.CreateTenantRequest{
		Name:        "test-tenant",
		DisplayName: "Test Tenant",
		Description: stringPtr("Test tenant for team tests"),
	})
	require.NoError(t, err)

	t.Run("existing team", func(t *testing.T) {
		// Create a team first
		team := Team{
			TenantID:    tenant.ID,
			Name:        "get-test-team",
			DisplayName: "Get Test Team",
			LeadEmail:   "get-lead@company.com",
			CreatedBy:   "test-user",
		}

		created, err := service.CreateTeam(ctx, team)
		require.NoError(t, err)

		// Get the team
		result, err := service.GetTeam(ctx, created.ID)
		require.NoError(t, err)

		assert.Equal(t, created.ID, result.ID)
		assert.Equal(t, created.Name, result.Name)
		assert.Equal(t, created.DisplayName, result.DisplayName)
		assert.Equal(t, created.LeadEmail, result.LeadEmail)
		assert.Equal(t, created.CreatedBy, result.CreatedBy)
	})

	t.Run("non-existent team", func(t *testing.T) {
		nonExistentID := uuid.New()
		_, err := service.GetTeam(ctx, nonExistentID)
		require.Error(t, err)
		assert.Equal(t, ErrTeamNotFound, err)
	})
}

func TestTeamService_ListTeams(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	pool, cleanup := testutils.SetupTestDB(t, ctx)
	defer cleanup()

	service := NewService(pool)

	// Create test tenant
	tenantManager := database.NewTenantManager(pool)
	tenant, err := tenantManager.CreateTenant(ctx, &database.CreateTenantRequest{
		Name:        "test-tenant",
		DisplayName: "Test Tenant",
		Description: stringPtr("Test tenant for team tests"),
	})
	require.NoError(t, err)

	t.Run("empty list", func(t *testing.T) {
		teams, total, err := service.ListTeams(ctx, 10, 0)
		require.NoError(t, err)
		assert.Empty(t, teams)
		assert.Equal(t, 0, total)
	})

	t.Run("with teams", func(t *testing.T) {
		// Create multiple teams
		teamNames := []string{"list-team-1", "list-team-2", "list-team-3"}
		createdTeams := make([]Team, len(teamNames))

		for i, name := range teamNames {
			team := Team{
				TenantID:    tenant.ID,
				Name:        name,
				DisplayName: name + " Display",
				LeadEmail:   name + "@company.com",
				CreatedBy:   "test-user",
			}

			created, err := service.CreateTeam(ctx, team)
			require.NoError(t, err)
			createdTeams[i] = created
		}

		// List all teams
		teams, total, err := service.ListTeams(ctx, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(teams), len(teamNames))
		assert.GreaterOrEqual(t, total, len(teamNames))

		// Find our created teams
		for _, expectedTeam := range createdTeams {
			found := false
			for _, actualTeam := range teams {
				if actualTeam.ID == expectedTeam.ID {
					found = true
					assert.Equal(t, expectedTeam.Name, actualTeam.Name)
					assert.Equal(t, expectedTeam.DisplayName, actualTeam.DisplayName)
					assert.Equal(t, expectedTeam.LeadEmail, actualTeam.LeadEmail)
					break
				}
			}
			assert.True(t, found, "Team %s not found in list", expectedTeam.Name)
		}
	})

	t.Run("pagination", func(t *testing.T) {
		// List with limit
		teams, total, err := service.ListTeams(ctx, 2, 0)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(teams), 2)
		assert.GreaterOrEqual(t, total, len(teams))

		// List with offset
		if total > 2 {
			teams2, total2, err := service.ListTeams(ctx, 2, 2)
			require.NoError(t, err)
			assert.Equal(t, total, total2) // Total should be the same
			
			// Teams should be different (assuming we have more than 2)
			if len(teams) > 0 && len(teams2) > 0 {
				assert.NotEqual(t, teams[0].ID, teams2[0].ID)
			}
		}
	})
}

func TestTeamService_UpdateTeam(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	pool, cleanup := testutils.SetupTestDB(t, ctx)
	defer cleanup()

	service := NewService(pool)

	// Create test tenant
	tenantManager := database.NewTenantManager(pool)
	tenant, err := tenantManager.CreateTenant(ctx, &database.CreateTenantRequest{
		Name:        "test-tenant",
		DisplayName: "Test Tenant",
		Description: stringPtr("Test tenant for team tests"),
	})
	require.NoError(t, err)

	t.Run("successful update", func(t *testing.T) {
		// Create a team first
		team := Team{
			TenantID:    tenant.ID,
			Name:        "update-test-team",
			DisplayName: "Update Test Team",
			LeadEmail:   "update-lead@company.com",
			CreatedBy:   "test-user",
		}

		created, err := service.CreateTeam(ctx, team)
		require.NoError(t, err)

		// Update the team
		created.DisplayName = "Updated Display Name"
		created.Description = stringPtr("Updated description")
		created.Department = stringPtr("Updated Department")
		
		updated, err := service.UpdateTeam(ctx, created)
		require.NoError(t, err)

		assert.Equal(t, created.ID, updated.ID)
		assert.Equal(t, "Updated Display Name", updated.DisplayName)
		assert.Equal(t, "Updated description", *updated.Description)
		assert.Equal(t, "Updated Department", *updated.Department)
		assert.True(t, updated.UpdatedAt.After(updated.CreatedAt))
		assert.NotNil(t, updated.UpdatedBy)
		assert.Equal(t, "system", *updated.UpdatedBy)
	})

	t.Run("non-existent team", func(t *testing.T) {
		team := Team{
			ID:          uuid.New(),
			Name:        "non-existent",
			DisplayName: "Non Existent",
			LeadEmail:   "non@company.com",
		}

		_, err := service.UpdateTeam(ctx, team)
		require.Error(t, err)
		assert.Equal(t, ErrTeamNotFound, err)
	})

	t.Run("missing required fields", func(t *testing.T) {
		// Test missing ID
		team := Team{
			Name:        "test",
			DisplayName: "Test",
			LeadEmail:   "test@company.com",
		}

		_, err := service.UpdateTeam(ctx, team)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "team ID is required")

		// Test missing name
		team = Team{
			ID:          uuid.New(),
			DisplayName: "Test",
			LeadEmail:   "test@company.com",
		}

		_, err = service.UpdateTeam(ctx, team)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")

		// Test missing lead email
		team = Team{
			ID:          uuid.New(),
			Name:        "test",
			DisplayName: "Test",
		}

		_, err = service.UpdateTeam(ctx, team)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "lead_email is required")
	})
}

func TestTeamService_DeleteTeam(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	pool, cleanup := testutils.SetupTestDB(t, ctx)
	defer cleanup()

	service := NewService(pool)

	// Create test tenant
	tenantManager := database.NewTenantManager(pool)
	tenant, err := tenantManager.CreateTenant(ctx, &database.CreateTenantRequest{
		Name:        "test-tenant",
		DisplayName: "Test Tenant",
		Description: stringPtr("Test tenant for team tests"),
	})
	require.NoError(t, err)

	t.Run("successful deletion", func(t *testing.T) {
		// Create a team first
		team := Team{
			TenantID:    tenant.ID,
			Name:        "delete-test-team",
			DisplayName: "Delete Test Team",
			LeadEmail:   "delete-lead@company.com",
			CreatedBy:   "test-user",
		}

		created, err := service.CreateTeam(ctx, team)
		require.NoError(t, err)

		// Delete the team
		err = service.DeleteTeam(ctx, created.ID)
		require.NoError(t, err)

		// Verify it's deleted
		_, err = service.GetTeam(ctx, created.ID)
		require.Error(t, err)
		assert.Equal(t, ErrTeamNotFound, err)
	})

	t.Run("non-existent team", func(t *testing.T) {
		nonExistentID := uuid.New()
		err := service.DeleteTeam(ctx, nonExistentID)
		require.Error(t, err)
		assert.Equal(t, ErrTeamNotFound, err)
	})
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}