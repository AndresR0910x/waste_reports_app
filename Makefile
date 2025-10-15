# Makefile for Backend Residuos App
# Usage: make [target]

# Variables
APP_NAME := backend-residuos-app
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GO_VERSION := $(shell go version | awk '{print $$3}')

# Build flags
LDFLAGS := -w -s -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)

# Directories
BUILD_DIR := build
SCRIPTS_DIR := scripts
TESTS_DIR := tests

# Colors for terminal output
RESET := \033[0m
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m

.PHONY: help build test clean deps lint format run dev docker-build docker-run docker-compose-up docker-compose-down install-tools

# Default target
all: clean deps test build

# Help target
help: ## Show this help message
	@echo "$(BLUE)Backend Residuos App - Available targets:$(RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BLUE)Environment Variables:$(RESET)"
	@echo "  APP_NAME:    $(APP_NAME)"
	@echo "  VERSION:     $(VERSION)"
	@echo "  BUILD_TIME:  $(BUILD_TIME)"
	@echo "  GIT_COMMIT:  $(GIT_COMMIT)"
	@echo "  GO_VERSION:  $(GO_VERSION)"

# Development targets
dev: ## Start development server with hot reload
	@echo "$(BLUE)[INFO]$(RESET) Starting development server..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "$(YELLOW)[WARNING]$(RESET) 'air' not found. Installing..."; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

run: ## Run the application
	@echo "$(BLUE)[INFO]$(RESET) Running application..."
	go run ./cmd/api

# Build targets
build: ## Build the application for current platform
	@echo "$(BLUE)[INFO]$(RESET) Building application..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/api
	@echo "$(GREEN)[SUCCESS]$(RESET) Built $(BUILD_DIR)/$(APP_NAME)"

build-all: ## Build for all platforms
	@echo "$(BLUE)[INFO]$(RESET) Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	
	@echo "Building for Linux AMD64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 ./cmd/api
	
	@echo "Building for Linux ARM64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 ./cmd/api
	
	@echo "Building for Darwin AMD64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 ./cmd/api
	
	@echo "Building for Darwin ARM64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 ./cmd/api
	
	@echo "Building for Windows AMD64..."
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe ./cmd/api
	
	@echo "$(GREEN)[SUCCESS]$(RESET) Built all platform binaries"

# Test targets
test: ## Run all tests
	@echo "$(BLUE)[INFO]$(RESET) Running tests..."
	go test -v ./$(TESTS_DIR)/...

test-validation: ## Run validation service tests
	@echo "$(BLUE)[INFO]$(RESET) Running validation service tests..."
	go test -v ./$(TESTS_DIR)/auth/validation_service_test.go

test-coverage: ## Run tests with coverage
	@echo "$(BLUE)[INFO]$(RESET) Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./$(TESTS_DIR)/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)[SUCCESS]$(RESET) Coverage report generated: coverage.html"

test-race: ## Run tests with race detection
	@echo "$(BLUE)[INFO]$(RESET) Running tests with race detection..."
	go test -race -v ./$(TESTS_DIR)/...

# Code quality targets
lint: ## Run linter
	@echo "$(BLUE)[INFO]$(RESET) Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)[WARNING]$(RESET) 'golangci-lint' not found. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

format: ## Format code
	@echo "$(BLUE)[INFO]$(RESET) Formatting code..."
	go fmt ./...
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	else \
		echo "$(YELLOW)[WARNING]$(RESET) 'goimports' not found. Installing..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
		goimports -w .; \
	fi

vet: ## Run go vet
	@echo "$(BLUE)[INFO]$(RESET) Running go vet..."
	go vet ./...

# Dependency targets
deps: ## Download and tidy dependencies
	@echo "$(BLUE)[INFO]$(RESET) Downloading dependencies..."
	go mod download
	go mod tidy

deps-update: ## Update dependencies
	@echo "$(BLUE)[INFO]$(RESET) Updating dependencies..."
	go get -u ./...
	go mod tidy

# Docker targets
docker-build: ## Build Docker image
	@echo "$(BLUE)[INFO]$(RESET) Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) -t $(APP_NAME):latest .

docker-run: ## Run Docker container
	@echo "$(BLUE)[INFO]$(RESET) Running Docker container..."
	docker run -p 8080:8080 --name $(APP_NAME)-container $(APP_NAME):latest

docker-compose-up: ## Start services with docker-compose
	@echo "$(BLUE)[INFO]$(RESET) Starting services with docker-compose..."
	docker-compose up -d

docker-compose-down: ## Stop services with docker-compose
	@echo "$(BLUE)[INFO]$(RESET) Stopping services with docker-compose..."
	docker-compose down

docker-compose-logs: ## View docker-compose logs
	@echo "$(BLUE)[INFO]$(RESET) Viewing docker-compose logs..."
	docker-compose logs -f

	@echo "$(BLUE)[INFO]$(RESET) Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "$(GREEN)[SUCCESS]$(RESET) Cleaned build artifacts"

install-tools: ## Install development tools
	@echo "$(BLUE)[INFO]$(RESET) Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "$(GREEN)[SUCCESS]$(RESET) Development tools installed"

swagger: ## Generate Swagger documentation
	@echo "$(BLUE)[INFO]$(RESET) Generating Swagger documentation..."
	@if command -v swag > /dev/null; then \
		swag init -g ./cmd/api/main.go; \
	else \
		echo "$(YELLOW)[WARNING]$(RESET) 'swag' not found. Installing..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
		swag init -g ./cmd/api/main.go; \
	fi

# Database targets
db-migrate: ## Run database migrations
	@echo "$(BLUE)[INFO]$(RESET) Running database migrations..."
	# Add your migration command here
	@echo "$(YELLOW)[TODO]$(RESET) Implement database migration command"

db-seed: ## Seed database with sample data
	@echo "$(BLUE)[INFO]$(RESET) Seeding database..."
	# Add your seed command here
	@echo "$(YELLOW)[TODO]$(RESET) Implement database seed command"

# CI/CD simulation targets
ci: deps lint vet test build ## Run CI pipeline locally
	@echo "$(GREEN)[SUCCESS]$(RESET) CI pipeline completed successfully"

cd: build-all docker-build ## Run CD pipeline locally
	@echo "$(GREEN)[SUCCESS]$(RESET) CD pipeline completed successfully"

# Security targets
security: ## Run security checks
	@echo "$(BLUE)[INFO]$(RESET) Running security checks..."
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "$(YELLOW)[WARNING]$(RESET) 'gosec' not found. Installing..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
		gosec ./...; \
	fi

# Information targets
info: ## Show project information
	@echo "$(BLUE)Project Information:$(RESET)"
	@echo "  Name:        $(APP_NAME)"
	@echo "  Version:     $(VERSION)"
	@echo "  Build Time:  $(BUILD_TIME)"
	@echo "  Git Commit:  $(GIT_COMMIT)"
	@echo "  Go Version:  $(GO_VERSION)"
	@echo ""
	@echo "$(BLUE)Project Structure:$(RESET)"
	@find . -type f -name "*.go" | head -10 | sed 's/^/  /'
	@echo "  ..."
	@sudo systemctl enable clean-city-backend
	@echo "✅ Service installed. Start with: sudo systemctl start clean-city-backend"

# Generate API documentation
docs:
	@echo "📚 Generating API documentation..."
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g $(MAIN_PATH) -o docs; \
	else \
		echo "swag not installed. Installing..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
		swag init -g $(MAIN_PATH) -o docs; \
	fi
	@echo "✅ API documentation generated in docs/"

# Database backup (PostgreSQL)
db-backup:
	@echo "💾 Creating database backup..."
	@if [ -f ".env" ]; then \
		export $$(cat .env | grep -v '^#' | xargs); \
		pg_dump -h $$DB_HOST -p $$DB_PORT -U $$DB_USER $$DB_NAME > backup_$$(date +%Y%m%d_%H%M%S).sql; \
		echo "✅ Database backup created"; \
	else \
		echo "❌ .env file not found"; \
	fi

# Database restore (PostgreSQL)
db-restore:
	@echo "🔄 Restoring database..."
	@read -p "Enter backup file path: " backup_file; \
	if [ -f ".env" ]; then \
		export $$(cat .env | grep -v '^#' | xargs); \
		psql -h $$DB_HOST -p $$DB_PORT -U $$DB_USER $$DB_NAME < $$backup_file; \
		echo "✅ Database restored"; \
	else \
		echo "❌ .env file not found"; \
	fi

# Health check
health:
	@echo "🏥 Checking application health..."
	@curl -f http://localhost:8080/api/v1/health || echo "❌ Application not responding"

# Full CI pipeline
ci: deps lint test build
	@echo "✅ CI pipeline completed successfully"