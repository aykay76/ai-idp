package main

import (
	"context"
	"errors"
	"time"

	"github.com/aykay76/ai-idp/internal/logger"
)

func main() {
	// Initialize the logger with JSON format for demonstration
	logger.Init("debug", "json")

	// Basic logging
	logger.Info("Application starting up")
	logger.Debug("Debug information", "version", "1.0.0")

	// Structured logging with fields
	logger.WithFields(logger.LogFields{
		"service": "demo",
		"port":    8080,
		"env":     "development",
	}).Info("Server configuration loaded")

	// Component-based logging
	authLogger := logger.WithComponent("authentication")
	authLogger.Info("Authentication module initialized")

	// Operation tracking
	authLogger.WithOperation("login").
		WithField("user_id", "user-123").
		Info("User login attempt")

	// Error logging
	err := errors.New("database connection failed")
	logger.WithError(err).
		WithComponent("database").
		Error("Critical database error occurred")

	// HTTP request logging simulation
	start := time.Now()
	time.Sleep(10 * time.Millisecond) // Simulate request processing

	logger.GetGlobalLogger().
		WithHTTP("POST", "/api/v1/users", 201).
		WithDuration(time.Since(start)).
		WithField("user_id", "user-456").
		Info("HTTP request completed successfully")

	// Context-aware logging
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.FieldRequestID, "req-789")
	ctx = context.WithValue(ctx, logger.FieldTenantID, "tenant-acme")

	logger.WithContext(ctx).
		WithComponent("user-service").
		WithOperation("create_user").
		Info("Processing user creation request")

	// Different log levels
	logger.Debug("Detailed debugging information")
	logger.Info("General information")
	logger.Warn("Warning message")
	logger.Error("Error occurred")

	logger.Info("Application startup complete")
}
