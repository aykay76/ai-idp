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
	"github.com/aykay76/ai-idp/internal/logger"
	"github.com/aykay76/ai-idp/internal/middleware"
	"github.com/aykay76/ai-idp/internal/proxy"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	Version   string    `json:"version,omitempty"`
}

// getEnvWithDefault gets an environment variable or returns a default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	appLogger := logger.New(cfg.Logging.Level, cfg.Logging.Format)
	appLogger.WithFields(logger.LogFields{
		logger.FieldComponent: "api-gateway",
		"port":                cfg.Server.Port,
	}).Info("Starting API Gateway")

	// Create HTTP server mux
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		response := HealthResponse{
			Status:    "healthy",
			Timestamp: time.Now().UTC(),
			Service:   "ai-idp-api-gateway",
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
			Service:   "ai-idp-api-gateway",
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
			Service:   "ai-idp-api-gateway",
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

	// Setup proxy configuration
	proxyConfig := &proxy.ProxyConfig{
		ApplicationServiceURL: getEnvWithDefault("APPLICATION_SERVICE_URL", "http://localhost:8082"),
		TeamServiceURL:        getEnvWithDefault("TEAM_SERVICE_URL", "http://localhost:8083"),
		Logger:                appLogger,
	}

	// Create proxy handler
	proxyHandler := proxy.NewProxyHandler(proxyConfig)

	// Add proxy routes for API endpoints
	mux.Handle("/api/", proxyHandler)

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
			logger.FieldComponent: "api-gateway",
			"address":             server.Addr,
		}).Info("API Gateway server starting")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.WithFields(logger.LogFields{
				logger.FieldComponent: "api-gateway",
				logger.FieldError:     err.Error(),
			}).Error("Server failed to start")
			os.Exit(1)
		}
	}()

	appLogger.WithFields(logger.LogFields{
		logger.FieldComponent: "api-gateway",
		"port":                cfg.Server.Port,
	}).Info("API Gateway server started successfully")

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.WithFields(logger.LogFields{
		logger.FieldComponent: "api-gateway",
	}).Info("API Gateway server shutting down")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		appLogger.WithFields(logger.LogFields{
			logger.FieldComponent: "api-gateway",
			logger.FieldError:     err.Error(),
		}).Error("Server forced to shutdown")
		os.Exit(1)
	}

	appLogger.WithFields(logger.LogFields{
		logger.FieldComponent: "api-gateway",
	}).Info("API Gateway server stopped")
}
