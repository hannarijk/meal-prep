.PHONY: build up down logs test clean help

# Build all services
build:
	docker-compose build

# Start all services
up:
	docker-compose up -d

# Start with tools (pgAdmin)
up-tools:
	docker-compose --profile tools up -d

# Stop all services
down:
	docker-compose down

# Stop and remove volumes
down-clean:
	docker-compose down -v

# View logs
logs:
	docker-compose logs -f

# View specific service logs
logs-auth:
	docker-compose logs -f auth-service

logs-recipes:
	docker-compose logs -f recipe-catalogue-service

logs-recommendations:
	docker-compose logs -f recommendations-service

# Run all unit tests (fast, no external dependencies)
test-unit:
	@echo "🧪 Running unit tests..."
	go test ./services/.../service/... -v -race -short

# Run integration tests (real database)
test-integration:
	@echo "🔗 Running integration tests with real database..."
	go test ./test/integration/... -v -timeout=300s

# Run end-to-end tests (complete workflows)
test-e2e:
	@echo "🌐 Running end-to-end tests..."
	go test ./test/e2e/... -v -timeout=300s

# Run all tests
test-all: test-unit test-integration test-e2e
	@echo "✅ All test types completed successfully"

# Quick test (unit tests only)
test:
	@echo "⚡ Running quick unit tests..."
	go test ./services/... -short -race

# Test with coverage
test-coverage:
	@echo "📊 Running tests with coverage..."
	go test ./services/... -v -race -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Test specific service
test-service:
	@echo "🎯 Testing specific service..."
	@read -p "Enter service name (auth/recipe-catalogue/recommendations): " service; \
	go test ./services/$$service/... -v -race

# Clean test artifacts
clean-tests:
	rm -f coverage.out coverage.html
	go clean -testcache
	docker system prune -f --filter label=testcontainers

# Clean up everything
clean:
	docker-compose down -v
	docker system prune -f

# Show help
help:
	@echo "Available commands:"
	@echo "  build             - Build all Docker images"
	@echo "  up                - Start all services"
	@echo "  up-tools          - Start all services including pgAdmin"
	@echo "  down              - Stop all services"
	@echo "  down-clean        - Stop services and remove volumes"
	@echo "  logs              - View all logs"
	@echo "  logs-*            - View specific service logs"
	@echo "  clean             - Clean up everything"
	@echo "Test commands:"
	@echo "  test"             - Quick unit tests"
	@echo "  test-unit"        - All unit tests"
	@echo "  test-integration" - Integration tests with real DB"
	@echo "  test-e2e"         - End-to-end API tests"
	@echo "  test-all"         - All test types"
	@echo "  test-coverage"    - Tests with coverage report"
	@echo "  clean-tests"      - Clean test artifacts"