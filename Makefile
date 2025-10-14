# Clean City Backend - Makefile
.PHONY: help build run test clean setup migrate-up migrate-down migrate-status docker-build docker-run

# Default target
help:
	@echo "Available targets:"
	@echo "  setup        - Run initial project setup"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application in development mode"
	@echo "  test         - Run all tests"
	@echo "  test-cover   - Run tests with coverage"
	@echo "  clean        - Clean build artifacts"
	@echo "  migrate-up   - Run database migrations up"
	@echo "  migrate-down - Run database migrations down"
	@echo "  migrate-status - Check migration status"
	@echo "  lint         - Run code linting"
	@echo "  format       - Format code"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker"

# Variables
APP_NAME=clean-city-backend
BUILD_DIR=bin
MAIN_PATH=cmd/api/main.go
MIGRATE_PATH=cmd/migrate/main.go

# Setup project
setup:
	@echo "🚀 Setting up Clean City Backend..."
	@if [ -f "scripts/setup.ps1" ]; then \
		powershell -ExecutionPolicy Bypass -File scripts/setup.ps1; \
	else \
		echo "PowerShell setup script not found. Running manual setup..."; \
		$(MAKE) deps; \
		cp .env.example .env || true; \
		echo "✅ Basic setup complete. Please configure .env and Firebase credentials."; \
	fi

# Build application
build:
	@echo "🔨 Building application..."
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@go build -ldflags="-s -w" -o $(BUILD_DIR)/migrate $(MIGRATE_PATH)
	@echo "✅ Build complete: $(BUILD_DIR)/$(APP_NAME)"

# Run application in development
run:
	@echo "🚀 Running application..."
	@go run $(MAIN_PATH)

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test -v ./...

# Run tests with coverage
test-cover:
	@echo "🧪 Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@go clean -cache
	@echo "✅ Clean complete"

# Database migrations
migrate-up:
	@echo "⬆️  Running database migrations up..."
	@go run $(MIGRATE_PATH) -action=up

migrate-down:
	@echo "⬇️  Running database migrations down..."
	@go run $(MIGRATE_PATH) -action=down

migrate-status:
	@echo "📊 Checking migration status..."
	@go run $(MIGRATE_PATH) -action=status

migrate-reset:
	@echo "🔄 Resetting all migrations..."
	@go run $(MIGRATE_PATH) -action=reset

# Code quality
lint:
	@echo "🔍 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

format:
	@echo "✨ Formatting code..."
	@go fmt ./...
	@go mod tidy

# Dependencies
deps:
	@echo "📦 Downloading dependencies..."
	@go mod download
	@go mod tidy
	@go mod verify

# Security scan
security:
	@echo "🔒 Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Installing..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
		gosec ./...; \
	fi

# Docker targets
docker-build:
	@echo "🐳 Building Docker image..."
	@docker build -t $(APP_NAME):latest .

docker-run:
	@echo "🐳 Running with Docker..."
	@docker-compose up -d

docker-stop:
	@echo "🛑 Stopping Docker containers..."
	@docker-compose down

# Development helpers
dev-setup: deps
	@echo "🛠️  Setting up development environment..."
	@go install github.com/air-verse/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@echo "✅ Development tools installed"

dev-run:
	@echo "🔄 Running with hot reload..."
	@air

# Production build
prod-build:
	@echo "🏭 Building for production..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags="-s -w -X main.version=$$(git describe --tags --always)" \
		-o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(MAIN_PATH)
	@echo "✅ Production build complete: $(BUILD_DIR)/$(APP_NAME)-linux-amd64"

# Install as service (Linux systemd)
install-service:
	@echo "📋 Installing systemd service..."
	@sudo cp scripts/clean-city-backend.service /etc/systemd/system/
	@sudo systemctl daemon-reload
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