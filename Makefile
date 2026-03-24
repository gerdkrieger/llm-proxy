# =============================================================================
# LLM-PROXY MAKEFILE - ENTERPRISE EDITION
# =============================================================================
# Comprehensive build, test, and deployment automation
# =============================================================================

.PHONY: help setup dev dev-native dev-docker dev-docker-up dev-docker-down dev-docker-logs dev-docker-restart dev-docker-clean build test clean docker-up docker-down migrate-up migrate-down lint fmt vet security deps check install run logs

# -----------------------------------------------------------------------------
# VARIABLES
# -----------------------------------------------------------------------------

APP_NAME := llm-proxy
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GO_VERSION := $(shell go version | awk '{print $$3}')

# Build flags
LDFLAGS := -ldflags "\
	-s -w \
	-X main.Version=$(VERSION) \
	-X main.BuildTime=$(BUILD_TIME) \
	-X main.GitCommit=$(GIT_COMMIT) \
	-X main.GoVersion=$(GO_VERSION)"

# Directories
BIN_DIR := bin
DOCKER_DIR := deployments/docker
MIGRATIONS_DIR := migrations

# Database connection
DB_URL := postgres://$(shell grep DB_USER .env | cut -d '=' -f2):$(shell grep DB_PASSWORD .env | cut -d '=' -f2)@localhost:$(shell grep DB_PORT .env | cut -d '=' -f2)/$(shell grep DB_NAME .env | cut -d '=' -f2)?sslmode=disable

# Colors for output
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_BLUE := \033[34m

# -----------------------------------------------------------------------------
# HELP
# -----------------------------------------------------------------------------

help: ## Show this help message
	@echo '$(COLOR_BOLD)LLM-Proxy Makefile - Available Targets:$(COLOR_RESET)'
	@echo ''
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-20s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ''
	@echo '$(COLOR_BOLD)Examples:$(COLOR_RESET)'
	@echo '  make setup          # Initial project setup'
	@echo '  make dev-docker     # Start Docker development (RECOMMENDED)'
	@echo '  make dev            # Start native Go development'
	@echo '  make test           # Run all tests'
	@echo '  make docker-up      # Start Docker services only'
	@echo ''

# -----------------------------------------------------------------------------
# SETUP & INITIALIZATION
# -----------------------------------------------------------------------------

setup: ## Initial project setup (install deps, start docker, run migrations)
	@echo "$(COLOR_BLUE)Setting up LLM-Proxy...$(COLOR_RESET)"
	@$(MAKE) deps
	@$(MAKE) docker-up
	@echo "$(COLOR_YELLOW)Waiting for services to be ready...$(COLOR_RESET)"
	@sleep 10
	@$(MAKE) migrate-up
	@$(MAKE) check
	@echo "$(COLOR_GREEN)Setup complete! Run 'make dev' to start.$(COLOR_RESET)"

deps: ## Install Go dependencies
	@echo "$(COLOR_BLUE)Installing dependencies...$(COLOR_RESET)"
	@go mod download
	@go mod verify
	@go mod tidy
	@echo "$(COLOR_GREEN)Dependencies installed$(COLOR_RESET)"

install-tools: ## Install development tools (golangci-lint, migrate, etc.)
	@echo "$(COLOR_BLUE)Installing development tools...$(COLOR_RESET)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "$(COLOR_GREEN)Tools installed$(COLOR_RESET)"

# -----------------------------------------------------------------------------
# DEVELOPMENT
# -----------------------------------------------------------------------------

dev: ## Start development server with hot reload (native Go)
	@echo "$(COLOR_BLUE)Starting development server (native)...$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)Note: Use 'make dev-docker' for Docker-based development$(COLOR_RESET)"
	@go run cmd/server/main.go

dev-native: dev ## Alias for native Go development

dev-docker: ## Start local Docker development environment with hot-reload
	@echo "$(COLOR_BLUE)Starting Docker development environment...$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)This will start: Backend (Air), Admin-UI (Vite), PostgreSQL, Redis$(COLOR_RESET)"
	@if [ ! -f .env.local ]; then \
		echo "$(COLOR_YELLOW)Creating .env.local from .env.example...$(COLOR_RESET)"; \
		cp .env.example .env.local; \
		echo "$(COLOR_GREEN).env.local created. Please edit it with your settings.$(COLOR_RESET)"; \
	fi
	@docker compose -f docker-compose.dev.yml up --build
	@echo "$(COLOR_GREEN)Docker development environment started$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)Services:$(COLOR_RESET)"
	@echo "  - Backend API:    http://localhost:8080"
	@echo "  - Admin UI:       http://localhost:3005"
	@echo "  - PostgreSQL:     localhost:5433"
	@echo "  - Redis:          localhost:6380"
	@echo "  - Metrics:        http://localhost:9090/metrics"

dev-docker-up: ## Start Docker development environment (detached)
	@echo "$(COLOR_BLUE)Starting Docker development environment (detached)...$(COLOR_RESET)"
	@if [ ! -f .env.local ]; then \
		echo "$(COLOR_YELLOW)Creating .env.local from .env.example...$(COLOR_RESET)"; \
		cp .env.example .env.local; \
		echo "$(COLOR_GREEN).env.local created. Please edit it with your settings.$(COLOR_RESET)"; \
	fi
	@docker compose -f docker-compose.dev.yml up -d --build
	@echo "$(COLOR_GREEN)Docker development environment started$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)View logs with: make dev-docker-logs$(COLOR_RESET)"

dev-docker-down: ## Stop Docker development environment
	@echo "$(COLOR_BLUE)Stopping Docker development environment...$(COLOR_RESET)"
	@docker compose -f docker-compose.dev.yml down
	@echo "$(COLOR_GREEN)Docker development environment stopped$(COLOR_RESET)"

dev-docker-logs: ## View Docker development logs
	@docker compose -f docker-compose.dev.yml logs -f

dev-docker-restart: dev-docker-down dev-docker-up ## Restart Docker development environment

dev-docker-clean: ## Stop and remove all dev containers and volumes
	@echo "$(COLOR_YELLOW)WARNING: This will delete all development data!$(COLOR_RESET)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker compose -f docker-compose.dev.yml down -v; \
		echo "$(COLOR_GREEN)Docker development environment cleaned$(COLOR_RESET)"; \
	fi

run: ## Run the application (alias for dev)
	@$(MAKE) dev

build: ## Build production binary
	@echo "$(COLOR_BLUE)Building $(APP_NAME) $(VERSION)...$(COLOR_RESET)"
	@mkdir -p $(BIN_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) ./cmd/server
	@echo "$(COLOR_GREEN)Build complete: $(BIN_DIR)/$(APP_NAME)$(COLOR_RESET)"

build-all: ## Build for multiple platforms
	@echo "$(COLOR_BLUE)Building for multiple platforms...$(COLOR_RESET)"
	@mkdir -p $(BIN_DIR)
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-linux-amd64 ./cmd/server
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-linux-arm64 ./cmd/server
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-darwin-amd64 ./cmd/server
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)-darwin-arm64 ./cmd/server
	@echo "$(COLOR_GREEN)Multi-platform build complete$(COLOR_RESET)"

