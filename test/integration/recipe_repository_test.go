package integration

import (
	"meal-prep/services/recipe-catalogue/repository"
	"meal-prep/shared/models"
	"meal-prep/test/helpers"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// RecipeRepositoryIntegrationSuite tests recipe repository against real PostgreSQL database
// Uses testcontainers to spin up isolated database for each test run
// Tests cover CRUD operations, relationships, transactions, and edge cases
type RecipeRepositoryIntegrationSuite struct {
	suite.Suite
	testDB         *helpers.TestDatabase
	recipeRepo     repository.RecipeRepository
	categoryRepo   repository.CategoryRepository
	ingredientRepo repository.IngredientRepository

	// Store seeded data IDs for use in tests
	italianCategoryID  int
	healthyCategoryID  int
	tomatoIngredientID int
	pastaIngredientID  int
	cheeseIngredientID int
}

func (suite *RecipeRepositoryIntegrationSuite) SetupSuite() {
	helpers.SuppressTestLogs()
	suite.testDB = helpers.SetupPostgresContainer(suite.T())

	suite.recipeRepo = repository.NewRecipeRepository(suite.testDB.DB)
	suite.categoryRepo = repository.NewCategoryRepository(suite.testDB.DB)
	suite.ingredientRepo = repository.NewIngredientRepository(suite.testDB.DB)
}

func (suite *RecipeRepositoryIntegrationSuite) TearDownSuite() {
	suite.testDB.Cleanup(suite.T())
	helpers.RestoreTestLogs()
}

func (suite *RecipeRepositoryIntegrationSuite) SetupTest() {
	suite.testDB.CleanupTestData(suite.T())
	suite.seedTestData()
}

// seedTestData creates test categories and ingredients for recipe tests
func (suite *RecipeRepositoryIntegrationSuite) seedTestData() {
	// Create test categories using proper repository methods
	category1 := models.CreateCategoryRequest{
		Name:        "Italian",
		Description: helpers.StringPtr("Italian cuisine"),
	}
	createdCategory1, err := suite.categoryRepo.Create(category1)
	require.NoError(suite.T(), err)
	suite.italianCategoryID = createdCategory1.ID

	category2 := models.CreateCategoryRequest{
		Name:        "Healthy",
		Description: helpers.StringPtr("Healthy recipes"),
	}
	createdCategory2, err := suite.categoryRepo.Create(category2)
	require.NoError(suite.T(), err)
	suite.healthyCategoryID = createdCategory2.ID

	// Create test ingredients using proper pointer fields
	ingredient1 := models.CreateIngredientRequest{
		Name:        "Tomato",
		Description: helpers.StringPtr("Fresh tomatoes"),
		Category:    helpers.StringPtr("Vegetables"),
	}
	createdIngredient1, err := suite.ingredientRepo.CreateIngredient(ingredient1)
	require.NoError(suite.T(), err)
	suite.tomatoIngredientID = createdIngredient1.ID

	ingredient2 := models.CreateIngredientRequest{
		Name:        "Pasta",
		Description: helpers.StringPtr("Italian pasta"),
		Category:    helpers.StringPtr("Grains"),
	}
	createdIngredient2, err := suite.ingredientRepo.CreateIngredient(ingredient2)
	require.NoError(suite.T(), err)
	suite.pastaIngredientID = createdIngredient2.ID

	ingredient3 := models.CreateIngredientRequest{
		Name:        "Cheese",
		Description: helpers.StringPtr("Mozzarella cheese"),
		Category:    helpers.StringPtr("Dairy"),
	}
	createdIngredient3, err := suite.ingredientRepo.CreateIngredient(ingredient3)
	require.NoError(suite.T(), err)
	suite.cheeseIngredientID = createdIngredient3.ID
}

// =============================================================================
// BASIC CRUD OPERATIONS
// =============================================================================

func (suite *RecipeRepositoryIntegrationSuite) TestGetAll_EmptyDatabase() {
	// Clean all recipes first
	suite.testDB.CleanupTestData(suite.T())
	suite.seedTestData() // Only categories and ingredients

	recipes, err := suite.recipeRepo.GetAll()

	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), recipes)
}

