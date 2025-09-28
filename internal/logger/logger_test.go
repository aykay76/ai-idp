package logger

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		format   string
		expected string
	}{
		{"debug_json", "debug", "json", "debug"},
		{"info_json", "info", "json", "info"},
		{"warn_text", "warn", "text", "warn"},
		{"error_text", "error", "text", "error"},
		{"invalid_level", "invalid", "json", "info"}, // defaults to info
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.level, tt.format)
			if logger == nil {
				t.Fatal("Expected logger to be created")
			}
			// Test that logger is functional
			logger.Info("test message")
		})
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"debug", "DEBUG"},
		{"info", "INFO"},
		{"warn", "WARN"},
		{"warning", "WARN"},
		{"error", "ERROR"},
		{"invalid", "INFO"}, // defaults to info
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level := parseLevel(tt.input)
			if level.String() != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, level.String())
			}
		})
	}
}

func TestLoggerWithFields(t *testing.T) {
	logger := New("debug", "json")

	// Test WithField
	newLogger := logger.WithField("test_key", "test_value")
	if newLogger == nil {
		t.Fatal("Expected logger to be returned")
	}

	// Test WithFields
	fields := LogFields{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}
	fieldsLogger := logger.WithFields(fields)
	if fieldsLogger == nil {
		t.Fatal("Expected logger to be returned")
	}

	// Test empty fields
	emptyLogger := logger.WithFields(LogFields{})
	if emptyLogger != logger {
		t.Error("Expected same logger instance for empty fields")
	}
}

func TestLoggerWithContext(t *testing.T) {
	logger := New("debug", "json")

	// Test with context containing values
	ctx := context.Background()
	ctx = context.WithValue(ctx, FieldRequestID, "req-123")
	ctx = context.WithValue(ctx, FieldUserID, "user-456")
	ctx = context.WithValue(ctx, FieldTenantID, "tenant-789")

	ctxLogger := logger.WithContext(ctx)
	if ctxLogger == nil {
		t.Fatal("Expected logger to be returned")
	}

	// Test with empty context
	emptyCtx := context.Background()
	emptyLogger := logger.WithContext(emptyCtx)
	if emptyLogger == nil {
		t.Fatal("Expected logger to be returned")
	}
}

func TestLoggerHelperMethods(t *testing.T) {
	logger := New("debug", "json")

	// Test WithComponent
	compLogger := logger.WithComponent("test-component")
	if compLogger == nil {
		t.Fatal("Expected logger to be returned")
	}

	// Test WithOperation
	opLogger := logger.WithOperation("test-operation")
	if opLogger == nil {
		t.Fatal("Expected logger to be returned")
	}

	// Test WithError
	testErr := errors.New("test error")
	errLogger := logger.WithError(testErr)
	if errLogger == nil {
		t.Fatal("Expected logger to be returned")
	}

	// Test WithError with nil
	nilErrLogger := logger.WithError(nil)
	if nilErrLogger != logger {
		t.Error("Expected same logger instance for nil error")
	}

	// Test WithDuration
	duration := time.Second * 5
	durLogger := logger.WithDuration(duration)
	if durLogger == nil {
		t.Fatal("Expected logger to be returned")
	}

	// Test WithHTTP
	httpLogger := logger.WithHTTP("GET", "/test", 200)
	if httpLogger == nil {
		t.Fatal("Expected logger to be returned")
	}
}

func TestGlobalLogger(t *testing.T) {
	// Test Init
	Init("debug", "json")

	global := GetGlobalLogger()
	if global == nil {
		t.Fatal("Expected global logger to be initialized")
	}

	// Test global functions
	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")

	// Test With functions
	withLogger := With("key", "value")
	if withLogger == nil {
		t.Fatal("Expected logger to be returned")
	}

	fieldsLogger := WithFields(LogFields{"key": "value"})
	if fieldsLogger == nil {
		t.Fatal("Expected logger to be returned")
	}

	fieldLogger := WithField("key", "value")
	if fieldLogger == nil {
		t.Fatal("Expected logger to be returned")
	}

	compLogger := WithComponent("test")
	if compLogger == nil {
		t.Fatal("Expected logger to be returned")
	}

	errLogger := WithError(errors.New("test"))
	if errLogger == nil {
		t.Fatal("Expected logger to be returned")
	}

	ctxLogger := WithContext(context.Background())
	if ctxLogger == nil {
		t.Fatal("Expected logger to be returned")
	}
}

func TestIsLevelEnabled(t *testing.T) {
	tests := []struct {
		name        string
		loggerLevel string
		testLevel   string
		enabled     bool
	}{
		{"debug_logger_debug_level", "debug", "debug", true},
		{"debug_logger_info_level", "debug", "info", true},
		{"debug_logger_error_level", "debug", "error", true},
		{"info_logger_debug_level", "info", "debug", false},
		{"info_logger_info_level", "info", "info", true},
		{"info_logger_error_level", "info", "error", true},
		{"error_logger_debug_level", "error", "debug", false},
		{"error_logger_info_level", "error", "info", false},
		{"error_logger_error_level", "error", "error", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.loggerLevel, "json")
			testLevel := parseLevel(tt.testLevel)
			enabled := logger.IsLevelEnabled(testLevel)

			if enabled != tt.enabled {
				t.Errorf("Expected %v, got %v for logger level %s, test level %s",
					tt.enabled, enabled, tt.loggerLevel, tt.testLevel)
			}
		})
	}
}

func TestJSONOutput(t *testing.T) {
	// This test would require capturing stdout, which is complex
	// For now, we just ensure the logger creates without errors
	logger := New("info", "json")

	// Test that logging doesn't panic
	logger.Info("test message", "key", "value")
	logger.WithField("test", "value").Info("test with field")
	logger.WithFields(LogFields{"key1": "value1", "key2": 42}).Info("test with fields")
}

func TestTextOutput(t *testing.T) {
	// Similar to JSON test - ensure text format works
	logger := New("info", "text")

	// Test that logging doesn't panic
	logger.Info("test message", "key", "value")
	logger.WithField("test", "value").Info("test with field")
	logger.WithFields(LogFields{"key1": "value1", "key2": 42}).Info("test with fields")
}

func TestFieldConstants(t *testing.T) {
	// Test that all field constants are defined and unique
	fields := []string{
		FieldRequestID,
		FieldUserID,
		FieldTenantID,
		FieldComponent,
		FieldOperation,
		FieldError,
		FieldDuration,
		FieldHTTPMethod,
		FieldHTTPPath,
		FieldHTTPStatus,
	}

	// Check that all fields are non-empty
	for _, field := range fields {
		if field == "" {
			t.Error("Field constant is empty")
		}
	}

	// Check for uniqueness
	seen := make(map[string]bool)
	for _, field := range fields {
		if seen[field] {
			t.Errorf("Duplicate field constant: %s", field)
		}
		seen[field] = true
	}
}
