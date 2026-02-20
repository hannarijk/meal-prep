package integration

import (
	"testing"

	"meal-prep/services/auth/repository"
	"meal-prep/test/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

type AuthRepositoryIntegrationSuite struct {
	suite.Suite
	testDB *helpers.TestDatabase
	repo   repository.UserRepository
}

func (suite *AuthRepositoryIntegrationSuite) SetupSuite() {
	helpers.SuppressTestLogs()
	suite.testDB = helpers.SetupPostgresContainer(suite.T())
	suite.repo = repository.NewUserRepository(suite.testDB.DB)
}

func (suite *AuthRepositoryIntegrationSuite) TearDownSuite() {
	suite.testDB.Cleanup(suite.T())
	helpers.RestoreTestLogs()
}

func (suite *AuthRepositoryIntegrationSuite) SetupTest() {
	// Clean data before each test
	suite.testDB.CleanupTestData(suite.T())
}

func (suite *AuthRepositoryIntegrationSuite) TestCreate_Success_RealDatabase() {
	// This tests REAL SQL queries against REAL PostgreSQL

	email := "integration-test@example.com"
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	// Act
	user, err := suite.repo.Create(email, string(passwordHash))

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), email, user.Email)
	assert.True(suite.T(), user.ID > 0)
	assert.NotZero(suite.T(), user.CreatedAt)
	assert.NotZero(suite.T(), user.UpdatedAt)

	// Verify user was actually inserted in database
	exists, err := suite.repo.EmailExists(email)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
}

func (suite *AuthRepositoryIntegrationSuite) TestCreate_DuplicateEmail_RealConstraint() {
	email := "duplicate@example.com"
	passwordHash := "hashed_password"

	// Create first user - should succeed
	user1, err := suite.repo.Create(email, passwordHash)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user1)

	// Try to create user with same email - should fail due to UNIQUE constraint
	user2, err := suite.repo.Create(email, passwordHash+"different")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user2)
	assert.Contains(suite.T(), err.Error(), "duplicate key value") // PostgreSQL constraint error
}

func (suite *AuthRepositoryIntegrationSuite) TestEmailExists_RealQueries() {
	email := "exists-test@example.com"

	// Initially should not exist
	exists, err := suite.repo.EmailExists(email)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), exists)

	// Create user
	_, err = suite.repo.Create(email, "hashed_password")
	assert.NoError(suite.T(), err)

	// Now should exist
	exists, err = suite.repo.EmailExists(email)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
}

func (suite *AuthRepositoryIntegrationSuite) TestGetByEmail_RealData() {
	email := "getbyemail-test@example.com"
	passwordHash := "test_hash_value"

	// Create user first
	createdUser, err := suite.repo.Create(email, passwordHash)
	assert.NoError(suite.T(), err)

	// Retrieve user
	retrievedUser, hash, err := suite.repo.GetByEmail(email)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedUser)
	assert.Equal(suite.T(), createdUser.ID, retrievedUser.ID)
	assert.Equal(suite.T(), email, retrievedUser.Email)
	assert.Equal(suite.T(), passwordHash, hash)
}

func (suite *AuthRepositoryIntegrationSuite) TestGetByEmail_NotFound() {
	// Try to get non-existent user
	user, hash, err := suite.repo.GetByEmail("nonexistent@example.com")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
	assert.Empty(suite.T(), hash)
}

func (suite *AuthRepositoryIntegrationSuite) TestGetByID_RealData() {
	email := "getbyid-test@example.com"
	passwordHash := "test_hash_value"

	// Create user first
	createdUser, err := suite.repo.Create(email, passwordHash)
	assert.NoError(suite.T(), err)

	// Retrieve user by ID
	retrievedUser, err := suite.repo.GetByID(createdUser.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), retrievedUser)
	assert.Equal(suite.T(), createdUser.ID, retrievedUser.ID)
	assert.Equal(suite.T(), email, retrievedUser.Email)
	assert.NotZero(suite.T(), retrievedUser.CreatedAt)
	assert.NotZero(suite.T(), retrievedUser.UpdatedAt)
}

func (suite *AuthRepositoryIntegrationSuite) TestGetByID_NotFound() {
	// Try to get non-existent user by ID
	user, err := suite.repo.GetByID(999999)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
}

// Test SQL injection protection (real database will catch this)
func (suite *AuthRepositoryIntegrationSuite) TestEmailExists_SQLInjectionProtection() {
	maliciousEmail := "test'; DROP TABLE auth.users; --"

	exists, err := suite.repo.EmailExists(maliciousEmail)
	assert.NoError(suite.T(), err) // Should handle safely
	assert.False(suite.T(), exists)

	// Verify table still exists by creating a user
	_, err = suite.repo.Create("safe@example.com", "password")
	assert.NoError(suite.T(), err) // Table wasn't dropped
}

func TestAuthRepositoryIntegration(t *testing.T) {
	// Skip integration tests if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(AuthRepositoryIntegrationSuite))
}