func (suite *RecipeRepositoryIntegrationSuite) TestGetAll_WithRecipes() {
	// Create test recipes
	recipe1 := models.CreateRecipeRequest{
		Name:        "Spaghetti Carbonara",
		Description: "Classic Italian pasta dish",
		CategoryID:  suite.italianCategoryID,
	}
	_, err := suite.recipeRepo.Create(1, recipe1)
	require.NoError(suite.T(), err)

	recipe2 := models.CreateRecipeRequest{
		Name:        "Caesar Salad",
		Description: "Fresh healthy salad",
		CategoryID:  suite.healthyCategoryID,
	}
	_, err = suite.recipeRepo.Create(1, recipe2)
	require.NoError(suite.T(), err)

	recipes, err := suite.recipeRepo.GetAll()

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), recipes, 2)

	// Verify recipes are sorted by name
	assert.Equal(suite.T(), "Caesar Salad", recipes[0].Name)
	assert.Equal(suite.T(), "Spaghetti Carbonara", recipes[1].Name)

	// Verify category information is loaded
	assert.NotNil(suite.T(), recipes[0].Category)
	assert.Equal(suite.T(), "Healthy", recipes[0].Category.Name)
	assert.NotNil(suite.T(), recipes[1].Category)
	assert.Equal(suite.T(), "Italian", recipes[1].Category.Name)
}

func (suite *RecipeRepositoryIntegrationSuite) TestGetByID_ExistingRecipe() {
	// Create test recipe
	recipe := models.CreateRecipeRequest{
		Name:        "Margherita Pizza",
		Description: "Classic Italian pizza",
		CategoryID:  suite.italianCategoryID,
	}
	createdRecipe, err := suite.recipeRepo.Create(1, recipe)
	require.NoError(suite.T(), err)

	foundRecipe, err := suite.recipeRepo.GetByID(createdRecipe.ID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), foundRecipe)
	assert.Equal(suite.T(), createdRecipe.ID, foundRecipe.ID)
	assert.Equal(suite.T(), "Margherita Pizza", foundRecipe.Name)
	assert.Equal(suite.T(), "Classic Italian pizza", *foundRecipe.Description)
	assert.Equal(suite.T(), suite.italianCategoryID, *foundRecipe.CategoryID)
	assert.NotNil(suite.T(), foundRecipe.Category)
	assert.Equal(suite.T(), "Italian", foundRecipe.Category.Name)
}

func (suite *RecipeRepositoryIntegrationSuite) TestGetByID_NonExistentRecipe() {
	foundRecipe, err := suite.recipeRepo.GetByID(99999)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), foundRecipe)
}

func (suite *RecipeRepositoryIntegrationSuite) TestCreate_ValidRecipe() {
	recipe := models.CreateRecipeRequest{
		Name:        "Chicken Alfredo",
		Description: "Creamy chicken pasta",
		CategoryID:  suite.italianCategoryID,
	}

	createdRecipe, err := suite.recipeRepo.Create(1, recipe)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), createdRecipe)
	assert.True(suite.T(), createdRecipe.ID > 0)
	assert.Equal(suite.T(), "Chicken Alfredo", createdRecipe.Name)
	assert.Equal(suite.T(), "Creamy chicken pasta", *createdRecipe.Description)
	assert.Equal(suite.T(), suite.italianCategoryID, *createdRecipe.CategoryID)
	assert.NotZero(suite.T(), createdRecipe.CreatedAt)
	assert.NotZero(suite.T(), createdRecipe.UpdatedAt)

	// Verify recipe was actually inserted
	foundRecipe, err := suite.recipeRepo.GetByID(createdRecipe.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), createdRecipe.ID, foundRecipe.ID)
}

