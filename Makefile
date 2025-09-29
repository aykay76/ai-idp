# AI-IDP Platform Development Makefile
# Provides convenient commands for local development workflow

.PHONY: help setup dev dev-services dev-web test test-unit test-integration lint clean migrate migrate-up migrate-down migrate-reset logs build docker-build deps tidy check

# Default target
help: ## Show this help message
	@echo "AI-IDP Platform Development Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development Environment Setup
setup: ## Initial project setup - install dependencies and initialize database
	@echo "🚀 Setting up AI-IDP development environment..."
	@go mod download
	@make deps
	@podman-compose up -d postgres redis minio
	@sleep 10  # Wait for PostgreSQL to be ready
	@make migrate-up
	@echo "✅ Setup completed! Run 'make dev' to start development."

deps: ## Install/update all dependencies
	@echo "📦 Installing dependencies..."
	@go mod tidy
	@go mod download
	@echo "📝 Note: Frontend dependencies will be installed when web directory is created"

# Development Servers
dev: ## Start all services for full development (infrastructure + services)
	@echo "🔧 Starting full development environment..."
	@podman-compose up -d postgres redis minio
	@sleep 5
	@make migrate-up
	@podman-compose --profile services up --build

dev-services: ## Start only backend services (API Gateway + microservices)
	@echo "🔧 Starting backend services..."
	@podman-compose up -d postgres redis minio
	@sleep 5
	@make migrate-up
	@podman-compose --profile services up --build api-gateway application-service team-service

dev-web: ## Start only frontend development server
	@echo "🎨 Starting frontend development server..."
	@podman-compose --profile web up web-dev

dev-infrastructure: ## Start only infrastructure services (DB, Redis, MinIO)
	@echo "🏗️  Starting infrastructure services..."
	@podman-compose up -d postgres redis minio

# Database Management
migrate: migrate-up ## Alias for migrate-up

migrate-up: ## Run database migrations (up)
	@echo "📊 Running database migrations up..."
	@podman-compose run --rm migrate

migrate-down: ## Rollback last database migration
	@echo "📊 Rolling back database migrations..."
	@podman-compose run --rm migrate -path /migrations -database "postgres://platform:platform_dev_password@postgres:5432/platform?sslmode=disable" down 1

migrate-reset: ## Reset database (drop and recreate with migrations)
	@echo "🔄 Resetting database..."
	@podman-compose down postgres
	@podman volume rm ai-idp_postgres_data || true
	@podman-compose up -d postgres
	@sleep 10
	@make migrate-up

