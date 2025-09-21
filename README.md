# AI-Powered Internal Developer Platform

A next-generation Internal Developer Platform that combines AI-driven infrastructure automation with developer-centric workflows.

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** - Backend services
- **Node.js 18+** - Frontend tooling  
- **Docker & Docker Compose** - Local development
- **PostgreSQL client tools** - Database access (optional)

### Initial Setup

```bash
# Clone and setup the project
git clone <repository-url>
cd ai-idp

# Initial setup (installs dependencies, starts infrastructure, runs migrations)
make setup

# Start all services for development
make dev
```

After setup completes, the platform will be available at:
- **API Gateway**: http://localhost:8080
- **Application Service**: http://localhost:8081  
- **PostgreSQL**: localhost:5432 (platform/platform_dev_password)
- **Redis**: localhost:6379 (redis_dev_password)
- **MinIO Console**: http://localhost:9001 (minioadmin/minioadmin123)

### Development Workflow

```bash
# Quick start for returning developers
make quick-start

# Start specific components
make dev-services    # Backend services only
make dev-web        # Frontend development server

# Run tests
make test           # All tests
make test-unit      # Unit tests only
make test-integration # Integration tests (requires DB)

# Code quality
make lint           # Run linters
make fmt            # Format code

# Database operations
make migrate        # Run migrations
make migrate-reset  # Reset database (⚠️ destroys data)
make db-shell       # Open PostgreSQL shell

# Utilities
make logs           # View all service logs
make status         # Check service status
make health         # Health check all services
make clean          # Stop and clean up
```

## 🏗️ Architecture Overview

### Platform Components

```
┌─────────────────────────────────────────────────────────────┐
│                        AI Layer                             │
│     (Natural Language → Infrastructure Definitions)        │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                    Platform Layer                          │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────┐ │
│  │ API Gateway │ │   Web UI    │ │    Control Plane       │ │
│  │   (8080)    │ │   (5173)    │ │   (Reconciliation)     │ │
│  └─────────────┘ └─────────────┘ └─────────────────────────┘ │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────┐ │
│  │Application  │ │   Team      │ │   Resource Providers   │ │
│  │Service(8081)│ │Service(8082)│ │  (GitHub, K8s, etc.)  │ │
│  └─────────────┘ └─────────────┘ └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                    Data Layer                              │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────┐ │
│  │ PostgreSQL  │ │    Redis    │ │   Tenant Databases     │ │
│  │ (Platform)  │ │   (Cache)   │ │    (Isolated)          │ │
│  └─────────────┘ └─────────────┘ └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### Native Go HTTP Architecture

- **Zero Framework Dependencies**: Uses only Go standard library for HTTP
- **Service Template Pattern**: Reusable templates for rapid service development
- **Tenant-Aware Middleware**: Built-in tenant isolation and validation
- **Standard CRUD Operations**: Consistent patterns across all resource types
- **Pluggable Resource Providers**: Easy to add new infrastructure integrations

### Multi-Tenant Architecture

- **Platform Database**: Single PostgreSQL database with component-specific schemas
- **Tenant Databases**: Isolated database per tenant for strong data separation
- **API Gateway**: Central routing with tenant-aware authentication
- **Resource Providers**: Pluggable components for different infrastructure types

## 📋 API Usage

### Creating an Application

```bash
# Create a new application
curl -X POST http://localhost:8081/api/v1/applications \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001" \
  -H "X-User-Email: developer@company.com" \
  -d '{
    "name": "my-awesome-app",
    "display_name": "My Awesome Application", 
    "description": "A revolutionary new application",
    "team_name": "platform-team",
    "owner_email": "developer@company.com",
    "lifecycle": "development"
  }'
```

### Listing Applications

```bash
# List all applications for a tenant
curl -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001" \
  http://localhost:8081/api/v1/applications

# Filter by team
curl -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001" \
  "http://localhost:8081/api/v1/applications?team_name=platform-team"
```

### Health Checks

```bash
# Check service health
curl http://localhost:8081/health
curl http://localhost:8081/readiness

# API Gateway health  
curl http://localhost:8080/health
```

## 🧪 Testing

### Running Tests

```bash
# Unit tests (fast, no external dependencies)
make test-unit
go test -short ./...

# Integration tests (requires running PostgreSQL)
make test-integration  
go test -tags=integration ./tests/...