func (suite *RecipeRepositoryIntegrationSuite) TestCreate_InvalidCategory() {
	recipe := models.CreateRecipeRequest{
		Name:        "Invalid Recipe",
		Description: "Recipe with invalid category",
		CategoryID:  99999, // Non-existent category
	}

	createdRecipe, err := suite.recipeRepo.Create(1, recipe)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), createdRecipe)
	assert.Contains(suite.T(), err.Error(), "foreign key constraint")
}

func (suite *RecipeRepositoryIntegrationSuite) TestUpdate_ExistingRecipe() {
	// Create initial recipe
	recipe := models.CreateRecipeRequest{
		Name:        "Original Name",
		Description: "Original description",
		CategoryID:  suite.italianCategoryID,
	}
	createdRecipe, err := suite.recipeRepo.Create(1, recipe)
	require.NoError(suite.T(), err)

	// Update recipe
	updateReq := models.UpdateRecipeRequest{
		Name:        "Updated Name",
		Description: "Updated description",
		CategoryID:  suite.healthyCategoryID, // Change category
	}

	updatedRecipe, err := suite.recipeRepo.Update(createdRecipe.ID, updateReq)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updatedRecipe)
	assert.Equal(suite.T(), createdRecipe.ID, updatedRecipe.ID)
	assert.Equal(suite.T(), "Updated Name", updatedRecipe.Name)
	assert.Equal(suite.T(), "Updated description", *updatedRecipe.Description)
	assert.Equal(suite.T(), suite.healthyCategoryID, *updatedRecipe.CategoryID)
	assert.True(suite.T(), updatedRecipe.UpdatedAt.After(createdRecipe.UpdatedAt))

	// Verify in database
	foundRecipe, err := suite.recipeRepo.GetByID(createdRecipe.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Name", foundRecipe.Name)
	assert.Equal(suite.T(), "Healthy", foundRecipe.Category.Name)
}

func (suite *RecipeRepositoryIntegrationSuite) TestUpdate_NonExistentRecipe() {
	updateReq := models.UpdateRecipeRequest{
		Name:        "Updated Name",
		Description: "Updated description",
		CategoryID:  suite.italianCategoryID,
	}

	updatedRecipe, err := suite.recipeRepo.Update(99999, updateReq)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), updatedRecipe)
}

func (suite *RecipeRepositoryIntegrationSuite) TestDelete_ExistingRecipe() {
	// Create recipe to delete
	recipe := models.CreateRecipeRequest{
		Name:        "Recipe to Delete",
		Description: "This will be deleted",
		CategoryID:  suite.italianCategoryID,
	}
	createdRecipe, err := suite.recipeRepo.Create(1, recipe)
	require.NoError(suite.T(), err)

	// Delete recipe
	err = suite.recipeRepo.Delete(createdRecipe.ID)

	assert.NoError(suite.T(), err)

	// Verify recipe is deleted
	foundRecipe, err := suite.recipeRepo.GetByID(createdRecipe.ID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), foundRecipe)
}

func (suite *RecipeRepositoryIntegrationSuite) TestDelete_NonExistentRecipe() {
	err := suite.recipeRepo.Delete(99999)

	assert.Error(suite.T(), err)
}

// =============================================================================
// CATEGORY-BASED OPERATIONS
// =============================================================================