migrate-create: ## Create new migration (usage: make migrate-create NAME=migration_name)
	@if [ -z "$(NAME)" ]; then echo "❌ Usage: make migrate-create NAME=migration_name"; exit 1; fi
	@echo "📝 Creating migration: $(NAME)"
	@timestamp=$$(date +%s); \
	touch migrations/$${timestamp}_$(NAME).up.sql migrations/$${timestamp}_$(NAME).down.sql
	@echo "✅ Created migration files:"
	@ls -la migrations/*$(NAME)*

# Testing
test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	@echo "🧪 Running unit tests..."
	@go test -v -race -short ./...

test-integration: ## Run integration tests (requires running infrastructure)
	@echo "🔗 Running integration tests..."
	@podman-compose up -d postgres redis minio
	@sleep 10
	@make migrate-up
	@go test -v -race -tags=integration ./tests/...

test-coverage: ## Run tests with coverage report
	@echo "📈 Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Code Quality
lint: ## Run linters and formatters
	@echo "🔍 Running code quality checks..."
	@go fmt ./...
	@go vet ./...
	@golangci-lint run || echo "⚠️  Install golangci-lint for complete linting"
	@echo "📝 Note: Frontend linting will be available when web directory is created"

fmt: ## Format code
	@echo "✨ Formatting code..."
	@go fmt ./...
	@goimports -w . || echo "💡 Install goimports: go install golang.org/x/tools/cmd/goimports@latest"

# Building
build: ## Build all binaries
	@echo "🔨 Building all services..."
	@mkdir -p bin
	@go build -o bin/api-gateway ./cmd/api-gateway
	@go build -o bin/application-service ./cmd/application-service  
	@go build -o bin/team-service ./cmd/team-service
	@echo "✅ Binaries built in ./bin/"

podman-build: ## Build all Podman images
	@echo "🐳 Building Podman images..."
	@podman-compose build

# Development Utilities
logs: ## Show logs from all services
	@podman-compose logs -f

logs-api: ## Show API Gateway logs
	@podman-compose logs -f api-gateway

logs-db: ## Show database logs  
	@podman-compose logs -f postgres

shell: ## Open shell in development container
	@podman-compose --profile dev run --rm dev-tools sh

db-shell: ## Open PostgreSQL shell
	@podman-compose exec postgres psql -U platform -d platform

redis-shell: ## Open Redis CLI
	@podman-compose exec redis redis-cli -a redis_dev_password

# Cleanup
clean: ## Stop all services and clean up
	@echo "🧹 Cleaning up development environment..."
	@podman-compose down
	@podman system prune -f

clean-data: ## Clean up including persistent data (⚠️  destroys all data)
	@echo "🗑️  Cleaning up data volumes (this will destroy all data)..."
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		podman-compose down -v; \
		podman volume prune -f; \
		echo "✅ Data cleaned up"; \
	else \
		echo "❌ Aborted"; \
	fi

# Status and Info
status: ## Show status of all services
	@echo "📊 Service Status:"
	@podman-compose ps

health: ## Check health of all services
	@echo "🏥 Health Check:"
	@curl -f http://localhost:8080/health 2>/dev/null && echo "✅ API Gateway: healthy" || echo "❌ API Gateway: unhealthy"
	@curl -f http://localhost:8081/health 2>/dev/null && echo "✅ Application Service: healthy" || echo "❌ Application Service: unhealthy"
	@curl -f http://localhost:8082/health 2>/dev/null && echo "✅ Team Service: healthy" || echo "❌ Team Service: unhealthy"
	@podman-compose exec -T postgres pg_isready -U platform -d platform && echo "✅ PostgreSQL: healthy" || echo "❌ PostgreSQL: unhealthy"
	@podman-compose exec -T redis redis-cli -a redis_dev_password ping && echo "✅ Redis: healthy" || echo "❌ Redis: unhealthy"

info: ## Show development environment information
	@echo "ℹ️  AI-IDP Development Environment"
	@echo "=================================="
	@echo "API Gateway:        http://localhost:8080"
	@echo "Application Service: http://localhost:8081"  
	@echo "Team Service:       http://localhost:8082"
	@echo "GitHub Provider:    http://localhost:8083"
	@echo "Frontend (dev):     http://localhost:5173"
	@echo "PostgreSQL:         localhost:5432"
	@echo "Redis:              localhost:6379"
	@echo "MinIO Console:      http://localhost:9001"
	@echo ""
	@echo "Default credentials:"
	@echo "  PostgreSQL: platform/platform_dev_password"
	@echo "  Redis: redis_dev_password"
	@echo "  MinIO: minioadmin/minioadmin123"

# Development helpers
watch: ## Watch for changes and restart services (requires entr)
	@echo "👀 Watching for changes... (requires 'entr' tool)"
	@find . -name "*.go" | entr -r make build

generate: ## Run go generate for all packages  
	@echo "⚡ Running go generate..."
	@go generate ./...

tidy: ## Tidy up go modules and format
	@echo "🧹 Tidying up..."
	@go mod tidy
	@make fmt

# Podman development workflow
podman-dev: ## Run development in Podman containers (full isolation)
	@echo "🐳 Starting containerized development..."
	@podman-compose --profile dev --profile services --profile web up --build

# Quick development commands
quick-start: ## Quick start for returning developers (assumes setup done)
	@echo "⚡ Quick starting development environment..."
	@podman-compose up -d postgres redis minio
	@sleep 3
	@make migrate-up
	@echo "✅ Infrastructure ready! Use 'make dev-services' for backend or 'make dev-web' for frontend."

# Safety checks
check: lint test-unit ## Run safety checks (lint + unit tests)
	@echo "✅ All checks passed!"

# Default development workflow documentation
workflow: ## Show recommended development workflow
	@echo "📋 Recommended Development Workflow"
	@echo "================================="
	@echo "1. Initial setup:       make setup"
	@echo "2. Start development:   make dev"
	@echo "3. Run tests:          make test"  
	@echo "4. Check code quality: make lint"
	@echo "5. Build for testing:  make build"
	@echo ""
	@echo "For returning developers:"
	@echo "1. Quick start:        make quick-start"
	@echo "2. Start services:     make dev-services"
	@echo "3. Start frontend:     make dev-web"
	@echo ""
	@echo "Useful during development:"
	@echo "- View logs:           make logs"
	@echo "- Check status:        make status"
	@echo "- Database shell:      make db-shell"
	@echo "- Reset database:      make migrate-reset"