# -----------------------------------------------------------------------------
# TESTING
# -----------------------------------------------------------------------------

test: ## Run all tests
	@echo "$(COLOR_BLUE)Running tests...$(COLOR_RESET)"
	@go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@echo "$(COLOR_GREEN)Tests complete$(COLOR_RESET)"

test-unit: ## Run unit tests only
	@echo "$(COLOR_BLUE)Running unit tests...$(COLOR_RESET)"
	@go test -v -short ./...

test-integration: ## Run integration tests
	@echo "$(COLOR_BLUE)Running integration tests...$(COLOR_RESET)"
	@go test -v -run Integration ./tests/integration/...

test-coverage: ## Run tests with coverage report
	@echo "$(COLOR_BLUE)Generating coverage report...$(COLOR_RESET)"
	@go test -coverprofile=coverage.txt -covermode=atomic ./...
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "$(COLOR_GREEN)Coverage report: coverage.html$(COLOR_RESET)"

bench: ## Run benchmarks
	@echo "$(COLOR_BLUE)Running benchmarks...$(COLOR_RESET)"
	@go test -bench=. -benchmem ./...

# -----------------------------------------------------------------------------
# CODE QUALITY
# -----------------------------------------------------------------------------

lint: ## Run linter (golangci-lint)
	@echo "$(COLOR_BLUE)Running linter...$(COLOR_RESET)"
	@golangci-lint run --timeout=5m ./...

fmt: ## Format code with gofmt
	@echo "$(COLOR_BLUE)Formatting code...$(COLOR_RESET)"
	@gofmt -s -w .
	@goimports -w .
	@echo "$(COLOR_GREEN)Code formatted$(COLOR_RESET)"

vet: ## Run go vet
	@echo "$(COLOR_BLUE)Running go vet...$(COLOR_RESET)"
	@go vet ./...

security: ## Run security scan (gosec)
	@echo "$(COLOR_BLUE)Running security scan...$(COLOR_RESET)"
	@gosec -quiet ./...

check: lint vet ## Run all code quality checks