func (suite *RecipeRepositoryIntegrationSuite) TestGetByCategory_WithRecipes() {
	// Create recipes in different categories
	italianRecipe := models.CreateRecipeRequest{
		Name:        "Pasta Primavera",
		Description: "Spring pasta",
		CategoryID:  suite.italianCategoryID,
	}
	suite.recipeRepo.Create(1, italianRecipe)

	healthyRecipe1 := models.CreateRecipeRequest{
		Name:        "Green Smoothie",
		Description: "Healthy smoothie",
		CategoryID:  suite.healthyCategoryID,
	}
	suite.recipeRepo.Create(1, healthyRecipe1)

	healthyRecipe2 := models.CreateRecipeRequest{
		Name:        "Quinoa Bowl",
		Description: "Nutritious bowl",
		CategoryID:  suite.healthyCategoryID,
	}
	suite.recipeRepo.Create(1, healthyRecipe2)

	// Get recipes by category
	healthyRecipes, err := suite.recipeRepo.GetByCategory(suite.healthyCategoryID)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), healthyRecipes, 2)
	assert.Equal(suite.T(), "Green Smoothie", healthyRecipes[0].Name)
	assert.Equal(suite.T(), "Quinoa Bowl", healthyRecipes[1].Name)

	// Verify all returned recipes have correct category
	for _, recipe := range healthyRecipes {
		assert.Equal(suite.T(), suite.healthyCategoryID, *recipe.CategoryID)
	}
}

func (suite *RecipeRepositoryIntegrationSuite) TestGetByCategory_EmptyCategory() {
	recipes, err := suite.recipeRepo.GetByCategory(suite.italianCategoryID)

	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), recipes)
}

func (suite *RecipeRepositoryIntegrationSuite) TestGetByCategory_NonExistentCategory() {
	recipes, err := suite.recipeRepo.GetByCategory(99999)

	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), recipes)
}

// =============================================================================
// RECIPES WITH INGREDIENTS OPERATIONS
// =============================================================================