# All tests
make test
```

### Test Database

Tests use a separate test database configuration:
- **Test DB URL**: `TEST_DATABASE_URL` environment variable
- **Default**: Uses the same PostgreSQL instance as development
- **Isolation**: Each integration test can use isolated databases if needed

### Writing Tests

```go
func TestMyService_CreateResource(t *testing.T) {
    // Skip in short mode for integration tests
    testutils.SkipIfShort(t)
    
    // Setup test database
    ctx := context.Background()
    pool, cleanup := testutils.SetupTestDB(t, ctx)
    defer cleanup()
    
    // Your test logic here...
}
```

## 📁 Project Structure

```
ai-idp/
├── cmd/                      # Main applications
│   ├── api-gateway/         # API Gateway service
│   ├── application-service/ # Application CRUD service  
│   ├── team-service/        # Team CRUD service
│   └── github-provider/     # GitHub integration
├── internal/                # Private application code
│   ├── database/           # Database utilities & migrations
│   ├── server/             # HTTP server framework  
│   ├── auth/               # Authentication middleware
│   ├── applications/       # Application service logic
│   └── testutils/          # Testing utilities
├── pkg/                    # Public libraries
│   ├── api/               # Generated API types
│   └── schemas/           # YAML schemas
├── web/                   # Frontend code
│   ├── components/        # Web Components
│   └── design-system/     # CSS design tokens
├── migrations/            # Database migrations
├── scripts/              # Development scripts  
├── docs/                 # Documentation
├── docker-compose.yml    # Local development environment
├── Makefile             # Development commands
└── go.mod               # Go dependencies
```

## 🔧 Configuration

### Environment Variables

**Application Service:**
- `PORT` - Service port (default: 8081)
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string  
- `ENVIRONMENT` - Environment name (development, staging, production)
- `DEBUG` - Enable debug logging (true/false)

**Database:**
- `TEST_DATABASE_URL` - Test database connection (for tests)

### Development vs Production

The platform automatically adapts based on the `ENVIRONMENT` variable:
- **development**: Debug mode, CORS enabled, verbose logging
- **production**: Release mode, security hardened, structured logging

## 🚦 Getting Started - Development Scenarios

### Scenario 1: Backend Developer

```bash
# Start infrastructure only
make dev-infrastructure

# Run application service locally
cd cmd/application-service
go run main.go

# Test the service
curl http://localhost:8081/health
```

### Scenario 2: Frontend Developer

```bash
# Start backend services
make dev-services

# Start frontend development (in separate terminal)
make dev-web

# Frontend will be at http://localhost:5173
```

### Scenario 3: Full Stack Development

```bash
# Start everything
make dev

# All services running:
# - API Gateway: http://localhost:8080
# - Services: http://localhost:8081-8083  
# - Frontend: http://localhost:5173
```

### Scenario 4: Database Development

```bash
# Start database only
make dev-infrastructure

# Open database shell
make db-shell

# Create new migration
make migrate-create NAME=add_new_feature

# Reset database for clean state
make migrate-reset
```

## 🤝 Contributing

1. **Setup**: Run `make setup` for initial environment
2. **Code**: Follow standard Go conventions and project structure
3. **Test**: Write tests for new features (`make test`)
4. **Lint**: Ensure code quality (`make lint`)  
5. **Document**: Update README and inline docs as needed

### Adding a New Service (Using Service Template)

The platform provides a reusable service template that makes creating new services incredibly easy:

**1. Create the main service file:**
```go
// cmd/team-service/main.go
package main

import (
    "context"
    "log"
    "github.com/aykay76/ai-idp/internal/server"
    "github.com/aykay76/ai-idp/internal/teams"
)

func main() {
    config := server.LoadConfigFromEnv("team-service", "8082")
    template := server.NewResourceTemplate("team-service", "teams", config)
    
    ctx := context.Background()
    if err := template.SetupDatabase(ctx); err != nil {
        log.Fatalf("Failed to setup database: %v", err)
    }
    
    teamService := teams.NewService(template.GetDB())
    teamHandlers := teams.NewTeamHandlers(teamService)
    
    // CRUD routes registered automatically!
    template.RegisterResourceHandlers(teamHandlers.GetCRUDHandlers())
    
    if err := template.Run(ctx); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}
```

**2. Implement your service logic:**
```go
// internal/teams/service.go - Copy application service pattern
// internal/teams/handlers.go - Copy application handlers pattern
```

**3. Add to infrastructure:**
- Add service to `docker-compose.yml`
- Update `Makefile` targets
- The template handles everything else automatically!

**Benefits of the Template Approach:**
- ✅ **5 minutes to working service** - Copy, adapt, run
- ✅ **Consistent patterns** - Same structure across all services
- ✅ **Built-in tenant isolation** - Middleware handles multi-tenancy
- ✅ **Standard endpoints** - Health, readiness, metrics included
- ✅ **CRUD operations** - Standard REST patterns pre-built
- ✅ **Error handling** - Consistent error responses
- ✅ **Validation** - Request validation built-in

### Database Changes

1. Create migration: `make migrate-create NAME=description`
2. Edit `.up.sql` and `.down.sql` files in `migrations/`
3. Test migration: `make migrate-reset && make migrate`
4. Update relevant Go structs and queries

## 📚 Next Steps

- [ ] **Add Authentication**: Implement JWT-based auth with configurable providers
- [ ] **API Gateway**: Create central routing and rate limiting
- [ ] **Team Service**: Implement team management CRUD operations  
- [ ] **GitHub Provider**: Add Git repository integration
- [ ] **Frontend UI**: Build Web Components-based interface
- [ ] **Control Plane**: Implement reconciliation loops
- [ ] **AI Integration**: Connect to existing AI layer
- [ ] **Multi-Cloud**: Add Azure/AWS resource providers

## 🆘 Troubleshooting

### Database Connection Issues

```bash
# Check if PostgreSQL is running
make status

# Reset database if corrupted
make migrate-reset

# Check database logs
make logs-db
```

### Service Start Issues

```bash
# Check service health
make health

# View service logs
make logs
make logs-api  # API Gateway specific

# Restart services
make clean && make dev
```

### Test Failures

```bash
# Run tests with verbose output
go test -v ./...

# Run specific test
go test -run TestApplicationService_CreateApplication ./internal/applications/

# Skip integration tests
go test -short ./...
```

---

For more detailed information, see the `/docs` directory or run `make help` for all available commands.
