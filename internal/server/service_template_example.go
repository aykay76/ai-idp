package server

/*
This file provides examples of how to use the service template for creating new services.

Example 1: Simple Resource Service (Team Service)
===============================================

Create: cmd/team-service/main.go
```go
package main

import (
	"context"
	"log"

	"github.com/aykay76/ai-idp/internal/server"
	"github.com/aykay76/ai-idp/internal/teams" // Your team service package
)

func main() {
	// Load configuration
	config := server.LoadConfigFromEnv("team-service", "8082")

	// Create resource template
	template := server.NewResourceTemplate("team-service", "teams", config)

	// Setup database
	ctx := context.Background()
	if err := template.SetupDatabase(ctx); err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	// Create service and handlers
	teamService := teams.NewService(template.GetDB())
	teamHandlers := teams.NewTeamHandlers(teamService)

	// Register CRUD routes automatically
	template.RegisterResourceHandlers(teamHandlers.GetCRUDHandlers())

	// Add custom routes if needed
	template.HandleFunc("GET /api/v1/teams/stats",
		server.WithTenantValidation(teamHandlers.GetTeamStats))

	// Start server
	if err := template.Run(ctx); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
```

Example 2: Resource Provider Service (GitHub Provider)
====================================================

Create: cmd/github-provider/main.go
```go
package main

import (
	"context"
	"log"

	"github.com/aykay76/ai-idp/internal/server"
	"github.com/aykay76/ai-idp/internal/github" // Your GitHub provider package
)

func main() {
	// Load configuration
	config := server.LoadConfigFromEnv("github-provider", "8083")

	// Add GitHub-specific config
	config.GitHubAppID = getEnv("GITHUB_APP_ID", "")
	config.GitHubPrivateKey = getEnv("GITHUB_PRIVATE_KEY", "")

	// Create service template (not resource template since it's not CRUD)
	template := server.NewServiceTemplate("github-provider", "1.0.0", config)

	// Setup database
	ctx := context.Background()
	if err := template.SetupDatabase(ctx); err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	// Create GitHub service and handlers
	githubService := github.NewService(template.GetDB(), config)
	githubHandlers := github.NewGitHubHandlers(githubService)

	// Register custom routes (not CRUD)
	template.HandleFunc("POST /api/v1/repositories",
		server.WithTenantValidation(githubHandlers.CreateRepository))
	template.HandleFunc("GET /api/v1/repositories",
		server.WithTenantValidation(githubHandlers.ListRepositories))
	template.HandleFunc("POST /api/v1/webhooks/github",
		githubHandlers.HandleWebhook) // No tenant validation for webhooks

	// Provider-specific endpoints
	template.HandleFunc("POST /api/v1/github/sync",
		server.WithTenantValidation(githubHandlers.SyncRepositories))

	// Start server
	if err := template.Run(ctx); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
```

Example 3: Handlers Implementation Pattern
========================================

Create: internal/teams/handlers.go
```go
package teams

import (
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/aykay76/ai-idp/internal/server"
)

type TeamHandlers struct {
	service  *Service
	validate *validator.Validate
}

func NewTeamHandlers(service *Service) *TeamHandlers {
	return &TeamHandlers{
		service:  service,
		validate: validator.New(),
	}
}

// Implement CRUD handlers interface
func (h *TeamHandlers) GetCRUDHandlers() *server.CRUDHandlers {
	return &server.CRUDHandlers{
		Create: server.WithTenantValidation(h.CreateTeam),
		List:   server.WithTenantValidation(h.ListTeams),
		Get:    server.WithTenantValidation(h.GetTeam),
		Update: server.WithTenantValidation(h.UpdateTeam),
		Delete: server.WithTenantValidation(h.DeleteTeam),
	}
}

func (h *TeamHandlers) CreateTeam(w http.ResponseWriter, r *http.Request) {
	tenantCtx, _ := server.GetTenantFromContext(r)

	var req CreateTeamRequest
	if err := server.ParseJSONBody(r, &req); err != nil {
		server.RespondWithValidationError(w, err)
		return
	}

	if err := h.validate.Struct(&req); err != nil {
		server.RespondWithValidationError(w, err)
		return
	}

	team, err := h.service.CreateTeam(r.Context(), tenantCtx.TenantID, &req, tenantCtx.UserID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			server.RespondWithError(w, http.StatusConflict, err, "Team name already exists")
			return
		}
		server.RespondWithInternalError(w, err)
		return
	}

	server.RespondCreated(w, team)
}

// ... implement other CRUD handlers following the same pattern

func (h *TeamHandlers) GetTeamStats(w http.ResponseWriter, r *http.Request) {
	tenantCtx, _ := server.GetTenantFromContext(r)

	stats, err := h.service.GetTeamStats(r.Context(), tenantCtx.TenantID)
	if err != nil {
		server.RespondWithInternalError(w, err)
		return
	}

	server.RespondWithData(w, stats)
}
```

Example 4: Service Layer Pattern
===============================

Create: internal/teams/service.go
```go
package teams

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/aykay76/ai-idp/internal/database"
)

type Service struct {
	db *database.Pool
}

func NewService(db *database.Pool) *Service {
	return &Service{db: db}
}

type Team struct {
	ID          uuid.UUID `json:"id" db:"id"`
	TenantID    uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Name        string    `json:"name" db:"name"`
	DisplayName string    `json:"display_name" db:"display_name"`
	// ... other fields
}

type CreateTeamRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=255"`
	DisplayName string `json:"display_name" validate:"required,min=2,max=255"`
	// ... other fields
}

func (s *Service) CreateTeam(ctx context.Context, tenantID uuid.UUID, req *CreateTeamRequest, createdBy string) (*Team, error) {
	// Implementation follows the same pattern as application service
	team := &Team{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		// ... set other fields
	}

	query := `
		INSERT INTO resource_management.teams (id, tenant_id, name, display_name, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`

	err := s.db.QueryRow(ctx, query, team.ID, team.TenantID, team.Name, team.DisplayName, createdBy).
		Scan(&team.CreatedAt, &team.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	return team, nil
}

// ... implement other service methods
```

Key Benefits of This Template Approach:
=====================================

1. **Consistent Structure**: All services follow the same patterns
2. **Reduced Boilerplate**: Common functionality is handled by the template
3. **Easy Testing**: Standard patterns for mocking and testing
4. **Tenant Isolation**: Built-in tenant context validation
5. **Error Handling**: Consistent error responses across services
6. **Observability**: Standard health, readiness, metrics endpoints
7. **Middleware**: Logging, recovery, CORS handled automatically

Quick Start Checklist for New Services:
=======================================

□ 1. Copy cmd/application-service/ to cmd/your-service/
□ 2. Update service name and port in main.go
□ 3. Create internal/yourservice/service.go (business logic)
□ 4. Create internal/yourservice/handlers.go (HTTP handlers)
□ 5. Add database tables/schemas for your resource
□ 6. Implement the CRUDHandlers interface
□ 7. Add service to docker-compose.yml
□ 8. Add service to Makefile targets
□ 9. Write tests following the same patterns
□ 10. Update README.md with your service documentation

The template handles everything else automatically!
*/