func (suite *RecipeRepositoryIntegrationSuite) TestCreateWithIngredients_SuccessfulCreation() {
	req := models.CreateRecipeWithIngredientsRequest{
		Name:        "Pasta with Tomatoes",
		Description: helpers.StringPtr("Simple pasta dish"),
		CategoryID:  suite.italianCategoryID,
		Ingredients: []models.AddRecipeIngredientRequest{
			{
				IngredientID: suite.tomatoIngredientID,
				Quantity:     2.0,
				Unit:         "pieces",
				Notes:        helpers.StringPtr("Fresh and ripe"),
			},
			{
				IngredientID: suite.pastaIngredientID,
				Quantity:     200.0,
				Unit:         "grams",
				Notes:        nil, // No notes
			},
		},
	}

	result, err := suite.recipeRepo.CreateWithIngredients(1, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.True(suite.T(), result.Recipe.ID > 0)
	assert.Equal(suite.T(), "Pasta with Tomatoes", result.Recipe.Name)
	assert.Len(suite.T(), result.Ingredients, 2)

	// Verify ingredients details
	assert.Equal(suite.T(), suite.tomatoIngredientID, result.Ingredients[0].IngredientID)
	assert.Equal(suite.T(), 2.0, result.Ingredients[0].Quantity)
	assert.Equal(suite.T(), "pieces", result.Ingredients[0].Unit)
	assert.NotNil(suite.T(), result.Ingredients[0].Notes)
	assert.Equal(suite.T(), "Fresh and ripe", *result.Ingredients[0].Notes)
	assert.Equal(suite.T(), "Tomato", result.Ingredients[0].Ingredient.Name)

	// Verify second ingredient has no notes (nil pointer)
	assert.Equal(suite.T(), suite.pastaIngredientID, result.Ingredients[1].IngredientID)
	assert.Nil(suite.T(), result.Ingredients[1].Notes)
}

func (suite *RecipeRepositoryIntegrationSuite) TestCreateWithIngredients_NoIngredients() {
	req := models.CreateRecipeWithIngredientsRequest{
		Name:        "Simple Recipe",
		Description: helpers.StringPtr("Recipe without ingredients"),
		CategoryID:  suite.italianCategoryID,
		Ingredients: []models.AddRecipeIngredientRequest{},
	}

	result, err := suite.recipeRepo.CreateWithIngredients(1, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.True(suite.T(), result.Recipe.ID > 0)
	assert.Empty(suite.T(), result.Ingredients)
}

func (suite *RecipeRepositoryIntegrationSuite) TestGetByIDWithIngredients_ExistingRecipe() {
	// Create recipe with ingredients
	req := models.CreateRecipeWithIngredientsRequest{
		Name:        "Cheese Pasta",
		Description: helpers.StringPtr("Pasta with cheese"),
		CategoryID:  suite.italianCategoryID,
		Ingredients: []models.AddRecipeIngredientRequest{
			{
				IngredientID: suite.pastaIngredientID,
				Quantity:     150.0,
				Unit:         "grams",
				Notes:        nil,
			},
			{
				IngredientID: suite.cheeseIngredientID,
				Quantity:     50.0,
				Unit:         "grams",
				Notes:        helpers.StringPtr("Grated"),
			},
		},
	}
	created, err := suite.recipeRepo.CreateWithIngredients(1, req)
	require.NoError(suite.T(), err)

	// Get recipe with ingredients
	result, err := suite.recipeRepo.GetByIDWithIngredients(created.Recipe.ID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), created.Recipe.ID, result.Recipe.ID)
	assert.Equal(suite.T(), "Cheese Pasta", result.Recipe.Name)
	assert.Len(suite.T(), result.Ingredients, 2)

	// Verify ingredients are loaded with full details
	for _, ingredient := range result.Ingredients {
		assert.NotNil(suite.T(), ingredient.Ingredient)
		assert.NotEmpty(suite.T(), ingredient.Ingredient.Name)
	}
}

func (suite *RecipeRepositoryIntegrationSuite) TestGetAllWithIngredients_MultipleRecipes() {
	// Create multiple recipes with ingredients
	recipe1 := models.CreateRecipeWithIngredientsRequest{
		Name:        "Tomato Pasta",
		Description: helpers.StringPtr("Pasta with tomatoes"),
		CategoryID:  suite.italianCategoryID,
		Ingredients: []models.AddRecipeIngredientRequest{
			{IngredientID: suite.tomatoIngredientID, Quantity: 2.0, Unit: "pieces", Notes: nil},
			{IngredientID: suite.pastaIngredientID, Quantity: 100.0, Unit: "grams", Notes: nil},
		},
	}
	suite.recipeRepo.CreateWithIngredients(1, recipe1)

	recipe2 := models.CreateRecipeWithIngredientsRequest{
		Name:        "Cheese Salad",
		Description: helpers.StringPtr("Salad with cheese"),
		CategoryID:  suite.healthyCategoryID,
		Ingredients: []models.AddRecipeIngredientRequest{
			{IngredientID: suite.cheeseIngredientID, Quantity: 30.0, Unit: "grams", Notes: helpers.StringPtr("Crumbled")},
		},
	}
	suite.recipeRepo.CreateWithIngredients(1, recipe2)

	results, err := suite.recipeRepo.GetAllWithIngredients()

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 2)

	// Verify both recipes have their ingredients loaded
	for _, result := range results {
		assert.NotEmpty(suite.T(), result.Ingredients)
		for _, ingredient := range result.Ingredients {
			assert.NotNil(suite.T(), ingredient.Ingredient)
		}
	}
}

