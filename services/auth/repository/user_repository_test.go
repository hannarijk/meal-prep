package repository

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"meal-prep/shared/database"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// UserRepositoryTestSuite - Using test suites for complex repository testing
type UserRepositoryTestSuite struct {
	suite.Suite
	db   *database.DB
	mock sqlmock.Sqlmock
	repo UserRepository
}

func (suite *UserRepositoryTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	require.NoError(suite.T(), err)

	suite.db = &database.DB{DB: db}
	suite.mock = mock
	suite.repo = NewUserRepository(suite.db)
}

func (suite *UserRepositoryTestSuite) TearDownTest() {
	suite.db.Close()
}

func (suite *UserRepositoryTestSuite) TestCreate_Success() {
	// Arrange
	email := "test@example.com"
	passwordHash := "hashed_password_123"
	now := time.Now()

	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO auth.users (email, password_hash, created_at, updated_at) 
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) 
		RETURNING id, email, created_at, updated_at`)).
		WithArgs(email, passwordHash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "created_at", "updated_at"}).
			AddRow(1, email, now, now))

	// Act
	user, err := suite.repo.Create(email, passwordHash)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), 1, user.ID)
	assert.Equal(suite.T(), email, user.Email)
	assert.WithinDuration(suite.T(), now, user.CreatedAt, time.Second)
	assert.WithinDuration(suite.T(), now, user.UpdatedAt, time.Second)

	// ✅ Correct way to verify SQL mock expectations
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *UserRepositoryTestSuite) TestCreate_SQLSyntaxError() {
	// Arrange
	email := "test@example.com"
	passwordHash := "hashed_password"

	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO auth.users (email, password_hash, created_at, updated_at) 
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) 
		RETURNING id, email, created_at, updated_at`)).
		WithArgs(email, passwordHash).
		WillReturnError(errors.New(`pq: syntax error at or near "RETURNING"`))

	// Act
	user, err := suite.repo.Create(email, passwordHash)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
	assert.Contains(suite.T(), err.Error(), "syntax error")

	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *UserRepositoryTestSuite) TestCreate_DuplicateKeyError() {
	// Arrange
	email := "duplicate@example.com"
	passwordHash := "hashed_password"

	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO auth.users (email, password_hash, created_at, updated_at) 
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) 
		RETURNING id, email, created_at, updated_at`)).
		WithArgs(email, passwordHash).
		WillReturnError(errors.New(`pq: duplicate key value violates unique constraint "users_email_key"`))

	// Act
	user, err := suite.repo.Create(email, passwordHash)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
	assert.Contains(suite.T(), err.Error(), "duplicate key")

	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *UserRepositoryTestSuite) TestEmailExists_True() {
	// Arrange
	email := "existing@example.com"

	suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM auth.users WHERE email = $1)")).
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Act
	exists, err := suite.repo.EmailExists(email)

	// Assert
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)

	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *UserRepositoryTestSuite) TestEmailExists_False() {
	// Arrange
	email := "new@example.com"

	suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM auth.users WHERE email = $1)")).
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Act
	exists, err := suite.repo.EmailExists(email)

	// Assert
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), exists)

	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *UserRepositoryTestSuite) TestEmailExists_DatabaseError() {
	// Arrange
	email := "test@example.com"

	suite.mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM auth.users WHERE email = $1)")).
		WithArgs(email).
		WillReturnError(errors.New("database connection lost"))

	// Act
	exists, err := suite.repo.EmailExists(email)

	// Assert
	assert.Error(suite.T(), err)
	assert.False(suite.T(), exists)
	assert.Contains(suite.T(), err.Error(), "database connection lost")

	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *UserRepositoryTestSuite) TestGetByEmail_Success() {
	// Arrange
	email := "found@example.com"
	passwordHash := "stored_hash_value"
	now := time.Now()

	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, email, password_hash, created_at, updated_at 
		FROM auth.users WHERE email = $1`)).
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(1, email, passwordHash, now, now))

	// Act
	user, hash, err := suite.repo.GetByEmail(email)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), 1, user.ID)
	assert.Equal(suite.T(), email, user.Email)
	assert.Equal(suite.T(), passwordHash, hash)
	assert.WithinDuration(suite.T(), now, user.CreatedAt, time.Second)

	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *UserRepositoryTestSuite) TestGetByEmail_NotFound() {
	// Arrange
	email := "notfound@example.com"

	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, email, password_hash, created_at, updated_at 
		FROM auth.users WHERE email = $1`)).
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)

	// Act
	user, hash, err := suite.repo.GetByEmail(email)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), sql.ErrNoRows, err)
	assert.Nil(suite.T(), user)
	assert.Empty(suite.T(), hash)

	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *UserRepositoryTestSuite) TestGetByEmail_ScanError() {
	// Arrange
	email := "test@example.com"

	// Return wrong data type for ID column (string instead of int)
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, email, password_hash, created_at, updated_at 
		FROM auth.users WHERE email = $1`)).
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow("invalid_id", email, "hash", time.Now(), time.Now()))

	// Act
	user, hash, err := suite.repo.GetByEmail(email)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
	assert.Empty(suite.T(), hash)

	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// Run the test suite
func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

// Table-driven tests for edge cases
func TestUserRepository_EdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		email           string
		expectedQueries int
		expectError     bool
		description     string
	}{
		{
			name:            "normal_email",
			email:           "test@example.com",
			expectedQueries: 1,
			expectError:     false,
			description:     "Normal email should execute one query",
		},
		{
			name:            "empty_email",
			email:           "",
			expectedQueries: 1,
			expectError:     false,
			description:     "Empty email should still execute query",
		},
		{
			name:            "sql_injection_attempt",
			email:           "'; DROP TABLE users; --",
			expectedQueries: 1,
			expectError:     false,
			description:     "SQL injection attempts should be safely parameterized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mock for each test
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			repo := NewUserRepository(&database.DB{DB: db})

			// Setup expectation
			if tt.expectedQueries > 0 {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM auth.users WHERE email = $1)")).
					WithArgs(tt.email).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
			}

			// Act
			exists, err := repo.EmailExists(tt.email)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.False(t, exists)
			}

			// ✅ Correct expectation verification
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// Benchmark repository operations
func BenchmarkUserRepository_EmailExists(b *testing.B) {
	db, mock, err := sqlmock.New()
	require.NoError(b, err)
	defer db.Close()

	repo := NewUserRepository(&database.DB{DB: db})

	// Setup mock expectation for benchmarking
	for i := 0; i < b.N; i++ {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM auth.users WHERE email = $1)")).
			WithArgs("bench@example.com").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := repo.EmailExists("bench@example.com")
		if err != nil {
			b.Fatal(err)
		}
	}
}
