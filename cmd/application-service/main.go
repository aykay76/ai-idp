package main

import (
	"context"
	"log"

	"github.com/aykay76/ai-idp/internal/applications"
	"github.com/aykay76/ai-idp/internal/server"
)

func main() {
	// Load configuration from environment
	config := server.LoadConfigFromEnv("application-service", "8081")

	// Validate configuration
	if err := config.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// Create resource template for applications
	template := server.NewResourceTemplate("application-service", "applications", config)

	// Setup database
	ctx := context.Background()
	if err := template.SetupDatabase(ctx); err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	// Setup application service
	appService := applications.NewService(template.GetDB())
	appHandlers := applications.NewApplicationHandlers(appService)

	// Register CRUD routes
	template.RegisterResourceHandlers(appHandlers.GetCRUDHandlers())

	// Register additional routes
	template.HandleFunc("GET /api/v1/applications/by-team/{teamName}",
		server.WithTenantValidation(appHandlers.GetApplicationsByTeam))
	template.HandleFunc("GET /api/v1/applications/stats",
		server.WithTenantValidation(appHandlers.GetApplicationStats))

	// Start server
	if err := template.Run(ctx); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
