package logger

import (
	"context"
	"log/slog"
	"os"
	"time"
)

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
	level slog.Level
}

// LogFields represents a map of structured log fields
type LogFields map[string]interface{}

// Common log field keys
const (
	FieldRequestID  = "request_id"
	FieldUserID     = "user_id"
	FieldTenantID   = "tenant_id"
	FieldComponent  = "component"
	FieldOperation  = "operation"
	FieldError      = "error"
	FieldDuration   = "duration"
	FieldHTTPMethod = "http_method"
	FieldHTTPPath   = "http_path"
	FieldHTTPStatus = "http_status"
)

// New creates a new logger instance with the specified level and format
func New(level string, format string) *Logger {
	logLevel := parseLevel(level)

	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: logLevel <= slog.LevelDebug, // Add source info for debug level
	}

	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
		level:  logLevel,
	}
}

// parseLevel converts string level to slog.Level
func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// IsLevelEnabled checks if the given level is enabled for this logger
func (l *Logger) IsLevelEnabled(level slog.Level) bool {
	return l.level <= level
}

// WithContext returns a logger with context values extracted
func (l *Logger) WithContext(ctx context.Context) *Logger {
	logger := l.Logger

	// Extract common context values and add them as fields
	if requestID := ctx.Value(FieldRequestID); requestID != nil {
		logger = logger.With(FieldRequestID, requestID)
	}
	if userID := ctx.Value(FieldUserID); userID != nil {
		logger = logger.With(FieldUserID, userID)
	}
	if tenantID := ctx.Value(FieldTenantID); tenantID != nil {
		logger = logger.With(FieldTenantID, tenantID)
	}

	return &Logger{
		Logger: logger,
		level:  l.level,
	}
}

// WithFields returns a logger with additional fields
func (l *Logger) WithFields(fields LogFields) *Logger {
	if len(fields) == 0 {
		return l
	}

	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}

	return &Logger{
		Logger: l.Logger.With(args...),
		level:  l.level,
	}
}

// WithField returns a logger with a single additional field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		Logger: l.Logger.With(key, value),
		level:  l.level,
	}
}

// WithComponent adds a component field to the logger
func (l *Logger) WithComponent(component string) *Logger {
	return l.WithField(FieldComponent, component)
}

// WithOperation adds an operation field to the logger
func (l *Logger) WithOperation(operation string) *Logger {
	return l.WithField(FieldOperation, operation)
}

// WithError adds an error field to the logger
func (l *Logger) WithError(err error) *Logger {
	if err == nil {
		return l
	}
	return l.WithField(FieldError, err.Error())
}

// WithDuration adds a duration field to the logger
func (l *Logger) WithDuration(d time.Duration) *Logger {
	return l.WithField(FieldDuration, d.String())
}

// WithHTTP adds HTTP-related fields to the logger
func (l *Logger) WithHTTP(method, path string, status int) *Logger {
	return l.WithFields(LogFields{
		FieldHTTPMethod: method,
		FieldHTTPPath:   path,
		FieldHTTPStatus: status,
	})
}

// Fatal logs at Fatal level and exits the program
func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.Error(msg, args...)
	os.Exit(1)
}

// Panic logs at Error level and panics
func (l *Logger) Panic(msg string, args ...interface{}) {
	l.Error(msg, args...)
	panic(msg)
}

// With returns a logger with the given arguments added as fields
func (l *Logger) With(args ...interface{}) *Logger {
	return &Logger{
		Logger: l.Logger.With(args...),
		level:  l.level,
	}
}

// Global logger instance
var globalLogger *Logger

// Init initializes the global logger with the specified level and format
func Init(level, format string) {
	globalLogger = New(level, format)
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	if globalLogger == nil {
		globalLogger = New("info", "json")
	}
	return globalLogger
}

// Global convenience functions
func Debug(msg string, args ...interface{}) {
	GetGlobalLogger().Debug(msg, args...)
}

func Info(msg string, args ...interface{}) {
	GetGlobalLogger().Info(msg, args...)
}

func Warn(msg string, args ...interface{}) {
	GetGlobalLogger().Warn(msg, args...)
}

func Error(msg string, args ...interface{}) {
	GetGlobalLogger().Error(msg, args...)
}

func Fatal(msg string, args ...interface{}) {
	GetGlobalLogger().Fatal(msg, args...)
}

func Panic(msg string, args ...interface{}) {
	GetGlobalLogger().Panic(msg, args...)
}

func With(args ...interface{}) *Logger {
	return &Logger{
		Logger: GetGlobalLogger().Logger.With(args...),
		level:  GetGlobalLogger().level,
	}
}

func WithFields(fields LogFields) *Logger {
	return GetGlobalLogger().WithFields(fields)
}

func WithField(key string, value interface{}) *Logger {
	return GetGlobalLogger().WithField(key, value)
}

func WithComponent(component string) *Logger {
	return GetGlobalLogger().WithComponent(component)
}

func WithError(err error) *Logger {
	return GetGlobalLogger().WithError(err)
}

func WithContext(ctx context.Context) *Logger {
	return GetGlobalLogger().WithContext(ctx)
}
