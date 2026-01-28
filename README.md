# The Meal-Prep Project

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A production-ready microservices system for weekly meal planning and recipe recommendations.

## Architecture overview

```mermaid
graph TB
    subgraph "Client Applications"
        WEB[Web App]
        MOBILE[Mobile App]
    end

    subgraph "API Gateway / Load Balancer"
        LB[Load Balancer]
    end

    subgraph "Microservices"
        AUTH[Auth Service<br/>:8001]
        RECIPES[Recipe Catalogue<br/>:8002] 
        REC[Recommendations<br/>:8003]
    end

    subgraph "Data Layer"
        DB[(PostgreSQL<br/>Database)]
    end

    WEB --> LB
    MOBILE --> LB
    LB --> AUTH
    LB --> RECIPES
    LB --> REC
    
    AUTH --> DB
    RECIPES --> DB
    REC --> DB

    REC -.->|JWT Validation| AUTH
    RECIPES -.->|JWT Validation| AUTH
```

### Service Communication

- **Authentication Flow**: JWT tokens issued by Auth service, validated by other services
- **Data Independence**: Each service manages its own database schema
- **Loose Coupling**: Services communicate via REST APIs with standardized error responses

## Quick Start

### Prerequisites

- [Go 1.23+](https://golang.org/dl/)
- [Docker & Docker Compose](https://docs.docker.com/get-docker/)
- Make (optional, for convenience commands)

### Installation

```bash
# Clone the repository
git clone https://github.com/hannarijk/meal-prep.git
cd meal-prep

# Start with Docker (recommended)
make up

# Or start individual services for development
go run ./services/auth &
go run ./services/recipe-catalogue &
go run ./services/recommendations &

# Stop with Docker

# Option 1: Stop all services
make down

# Option 2: Stop all services and clean up containers and volumes
make clean
```

### Verify Installation

```bash
# Check all services are healthy
curl http://localhost:8001/health  # Auth Service
curl http://localhost:8002/health  # Recipe Catalogue  
curl http://localhost:8003/health  # Recommendations

# Expected response: {"status": "healthy", "service": "service-name"}
```

## API Documentation

### Authentication Service (Port 8001)

| Endpoint | Method | Description | Auth Required |
|----------|--------|-------------|---------------|
| `/health` | GET | Health check | No |
| `/register` | POST | User registration | No |
| `/login` | POST | User authentication | No |

#### Registration

```bash
curl -X POST http://localhost:8001/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123"
  }'
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "created_at": "2025-08-28T12:00:00Z"
  }
}
```

### Recipe Catalogue Service (Port 8002)

#### Recipe Management

| Endpoint | Method | Description         | Auth Required |
|----------|--------|---------------------|---------------|
| `/health` | GET | Health check        | No |
| `/recipes` | GET | List all recipes    | No |
| `/recipes` | POST | Create new recipe   | **Yes** |
| `/recipes/{id}` | GET | Get specific recipe | No |
| `/recipes/{id}` | PUT | Update recipe       | **Yes** |
| `/recipes/{id}` | DELETE | Delete recipe       | **Yes** |
| `/categories` | GET | List categories     | No |
| `/categories/{id}/recipes` | GET | Recipes by category | No |

#### Ingredient Management

| Endpoint | Method | Description | Auth Required |
|----------|--------|-------------|---------------|
| `/ingredients` | GET | List all ingredients | No |
| `/ingredients` | POST | Create new ingredient | **Yes** |
| `/ingredients/{id}` | GET | Get specific ingredient | No |
| `/ingredients/{id}` | PUT | Update ingredient | **Yes** |
| `/ingredients/{id}` | DELETE | Delete ingredient | **Yes** |
| `/ingredients/{id}/recipes` | GET | Recipes using ingredient | No |

#### Recipe-Ingredient Relationships

| Endpoint | Method | Description | Auth Required |
|----------|--------|-------------|---------------|
| `/recipes/{id}/ingredients` | GET | Get recipe ingredients | No |
| `/recipes/{id}/ingredients` | POST | Add ingredient to recipe | **Yes** |
| `/recipes/{id}/ingredients` | PUT | Set all recipe ingredients | **Yes** |
| `/recipes/{recipeId}/ingredients/{ingredientId}` | PUT | Update recipe ingredient | **Yes** |
| `/recipes/{recipeId}/ingredients/{ingredientId}` | DELETE | Remove ingredient from recipe | **Yes** |

#### Grocery Lists

| Endpoint        | Method | Description                        | Auth Required |
|-----------------|--------|------------------------------------|---------------|
| `/grocery-list` | POST | Generate grocery list from recipes | **Yes** |

#### Create Recipe (Protected)

```bash
curl -X POST http://localhost:8002/recipes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "Avocado Toast",
    "description": "Healthy breakfast with fresh avocado",
    "category_id": 1
  }'
```

#### Create Ingredient

```bash
curl -X POST http://localhost:8002/ingredients \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "Avocado",
    "description": "Fresh Hass avocado",
    "category": "Produce"
  }'
```

#### Search Ingredients

```bash
# Search by name/description
curl "http://localhost:8002/ingredients?search=avocado"

# Filter by category
curl "http://localhost:8002/ingredients?category=Produce"
```

#### Add Ingredient to Recipe

```bash
curl -X POST http://localhost:8002/recipes/1/ingredients \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "ingredient_id": 5,
    "quantity": 2,
    "unit": "pieces",
    "notes": "Ripe and soft"
  }'
```

#### Generate Grocery List

```bash
curl -X POST http://localhost:8002/grocery-list \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "recipe_ids": [1, 3, 7, 12]
  }'
```

**Response:**
```json
[
  {
    "ingredient_id": 5,
    "ingredient": {
      "id": 5,
      "name": "Avocado",
      "category": "Produce"
    },
    "total_quantity": 4,
    "unit": "pieces",
    "recipes": ["Avocado Toast", "Guacamole", "Green Smoothie"]
  }
]
```

### Recommendations Service (Port 8003)

| Endpoint | Method | Description | Auth Required |
|----------|--------|-------------|---------------|
| `/health` | GET | Health check | No |
| `/recommendations` | GET | Get personalized recommendations | **Yes** |
| `/preferences` | GET | Get user preferences | **Yes** |
| `/preferences` | PUT | Update user preferences | **Yes** |
| `/cooking` | POST | Log cooking activity | **Yes** |
| `/cooking/history` | GET | Get cooking history | **Yes** |

#### Get Recommendations

```bash
# Get hybrid recommendations (default)
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  "http://localhost:8003/recommendations?limit=5"

# Get time-decay based recommendations  
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  "http://localhost:8003/recommendations?algorithm=time_decay&limit=10"

# Get preference-based recommendations
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  "http://localhost:8003/recommendations?algorithm=preference&limit=8"
```

#### Set Food Preferences

```bash
curl -X PUT http://localhost:8003/preferences \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "preferred_categories": [1, 3, 5]
  }'
```

#### Log Cooking Activity

```bash
curl -X POST http://localhost:8003/cooking \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "recipe_id": 5,
    "rating": 4
  }'
```

## Recommendation Algorithms

### Time Decay Algorithm
Prioritizes recipes based on how long it's been since you last cooked them:

- **Recently cooked** (< 7 days): Low priority (0.1)
- **Good to revisit** (7-30 days): Medium priority (0.7)
- **Perfect time** (30-90 days): High priority (1.0)
- **Long time favorites** (90+ days): Highest priority (1.2)
- **Never cooked**: Medium priority (0.5)

### Hybrid Algorithm (Default)
Combines time decay and user preferences:
- **Time Decay Weight**: 60%
- **Preference Weight**: 40%
- **Final Score**: `(time_score × 0.6) + (preference_score × 0.4)`

### Preference Algorithm
Pure preference-based recommendations using your selected categories.

## Development

### Project Structure

```
meal-prep/
├── services/                 # Microservices
│   ├── auth/                # Authentication service
│   ├── recipe-catalogue/      # Recipe management
│   └── recommendations/     # AI recommendations
├── shared/                  # Shared libraries
│   ├── database/           # Database connections
│   ├── middleware/         # HTTP middleware
│   ├── models/            # Data models
│   ├── utils/             # Utilities (JWT, etc.)
│   └── logging/           # Structured logging
├── test/                   # Integration & E2E tests
├── scripts/               # Database scripts
└── deployments/          # Kubernetes manifests
```

### Local Development

```bash
# Start dependencies only
docker compose up postgres -d

# Run services individually for development
go run ./services/auth
go run ./services/recipe-catalogue  
go run ./services/recommendations
```

### Environment Variables

Create a `.env` file in the project root:

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres123
DB_NAME=mealprep

# JWT Configuration  
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Service Ports
AUTH_PORT=8001
RECIPE_CATALOGUE_PORT=8002
RECOMMENDATIONS_PORT=8003

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

### Building

```bash
# Build all services
make build

# Build specific service
go build -o bin/auth ./services/auth
go build -o bin/recipe-catalogue ./services/recipe-catalogue
go build -o bin/recommendations ./services/recommendations
```

## Testing

The project follows Go testing best practices with three test layers:

### Unit Tests (Business Logic)
```bash
# Run all unit tests
make test-unit

# Test specific service
go test ./services/auth/... -v

# With coverage
make test-coverage
```

### Integration Tests (Real Database)
```bash
# Run integration tests with testcontainers
make test-integration

# Skip integration tests (for CI speed)
go test ./... -short
```

### End-to-End Tests (Complete Workflows)
```bash
# Run full API workflow tests
make test-e2e

# Run all test types
make test-all
```

### Test Coverage

```bash
# Generate coverage report
make test-coverage
open coverage.html
```

## Deployment
```bash
# Production deployment
docker compose -f docker-compose.yml up -d
```

## Monitoring

### Structured Logging

All services output structured JSON logs suitable for log aggregation:

```json
{
  "timestamp": "2025-08-28T12:00:00Z",
  "level": "info", 
  "message": "Request completed",
  "service": "auth-service",
  "request_id": "abc-123-def",
  "user_id": 42,
  "method": "POST",
  "uri": "/login",
  "status_code": 200,
  "duration_ms": 45
}
```

### Log Levels

- `DEBUG`: Detailed information for debugging
- `INFO`: General operational messages
- `WARN`: Warning conditions
- `ERROR`: Error conditions requiring attention

Set via `LOG_LEVEL` environment variable.
```bash
  LOG_LEVEL=debug make up
```

## Contributing

### Code Standards

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use [gofmt](https://golang.org/cmd/gofmt/) for formatting
- Run [golint](https://github.com/golang/lint) and [go vet](https://golang.org/cmd/vet/)
- Write tests for all new features
- Update documentation for API changes

### Development Workflow

1. **Fork** the repository
2. **Create** a feature branch: `git checkout -b feature/amazing-feature`
3. **Write** tests for your changes
4. **Ensure** tests pass: `make test-all`
5. **Commit** with conventional commits: `feat: add amazing feature`
6. **Push** to your branch: `git push origin feature/amazing-feature`
7. **Open** a Pull Request

### Running Tests

```bash
# Before committing - run all tests
make test-all

# Quick validation
make test-unit

# Full integration testing
make test-integration
```

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make build` | Build Docker images |
| `make up` | Start all services |
| `make down` | Stop all services |
| `make logs` | View all service logs |
| `make test-unit` | Run unit tests |
| `make test-integration` | Run integration tests |
| `make test-e2e` | Run end-to-end tests |
| `make test-all` | Run all test types |
| `make clean` | Clean up containers and volumes |

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DB_HOST` | PostgreSQL host | `localhost` | Yes |
| `DB_PORT` | PostgreSQL port | `5432` | No |
| `DB_USER` | Database username | `postgres` | Yes |
| `DB_PASSWORD` | Database password | - | Yes |
| `DB_NAME` | Database name | `mealprep` | Yes |
| `JWT_SECRET` | JWT signing secret | - | Yes |
| `LOG_LEVEL` | Logging level | `info` | No |
| `LOG_FORMAT` | Log format (`json`/`text`) | `json` | No |

### Service Ports

| Service          | Port | Description |
|------------------|------|-------------|
| Auth             | 8001 | User authentication |
| Recipe Catalogue | 8002 | Recipe management |
| Recommendations  | 8003 | AI recommendations |
| PostgreSQL       | 5432 | Database |

## Examples

### Complete User Journey
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJlbWFpbCI6InVzZXJAZXhhbXBsZS5jb20iLCJleHAiOjE3NTY3NDA5NzEsIm5iZiI6MTc1NjY1NDU3MSwiaWF0IjoxNzU2NjU0NTcxfQ.U4brcR8gdfHXNjTehsmDLUKtEOQXws3c9df67BYfCh8
```bash
# 1. Register new user
RESPONSE=$(curl -s -X POST http://localhost:8001/register \
  -H "Content-Type: application/json" \
  -d '{"email":"chef@example.com","password":"mycookingpassword"}')

# Extract token
TOKEN=$(echo $RESPONSE | jq -r '.token')

# 2. Browse available recipes
curl -s http://localhost:8002/recipes | jq '.[:3]'

# 3. Set food preferences
curl -X PUT http://localhost:8003/preferences \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"preferred_categories": [1, 3]}'

# 4. Get personalized recommendations  
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8003/recommendations?limit=5" | jq

# 5. Log cooking activity
curl -X POST http://localhost:8003/cooking \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"recipe_id": 2, "rating": 5}'

# 6. Get updated recommendations (now considers cooking history)
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8003/recommendations?algorithm=time_decay&limit=5" | jq
```

### Creating Custom Recipes

```bash
# Create breakfast recipe
curl -X POST http://localhost:8002/recipes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Protein Pancakes",
    "description": "High-protein pancakes with oats and eggs",
    "category_id": 1
  }'
```

### Error Testing
```bash
# Test without authentication
curl http://localhost:8003/recommendations
````

```bash
# Test invalid algorithm
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8003/recommendations?algorithm=invalid"
````
Expected Result: Success (200) with hybrid recommendations:
```json
{
  "recipes": [...],
  "algorithm": "hybrid",
  "generated_at": "2025-08-28T...",
  "total_scored": 5
}
```
Why: The service has a validateAlgorithm() method that defaults to "hybrid" for invalid algorithms instead of throwing an error. This is actually good UX - graceful degradation.

```bash
# Test invalid recipe ID
curl -X POST http://localhost:8003/cooking \
-H "Content-Type: application/json" \
-H "Authorization: Bearer $TOKEN" \
-d '{"recipe_id": 999, "rating": 5}'
```
Expected Result: Status 400 Bad Request:
```json
{
  "error": "recommendations_error",
  "code": 400,
  "message": "recipe not found"
}
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built with [Go](https://golang.org/) and industry-standard libraries
- Database powered by [PostgreSQL](https://www.postgresql.org/)
- Containerized with [Docker](https://www.docker.com/)
- Testing with [Testify](https://github.com/stretchr/testify) and [Testcontainers](https://testcontainers.com/)
- Recommendation algorithms inspired by Netflix and Spotify approaches

## Support

For questions, issues, or contributions, please:

1. Check existing [Issues](https://github.com/hannarijk/meal-prep/issues)
2. Open a new issue with detailed description
3. Join discussions in [Discussions](https://github.com/hannarijk/meal-prep/discussions)

**Built with ❤️ for developers who love Go and great food!**