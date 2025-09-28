package main

import (
	"context"
	"errors"

	"github.com/aykay76/ai-idp/internal/logger"
)

func main() {
	// Initialize the logger with text format for demonstration
	logger.Init("info", "text")

	logger.Info("Starting demonstration with text format")

	// Structured logging with fields
	logger.WithFields(logger.LogFields{
		"service": "demo",
		"format":  "text",
	}).Info("Text format logging demonstration")

	// Component and operation logging
	logger.WithComponent("authentication").
		WithOperation("login").
		WithField("user_id", "user-123").
		Info("User authentication successful")

	// Error logging
	err := errors.New("example error for demonstration")
	logger.WithError(err).
		WithComponent("database").
		Error("Example error log entry")

	// Context-aware logging
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.FieldRequestID, "req-456")

	logger.WithContext(ctx).
		WithComponent("user-service").
		Info("Processing with context")

	logger.Info("Text format demonstration complete")
}
