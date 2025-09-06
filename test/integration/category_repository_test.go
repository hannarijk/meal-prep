package integration

import (
	"database/sql"
	"testing"

	"meal-prep/services/recipe-catalogue/repository"
	"meal-prep/shared/models"
	"meal-prep/test/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// CategoryRepositoryIntegrationSuite tests category repository against real PostgreSQL database
// Uses testcontainers to spin up isolated database for each test run
// Tests cover CRUD operations, constraint validation, and edge cases
type CategoryRepositoryIntegrationSuite struct {
	suite.Suite
	testDB       *helpers.TestDatabase
	categoryRepo repository.CategoryRepository
	recipeRepo   repository.RecipeRepository // Needed for foreign key constraint testing
}

func (suite *CategoryRepositoryIntegrationSuite) SetupSuite() {
	helpers.SuppressTestLogs()
	suite.testDB = helpers.SetupPostgresContainer(suite.T())

	suite.categoryRepo = repository.NewCategoryRepository(suite.testDB.DB)
	suite.recipeRepo = repository.NewRecipeRepository(suite.testDB.DB)
}

func (suite *CategoryRepositoryIntegrationSuite) TearDownSuite() {
	suite.testDB.Cleanup(suite.T())
	helpers.RestoreTestLogs()
}

func (suite *CategoryRepositoryIntegrationSuite) SetupTest() {
	suite.testDB.CleanupTestData(suite.T())
}

// =============================================================================
// BASIC CRUD OPERATIONS
// =============================================================================

func (suite *CategoryRepositoryIntegrationSuite) TestGetAll_EmptyDatabase() {
	categories, err := suite.categoryRepo.GetAll()

	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), categories)
}

func (suite *CategoryRepositoryIntegrationSuite) TestGetAll_WithCategories() {
	// Create test categories
	category1 := models.CreateCategoryRequest{
		Name:        "Italian",
		Description: helpers.StringPtr("Italian cuisine"),
	}
	_, err := suite.categoryRepo.Create(category1)
	require.NoError(suite.T(), err)

	category2 := models.CreateCategoryRequest{
		Name:        "Asian",
		Description: helpers.StringPtr("Asian cuisine"),
	}
	_, err = suite.categoryRepo.Create(category2)
	require.NoError(suite.T(), err)

	categories, err := suite.categoryRepo.GetAll()

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), categories, 2)

	// Verify categories are sorted by name (Asian comes before Italian)
	assert.Equal(suite.T(), "Asian", categories[0].Name)
	assert.Equal(suite.T(), "Italian", categories[1].Name)

	// Verify category details
	assert.NotNil(suite.T(), categories[0].Description)
	assert.Equal(suite.T(), "Asian cuisine", *categories[0].Description)
	assert.NotNil(suite.T(), categories[1].Description)
	assert.Equal(suite.T(), "Italian cuisine", *categories[1].Description)
}

func (suite *CategoryRepositoryIntegrationSuite) TestGetByID_ExistingCategory() {
	// Create test category
	category := models.CreateCategoryRequest{
		Name:        "Mediterranean",
		Description: helpers.StringPtr("Mediterranean cuisine"),
	}
	created, err := suite.categoryRepo.Create(category)
	require.NoError(suite.T(), err)

	found, err := suite.categoryRepo.GetByID(created.ID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), found)
	assert.Equal(suite.T(), created.ID, found.ID)
	assert.Equal(suite.T(), "Mediterranean", found.Name)
	assert.NotNil(suite.T(), found.Description)
	assert.Equal(suite.T(), "Mediterranean cuisine", *found.Description)
	assert.NotZero(suite.T(), found.CreatedAt)
	assert.NotZero(suite.T(), found.UpdatedAt)
}

func (suite *CategoryRepositoryIntegrationSuite) TestGetByID_NonExistentCategory() {
	found, err := suite.categoryRepo.GetByID(99999)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), found)
	assert.Equal(suite.T(), sql.ErrNoRows, err)
}

func (suite *CategoryRepositoryIntegrationSuite) TestExists_ExistingCategory() {
	// Create test category
	category := models.CreateCategoryRequest{
		Name:        "French",
		Description: helpers.StringPtr("French cuisine"),
	}
	created, err := suite.categoryRepo.Create(category)
	require.NoError(suite.T(), err)

	exists, err := suite.categoryRepo.Exists(created.ID)

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
}

