.PHONY: help build build-prod run dev clean test test-coverage test-integration docker-up docker-down docker-logs docker-clean docker-rebuild docs lint lint-fix tidy migrate-up migrate-down migrate-create install-tools security-scan run-local run-dev run-uat run-prod database-up database-down first-setup check-port

# Default target
.DEFAULT_GOAL := help

# ============================================================================
BINARY_NAME=server
MAIN_FILE=cmd/api/main.go
DOCKER_COMPOSE_FILE=docker-compose.yml

# Environment-specific config files
CONFIG_LOCAL=config/config.local.yaml
CONFIG_DEV=config/config.dev.yaml
CONFIG_UAT=config/config.uat.yaml
CONFIG_PROD=config/config.prod.yaml

# Default environment (can be overridden: make run ENV=DEV)
ENV ?= DEV

# DB Config for migration
DB_DSN ?= "postgres://$${DB_USER}:$${DB_PASSWORD}@$${DB_HOST}:$${DB_PORT}/$${DB_NAME}?sslmode=$${DB_SSL_MODE}"

# Build flags for production
BUILD_FLAGS=-ldflags="-s -w -X main.Version=$$(git describe --tags --always --dirty 2>/dev/null || echo dev) -X main.BuildTime=$$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# ============================================================================
help: ## Show this help message (all available commands)
	@echo "╔════════════════════════════════════════════════════════════════╗"
	@echo "║          GO-COMMUNITY - Makefile Commands                      ║"
	@echo "╚════════════════════════════════════════════════════════════════╝"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2} /^##@/ {printf "\n\033[1m%s\033[0m\n", substr($$0, 5)}' $(MAKEFILE_LIST)
	@echo ""

# ============================================================================
##@ First Time Setup

first-setup: ## Run initial setup (install tools, start docker DB, run migrations)
	@echo "� Starting first-time setup..."
	@$(MAKE) install-tools
	@echo "🐳 Starting Docker containers (database)..."
	@docker compose up postgres -d
	@echo "⏳ Waiting for database to be ready (5 seconds)..."
	@sleep 5
	@echo "📊 Running database migrations... (Assuming DEV credentials)"
	@DB_USER=postgres DB_PASSWORD=postgres DB_HOST=localhost DB_PORT=5888 DB_NAME=community_db DB_SSL_MODE=disable $(MAKE) migrate-up
	@echo "✅ First-time setup complete! You can now run 'make dev' to start the server."

# ============================================================================
##@ Build & Development

build: ## Build the application binary into bin/
	@echo "🔨 Building $$(BINARY_NAME)..."
	@mkdir -p bin
	@go build -o bin/$(BINARY_NAME) $(MAIN_FILE)
	@echo "✅ Build complete: bin/$(BINARY_NAME)"

build-prod: ## Build optimized production binary into bin/
	@echo "🔨 Building production binary with optimizations..."
	@mkdir -p bin
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/$(BINARY_NAME) $(MAIN_FILE)
	@echo "✅ Production build complete: bin/$(BINARY_NAME)"
	@ls -lh bin/$(BINARY_NAME)

dev:
	@echo "Create or Update swagger documentation"
	@$(MAKE) docs
	@echo "🔥 Starting hot-reload development server..."
	@if ! command -v air > /dev/null; then \
		echo "⚠️  'air' not found. Run 'make install-tools'"; \
		exit 1; \
	fi
	@ENV=${ENV} air

# ============================================================================
##@ Execution Settings

run: docs tidy ## Run the application (ENV can be overridden, e.g. make run ENV=DEV)
	@echo "🚀 Running application with ENV=$(ENV)..."
	@export ENV="$(ENV)" && go run $(MAIN_FILE) | jq -R '. as $$line | try (fromjson) catch $$line'

run-local: ## Run application with LOCAL config
	@$(MAKE) run ENV=local

run-dev: ## Run application with DEV config
	@$(MAKE) run ENV=DEV

run-uat: ## Run application with UAT config
	@$(MAKE) run ENV=uat

run-prod: ## Run application with PROD config
	@$(MAKE) run ENV=prod

# ============================================================================
##@ Testing & Quality

test: ## Run all unit tests
	@echo "🧪 Running tests..."
	@go test -v -race ./...

test-coverage: ## Run tests and generate HTML coverage report
	@echo "🧪 Running tests with coverage..."
	@mkdir -p coverage
	@go test -v -race -coverprofile=coverage/coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "✅ Coverage report generated: coverage/coverage.html"
	@go tool cover -func=coverage/coverage.out | grep total

test-integration: ## Run integration tests
	@echo "🧪 Running integration tests..."
	@go test -v -race -tags=integration ./tests/...

