package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/aykay76/ai-idp/internal/applications"
	"github.com/aykay76/ai-idp/internal/config"
	"github.com/aykay76/ai-idp/internal/database"
	"github.com/aykay76/ai-idp/internal/logger"
	"github.com/aykay76/ai-idp/internal/middleware"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	Version   string    `json:"version,omitempty"`
}

func main() {
	// Load configuration
	cfg := config.LoadWithDefaults("application-service", "8082")

	// Initialize logger
	appLogger := logger.New(cfg.Logging.Level, cfg.Logging.Format)
	appLogger.WithFields(logger.LogFields{
		logger.FieldComponent: "application-service",
		"port":                cfg.Server.Port,
	}).Info("Starting Application Service")

	// Setup database connection
	ctx := context.Background()
	dbConfig := database.DefaultConfig(cfg.Database.URL)
	dbPool, err := database.NewPool(ctx, dbConfig)
	if err != nil {
		appLogger.WithFields(logger.LogFields{
			logger.FieldComponent: "application-service",
			logger.FieldError:     err.Error(),
		}).Fatal("Failed to connect to database")
	}
	defer dbPool.Close()

	appLogger.WithFields(logger.LogFields{
		logger.FieldComponent: "application-service",
	}).Info("Database connection established")

	// Initialize application service
	appService := applications.NewService(dbPool)
	appHandlers := applications.NewHandlers(appService, appLogger)

	// Create HTTP server mux
	mux := http.NewServeMux()

	// Health check endpoints
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		response := HealthResponse{
			Status:    "healthy",
			Timestamp: time.Now().UTC(),
			Service:   "ai-idp-application-service",
			Version:   os.Getenv("VERSION"),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			appLogger.WithFields(logger.LogFields{
				logger.FieldError: err.Error(),
			}).Error("Failed to encode health response")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})

	// Application endpoints
	mux.HandleFunc("GET /api/v1/applications", appHandlers.ListApplications)
	mux.HandleFunc("POST /api/v1/applications", appHandlers.CreateApplication)
	mux.HandleFunc("GET /api/v1/applications/{id}", appHandlers.GetApplication)
	mux.HandleFunc("PUT /api/v1/applications/{id}", appHandlers.UpdateApplication)
	mux.HandleFunc("DELETE /api/v1/applications/{id}", appHandlers.DeleteApplication)

	// Apply middleware chain
	handler := middleware.RequestID(mux)
	handler = middleware.Logging(appLogger)(handler)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		appLogger.WithFields(logger.LogFields{
			logger.FieldComponent: "application-service",
			"address":             server.Addr,
		}).Info("Application service starting")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.WithFields(logger.LogFields{
				logger.FieldComponent: "application-service",
				logger.FieldError:     err.Error(),
			}).Fatal("Server failed to start")
		}
	}()

	appLogger.WithFields(logger.LogFields{
		logger.FieldComponent: "application-service",
		"port":                cfg.Server.Port,
	}).Info("Application service started successfully")

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.WithFields(logger.LogFields{
		logger.FieldComponent: "application-service",
	}).Info("Application service shutting down")

	// Create a context with timeout for shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(shutdownCtx); err != nil {
		appLogger.WithFields(logger.LogFields{
			logger.FieldComponent: "application-service",
			logger.FieldError:     err.Error(),
		}).Error("Server forced to shutdown")
		os.Exit(1)
	}

	appLogger.WithFields(logger.LogFields{
		logger.FieldComponent: "application-service",
	}).Info("Application service stopped")
}

// parseQuery helper function for parsing query parameters
func parseQuery(r *http.Request, key string, defaultValue int) int {
	if value := r.URL.Query().Get(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