func (suite *CategoryRepositoryIntegrationSuite) TestExists_NonExistentCategory() {
	exists, err := suite.categoryRepo.Exists(99999)

	assert.NoError(suite.T(), err)
	assert.False(suite.T(), exists)
}

// =============================================================================
// CREATE OPERATIONS
// =============================================================================

func (suite *CategoryRepositoryIntegrationSuite) TestCreate_Success() {
	req := models.CreateCategoryRequest{
		Name:        "Spanish",
		Description: helpers.StringPtr("Spanish cuisine"),
	}

	category, err := suite.categoryRepo.Create(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), category)
	assert.True(suite.T(), category.ID > 0)
	assert.Equal(suite.T(), "Spanish", category.Name)
	assert.NotNil(suite.T(), category.Description)
	assert.Equal(suite.T(), "Spanish cuisine", *category.Description)
	assert.NotZero(suite.T(), category.CreatedAt)
	assert.NotZero(suite.T(), category.UpdatedAt)

	// Verify category was actually created in database
	found, err := suite.categoryRepo.GetByID(category.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), category.ID, found.ID)
	assert.Equal(suite.T(), category.Name, found.Name)

	// Verify exists check works
	exists, err := suite.categoryRepo.Exists(category.ID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
}

func (suite *CategoryRepositoryIntegrationSuite) TestCreate_NilDescription() {
	req := models.CreateCategoryRequest{
		Name:        "Minimal Category",
		Description: nil, // No description
	}

	category, err := suite.categoryRepo.Create(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), category)
	assert.Equal(suite.T(), "Minimal Category", category.Name)
	assert.Nil(suite.T(), category.Description)

	// Verify in database
	found, err := suite.categoryRepo.GetByID(category.ID)
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), found.Description)
}

func (suite *CategoryRepositoryIntegrationSuite) TestCreate_EmptyDescription() {
	req := models.CreateCategoryRequest{
		Name:        "Empty Desc Category",
		Description: helpers.StringPtr(""), // Empty string description
	}

	category, err := suite.categoryRepo.Create(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), category)
	assert.Equal(suite.T(), "Empty Desc Category", category.Name)
	assert.NotNil(suite.T(), category.Description)
	assert.Equal(suite.T(), "", *category.Description)
}

func (suite *CategoryRepositoryIntegrationSuite) TestCreate_DuplicateName_ShouldFail() {
	// Create first category
	req1 := models.CreateCategoryRequest{
		Name:        "Duplicate Name",
		Description: helpers.StringPtr("First category"),
	}
	category1, err := suite.categoryRepo.Create(req1)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), category1)

	// Try to create second category with same name
	req2 := models.CreateCategoryRequest{
		Name:        "Duplicate Name", // Same name
		Description: helpers.StringPtr("Second category"),
	}
	category2, err := suite.categoryRepo.Create(req2)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), category2)
	assert.Contains(suite.T(), err.Error(), "duplicate key value") // PostgreSQL unique constraint error
}

// =============================================================================
// UPDATE OPERATIONS
// =============================================================================

func (suite *CategoryRepositoryIntegrationSuite) TestUpdate_Success() {
	// Create initial category
	createReq := models.CreateCategoryRequest{
		Name:        "Original Name",
		Description: helpers.StringPtr("Original description"),
	}
	created, err := suite.categoryRepo.Create(createReq)
	require.NoError(suite.T(), err)

	// Update category
	updateReq := models.UpdateCategoryRequest{
		Name:        helpers.StringPtr("Updated Name"),
		Description: helpers.StringPtr("Updated description"),
	}

	updated, err := suite.categoryRepo.Update(created.ID, updateReq)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updated)
	assert.Equal(suite.T(), created.ID, updated.ID)
	assert.Equal(suite.T(), "Updated Name", updated.Name)
	assert.NotNil(suite.T(), updated.Description)
	assert.Equal(suite.T(), "Updated description", *updated.Description)
	assert.Equal(suite.T(), created.CreatedAt, updated.CreatedAt) // CreatedAt should not change
	assert.True(suite.T(), updated.UpdatedAt.After(created.UpdatedAt))

	// Verify in database
	found, err := suite.categoryRepo.GetByID(created.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Name", found.Name)
	assert.Equal(suite.T(), "Updated description", *found.Description)
}

