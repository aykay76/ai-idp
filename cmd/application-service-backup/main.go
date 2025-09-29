package main

import (
	"context"

	"github.com/aykay76/ai-idp/internal/applications"
	"github.com/aykay76/ai-idp/internal/logger"
	"github.com/aykay76/ai-idp/internal/server"
)

func main() {
	// Load configuration from environment
	config := server.LoadConfigFromEnv("application-service", "8081")

	// Initialize structured logging
	logger.Init(config.Logging.Level, config.Logging.Format)
	log := logger.WithComponent("application-service")

	// Validate configuration
	if err := config.Validate(); err != nil {
		log.WithError(err).Fatal("Configuration validation failed")
	}

	log.Info("Starting application service", "port", config.Server.Port)

	// Create resource template for applications
	template := server.NewResourceTemplate("application-service", "applications", config)

	// Setup database
	ctx := context.Background()
	if err := template.SetupDatabase(ctx); err != nil {
		log.WithError(err).Fatal("Failed to setup database")
	}

	log.Info("Database setup completed successfully")

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

	log.Info("Routes registered successfully")

	// Start server
	if err := template.Run(ctx); err != nil {
		log.WithError(err).Fatal("Server failed")
	}
}
