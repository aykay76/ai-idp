package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aykay76/ai-idp/internal/config"
	"github.com/aykay76/ai-idp/internal/database"
	"github.com/aykay76/ai-idp/internal/logger"
	"github.com/aykay76/ai-idp/internal/middleware"
	"github.com/aykay76/ai-idp/internal/teams"
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
	cfg := config.LoadWithDefaults("team-service", "8083")

	// Initialize logger
	appLogger := logger.New(cfg.Logging.Level, cfg.Logging.Format)
	appLogger.WithFields(logger.LogFields{
		logger.FieldComponent: "team-service",
		"port":                cfg.Server.Port,
	}).Info("Starting Team Service")

	// Setup database connection
	ctx := context.Background()
	dbConfig := database.DefaultConfig(cfg.Database.URL)
	dbPool, err := database.NewPool(ctx, dbConfig)
	if err != nil {
		appLogger.WithFields(logger.LogFields{
			logger.FieldComponent: "team-service",
			logger.FieldError:     err.Error(),
		}).Fatal("Failed to connect to database")
	}
	defer dbPool.Close()

	// Initialize team service
	teamService := teams.NewService(dbPool)
	teamHandlers := teams.NewHandlers(teamService, appLogger)

	// Create HTTP server mux
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		response := HealthResponse{
			Status:    "healthy",
			Timestamp: time.Now().UTC(),
			Service:   "ai-idp-team-service",
			Version:   os.Getenv("VERSION"),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			appLogger.WithFields(logger.LogFields{
				logger.FieldError: err.Error(),
			}).Error("Failed to encode health response")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		appLogger.WithFields(logger.LogFields{
			logger.FieldHTTPMethod: r.Method,
			logger.FieldHTTPPath:   r.URL.Path,
			logger.FieldHTTPStatus: http.StatusOK,
		}).Debug("Health check requested")
	})

	// Readiness check endpoint (for Kubernetes)
	mux.HandleFunc("GET /readiness", func(w http.ResponseWriter, r *http.Request) {
		response := HealthResponse{
			Status:    "ready",
			Timestamp: time.Now().UTC(),
			Service:   "ai-idp-team-service",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			appLogger.WithFields(logger.LogFields{
				logger.FieldError: err.Error(),
			}).Error("Failed to encode readiness response")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		appLogger.WithFields(logger.LogFields{
			logger.FieldHTTPMethod: r.Method,
			logger.FieldHTTPPath:   r.URL.Path,
			logger.FieldHTTPStatus: http.StatusOK,
		}).Debug("Readiness check requested")
	})

	// Liveness check endpoint (for Kubernetes)
	mux.HandleFunc("GET /liveness", func(w http.ResponseWriter, r *http.Request) {
		response := HealthResponse{
			Status:    "alive",
			Timestamp: time.Now().UTC(),
			Service:   "ai-idp-team-service",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			appLogger.WithFields(logger.LogFields{
				logger.FieldError: err.Error(),
			}).Error("Failed to encode liveness response")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		appLogger.WithFields(logger.LogFields{
			logger.FieldHTTPMethod: r.Method,
			logger.FieldHTTPPath:   r.URL.Path,
			logger.FieldHTTPStatus: http.StatusOK,
		}).Debug("Liveness check requested")
	})

	// Team API endpoints
	mux.HandleFunc("POST /api/v1/teams", teamHandlers.CreateTeam)
	mux.HandleFunc("GET /api/v1/teams/{id}", teamHandlers.GetTeam)
	mux.HandleFunc("PUT /api/v1/teams/{id}", teamHandlers.UpdateTeam)
	mux.HandleFunc("DELETE /api/v1/teams/{id}", teamHandlers.DeleteTeam)
	mux.HandleFunc("GET /api/v1/teams", teamHandlers.ListTeams)

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
			logger.FieldComponent: "team-service",
			"address":             server.Addr,
		}).Info("Team Service server starting")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.WithFields(logger.LogFields{
				logger.FieldComponent: "team-service",
				logger.FieldError:     err.Error(),
			}).Error("Server failed to start")
			os.Exit(1)
		}
	}()

	appLogger.WithFields(logger.LogFields{
		logger.FieldComponent: "team-service",
		"port":                cfg.Server.Port,
	}).Info("Team Service server started successfully")

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.WithFields(logger.LogFields{
		logger.FieldComponent: "team-service",
	}).Info("Team Service server shutting down")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		appLogger.WithFields(logger.LogFields{
			logger.FieldComponent: "team-service",
			logger.FieldError:     err.Error(),
		}).Error("Server forced to shutdown")
		os.Exit(1)
	}

	appLogger.WithFields(logger.LogFields{
		logger.FieldComponent: "team-service",
	}).Info("Team Service server stopped")
}
