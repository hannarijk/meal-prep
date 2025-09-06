package repository

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"meal-prep/shared/database"
)

// CategoryRepositoryTestSuite tests all category repository data access operations
type CategoryRepositoryTestSuite struct {
	suite.Suite
	db   *database.DB
	mock sqlmock.Sqlmock
	repo CategoryRepository
}

func (suite *CategoryRepositoryTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	require.NoError(suite.T(), err)

	suite.db = &database.DB{DB: db}
	suite.mock = mock
	suite.repo = NewCategoryRepository(suite.db)
}

func (suite *CategoryRepositoryTestSuite) TearDownTest() {
	suite.db.Close()
}

// =============================================================================
// GET ALL CATEGORIES TESTS
// =============================================================================

func (suite *CategoryRepositoryTestSuite) TestGetAll_ReturnsCategories() {
	// Arrange
	now := time.Now()
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, description, created_at, updated_at FROM recipe_catalogue.categories ORDER BY name`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}).
			AddRow(1, "Italian", "Italian cuisine", now, now).
			AddRow(2, "Mexican", "Mexican cuisine", now, now).
			AddRow(3, "Vegetarian", nil, now, now))

	// Act
	categories, err := suite.repo.GetAll()

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), categories, 3)

	// Verify first category
	assert.Equal(suite.T(), 1, categories[0].ID)
	assert.Equal(suite.T(), "Italian", categories[0].Name)
	assert.NotNil(suite.T(), categories[0].Description)
	assert.Equal(suite.T(), "Italian cuisine", *categories[0].Description)

	// Verify category with nil description
	assert.Equal(suite.T(), 3, categories[2].ID)
	assert.Equal(suite.T(), "Vegetarian", categories[2].Name)
	assert.Nil(suite.T(), categories[2].Description)

	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *CategoryRepositoryTestSuite) TestGetAll_EmptyResult() {
	// Arrange
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, description, created_at, updated_at FROM recipe_catalogue.categories ORDER BY name`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}))

	// Act
	categories, err := suite.repo.GetAll()

	// Assert
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), categories)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *CategoryRepositoryTestSuite) TestGetAll_DatabaseError() {
	// Arrange
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, description, created_at, updated_at FROM recipe_catalogue.categories ORDER BY name`)).
		WillReturnError(errors.New("connection lost"))

	// Act
	categories, err := suite.repo.GetAll()

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), categories)
	assert.Contains(suite.T(), err.Error(), "connection lost")
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *CategoryRepositoryTestSuite) TestGetAll_ScanError() {
	// Arrange
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, description, created_at, updated_at FROM recipe_catalogue.categories ORDER BY name`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}).
			AddRow("invalid_id", "Italian", "Italian cuisine", time.Now(), time.Now())) // Invalid ID type

	// Act
	categories, err := suite.repo.GetAll()

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), categories)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// =============================================================================
// GET BY ID TESTS
// =============================================================================