func (suite *CategoryRepositoryIntegrationSuite) TestUpdate_PartialUpdate_NameOnly() {
	// Create initial category
	createReq := models.CreateCategoryRequest{
		Name:        "Test Category",
		Description: helpers.StringPtr("Original description"),
	}
	created, err := suite.categoryRepo.Create(createReq)
	require.NoError(suite.T(), err)

	// Update only name (description should remain unchanged)
	updateReq := models.UpdateCategoryRequest{
		Name:        helpers.StringPtr("New Name Only"),
		Description: nil, // Don't update description
	}

	updated, err := suite.categoryRepo.Update(created.ID, updateReq)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "New Name Only", updated.Name)
	assert.NotNil(suite.T(), updated.Description)
	assert.Equal(suite.T(), "Original description", *updated.Description) // Should remain unchanged
}

func (suite *CategoryRepositoryIntegrationSuite) TestUpdate_PartialUpdate_DescriptionOnly() {
	// Create initial category
	createReq := models.CreateCategoryRequest{
		Name:        "Original Name",
		Description: helpers.StringPtr("Original description"),
	}
	created, err := suite.categoryRepo.Create(createReq)
	require.NoError(suite.T(), err)

	// Update only description (name should remain unchanged)
	updateReq := models.UpdateCategoryRequest{
		Name:        nil, // Don't update name
		Description: helpers.StringPtr("New Description Only"),
	}

	updated, err := suite.categoryRepo.Update(created.ID, updateReq)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Original Name", updated.Name) // Should remain unchanged
	assert.NotNil(suite.T(), updated.Description)
	assert.Equal(suite.T(), "New Description Only", *updated.Description)
}

func (suite *CategoryRepositoryIntegrationSuite) TestUpdate_SetDescriptionToNull() {
	// Create initial category with description
	createReq := models.CreateCategoryRequest{
		Name:        "Category With Description",
		Description: helpers.StringPtr("Has description"),
	}
	created, err := suite.categoryRepo.Create(createReq)
	require.NoError(suite.T(), err)

	// Update to remove description (set to null)
	updateReq := models.UpdateCategoryRequest{
		Name:        nil,                   // Don't update name
		Description: helpers.StringPtr(""), // Empty string to set null
	}

	updated, err := suite.categoryRepo.Update(created.ID, updateReq)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Category With Description", updated.Name)
	assert.NotNil(suite.T(), updated.Description)
	assert.Equal(suite.T(), "", *updated.Description) // Should be empty string
}

func (suite *CategoryRepositoryIntegrationSuite) TestUpdate_NonExistentCategory() {
	updateReq := models.UpdateCategoryRequest{
		Name:        helpers.StringPtr("Updated Name"),
		Description: helpers.StringPtr("Updated description"),
	}

	updated, err := suite.categoryRepo.Update(99999, updateReq)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), updated)
	assert.Equal(suite.T(), sql.ErrNoRows, err)
}

func (suite *CategoryRepositoryIntegrationSuite) TestUpdate_DuplicateName_ShouldFail() {
	// Create two categories
	req1 := models.CreateCategoryRequest{
		Name:        "First Category",
		Description: helpers.StringPtr("First description"),
	}
	_, err := suite.categoryRepo.Create(req1)
	require.NoError(suite.T(), err)

	req2 := models.CreateCategoryRequest{
		Name:        "Second Category",
		Description: helpers.StringPtr("Second description"),
	}
	category2, err := suite.categoryRepo.Create(req2)
	require.NoError(suite.T(), err)

	// Try to update second category to have same name as first
	updateReq := models.UpdateCategoryRequest{
		Name:        helpers.StringPtr("First Category"), // Duplicate name
		Description: helpers.StringPtr("Updated description"),
	}

	updated, err := suite.categoryRepo.Update(category2.ID, updateReq)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), updated)
	assert.Contains(suite.T(), err.Error(), "duplicate key value") // PostgreSQL unique constraint error
}

// =============================================================================
// DELETE OPERATIONS
// =============================================================================

func (suite *CategoryRepositoryIntegrationSuite) TestDelete_Success() {
	// Create category to delete
	createReq := models.CreateCategoryRequest{
		Name:        "Category To Delete",
		Description: helpers.StringPtr("This will be deleted"),
	}
	created, err := suite.categoryRepo.Create(createReq)
	require.NoError(suite.T(), err)

	// Verify category exists before deletion
	exists, err := suite.categoryRepo.Exists(created.ID)
	require.NoError(suite.T(), err)
	require.True(suite.T(), exists)

	// Delete category
	err = suite.categoryRepo.Delete(created.ID)

	assert.NoError(suite.T(), err)

	// Verify category is deleted
	found, err := suite.categoryRepo.GetByID(created.ID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), found)
	assert.Equal(suite.T(), sql.ErrNoRows, err)

	// Verify exists check returns false
	exists, err = suite.categoryRepo.Exists(created.ID)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), exists)
}

