.PHONY: build up down clean down-clean logs logs-auth logs-recipes logs-recommendations \
		test-unit test-integration test-e2e test-all test test-coverage test-service clean-tests \
		migrate-build postgres-up migrate-up migrate-up-% migrate-status deploy-fresh deploy-update \
        migrate-auth migrate-recipe migrate-recommendations migrate-all \
        flyway-auth-info flyway-recipe-info flyway-recommendations-info \
        flyway-auth-repair flyway-recipe-repair flyway-recommendations-repair \
        flyway-auth-clean flyway-recipe-clean flyway-recommendations-clean \
        help

# Build all services
build:
	docker-compose build

# Start all services
up:
	docker-compose up -d

# Stop all services
down:
	docker-compose down

# Stop and remove volumes
down-clean:
	docker-compose down -v

# Stop and remove volumes, including orphan volumes (aka clean up everything)
clean:
	docker-compose down -v
	docker system prune -f

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
	@echo "Running unit tests..."
	#go test ./services/.../service/... -v -race -short
	go test ./services/... -v -race -short

# Run integration tests (real database)
test-integration:
	@echo "Running integration tests with real database..."
	go test ./test/integration/... -v -timeout=300s

# Run end-to-end tests (complete workflows)
test-e2e:
	@echo "Running end-to-end tests..."
	go test ./test/e2e/... -v -timeout=300s

# Run all tests
test-all: test-unit test-integration test-e2e
	@echo "All test types completed successfully"

# Quick test (unit tests only)
test:
	@echo "Running quick unit tests..."
	go test ./services/... -short -race

# Test with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test ./services/... -v -race -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Test specific service
test-service:
	@echo "Testing specific service..."
	@read -p "Enter service name (auth/recipe-catalogue/recommendations): " service; \
	go test ./services/$$service/... -v -race

# Clean test artifacts
clean-tests:
	rm -f coverage.out coverage.html
	go clean -testcache
	docker system prune -f --filter label=testcontainers

# Show help
help:
	@echo "Available commands:"
	@echo "  build            			  - Build all Docker images"
	@echo "  up               			  - Start all services"
	@echo "  down             			  - Stop all services"
	@echo "  down-clean       			  - Stop services and remove volumes"
	@echo "  clean            			  - Stop and remove volumes, including orphan volumes (aka clean up everything)"
	@echo "  logs             			  - View all logs"
	@echo "  logs-*           			  - View specific service logs"
	@echo "Test commands:"
	@echo "  test             			  - Quick unit tests"
	@echo "  test-unit        			  - All unit tests"
	@echo "  test-integration 			  - Integration tests with real DB"
	@echo "  test-e2e         			  - End-to-end API tests"
	@echo "  test-all         			  - All test types"
	@echo "  test-coverage    			  - Tests with coverage report"
	@echo "  clean-tests      			  - Clean test artifacts"
	@echo "Migration commands:"
	@echo "  migrate-all      			  - Run all Flyway migrations (all schemas)"
	@echo "  migrate-auth|recipe|recommendations"
	@echo "  flyway-*-info|repair|clean   - Inspect/repair/clean schemas (clean is destructive!)"

# -------------------------------
# Flyway (DB migrations)
# -------------------------------

# Run migrations per schema (one-shot containers)
# run --rm starts a container for this service, executes Flyway, and removes the container immediately after.
migrate-auth:
	docker-compose run --rm flyway-auth

migrate-recipe-catalogue:
	docker-compose run --rm flyway-recipe-catalogue

migrate-recommendations:
	docker-compose run --rm flyway-recommendations

migrate-all: migrate-auth migrate-recipe-catalogue migrate-recommendations

# Info / Repair / Clean (per schema)
flyway-auth-info:
	docker-compose run --rm flyway-auth info
flyway-recipe-catalogue-info:
	docker-compose run --rm flyway-recipe-catalogue info
flyway-recommendations-info:
	docker-compose run --rm flyway-recommendations info

flyway-auth-repair:
	docker-compose run --rm flyway-auth repair
flyway-recipe-catalogue-repair:
	docker-compose run --rm flyway-recipe-catalogue repair
flyway-recommendations-repair:
	docker-compose run --rm flyway-recommendations repair

# DANGEROUS: drops all objects in the schema
flyway-auth-clean:
	docker-compose run --rm flyway-auth clean
flyway-recipe-catalogue-clean:
	docker-compose run --rm flyway-recipe-catalogue clean
flyway-recommendations-clean:
	docker-compose run --rm flyway-recommendations clean