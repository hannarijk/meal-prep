package helpers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"meal-prep/shared/database"
	"meal-prep/shared/logging"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// TestDatabase represents a test database instance
type TestDatabase struct {
	Container *postgres.PostgresContainer
	DB        *database.DB
	ConnStr   string
}

// SetupPostgresContainer starts a PostgreSQL container for integration tests
func SetupPostgresContainer(t *testing.T) *TestDatabase {
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	ctx := context.Background()

	// Start PostgreSQL container
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %v", err)
	}

	// Get connection string
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	// Connect to database
	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	db := &database.DB{DB: sqlDB}

	// Initialize test schema
	if err := initializeTestSchema(db); err != nil {
		t.Fatalf("Failed to initialize test schema: %v", err)
	}

	return &TestDatabase{
		Container: pgContainer,
		DB:        db,
		ConnStr:   connStr,
	}
}

// Cleanup terminates the test database
func (td *TestDatabase) Cleanup(t *testing.T) {
	ctx := context.Background()

	if td.DB != nil {
		td.DB.Close()
	}

	if td.Container != nil {
		if err := td.Container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}
}

// CleanupTestData removes all test data from database tables
func (td *TestDatabase) CleanupTestData(t *testing.T) {
	queries := []string{
		"DELETE FROM auth.users",
		"DELETE FROM recipe_catalogue.recipe_ingredients",
		"DELETE FROM recipe_catalogue.recipes",
		"DELETE FROM recipe_catalogue.ingredients",
		"DELETE FROM recipe_catalogue.categories",
		"DELETE FROM recommendations.cooking_history",
		"DELETE FROM recommendations.user_preferences",
		"DELETE FROM recommendations.recommendation_history",
	}

	for _, query := range queries {
		if _, err := td.DB.Exec(query); err != nil {
			t.Logf("Warning: Failed to cleanup table with query '%s': %v", query, err)
		}
	}
}

// initializeTestSchema creates the test database schema
func initializeTestSchema(db *database.DB) error {
	schemaSQL := `
		-- Create schemas
		CREATE SCHEMA IF NOT EXISTS auth;
		CREATE SCHEMA IF NOT EXISTS recipe_catalogue;
		CREATE SCHEMA IF NOT EXISTS recommendations;

		-- Auth tables
		CREATE TABLE IF NOT EXISTS auth.users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		-- Recipe catalogue tables
		CREATE TABLE IF NOT EXISTS recipe_catalogue.categories (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL UNIQUE,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			-- Basic constraints
    		CONSTRAINT categories_name_not_empty CHECK (length(trim(name)) > 0)
		);

		CREATE TABLE IF NOT EXISTS recipe_catalogue.recipes (
			id SERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			description TEXT,
			category_id INTEGER REFERENCES recipe_catalogue.categories(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS recipe_catalogue.ingredients (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL UNIQUE,
			description TEXT,
			category VARCHAR(50) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			-- Basic constraints
			CONSTRAINT ingredients_name_not_empty CHECK (length(trim(name)) > 0),
			CONSTRAINT ingredients_category_valid CHECK (category IN ('Meat', 'Vegetables', 'Dairy', 'Grains', 'Spices', 'Oils', 'Fish', 'Fruits'))
		);

		CREATE TABLE IF NOT EXISTS recipe_catalogue.recipe_ingredients (
			id SERIAL PRIMARY KEY,
			recipe_id INTEGER NOT NULL REFERENCES recipe_catalogue.recipes(id) ON DELETE CASCADE,
			ingredient_id INTEGER NOT NULL REFERENCES recipe_catalogue.ingredients(id) ON DELETE RESTRICT,
			quantity DECIMAL(8,2) NOT NULL,
			unit VARCHAR(20) NOT NULL,
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			-- Business constraints
			CONSTRAINT recipe_ingredients_quantity_positive CHECK (quantity > 0),
			CONSTRAINT recipe_ingredients_unit_not_empty CHECK (length(trim(unit)) > 0),
			CONSTRAINT recipe_ingredients_unique_per_recipe UNIQUE (recipe_id, ingredient_id)
		);

		-- Recommendations tables
		CREATE TABLE IF NOT EXISTS recommendations.user_preferences (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			preferred_categories INTEGER[],
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id)
		);

		CREATE TABLE IF NOT EXISTS recommendations.cooking_history (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			recipe_id INTEGER NOT NULL,
			cooked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			rating INTEGER CHECK (rating BETWEEN 1 AND 5)
		);

		CREATE TABLE IF NOT EXISTS recommendations.recommendation_history (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			recipe_id INTEGER NOT NULL,
			recommended_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			clicked BOOLEAN DEFAULT FALSE,
			algorithm_used VARCHAR(50) 
		);
	`

	_, err := db.Exec(schemaSQL)
	return err
}

// SuppressTestLogs reduces log noise during tests
func SuppressTestLogs() {
	os.Setenv("LOG_LEVEL", "error")
	logging.Init("test-service")
}

// RestoreTestLogs restores normal logging after tests
func RestoreTestLogs() {
	os.Setenv("LOG_LEVEL", "info")
}

// =============================================================================
// HTTP REQUEST HELPERS
// =============================================================================

// TestHttpClient with Kong header injection
type TestHttpClient struct {
	Client *http.Client
	userID int
	email  string
}

// NewTestHttpClient creates client that simulates Kong
func NewTestHttpClient(client *http.Client, userID int, email string) *TestHttpClient {
	return &TestHttpClient{
		Client: client,
		userID: userID,
		email:  email,
	}
}

// MakeRequest performs an HTTP request without authentication
func (c *TestHttpClient) MakeRequest(t *testing.T, method, url string, payload interface{}) *http.Response {
	t.Helper()

	var body *bytes.Buffer
	if payload != nil {
		jsonPayload, err := json.Marshal(payload)
		require.NoError(t, err, "Failed to marshal request payload")
		body = bytes.NewBuffer(jsonPayload)
	} else {
		body = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, url, body)
	require.NoError(t, err, "Failed to create HTTP request")

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.Client.Do(req)
	require.NoError(t, err, "HTTP request failed")

	return resp
}

// MakeAuthenticatedRequest simulates Kong's behavior
func (c *TestHttpClient) MakeAuthenticatedRequest(t *testing.T, method, url string, payload interface{}) *http.Response {
	t.Helper()

	var body *bytes.Buffer
	if payload != nil {
		jsonPayload, err := json.Marshal(payload)
		require.NoError(t, err, "Failed to marshal request payload")
		body = bytes.NewBuffer(jsonPayload)
	} else {
		body = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, url, body)
	require.NoError(t, err, "Failed to create HTTP request")

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	InjectKongHeaders(req, c.userID, c.email)

	resp, err := c.Client.Do(req)
	require.NoError(t, err, "HTTP request failed")

	return resp
}

// =============================================================================
// GENERIC HELPERS
// =============================================================================

func InjectKongHeaders(req *http.Request, userID int, email string) {
	token := generateTestJWT(userID, email)
	req.Header.Set("Authorization", "Bearer "+token)
}

// generateTestJWT creates a JWT with the given claims.
// The middleware uses ParseUnverified, so the signature is not validated.
func generateTestJWT(userID int, email string) string {
	type testClaims struct {
		UserID int    `json:"user_id"`
		Email  string `json:"email"`
		jwt.RegisteredClaims
	}
	claims := testClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-secret"))
	return token
}

func StringPtr(s string) *string {
	return &s
}