func (suite *RecipeRepositoryIntegrationSuite) TestUpdateWithIngredients_Success() {
	// Create initial recipe with ingredients
	createReq := models.CreateRecipeWithIngredientsRequest{
		Name:        "Original Recipe",
		Description: helpers.StringPtr("Original description"),
		CategoryID:  suite.italianCategoryID,
		Ingredients: []models.AddRecipeIngredientRequest{
			{IngredientID: suite.tomatoIngredientID, Quantity: 1.0, Unit: "piece", Notes: nil},
		},
	}
	created, err := suite.recipeRepo.CreateWithIngredients(1, createReq)
	require.NoError(suite.T(), err)

	// Update recipe with different ingredients
	updateReq := models.UpdateRecipeWithIngredientsRequest{
		Name:        "Updated Recipe",
		Description: "Updated description",
		CategoryID:  suite.healthyCategoryID,
		Ingredients: []models.AddRecipeIngredientRequest{
			{IngredientID: suite.cheeseIngredientID, Quantity: 50.0, Unit: "grams", Notes: helpers.StringPtr("Grated")},
			{IngredientID: suite.pastaIngredientID, Quantity: 200.0, Unit: "grams", Notes: nil},
		},
	}

	updated, err := suite.recipeRepo.UpdateWithIngredients(created.Recipe.ID, updateReq)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updated)
	assert.Equal(suite.T(), "Updated Recipe", updated.Recipe.Name)
	assert.Equal(suite.T(), "Updated description", *updated.Recipe.Description)
	assert.Equal(suite.T(), suite.healthyCategoryID, *updated.Recipe.CategoryID)
	assert.Len(suite.T(), updated.Ingredients, 2)

	// Verify old ingredients were replaced
	ingredientIDs := []int{updated.Ingredients[0].IngredientID, updated.Ingredients[1].IngredientID}
	assert.Contains(suite.T(), ingredientIDs, suite.cheeseIngredientID)
	assert.Contains(suite.T(), ingredientIDs, suite.pastaIngredientID)
	assert.NotContains(suite.T(), ingredientIDs, suite.tomatoIngredientID) // Should be removed
}

func (suite *RecipeRepositoryIntegrationSuite) TestGetByCategoryWithIngredients_Success() {
	// Create recipes with ingredients in specific category
	recipe1 := models.CreateRecipeWithIngredientsRequest{
		Name:       "Italian Recipe 1",
		CategoryID: suite.italianCategoryID,
		Ingredients: []models.AddRecipeIngredientRequest{
			{IngredientID: suite.tomatoIngredientID, Quantity: 2.0, Unit: "pieces", Notes: nil},
		},
	}
	suite.recipeRepo.CreateWithIngredients(1, recipe1)

	recipe2 := models.CreateRecipeWithIngredientsRequest{
		Name:       "Italian Recipe 2",
		CategoryID: suite.italianCategoryID,
		Ingredients: []models.AddRecipeIngredientRequest{
			{IngredientID: suite.pastaIngredientID, Quantity: 100.0, Unit: "grams", Notes: nil},
		},
	}
	suite.recipeRepo.CreateWithIngredients(1, recipe2)

	// Create recipe in different category
	recipe3 := models.CreateRecipeWithIngredientsRequest{
		Name:       "Healthy Recipe",
		CategoryID: suite.healthyCategoryID,
		Ingredients: []models.AddRecipeIngredientRequest{
			{IngredientID: suite.cheeseIngredientID, Quantity: 30.0, Unit: "grams", Notes: nil},
		},
	}
	suite.recipeRepo.CreateWithIngredients(1, recipe3)

	// Get Italian recipes with ingredients
	results, err := suite.recipeRepo.GetByCategoryWithIngredients(suite.italianCategoryID)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 2)

	// Verify all recipes are Italian and have ingredients
	for _, result := range results {
		assert.Equal(suite.T(), suite.italianCategoryID, *result.Recipe.CategoryID)
		assert.NotEmpty(suite.T(), result.Ingredients)
	}
}

// =============================================================================
// SEARCH OPERATIONS
// =============================================================================

