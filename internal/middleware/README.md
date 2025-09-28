# Middleware Package

The `middleware` package provides HTTP middleware components for the AI-IDP platform.

## Available Middleware

### RequestID
Adds unique request IDs to each HTTP request for tracing and correlation.

### Logging
Logs HTTP requests with structured data including method, path, status, duration, and request ID.

### CORS
Handles Cross-Origin Resource Sharing (CORS) headers for web API access.

### Recovery
Recovers from panics in HTTP handlers and logs them appropriately.

## Usage

```go
import (
    "github.com/aykay76/ai-idp/internal/middleware"
    "github.com/aykay76/ai-idp/internal/logger"
)

log := logger.New("info", "json")

// Create middleware chain
mux := http.NewServeMux()
handler := middleware.RequestID(
    middleware.Logging(log)(
        middleware.CORS([]string{"*"})(
            middleware.Recovery(log)(mux),
        ),
    ),
)

// Start server
http.ListenAndServe(":8080", handler)
```

## Middleware Order

Recommended middleware order (from outer to inner):
1. Recovery (catches all panics)
2. RequestID (adds tracing)
3. Logging (logs requests)
4. CORS (handles CORS headers)
5. Your application handlers