func (suite *CategoryRepositoryIntegrationSuite) TestDelete_NonExistentCategory() {
	err := suite.categoryRepo.Delete(99999)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), sql.ErrNoRows, err)
}

func (suite *CategoryRepositoryIntegrationSuite) TestDelete_WithRecipes_ShouldFail() {
	// Create category
	categoryReq := models.CreateCategoryRequest{
		Name:        "Category With Recipe",
		Description: helpers.StringPtr("Has recipes"),
	}
	category, err := suite.categoryRepo.Create(categoryReq)
	require.NoError(suite.T(), err)

	// Create recipe in this category
	recipeReq := models.CreateRecipeRequest{
		Name:        "Recipe in Category",
		Description: "Recipe description",
		CategoryID:  category.ID,
	}
	_, err = suite.recipeRepo.Create(recipeReq)
	require.NoError(suite.T(), err)

	// Try to delete category - should fail due to foreign key constraint
	err = suite.categoryRepo.Delete(category.ID)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "foreign key constraint") // PostgreSQL constraint error

	// Verify category still exists
	found, err := suite.categoryRepo.GetByID(category.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), found)

	exists, err := suite.categoryRepo.Exists(category.ID)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
}

// =============================================================================
// DATA INTEGRITY AND EDGE CASES
// =============================================================================

func (suite *CategoryRepositoryIntegrationSuite) TestConcurrentCreation() {
	// Test concurrent category creation
	done := make(chan bool, 2)

	go func() {
		req := models.CreateCategoryRequest{
			Name:        "Concurrent Category 1",
			Description: helpers.StringPtr("Created concurrently"),
		}
		_, err := suite.categoryRepo.Create(req)
		assert.NoError(suite.T(), err)
		done <- true
	}()

	go func() {
		req := models.CreateCategoryRequest{
			Name:        "Concurrent Category 2",
			Description: helpers.StringPtr("Also created concurrently"),
		}
		_, err := suite.categoryRepo.Create(req)
		assert.NoError(suite.T(), err)
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify both categories were created
	categories, err := suite.categoryRepo.GetAll()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), categories, 2)
}

func (suite *CategoryRepositoryIntegrationSuite) TestLongNames() {
	// Test with very long category name (within reasonable limits)
	longName := "Very Long Category Name That Tests Database Field Length Limits But Should Still Work Fine"

	req := models.CreateCategoryRequest{
		Name:        longName,
		Description: helpers.StringPtr("Category with long name"),
	}

	category, err := suite.categoryRepo.Create(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), category)
	assert.Equal(suite.T(), longName, category.Name)

	// Verify in database
	found, err := suite.categoryRepo.GetByID(category.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), longName, found.Name)
}

func (suite *CategoryRepositoryIntegrationSuite) TestSpecialCharacters() {
	// Test with special characters in name and description
	specialName := "Fusion & Modern Cuisine (Asian-European)"
	specialDesc := "Includes: àáâãäå, üöäß, and other special chars: !@#$%^&*()"

	req := models.CreateCategoryRequest{
		Name:        specialName,
		Description: helpers.StringPtr(specialDesc),
	}

	category, err := suite.categoryRepo.Create(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), category)
	assert.Equal(suite.T(), specialName, category.Name)
	assert.NotNil(suite.T(), category.Description)
	assert.Equal(suite.T(), specialDesc, *category.Description)

	// Verify in database
	found, err := suite.categoryRepo.GetByID(category.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), specialName, found.Name)
	assert.Equal(suite.T(), specialDesc, *found.Description)
}

func (suite *CategoryRepositoryIntegrationSuite) TestEmptyNameShouldFail() {
	// Test with empty name - should fail due to database constraints
	req := models.CreateCategoryRequest{
		Name:        "", // Empty name
		Description: helpers.StringPtr("Valid description"),
	}

	category, err := suite.categoryRepo.Create(req)

	// This should fail - exact error depends on database constraints
	// Might be a NOT NULL constraint or a CHECK constraint
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), category)
}

// Run the test suite
func TestCategoryRepositoryIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(CategoryRepositoryIntegrationSuite))
}
