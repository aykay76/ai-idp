# Types Package

The `types` package provides core domain types, constants, and data structures used across the AI-IDP platform.

## Core Types

### Application
Represents an application configuration in the platform:
- ID: Unique identifier
- Name: Human-readable application name
- Team: Owning team
- Repository: Git repository URL
- Config: YAML configuration
- Status: Current reconciliation status
- CreatedAt/UpdatedAt: Timestamps

### ReconcileStatus
Enumeration of reconciliation states:
- `StatusPending`: Waiting to be processed
- `StatusInProgress`: Currently being reconciled
- `StatusCompleted`: Successfully completed
- `StatusFailed`: Failed with errors

### ContextKey
Type-safe context keys to avoid string collisions:
- `RequestIDKey`: For storing request IDs
- `UserIDKey`: For storing user identifiers
- `TenantIDKey`: For storing tenant information

### APIResponse
Standardized API response structure:
- Success: Boolean status
- Message: Human-readable message
- Data: Response payload
- Error: Error details (if any)

## Usage

```go
import (
    "context"
    "github.com/aykay76/ai-idp/internal/types"
)

// Using context keys
ctx := context.WithValue(context.Background(), types.RequestIDKey, "123")
requestID := ctx.Value(types.RequestIDKey).(string)

// Creating API responses
response := types.APIResponse{
    Success: true,
    Message: "Application created successfully",
    Data:    application,
}

// Working with applications
app := types.Application{
    Name:   "my-service",
    Team:   "platform",
    Status: types.StatusPending,
}
```

## Constants

The package defines platform-wide constants for status values and ensures type safety across the codebase.