lint: ## Run linter (golangci-lint)
	@echo "🔍 Running linter..."
	@if ! command -v golangci-lint > /dev/null; then \
		echo "⚠️  golangci-lint not installed. Run 'make install-tools'"; \
		exit 1; \
	fi
	@golangci-lint run ./...

lint-fix: ## Run linter and auto-fix simple issues
	@echo "🔧 Running linter with auto-fix..."
	@golangci-lint run --fix ./...

security-scan: ## Run security vulnerability scan (gosec)
	@echo "🔒 Running security scan..."
	@if ! command -v gosec > /dev/null; then \
		echo "⚠️  Installing gosec..."; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	fi
	@gosec -fmt=json -out=security-report.json ./... || true
	@gosec ./...

# ============================================================================
##@ Documentation & Tools

docs: ## Generate Swagger API documentation
	@echo "📚 Generating Swagger documentation..."
	@if ! command -v swag > /dev/null; then \
		echo "⚠️  swag not installed. Installing..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@swag init -g $(MAIN_FILE)
	@echo "✅ Swagger docs generated"

install-tools: ## Install required development tools (air, swag, migrate, etc)
	@echo "🔧 Installing development tools..."
	@echo "Installing air (hot-reload)..."
	@go install github.com/air-verse/air@latest
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Installing swag (Swagger)..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Installing golang-migrate (migrations)..."
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "Installing gosec (security)..."
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@echo "✅ All tools installed successfully"

# ============================================================================
##@ Docker & Containerization

docker-up: ## Start services with Docker Compose
	@echo "🐳 Starting containers..."
	@docker compose up -d
	@echo "✅ Containers started"
	@docker compose ps

docker-down: ## Stop Docker Compose services
	@echo "🐳 Stopping containers..."
	@docker compose down
	@echo "✅ Containers stopped"

docker-logs: ## View container logs (tail 100)
	@docker compose logs -f --tail=100

docker-clean: ## Remove containers, volumes, and images
	@echo "🧹 Cleaning Docker resources..."
	@docker compose down -v --remove-orphans
	@docker system prune -f
	@echo "✅ Docker cleanup complete"

docker-rebuild: ## Rebuild docker compose and force recreate app container
	@echo "🐳 Rebuilding and starting app container..."
	@docker compose build
	@docker compose up app --force-recreate -d
	@echo "✅ App container rebuilt and started"

# ============================================================================
##@ Database & Migrations

database-up: ## Create the PostgreSQL community_db inside container
	@echo "🗄️  Creating database..."
	@docker exec -i community createdb --username=postgres --owner=postgres community_db
	@echo "✅ Database created"

database-down: ## Drop the PostgreSQL community_db inside container
	@echo "🗄️  Dropping database..."
	@docker exec -i community dropdb --username=postgres community_db
	@echo "✅ Database dropped"

migrate-up: ## Run pending database migrations
	@echo "📊 Running migrations..."
	@if ! command -v migrate > /dev/null; then \
		echo "⚠️  golang-migrate (migrate) not installed. Run 'make install-tools'"; \
		exit 1; \
	fi
	@migrate -path tests/integration/db/migrations/ -database $(DB_DSN) up
	@echo "✅ Migrations applied"

migrate-down: ## Rollback the last database migration
	@echo "📊 Rolling back migrations..."
	@if ! command -v migrate > /dev/null; then \
		echo "⚠️  golang-migrate (migrate) not installed. Run 'make install-tools'"; \
		exit 1; \
	fi
	@migrate -path tests/integration/db/migrations/ -database $(DB_DSN) down
	@echo "✅ Migration rolled back"

migrate-create: ## Create new migration files (Usage: make migrate-create NAME=xyz)
	@if [ -z "$(NAME)" ]; then \
		echo "❌ ERROR: Please specify NAME=<migration_name>"; \
		echo "Example: make migrate-create NAME=create_users_table"; \
		exit 1; \
	fi
	@if ! command -v migrate > /dev/null; then \
		echo "⚠️  golang-migrate (migrate) not installed. Run 'make install-tools'"; \
		exit 1; \
	fi
	@migrate create -ext sql -dir tests/integration/db/migrations/ -seq $(NAME)
	@echo "✅ Migration created in tests/integration/db/migrations/"

# ============================================================================
##@ Cleanup & Maintenance

clean: ## Remove generated binaries and test artifacts
	@echo "🧹 Cleaning up..."
	@rm -rf bin/
	@rm -rf coverage/
	@rm -f security-report.json
	@echo "✅ Cleanup complete"

tidy: ## Output and download Go modules cleanly
	@echo "📦 Tidying modules..."
	@go mod tidy
	@go mod verify
	@echo "✅ Modules tidied"

check-port: ## Check if anything is running on port 8080
	@lsof -i :8080 || true