func (suite *CategoryRepositoryTestSuite) TestGetByID_ReturnsCategory() {
	// Arrange
	now := time.Now()
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, description, created_at, updated_at FROM recipe_catalogue.categories WHERE id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}).
			AddRow(1, "Italian", "Italian cuisine", now, now))

	// Act
	category, err := suite.repo.GetByID(1)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), category)
	assert.Equal(suite.T(), 1, category.ID)
	assert.Equal(suite.T(), "Italian", category.Name)
	assert.NotNil(suite.T(), category.Description)
	assert.Equal(suite.T(), "Italian cuisine", *category.Description)
	assert.WithinDuration(suite.T(), now, category.CreatedAt, time.Second)
	assert.WithinDuration(suite.T(), now, category.UpdatedAt, time.Second)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *CategoryRepositoryTestSuite) TestGetByID_WithNullDescription() {
	// Arrange
	now := time.Now()
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, description, created_at, updated_at FROM recipe_catalogue.categories WHERE id = $1`)).
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}).
			AddRow(2, "Vegetarian", nil, now, now))

	// Act
	category, err := suite.repo.GetByID(2)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), category)
	assert.Equal(suite.T(), 2, category.ID)
	assert.Equal(suite.T(), "Vegetarian", category.Name)
	assert.Nil(suite.T(), category.Description)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *CategoryRepositoryTestSuite) TestGetByID_NotFound() {
	// Arrange
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, description, created_at, updated_at FROM recipe_catalogue.categories WHERE id = $1`)).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	// Act
	category, err := suite.repo.GetByID(999)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), sql.ErrNoRows, err)
	assert.Nil(suite.T(), category)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *CategoryRepositoryTestSuite) TestGetByID_DatabaseError() {
	// Arrange
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, description, created_at, updated_at FROM recipe_catalogue.categories WHERE id = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("database timeout"))

	// Act
	category, err := suite.repo.GetByID(1)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), category)
	assert.Contains(suite.T(), err.Error(), "database timeout")
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *CategoryRepositoryTestSuite) TestGetByID_ScanError() {
	// Arrange
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, description, created_at, updated_at FROM recipe_catalogue.categories WHERE id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "created_at", "updated_at"}).
			AddRow("invalid_id", "Italian", "Italian cuisine", time.Now(), time.Now())) // Invalid ID type

	// Act
	category, err := suite.repo.GetByID(1)

	// Assert
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), category)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// =============================================================================
// EXISTS TESTS
// =============================================================================

func (suite *CategoryRepositoryTestSuite) TestExists_True() {
	// Arrange
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT EXISTS(SELECT 1 FROM recipe_catalogue.categories WHERE id = $1)`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Act
	exists, err := suite.repo.Exists(1)

	// Assert
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *CategoryRepositoryTestSuite) TestExists_False() {
	// Arrange
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT EXISTS(SELECT 1 FROM recipe_catalogue.categories WHERE id = $1)`)).
		WithArgs(999).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Act
	exists, err := suite.repo.Exists(999)

	// Assert
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), exists)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *CategoryRepositoryTestSuite) TestExists_DatabaseError() {
	// Arrange
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT EXISTS(SELECT 1 FROM recipe_catalogue.categories WHERE id = $1)`)).
		WithArgs(1).
		WillReturnError(errors.New("connection failed"))

	// Act
	exists, err := suite.repo.Exists(1)

	// Assert
	assert.Error(suite.T(), err)
	assert.False(suite.T(), exists)
	assert.Contains(suite.T(), err.Error(), "connection failed")
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *CategoryRepositoryTestSuite) TestExists_ScanError() {
	// Arrange
	suite.mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT EXISTS(SELECT 1 FROM recipe_catalogue.categories WHERE id = $1)`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow("invalid_boolean"))

	// Act
	exists, err := suite.repo.Exists(1)

	// Assert
	assert.Error(suite.T(), err)
	assert.False(suite.T(), exists)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

// =============================================================================
// RUN TEST SUITE
// =============================================================================

func TestCategoryRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(CategoryRepositoryTestSuite))
}

// =============================================================================
// TABLE-DRIVEN TESTS FOR EDGE CASES
// =============================================================================

func TestCategoryRepository_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		categoryID  int
		description string
	}{
		{
			name:        "valid_positive_id",
			categoryID:  1,
			description: "Normal positive ID should work",
		},
		{
			name:        "zero_id",
			categoryID:  0,
			description: "Zero ID should be handled (though may not exist)",
		},
		{
			name:        "negative_id",
			categoryID:  -1,
			description: "Negative ID should be handled gracefully",
		},
		{
			name:        "large_id",
			categoryID:  999999,
			description: "Large ID should be handled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh mock for each test
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			repo := NewCategoryRepository(&database.DB{DB: db})

			// Setup expectation for Exists call
			mock.ExpectQuery(regexp.QuoteMeta(`
				SELECT EXISTS(SELECT 1 FROM recipe_catalogue.categories WHERE id = $1)`)).
				WithArgs(tt.categoryID).
				WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

			// Act
			exists, err := repo.Exists(tt.categoryID)

			// Assert
			assert.NoError(t, err, tt.description)
			assert.False(t, exists, tt.description)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