# -----------------------------------------------------------------------------
# DOCKER
# -----------------------------------------------------------------------------

docker-up: ## Start Docker services (PostgreSQL, Redis, Prometheus, Grafana)
	@echo "$(COLOR_BLUE)Starting Docker services...$(COLOR_RESET)"
	@cd $(DOCKER_DIR) && docker-compose up -d
	@echo "$(COLOR_GREEN)Docker services started$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)Services:$(COLOR_RESET)"
	@echo "  - PostgreSQL: localhost:5433"
	@echo "  - Redis:      localhost:6380"
	@echo "  - Prometheus: localhost:9090"
	@echo "  - Grafana:    localhost:3001 (admin/admin)"

docker-down: ## Stop Docker services
	@echo "$(COLOR_BLUE)Stopping Docker services...$(COLOR_RESET)"
	@cd $(DOCKER_DIR) && docker-compose down
	@echo "$(COLOR_GREEN)Docker services stopped$(COLOR_RESET)"

docker-restart: docker-down docker-up ## Restart Docker services

docker-logs: ## View Docker logs
	@cd $(DOCKER_DIR) && docker-compose logs -f

docker-clean: ## Stop services and remove volumes (WARNING: deletes data!)
	@echo "$(COLOR_YELLOW)WARNING: This will delete all data!$(COLOR_RESET)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		cd $(DOCKER_DIR) && docker-compose down -v; \
		echo "$(COLOR_GREEN)Docker services and volumes removed$(COLOR_RESET)"; \
	fi

docker-build: ## Build Docker image for LLM-Proxy
	@echo "$(COLOR_BLUE)Building Docker image...$(COLOR_RESET)"
	@docker build -t $(APP_NAME):$(VERSION) -f $(DOCKER_DIR)/Dockerfile .
	@docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest
	@echo "$(COLOR_GREEN)Docker image built: $(APP_NAME):$(VERSION)$(COLOR_RESET)"

# -----------------------------------------------------------------------------
# DATABASE MIGRATIONS
# -----------------------------------------------------------------------------

migrate-up: ## Run database migrations (up)
	@echo "$(COLOR_BLUE)Running migrations (up)...$(COLOR_RESET)"
	@migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up
	@echo "$(COLOR_GREEN)Migrations applied$(COLOR_RESET)"

migrate-down: ## Rollback last migration
	@echo "$(COLOR_YELLOW)Rolling back last migration...$(COLOR_RESET)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1; \
		echo "$(COLOR_GREEN)Migration rolled back$(COLOR_RESET)"; \
	fi

migrate-reset: ## Reset database (down all, then up)
	@echo "$(COLOR_YELLOW)WARNING: This will reset the database!$(COLOR_RESET)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down -all; \
		migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up; \
		echo "$(COLOR_GREEN)Database reset complete$(COLOR_RESET)"; \
	fi

migrate-status: ## Show migration status
	@echo "$(COLOR_BLUE)Migration status:$(COLOR_RESET)"
	@migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" version

migrate-create: ## Create new migration (usage: make migrate-create NAME=add_users)
	@if [ -z "$(NAME)" ]; then \
		echo "$(COLOR_YELLOW)Usage: make migrate-create NAME=migration_name$(COLOR_RESET)"; \
		exit 1; \
	fi
	@migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)
	@echo "$(COLOR_GREEN)Migration created$(COLOR_RESET)"

# -----------------------------------------------------------------------------
# PRODUCTION DEPLOYMENT
# -----------------------------------------------------------------------------

deploy-prod: ## Deploy production stack (full build + start)
	@echo "$(COLOR_BLUE)Deploying production stack...$(COLOR_RESET)"
	@cd $(DOCKER_DIR) && ./deploy.sh deploy
	@echo "$(COLOR_GREEN)Production deployment complete!$(COLOR_RESET)"

deploy-update: ## Update production deployment (rebuild + restart)
	@echo "$(COLOR_BLUE)Updating production deployment...$(COLOR_RESET)"
	@cd $(DOCKER_DIR) && ./deploy.sh update
	@echo "$(COLOR_GREEN)Update complete!$(COLOR_RESET)"

deploy-start: ## Start production services
	@echo "$(COLOR_BLUE)Starting production services...$(COLOR_RESET)"
	@cd $(DOCKER_DIR) && ./deploy.sh start

deploy-stop: ## Stop production services
	@echo "$(COLOR_BLUE)Stopping production services...$(COLOR_RESET)"
	@cd $(DOCKER_DIR) && ./deploy.sh stop

deploy-status: ## Show production service status
	@cd $(DOCKER_DIR) && ./deploy.sh status

