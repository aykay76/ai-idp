# Logger Package

The `logger` package provides structured logging capabilities using Go's standard `slog` package with enhanced features for the AI-IDP platform.

## Features

- Structured logging with JSON and text formats
- Configurable log levels (debug, info, warn, error, fatal, panic)
- Context-aware logging with automatic field extraction
- Rich field-based logging with predefined field names
- HTTP request logging support
- Component and operation tracking
- Error and duration tracking
- Global logger instance for convenience
- Source location tracking for debug level

## Usage

### Basic Usage

```go
import "github.com/aykay76/ai-idp/internal/logger"

// Initialize global logger
logger.Init("info", "json")

// Use global functions
logger.Info("Server starting", "port", 8080)
logger.Error("Database connection failed", "error", err)
logger.Fatal("Critical error, exiting") // Logs and exits
```

### Structured Logging with Fields

```go
// Use predefined field helpers
logger.WithComponent("auth").
    WithOperation("login").
    WithField("user_id", "123").
    Info("User authentication successful")

// Use field maps
logger.WithFields(logger.LogFields{
    "user_id": "123",
    "tenant_id": "acme-corp",
    "operation": "create_application",
}).Info("Application created successfully")
```

### Context-Aware Logging

```go
// Add values to context
ctx = context.WithValue(ctx, logger.FieldRequestID, "req-123")
ctx = context.WithValue(ctx, logger.FieldUserID, "user-456")
ctx = context.WithValue(ctx, logger.FieldTenantID, "tenant-789")

// Logger will automatically extract and include these fields
logger.WithContext(ctx).Info("Processing request")
```

### HTTP Request Logging

```go
logger.WithHTTP("POST", "/api/v1/applications", 201).
    WithDuration(time.Since(start)).
    Info("HTTP request processed")
```

### Error Logging

```go
if err != nil {
    logger.WithError(err).
        WithComponent("database").
        Error("Failed to query database")
}
```

### Instance-Based Logging

```go
// Create logger instance with specific configuration
log := logger.New("debug", "text")
log.WithField("service", "application-service").Info("Service started")
```

## Predefined Field Names

The logger provides constants for common field names to ensure consistency:

- `logger.FieldRequestID` - HTTP request ID
- `logger.FieldUserID` - User identifier
- `logger.FieldTenantID` - Tenant identifier
- `logger.FieldComponent` - Component/service name
- `logger.FieldOperation` - Operation being performed
- `logger.FieldError` - Error message
- `logger.FieldDuration` - Operation duration
- `logger.FieldHTTPMethod` - HTTP method
- `logger.FieldHTTPPath` - HTTP path
- `logger.FieldHTTPStatus` - HTTP status code

## Log Levels

- `debug` - Detailed information for debugging (includes source location)
- `info` - General information about program execution
- `warn` - Warning messages
- `error` - Error messages
- `fatal` - Fatal errors that cause program termination
- `panic` - Panic-level errors that cause program panic

## Output Formats

- `json` - Structured JSON output (recommended for production)
- `text` - Human-readable text output (good for development)

## Integration with Config

The logger integrates with the configuration system:

```go
config := config.Load()
logger.Init(config.Logging.Level, config.Logging.Format)
```