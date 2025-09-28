# Config Package

The `config` package provides centralized configuration management for the AI-IDP platform with comprehensive environment variable support, validation, and structured configuration sections.

## Features

- **Structured Configuration**: Organized into logical sections (Server, Database, Redis, Logging, Security, GitHub)
- **Environment Variable Loading**: Automatic loading from environment variables with sensible defaults
- **Comprehensive Validation**: Built-in validation with detailed error messages
- **Multiple Data Types**: Support for strings, integers, booleans, durations, and more
- **Environment Detection**: Built-in methods for development/production environment detection
- **Backward Compatibility**: Helper methods for legacy code migration
- **Type Safety**: Strongly typed configuration structures

## Configuration Structure

```go
type Config struct {
    Environment string           // Application environment (development/staging/production)
    ServiceName string           // Service identifier
    
    Server   ServerConfig        // HTTP server configuration
    Database DatabaseConfig      // Database connection settings
    Redis    RedisConfig         // Redis cache configuration
    Logging  LoggingConfig       // Application logging settings
    Security SecurityConfig      // Security-related settings
    GitHub   GitHubConfig        // GitHub integration settings
}
```

## Usage

### Basic Usage

```go
import "github.com/aykay76/ai-idp/internal/config"

// Load configuration with defaults
cfg := config.Load()

// Validate configuration
}

// Access database configuration
db, err := database.NewPool(ctx, &database.Config{
    URL:             cfg.Database.URL,
    MaxConnections:  cfg.Database.MaxConnections,
    MinConnections:  cfg.Database.MinConnections,
    ConnectTimeout:  cfg.Database.ConnectTimeout,
})
```

### Backward Compatibility

For legacy code migration, helper methods are available:

```go
cfg := config.Load()

// Legacy access methods
dbURL := cfg.DatabaseURL()        // Returns cfg.Database.URL
redisURL := cfg.RedisURL()        // Returns cfg.Redis.URL
port := cfg.Port()                // Returns cfg.Server.Port
secret := cfg.JWTSecret()         // Returns cfg.Security.JWTSecret
```

## Environment Variables

The configuration system supports the following environment variables with their defaults:

### General Settings
- `ENVIRONMENT`: Application environment - development, staging, or production (default: "development")
- `SERVICE_NAME`: Service identifier (default: "ai-idp")

### Server Configuration
- `PORT`: HTTP server port (default: "8080")
- `HOST`: Server bind address (default: "0.0.0.0")
- `SHUTDOWN_TIMEOUT`: Graceful shutdown timeout (default: "30s")
- `DEBUG`: Enable debug mode - true/false (default: false)

### Database Configuration
- `DATABASE_URL`: PostgreSQL connection string (required)
- `DB_MAX_CONNECTIONS`: Maximum database connections (default: 25)
- `DB_MIN_CONNECTIONS`: Minimum database connections (default: 5)
- `DB_CONNECT_TIMEOUT`: Database connection timeout (default: "10s")
- `DB_MAX_IDLE_TIME`: Maximum connection idle time (default: "30m")

### Redis Configuration
- `REDIS_URL`: Redis connection string (default: "redis://:redis_dev_password@localhost:6379/0")
- `REDIS_PASSWORD`: Redis password (default: "")
- `REDIS_DB`: Redis database number (default: 0)

### Logging Configuration
- `LOG_LEVEL`: Logging level - debug, info, warn, error, fatal, panic (default: "info")
- `LOG_FORMAT`: Log format - json or text (default: "json")

### Security Configuration
- `JWT_SECRET`: JWT signing secret (required in production, default: "dev_jwt_secret_change_in_production")

### GitHub Integration
- `GITHUB_APP_ID`: GitHub App ID for integration
- `GITHUB_PRIVATE_KEY`: GitHub App private key content

## Environment Variable Formats

### Duration Values
Duration values can be specified in Go duration format or as seconds:
```bash
SHUTDOWN_TIMEOUT=30s           # 30 seconds
SHUTDOWN_TIMEOUT=5m            # 5 minutes
SHUTDOWN_TIMEOUT=60            # 60 seconds (fallback)
```

### Boolean Values
Boolean values accept multiple formats:
```bash
DEBUG=true          # true
DEBUG=1             # true
DEBUG=yes           # true
DEBUG=on            # true
DEBUG=false         # false
DEBUG=0             # false
DEBUG=no            # false
DEBUG=off           # false
```

## Configuration Validation

The config package includes comprehensive validation that checks:

- **Required Fields**: `DATABASE_URL` is always required; `JWT_SECRET` is required in production
- **Environment Values**: Must be one of: development, staging, production
- **Log Levels**: Must be one of: debug, info, warn, error, fatal, panic
- **Database Limits**: Max connections must be >= min connections
- **Port Specification**: Server port must be specified

```go
cfg := config.Load()
if err := cfg.Validate(); err != nil {
    log.Fatalf("Configuration validation failed: %v", err)
}
```

## Environment Detection

Built-in methods for environment-specific logic:

```go
cfg := config.Load()

if cfg.IsDevelopment() {
    // Enable debug features, verbose logging, etc.
    setupDevelopmentFeatures()
}

if cfg.IsProduction() {
    // Enable production optimizations, monitoring, etc.
    setupProductionFeatures()
}
```

## Testing

The package includes comprehensive tests covering:
- Default value loading
- Environment variable override
- Configuration validation
- Helper function behavior
- Error conditions

Run tests with:
```bash
go test ./internal/config
```

// Use structured configuration
fmt.Printf("Starting %s on port %s\n", cfg.ServiceName, cfg.Server.Port)
```

### Custom Service Configuration

```go
// Load with custom service name and default port
cfg := config.LoadWithDefaults("my-service", "9000")

// Environment detection
if cfg.IsDevelopment() {
    log.Println("Running in development mode")
}

if cfg.IsProduction() {
    log.Println("Running in production mode")
}
if err := cfg.Validate(); err != nil {
    log.Fatal(err)
}

// Use configuration
fmt.Printf("Server running on port %s\n", cfg.Port)
```

## Environment Variables

- `PORT` - Server port (default: 8080)
- `HOST` - Server host (default: 0.0.0.0)
- `ENVIRONMENT` - Environment (development/production, default: development)
- `DATABASE_URL` - Database connection string (required)
- `REDIS_URL` - Redis connection string (optional)
- `LOG_LEVEL` - Log level (default: info)
- `LOG_FORMAT` - Log format (json/text, default: json)
- `JWT_SECRET` - JWT signing secret (required in production)
- `SERVICE_NAME` - Service name (default: ai-idp)
- `SHUTDOWN_TIMEOUT` - Graceful shutdown timeout (default: 30s)