deploy-logs: ## Show production logs
	@cd $(DOCKER_DIR) && ./deploy.sh logs

deploy-health: ## Check production service health
	@cd $(DOCKER_DIR) && ./deploy.sh health

deploy-backup: ## Backup production data
	@cd $(DOCKER_DIR) && ./deploy.sh backup

deploy-clean: ## Clean production deployment (remove all)
	@cd $(DOCKER_DIR) && ./deploy.sh clean

build-prod-images: ## Build production Docker images
	@echo "$(COLOR_BLUE)Building production images...$(COLOR_RESET)"
	@cd $(DOCKER_DIR) && docker compose -f docker-compose.prod.yml build
	@echo "$(COLOR_GREEN)Images built successfully$(COLOR_RESET)"

# -----------------------------------------------------------------------------
# REMOTE DEPLOYMENT (Database Migrations & Filters)
# -----------------------------------------------------------------------------

deploy-migrations: ## Deploy database migrations to production server
	@echo "$(COLOR_BLUE)Deploying migrations to production...$(COLOR_RESET)"
	@./scripts/deployment/deploy-migrations.sh
	@echo "$(COLOR_GREEN)Migrations deployed successfully$(COLOR_RESET)"

deploy-filters: ## Deploy content filters to production server
	@echo "$(COLOR_BLUE)Deploying filters to production...$(COLOR_RESET)"
	@./scripts/deployment/deploy-filters.sh
	@echo "$(COLOR_GREEN)Filters deployed successfully$(COLOR_RESET)"

deploy: ## Deploy to production (build images, transfer via SSH, restart)
	@echo "$(COLOR_BLUE)Starting deployment...$(COLOR_RESET)"
	@./scripts/deployment/deploy.sh
	@echo "$(COLOR_GREEN)Deployment complete$(COLOR_RESET)"

deploy-auto: ## Run auto-deployment check (detects pending migrations/filters)
	@./scripts/deployment/auto-deploy-migrations.sh

setup-hooks: ## Install Git hooks for auto-deployment
	@echo "$(COLOR_BLUE)Setting up Git hooks...$(COLOR_RESET)"
	@./scripts/setup-git-hooks.sh

# -----------------------------------------------------------------------------
# UTILITIES
# -----------------------------------------------------------------------------

logs: ## Show application logs (requires running app)
	@tail -f logs/*.log

clean: ## Clean build artifacts and temporary files
	@echo "$(COLOR_BLUE)Cleaning...$(COLOR_RESET)"
	@rm -rf $(BIN_DIR)
	@rm -rf coverage.txt coverage.html
	@rm -rf tmp/
	@go clean -cache -testcache -modcache
	@echo "$(COLOR_GREEN)Cleaned$(COLOR_RESET)"

version: ## Show version information
	@echo "$(COLOR_BOLD)LLM-Proxy$(COLOR_RESET)"
	@echo "  Version:     $(VERSION)"
	@echo "  Git Commit:  $(GIT_COMMIT)"
	@echo "  Build Time:  $(BUILD_TIME)"
	@echo "  Go Version:  $(GO_VERSION)"

health-check: ## Check if services are healthy
	@echo "$(COLOR_BLUE)Checking service health...$(COLOR_RESET)"
	@echo -n "PostgreSQL: "
	@docker exec llm-proxy-postgres pg_isready -U proxy_user > /dev/null 2>&1 && echo "$(COLOR_GREEN)✓$(COLOR_RESET)" || echo "$(COLOR_YELLOW)✗$(COLOR_RESET)"
	@echo -n "Redis:      "
	@docker exec llm-proxy-redis redis-cli ping > /dev/null 2>&1 && echo "$(COLOR_GREEN)✓$(COLOR_RESET)" || echo "$(COLOR_YELLOW)✗$(COLOR_RESET)"
	@echo -n "LLM-Proxy:  "
	@curl -s http://localhost:8080/health > /dev/null 2>&1 && echo "$(COLOR_GREEN)✓$(COLOR_RESET)" || echo "$(COLOR_YELLOW)✗$(COLOR_RESET)"

# -----------------------------------------------------------------------------
# DOCUMENTATION
# -----------------------------------------------------------------------------

docs: ## Generate API documentation (Swagger)
	@echo "$(COLOR_BLUE)Generating API documentation...$(COLOR_RESET)"
	@swag init -g cmd/server/main.go -o api/swagger
	@echo "$(COLOR_GREEN)Documentation generated$(COLOR_RESET)"

# -----------------------------------------------------------------------------
# DEFAULT TARGET
# -----------------------------------------------------------------------------

.DEFAULT_GOAL := help