func (suite *RecipeRepositoryIntegrationSuite) TestSearchRecipesByIngredients_MatchingRecipes() {
	// Create recipes with specific ingredients
	tomatoPasta := models.CreateRecipeWithIngredientsRequest{
		Name:        "Tomato Pasta",
		Description: nil,
		CategoryID:  suite.italianCategoryID,
		Ingredients: []models.AddRecipeIngredientRequest{
			{IngredientID: suite.tomatoIngredientID, Quantity: 2.0, Unit: "pieces", Notes: nil},
			{IngredientID: suite.pastaIngredientID, Quantity: 100.0, Unit: "grams", Notes: nil},
		},
	}
	suite.recipeRepo.CreateWithIngredients(1, tomatoPasta)

	cheesePasta := models.CreateRecipeWithIngredientsRequest{
		Name:        "Cheese Pasta",
		Description: nil,
		CategoryID:  suite.italianCategoryID,
		Ingredients: []models.AddRecipeIngredientRequest{
			{IngredientID: suite.pastaIngredientID, Quantity: 100.0, Unit: "grams", Notes: nil},
			{IngredientID: suite.cheeseIngredientID, Quantity: 50.0, Unit: "grams", Notes: nil},
		},
	}
	suite.recipeRepo.CreateWithIngredients(1, cheesePasta)

	tomatoSalad := models.CreateRecipeWithIngredientsRequest{
		Name:        "Tomato Salad",
		Description: helpers.StringPtr("Fresh salad"),
		CategoryID:  suite.healthyCategoryID,
		Ingredients: []models.AddRecipeIngredientRequest{
			{IngredientID: suite.tomatoIngredientID, Quantity: 3.0, Unit: "pieces", Notes: helpers.StringPtr("Diced")},
		},
	}
	suite.recipeRepo.CreateWithIngredients(1, tomatoSalad)

	// Search for recipes containing tomato
	recipes, err := suite.recipeRepo.SearchRecipesByIngredients([]int{suite.tomatoIngredientID})

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), recipes, 2) // Tomato Pasta and Tomato Salad

	recipeNames := []string{recipes[0].Name, recipes[1].Name}
	assert.Contains(suite.T(), recipeNames, "Tomato Pasta")
	assert.Contains(suite.T(), recipeNames, "Tomato Salad")

	// Search for recipes containing both tomato and pasta
	recipes, err = suite.recipeRepo.SearchRecipesByIngredients([]int{suite.tomatoIngredientID, suite.pastaIngredientID})

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), recipes, 3) // Recipes include at least one of the ingredients
	allRecipeNames := []string{recipes[0].Name, recipes[1].Name, recipes[2].Name}
	assert.Contains(suite.T(), allRecipeNames, "Tomato Pasta")
	assert.Contains(suite.T(), allRecipeNames, "Tomato Salad")
	assert.Contains(suite.T(), allRecipeNames, "Cheese Pasta")
}

func (suite *RecipeRepositoryIntegrationSuite) TestSearchRecipesByIngredients_NoMatchingRecipes() {
	// Create recipe without the searched ingredient
	recipe := models.CreateRecipeWithIngredientsRequest{
		Name:        "Plain Pasta",
		Description: nil,
		CategoryID:  suite.italianCategoryID,
		Ingredients: []models.AddRecipeIngredientRequest{
			{IngredientID: suite.pastaIngredientID, Quantity: 100.0, Unit: "grams", Notes: nil},
		},
	}
	suite.recipeRepo.CreateWithIngredients(1, recipe)

	// Search for recipes with tomato
	recipes, err := suite.recipeRepo.SearchRecipesByIngredients([]int{suite.tomatoIngredientID})

	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), recipes)
}

func (suite *RecipeRepositoryIntegrationSuite) TestSearchRecipesByIngredientsWithIngredients_CompleteData() {
	// Create recipe with ingredients
	recipe := models.CreateRecipeWithIngredientsRequest{
		Name:        "Complete Recipe",
		Description: helpers.StringPtr("Recipe with all details"),
		CategoryID:  suite.italianCategoryID,
		Ingredients: []models.AddRecipeIngredientRequest{
			{IngredientID: suite.tomatoIngredientID, Quantity: 2.0, Unit: "pieces", Notes: helpers.StringPtr("Fresh")},
			{IngredientID: suite.pastaIngredientID, Quantity: 100.0, Unit: "grams", Notes: nil},
		},
	}
	suite.recipeRepo.CreateWithIngredients(1, recipe)

	// Search with complete ingredient data
	results, err := suite.recipeRepo.SearchRecipesByIngredientsWithIngredients([]int{suite.tomatoIngredientID})

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 1)
	assert.Equal(suite.T(), "Complete Recipe", results[0].Recipe.Name)
	assert.Len(suite.T(), results[0].Ingredients, 2)

	// Verify ingredient details are loaded
	for _, ingredient := range results[0].Ingredients {
		assert.NotNil(suite.T(), ingredient.Ingredient)
		assert.NotEmpty(suite.T(), ingredient.Ingredient.Name)
	}
}

