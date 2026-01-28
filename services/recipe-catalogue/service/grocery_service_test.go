package service

import (
	"meal-prep/services/recipe-catalogue/service/mocks"
	"meal-prep/services/recipe-catalogue/service/testdata"
	"meal-prep/shared/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

type groceryServiceTestSetup struct {
	service        GroceryService
	ingredientRepo *mocks.MockIngredientRepository
	recipeRepo     *mocks.MockRecipeRepository
}

func setupGroceryServiceTest() *groceryServiceTestSetup {
	ingredientRepo := new(mocks.MockIngredientRepository)
	recipeRepo := new(mocks.MockRecipeRepository)

	service := NewGroceryService(ingredientRepo, recipeRepo)

	return &groceryServiceTestSetup{
		service:        service,
		ingredientRepo: ingredientRepo,
		recipeRepo:     recipeRepo,
	}
}

// =============================================================================
// GENERATE GROCERY LIST TESTS
// =============================================================================

func TestIngredientService_GenerateGroceryList_Success(t *testing.T) {
	setup := setupGroceryServiceTest()
	request := testdata.NewGroceryListRequestBuilder().WithRecipeIDs([]int{1, 2}).Build()

	ingredientsMap := map[int][]models.RecipeIngredient{
		1: {testdata.NewRecipeIngredientBuilder().WithRecipeID(1).WithIngredientID(1).WithQuantity(100.0).WithUnit("grams").Build()},
		2: {testdata.NewRecipeIngredientBuilder().WithRecipeID(2).WithIngredientID(1).WithQuantity(50.0).WithUnit("grams").Build()},
	}

	recipe1 := testdata.NewRecipeBuilder().WithID(1).WithName("Recipe 1").BuildPtr()
	recipe2 := testdata.NewRecipeBuilder().WithID(2).WithName("Recipe 2").BuildPtr()

	setup.ingredientRepo.On("GetIngredientsForRecipes", []int{1, 2}).Return(ingredientsMap, nil)
	setup.recipeRepo.On("GetByID", 1).Return(recipe1, nil)
	setup.recipeRepo.On("GetByID", 2).Return(recipe2, nil)

	result, err := setup.service.GenerateGroceryList(request)

	assert.NoError(t, err)
	assert.Len(t, result, 1)                        // Same ingredient from both recipes should be aggregated
	assert.Equal(t, 150.0, result[0].TotalQuantity) // 100 + 50
	assert.Contains(t, result[0].Recipes, "Recipe 1")
	assert.Contains(t, result[0].Recipes, "Recipe 2")
	setup.ingredientRepo.AssertExpectations(t)
	setup.recipeRepo.AssertExpectations(t)
}

func TestIngredientService_GenerateGroceryList_EmptyRecipeList(t *testing.T) {
	setup := setupGroceryServiceTest()
	request := testdata.NewGroceryListRequestBuilder().WithNoRecipes().Build()

	result, err := setup.service.GenerateGroceryList(request)

	assert.NoError(t, err)
	assert.Empty(t, result)
	setup.ingredientRepo.AssertNotCalled(t, "GetIngredientsForRecipes")
}

func TestIngredientService_GenerateGroceryList_DifferentUnits(t *testing.T) {
	setup := setupGroceryServiceTest()
	request := testdata.NewGroceryListRequestBuilder().WithRecipeIDs([]int{1, 2}).Build()

	// Same ingredient but different units
	ingredientsMap := map[int][]models.RecipeIngredient{
		1: {testdata.NewRecipeIngredientBuilder().WithRecipeID(1).WithIngredientID(1).WithQuantity(100.0).WithUnit("grams").Build()},
		2: {testdata.NewRecipeIngredientBuilder().WithRecipeID(2).WithIngredientID(1).WithQuantity(1.0).WithUnit("cup").Build()},
	}

	recipe1 := testdata.NewRecipeBuilder().WithID(1).WithName("Recipe 1").BuildPtr()
	recipe2 := testdata.NewRecipeBuilder().WithID(2).WithName("Recipe 2").BuildPtr()

	setup.ingredientRepo.On("GetIngredientsForRecipes", []int{1, 2}).Return(ingredientsMap, nil)
	setup.recipeRepo.On("GetByID", 1).Return(recipe1, nil)
	setup.recipeRepo.On("GetByID", 2).Return(recipe2, nil)

	result, err := setup.service.GenerateGroceryList(request)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, -1.0, result[0].TotalQuantity) // Flag for manual calculation
	setup.ingredientRepo.AssertExpectations(t)
	setup.recipeRepo.AssertExpectations(t)
}
