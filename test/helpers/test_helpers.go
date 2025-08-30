package helpers

import (
	"context"
	"database/sql"
	"github.com/testcontainers/testcontainers-go/wait"
	"meal-prep/shared/database"
	"meal-prep/shared/logging"
	"os"
	"testing"
	"time"

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
		"DELETE FROM dish_catalogue.dishes",
		"DELETE FROM dish_catalogue.categories",
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
		CREATE SCHEMA IF NOT EXISTS dish_catalogue;
		CREATE SCHEMA IF NOT EXISTS recommendations;

		-- Auth tables
		CREATE TABLE IF NOT EXISTS auth.users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		-- Dish catalogue tables
		CREATE TABLE IF NOT EXISTS dish_catalogue.categories (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL UNIQUE,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS dish_catalogue.dishes (
			id SERIAL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			description TEXT,
			category_id INTEGER REFERENCES dish_catalogue.categories(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
			dish_id INTEGER NOT NULL,
			cooked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			rating INTEGER CHECK (rating BETWEEN 1 AND 5)
		);

		CREATE TABLE IF NOT EXISTS recommendations.recommendation_history (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			dish_id INTEGER NOT NULL,
			recommended_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			clicked BOOLEAN DEFAULT FALSE,
			algorithm_used VARCHAR(50) 
		);

		-- Insert test categories
		INSERT INTO dish_catalogue.categories (name, description) VALUES 
		('Breakfast', 'Morning meals'),
		('Lunch', 'Midday meals'),
		('Dinner', 'Evening meals')
		ON CONFLICT (name) DO NOTHING;

		-- Insert test dishes
		INSERT INTO dish_catalogue.dishes (name, description, category_id) VALUES 
		('Scrambled Eggs', 'Classic breakfast eggs', 1),
		('Caesar Salad', 'Fresh romaine salad', 2),
		('Grilled Chicken', 'Juicy grilled chicken', 3)
		ON CONFLICT DO NOTHING;
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