// =============================================================================
// ERROR HANDLING AND EDGE CASES
// =============================================================================

func (suite *RecipeRepositoryIntegrationSuite) TestTransactionRollback_CreateWithIngredientsFailure() {
	req := models.CreateRecipeWithIngredientsRequest{
		Name:        "Failed Recipe",
		Description: helpers.StringPtr("This should fail"),
		CategoryID:  suite.italianCategoryID,
		Ingredients: []models.AddRecipeIngredientRequest{
			{
				IngredientID: 99999, // Non-existent ingredient
				Quantity:     100.0,
				Unit:         "grams",
				Notes:        nil,
			},
		},
	}

	result, err := suite.recipeRepo.CreateWithIngredients(1, req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)

	// Verify no partial data was created (transaction was rolled back)
	recipes, _ := suite.recipeRepo.GetAll()
	for _, recipe := range recipes {
		assert.NotEqual(suite.T(), "Failed Recipe", recipe.Name)
	}
}

func (suite *RecipeRepositoryIntegrationSuite) TestConcurrentAccess_MultipleCreates() {
	// Test concurrent recipe creation to verify database constraints
	done := make(chan bool, 2)

	go func() {
		recipe := models.CreateRecipeRequest{
			Name:       "Concurrent Recipe 1",
			CategoryID: suite.italianCategoryID,
		}
		_, err := suite.recipeRepo.Create(1, recipe)
		assert.NoError(suite.T(), err)
		done <- true
	}()

	go func() {
		recipe := models.CreateRecipeRequest{
			Name:       "Concurrent Recipe 2",
			CategoryID: suite.healthyCategoryID,
		}
		_, err := suite.recipeRepo.Create(1, recipe)
		assert.NoError(suite.T(), err)
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify both recipes were created
	recipes, err := suite.recipeRepo.GetAll()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), recipes, 2)
}

func (suite *RecipeRepositoryIntegrationSuite) TestPointerFields_NilValueHandling() {
	// Test that nil pointer fields are handled correctly
	req := models.CreateRecipeWithIngredientsRequest{
		Name:        "Minimal Recipe",
		Description: nil, // Nil description pointer
		CategoryID:  suite.italianCategoryID,
		Ingredients: []models.AddRecipeIngredientRequest{
			{
				IngredientID: suite.tomatoIngredientID,
				Quantity:     1.0,
				Unit:         "piece",
				Notes:        nil, // Nil notes pointer
			},
			{
				IngredientID: suite.pastaIngredientID,
				Quantity:     100.0,
				Unit:         "grams",
				Notes:        helpers.StringPtr("Some notes"), // Non-nil notes
			},
		},
	}

	result, err := suite.recipeRepo.CreateWithIngredients(1, req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)

	// Verify recipe description is nil
	assert.Nil(suite.T(), result.Recipe.Description)

	// Verify ingredient notes handling
	assert.Len(suite.T(), result.Ingredients, 2)
	assert.Nil(suite.T(), result.Ingredients[0].Notes)    // First ingredient: nil notes
	assert.NotNil(suite.T(), result.Ingredients[1].Notes) // Second ingredient: has notes
	assert.Equal(suite.T(), "Some notes", *result.Ingredients[1].Notes)

	// Verify we can retrieve the recipe and pointer fields are still correct
	retrieved, err := suite.recipeRepo.GetByIDWithIngredients(result.Recipe.ID)
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), retrieved.Recipe.Description)
	assert.Nil(suite.T(), retrieved.Ingredients[0].Notes)
	assert.NotNil(suite.T(), retrieved.Ingredients[1].Notes)
}

// Run the test suite
func TestRecipeRepositoryIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(RecipeRepositoryIntegrationSuite))
}
