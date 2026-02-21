package service

import (
	"database/sql"
	"errors"
	"fmt"
	"meal-prep/services/recipe-catalogue/domain"
	"testing"

	"meal-prep/services/recipe-catalogue/service/mocks"
	"meal-prep/services/recipe-catalogue/service/testdata"
	"meal-prep/shared/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type recipeServiceTestSetup struct {
	service        RecipeService
	recipeRepo     *mocks.MockRecipeRepository
	categoryRepo   *mocks.MockCategoryRepository
	ingredientRepo *mocks.MockIngredientRepository
}

func setupRecipeServiceTest() *recipeServiceTestSetup {
	recipeRepo := new(mocks.MockRecipeRepository)
	categoryRepo := new(mocks.MockCategoryRepository)
	ingredientRepo := new(mocks.MockIngredientRepository)

	service := NewRecipeService(recipeRepo, categoryRepo, ingredientRepo)

	return &recipeServiceTestSetup{
		service:        service,
		recipeRepo:     recipeRepo,
		categoryRepo:   categoryRepo,
		ingredientRepo: ingredientRepo,
	}
}

// =============================================================================
// GET ALL RECIPES TESTS
// =============================================================================

func TestRecipeService_GetAllRecipes_Success(t *testing.T) {
	setup := setupRecipeServiceTest()
	expectedRecipes := []models.Recipe{
		testdata.NewRecipeBuilder().WithID(1).WithName("Recipe 1").Build(),
		testdata.NewRecipeBuilder().WithID(2).WithName("Recipe 2").Build(),
	}

	setup.recipeRepo.On("GetAll").Return(expectedRecipes, nil)

	result, err := setup.service.GetAllRecipes()

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Recipe 1", result[0].Name)
	setup.recipeRepo.AssertExpectations(t)
}

func TestRecipeService_GetAllRecipes_RepositoryError(t *testing.T) {
	setup := setupRecipeServiceTest()
	expectedError := errors.New("database error")

	setup.recipeRepo.On("GetAll").Return(nil, expectedError)

	result, err := setup.service.GetAllRecipes()

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")
	setup.recipeRepo.AssertExpectations(t)
}

// =============================================================================
// GET RECIPE BY ID TESTS
// =============================================================================

func TestRecipeService_GetRecipeByID_Success(t *testing.T) {
	setup := setupRecipeServiceTest()
	expectedRecipe := testdata.NewRecipeBuilder().WithID(1).BuildPtr()

	setup.recipeRepo.On("GetByID", 1).Return(expectedRecipe, nil)

	result, err := setup.service.GetRecipeByID(1)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.ID)
	setup.recipeRepo.AssertExpectations(t)
}

func TestRecipeService_GetRecipeByID_InvalidID(t *testing.T) {
	invalidIDs := []int{0, -1, -999}

	for _, id := range invalidIDs {
		t.Run(fmt.Sprintf("id_%d", id), func(t *testing.T) {
			setup := setupRecipeServiceTest()

			result, err := setup.service.GetRecipeByID(id)

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Equal(t, domain.ErrRecipeNotFound, err)
			setup.recipeRepo.AssertNotCalled(t, "GetByID")
		})
	}
}

func TestRecipeService_GetRecipeByID_NotFound(t *testing.T) {
	setup := setupRecipeServiceTest()

	setup.recipeRepo.On("GetByID", 999).Return(nil, sql.ErrNoRows)

	result, err := setup.service.GetRecipeByID(999)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrRecipeNotFound, err)
	setup.recipeRepo.AssertExpectations(t)
}

// =============================================================================
// GET RECIPES BY CATEGORY TESTS
// =============================================================================

func TestRecipeService_GetRecipesByCategory_Success(t *testing.T) {
	setup := setupRecipeServiceTest()
	categoryID := 1
	expectedRecipes := []models.Recipe{
		testdata.NewRecipeBuilder().WithID(1).WithCategoryID(categoryID).Build(),
	}

	setup.categoryRepo.On("Exists", categoryID).Return(true, nil)
	setup.recipeRepo.On("GetByCategory", categoryID).Return(expectedRecipes, nil)

	result, err := setup.service.GetRecipesByCategory(categoryID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, categoryID, *result[0].CategoryID)
	setup.categoryRepo.AssertExpectations(t)
	setup.recipeRepo.AssertExpectations(t)
}

func TestRecipeService_GetRecipesByCategory_InvalidCategory(t *testing.T) {
	invalidIDs := []int{0, -1}

	for _, id := range invalidIDs {
		t.Run(fmt.Sprintf("category_%d", id), func(t *testing.T) {
			setup := setupRecipeServiceTest()

			result, err := setup.service.GetRecipesByCategory(id)

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Equal(t, domain.ErrInvalidCategory, err)
			setup.categoryRepo.AssertNotCalled(t, "Exists")
		})
	}
}

func TestRecipeService_GetRecipesByCategory_CategoryNotFound(t *testing.T) {
	setup := setupRecipeServiceTest()
	categoryID := 999

	setup.categoryRepo.On("Exists", categoryID).Return(false, nil)

	result, err := setup.service.GetRecipesByCategory(categoryID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrCategoryNotFound, err)
	setup.categoryRepo.AssertExpectations(t)
}

// =============================================================================
// CREATE RECIPE TESTS
// =============================================================================

func TestRecipeService_CreateRecipe_Success(t *testing.T) {
	setup := setupRecipeServiceTest()
	request := testdata.NewCreateRecipeRequestBuilder().
		WithName("Test Recipe").
		WithCategoryID(1).
		Build()
	expectedRecipe := testdata.NewRecipeBuilder().WithName("Test Recipe").BuildPtr()

	setup.categoryRepo.On("Exists", 1).Return(true, nil)
	setup.recipeRepo.On("Create", 1, mock.AnythingOfType("models.CreateRecipeRequest")).
		Return(expectedRecipe, nil)

	result, err := setup.service.CreateRecipe(1, request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Recipe", result.Name)
	setup.categoryRepo.AssertExpectations(t)
	setup.recipeRepo.AssertExpectations(t)
}

func TestRecipeService_CreateRecipe_ValidationErrors(t *testing.T) {
	testCases := []struct {
		name          string
		request       models.CreateRecipeRequest
		expectedError error
	}{
		{
			name:          "empty_name",
			request:       testdata.NewCreateRecipeRequestBuilder().WithName("").Build(),
			expectedError: domain.ErrRecipeNameRequired,
		},
		{
			name:          "whitespace_name",
			request:       testdata.NewCreateRecipeRequestBuilder().WithName("   ").Build(),
			expectedError: domain.ErrRecipeNameRequired,
		},
		{
			name:          "zero_category",
			request:       testdata.NewCreateRecipeRequestBuilder().WithCategoryID(0).Build(),
			expectedError: domain.ErrInvalidCategory,
		},
		{
			name:          "negative_category",
			request:       testdata.NewCreateRecipeRequestBuilder().WithCategoryID(-1).Build(),
			expectedError: domain.ErrInvalidCategory,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setup := setupRecipeServiceTest()

			result, err := setup.service.CreateRecipe(1, tc.request)

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Equal(t, tc.expectedError, err)
			setup.categoryRepo.AssertNotCalled(t, "Exists")
			setup.recipeRepo.AssertNotCalled(t, "Create")
		})
	}
}

func TestRecipeService_CreateRecipe_CategoryNotFound(t *testing.T) {
	setup := setupRecipeServiceTest()
	request := testdata.NewCreateRecipeRequestBuilder().WithCategoryID(999).Build()

	setup.categoryRepo.On("Exists", 999).Return(false, nil)

	result, err := setup.service.CreateRecipe(1, request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrCategoryNotFound, err)
	setup.categoryRepo.AssertExpectations(t)
	setup.recipeRepo.AssertNotCalled(t, "Create")
}

// =============================================================================
// UPDATE RECIPE TESTS
// =============================================================================

func TestRecipeService_UpdateRecipe_Success(t *testing.T) {
	setup := setupRecipeServiceTest()
	recipeID := 1
	request := testdata.NewUpdateRecipeRequestBuilder().WithName("Updated Recipe").Build()
	updatedRecipe := testdata.NewRecipeBuilder().WithID(recipeID).WithName("Updated Recipe").BuildPtr()

	setup.recipeRepo.On("GetOwnerID", recipeID).Return(1, nil)
	setup.categoryRepo.On("Exists", request.CategoryID).Return(true, nil)
	setup.recipeRepo.On("Update", recipeID, mock.AnythingOfType("models.UpdateRecipeRequest")).
		Return(updatedRecipe, nil)

	result, err := setup.service.UpdateRecipe(1, recipeID, request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Updated Recipe", result.Name)
	setup.recipeRepo.AssertExpectations(t)
	setup.categoryRepo.AssertExpectations(t)
}

func TestRecipeService_UpdateRecipe_RequiresName(t *testing.T) {
	setup := setupRecipeServiceTest()
	recipeID := 1
	updateReq := models.UpdateRecipeRequest{
		Name:        "", // Should fail validation
		Description: "Some description",
		CategoryID:  1,
	}

	result, err := setup.service.UpdateRecipe(1, recipeID, updateReq)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrRecipeNameRequired, err)
	setup.recipeRepo.AssertNotCalled(t, "GetOwnerID")
}

func TestRecipeService_UpdateRecipe_InvalidCategoryId(t *testing.T) {
	setup := setupRecipeServiceTest()
	recipeID := 1
	updateReq := models.UpdateRecipeRequest{
		Name:        "Some name",
		Description: "",
		CategoryID:  0, // Should fail validation
	}

	result, err := setup.service.UpdateRecipe(1, recipeID, updateReq)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrInvalidCategory, err)
	setup.recipeRepo.AssertNotCalled(t, "GetOwnerID")
}

func TestRecipeService_UpdateRecipe_InvalidID(t *testing.T) {
	invalidIDs := []int{0, -1}

	for _, id := range invalidIDs {
		t.Run(fmt.Sprintf("id_%d", id), func(t *testing.T) {
			setup := setupRecipeServiceTest()
			request := testdata.NewUpdateRecipeRequestBuilder().Build()

			result, err := setup.service.UpdateRecipe(1, id, request)

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Equal(t, domain.ErrRecipeNotFound, err)
			setup.recipeRepo.AssertNotCalled(t, "GetOwnerID")
		})
	}
}

func TestRecipeService_UpdateRecipe_RecipeNotFound(t *testing.T) {
	setup := setupRecipeServiceTest()
	recipeID := 999
	request := testdata.NewUpdateRecipeRequestBuilder().Build()

	setup.recipeRepo.On("GetOwnerID", recipeID).Return(0, sql.ErrNoRows)

	result, err := setup.service.UpdateRecipe(1, recipeID, request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrRecipeNotFound, err)
	setup.recipeRepo.AssertExpectations(t)
}

// =============================================================================
// DELETE RECIPE TESTS
// =============================================================================

func TestRecipeService_DeleteRecipe_Success(t *testing.T) {
	setup := setupRecipeServiceTest()
	recipeID := 1

	setup.recipeRepo.On("GetOwnerID", recipeID).Return(1, nil)
	setup.recipeRepo.On("Delete", recipeID).Return(nil)

	err := setup.service.DeleteRecipe(1, recipeID)

	assert.NoError(t, err)
	setup.recipeRepo.AssertExpectations(t)
}

func TestRecipeService_DeleteRecipe_InvalidID(t *testing.T) {
	invalidIDs := []int{0, -1}

	for _, id := range invalidIDs {
		t.Run(fmt.Sprintf("id_%d", id), func(t *testing.T) {
			setup := setupRecipeServiceTest()

			err := setup.service.DeleteRecipe(1, id)

			assert.Error(t, err)
			assert.Equal(t, domain.ErrRecipeNotFound, err)
			setup.recipeRepo.AssertNotCalled(t, "Delete")
		})
	}
}

func TestRecipeService_DeleteRecipe_NotFound(t *testing.T) {
	setup := setupRecipeServiceTest()
	recipeID := 999

	setup.recipeRepo.On("GetOwnerID", recipeID).Return(0, sql.ErrNoRows)

	err := setup.service.DeleteRecipe(1, recipeID)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrRecipeNotFound, err)
	setup.recipeRepo.AssertExpectations(t)
}

// =============================================================================
// OWNERSHIP TESTS
// =============================================================================

func TestRecipeService_UpdateRecipe_Forbidden(t *testing.T) {
	setup := setupRecipeServiceTest()
	recipeID := 1
	request := testdata.NewUpdateRecipeRequestBuilder().WithName("Updated Recipe").Build()

	// Recipe exists but is owned by user 2, not the caller (user 1)
	setup.recipeRepo.On("GetOwnerID", recipeID).Return(2, nil)

	result, err := setup.service.UpdateRecipe(1, recipeID, request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrForbidden, err)
	setup.recipeRepo.AssertNotCalled(t, "Update")
}

func TestRecipeService_DeleteRecipe_Forbidden(t *testing.T) {
	setup := setupRecipeServiceTest()
	recipeID := 1

	// Recipe exists but is owned by user 2, not the caller (user 1)
	setup.recipeRepo.On("GetOwnerID", recipeID).Return(2, nil)

	err := setup.service.DeleteRecipe(1, recipeID)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrForbidden, err)
	setup.recipeRepo.AssertNotCalled(t, "Delete")
}

// =============================================================================
// GET ALL CATEGORIES TESTS
// =============================================================================

func TestRecipeService_GetAllCategories_Success(t *testing.T) {
	setup := setupRecipeServiceTest()
	expectedCategories := []models.Category{
		testdata.NewCategoryBuilder().WithID(1).WithName("Italian").Build(),
		testdata.NewCategoryBuilder().WithID(2).WithName("Mexican").Build(),
	}

	setup.categoryRepo.On("GetAll").Return(expectedCategories, nil)

	result, err := setup.service.GetAllCategories()

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Italian", result[0].Name)
	setup.categoryRepo.AssertExpectations(t)
}

func TestRecipeService_GetAllCategories_RepositoryError(t *testing.T) {
	setup := setupRecipeServiceTest()
	expectedError := errors.New("database error")

	setup.categoryRepo.On("GetAll").Return(nil, expectedError)

	result, err := setup.service.GetAllCategories()

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")
	setup.categoryRepo.AssertExpectations(t)
}

// =============================================================================
// GET ALL RECIPES WITH INGREDIENTS TESTS
// =============================================================================

func TestRecipeService_GetAllRecipesWithIngredients_Success(t *testing.T) {
	setup := setupRecipeServiceTest()
	expectedRecipes := []models.RecipeWithIngredients{
		testdata.NewRecipeWithIngredientsBuilder().Build(),
	}

	setup.recipeRepo.On("GetAllWithIngredients").Return(expectedRecipes, nil)

	result, err := setup.service.GetAllRecipesWithIngredients()

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	setup.recipeRepo.AssertExpectations(t)
}

func TestRecipeService_GetAllRecipesWithIngredients_RepositoryError(t *testing.T) {
	setup := setupRecipeServiceTest()
	expectedError := errors.New("database error")

	setup.recipeRepo.On("GetAllWithIngredients").Return(nil, expectedError)

	result, err := setup.service.GetAllRecipesWithIngredients()

	assert.Error(t, err)
	assert.Nil(t, result)
	setup.recipeRepo.AssertExpectations(t)
}

// =============================================================================
// GET RECIPE BY ID WITH INGREDIENTS TESTS
// =============================================================================

func TestRecipeService_GetRecipeByIDWithIngredients_Success(t *testing.T) {
	setup := setupRecipeServiceTest()
	expectedRecipe := testdata.NewRecipeWithIngredientsBuilder().BuildPtr()

	setup.recipeRepo.On("GetByIDWithIngredients", 1).Return(expectedRecipe, nil)

	result, err := setup.service.GetRecipeByIDWithIngredients(1)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	setup.recipeRepo.AssertExpectations(t)
}

func TestRecipeService_GetRecipeByIDWithIngredients_InvalidID(t *testing.T) {
	setup := setupRecipeServiceTest()

	result, err := setup.service.GetRecipeByIDWithIngredients(0)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrRecipeNotFound, err)
	setup.recipeRepo.AssertNotCalled(t, "GetByIDWithIngredients")
}

func TestRecipeService_GetRecipeByIDWithIngredients_NotFound(t *testing.T) {
	setup := setupRecipeServiceTest()

	setup.recipeRepo.On("GetByIDWithIngredients", 999).Return(nil, sql.ErrNoRows)

	result, err := setup.service.GetRecipeByIDWithIngredients(999)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrRecipeNotFound, err)
	setup.recipeRepo.AssertExpectations(t)
}

// =============================================================================
// GET RECIPES BY CATEGORY WITH INGREDIENTS TESTS
// =============================================================================

func TestRecipeService_GetRecipesByCategoryWithIngredients_Success(t *testing.T) {
	setup := setupRecipeServiceTest()
	categoryID := 1
	expectedRecipes := []models.RecipeWithIngredients{
		testdata.NewRecipeWithIngredientsBuilder().Build(),
	}

	setup.categoryRepo.On("Exists", categoryID).Return(true, nil)
	setup.recipeRepo.On("GetByCategoryWithIngredients", categoryID).Return(expectedRecipes, nil)

	result, err := setup.service.GetRecipesByCategoryWithIngredients(categoryID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	setup.categoryRepo.AssertExpectations(t)
	setup.recipeRepo.AssertExpectations(t)
}

func TestRecipeService_GetRecipesByCategoryWithIngredients_InvalidCategory(t *testing.T) {
	setup := setupRecipeServiceTest()

	result, err := setup.service.GetRecipesByCategoryWithIngredients(0)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrInvalidCategory, err)
	setup.categoryRepo.AssertNotCalled(t, "Exists")
}

// =============================================================================
// CREATE RECIPE WITH INGREDIENTS TESTS
// =============================================================================

func TestRecipeService_CreateRecipeWithIngredients_Success(t *testing.T) {
	setup := setupRecipeServiceTest()
	request := models.CreateRecipeWithIngredientsRequest{
		Name:       "Recipe with Ingredients",
		CategoryID: 1,
		Ingredients: []models.AddRecipeIngredientRequest{
			{IngredientID: 1, Quantity: 100.0, Unit: "grams"},
		},
	}
	expectedResult := testdata.NewRecipeWithIngredientsBuilder().BuildPtr()

	setup.categoryRepo.On("Exists", 1).Return(true, nil)
	setup.ingredientRepo.On("IngredientExists", 1).Return(true, nil)
	setup.recipeRepo.On("CreateWithIngredients", 1, mock.AnythingOfType("models.CreateRecipeWithIngredientsRequest")).
		Return(expectedResult, nil)

	result, err := setup.service.CreateRecipeWithIngredients(1, request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	setup.categoryRepo.AssertExpectations(t)
	setup.ingredientRepo.AssertExpectations(t)
	setup.recipeRepo.AssertExpectations(t)
}

func TestRecipeService_CreateRecipeWithIngredients_ValidationErrors(t *testing.T) {
	testCases := []struct {
		name          string
		request       models.CreateRecipeWithIngredientsRequest
		expectedError error
		setupMocks    func(*recipeServiceTestSetup) // Add setup function for conditional mocking
	}{
		{
			name:          "empty_name",
			request:       models.CreateRecipeWithIngredientsRequest{Name: "", CategoryID: 1},
			expectedError: domain.ErrRecipeNameRequired,
			setupMocks:    func(setup *recipeServiceTestSetup) {}, // No mocks needed - fails before category check
		},
		{
			name:          "invalid_category",
			request:       models.CreateRecipeWithIngredientsRequest{Name: "Valid", CategoryID: 0},
			expectedError: domain.ErrInvalidCategory,
			setupMocks:    func(setup *recipeServiceTestSetup) {}, // No mocks needed - fails on CategoryID <= 0
		},
		{
			name: "invalid_ingredient_id",
			request: models.CreateRecipeWithIngredientsRequest{
				Name:       "Valid",
				CategoryID: 1,
				Ingredients: []models.AddRecipeIngredientRequest{
					{IngredientID: 0, Quantity: 100.0, Unit: "grams"},
				},
			},
			expectedError: domain.ErrIngredientNotFound,
			setupMocks: func(setup *recipeServiceTestSetup) {
				// Need to mock category exists since CategoryID: 1 is valid
				setup.categoryRepo.On("Exists", 1).Return(true, nil)
			},
		},
		{
			name: "invalid_quantity",
			request: models.CreateRecipeWithIngredientsRequest{
				Name:       "Valid",
				CategoryID: 1,
				Ingredients: []models.AddRecipeIngredientRequest{
					{IngredientID: 1, Quantity: 0, Unit: "grams"},
				},
			},
			expectedError: domain.ErrInvalidQuantity,
			setupMocks: func(setup *recipeServiceTestSetup) {
				// Need to mock category exists since CategoryID: 1 is valid
				setup.categoryRepo.On("Exists", 1).Return(true, nil)
			},
		},
		{
			name: "empty_unit",
			request: models.CreateRecipeWithIngredientsRequest{
				Name:       "Valid",
				CategoryID: 1,
				Ingredients: []models.AddRecipeIngredientRequest{
					{IngredientID: 1, Quantity: 100.0, Unit: ""},
				},
			},
			expectedError: domain.ErrInvalidUnit,
			setupMocks: func(setup *recipeServiceTestSetup) {
				// Need to mock category exists since CategoryID: 1 is valid
				setup.categoryRepo.On("Exists", 1).Return(true, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setup := setupRecipeServiceTest()

			// Set up conditional mocks based on test case
			tc.setupMocks(setup)

			result, err := setup.service.CreateRecipeWithIngredients(1, tc.request)

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Equal(t, tc.expectedError, err)

			// Verify all mock expectations were met
			setup.categoryRepo.AssertExpectations(t)
			setup.ingredientRepo.AssertExpectations(t)
			setup.recipeRepo.AssertExpectations(t)
		})
	}
}

// =============================================================================
// UPDATE RECIPE WITH INGREDIENTS TESTS
// =============================================================================

func TestRecipeService_UpdateRecipeWithIngredients_Success(t *testing.T) {
	setup := setupRecipeServiceTest()
	recipeID := 1
	request := models.UpdateRecipeWithIngredientsRequest{
		Name:       "Recipe with Ingredients",
		CategoryID: 1,
		Ingredients: []models.AddRecipeIngredientRequest{
			{IngredientID: 1, Quantity: 150.0, Unit: "grams"},
		},
	}
	expectedResult := testdata.NewRecipeWithIngredientsBuilder().BuildPtr()

	setup.recipeRepo.On("GetOwnerID", recipeID).Return(1, nil)
	setup.categoryRepo.On("Exists", 1).Return(true, nil)
	setup.ingredientRepo.On("IngredientExists", 1).Return(true, nil)
	setup.recipeRepo.On("UpdateWithIngredients", recipeID, mock.AnythingOfType("models.UpdateRecipeWithIngredientsRequest")).
		Return(expectedResult, nil)

	result, err := setup.service.UpdateRecipeWithIngredients(1, recipeID, request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	setup.recipeRepo.AssertExpectations(t)
	setup.ingredientRepo.AssertExpectations(t)
}

func TestRecipeService_UpdateRecipeWithIngredients_InvalidID(t *testing.T) {
	setup := setupRecipeServiceTest()
	request := models.UpdateRecipeWithIngredientsRequest{}

	result, err := setup.service.UpdateRecipeWithIngredients(1, 0, request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrRecipeNotFound, err)
	setup.recipeRepo.AssertNotCalled(t, "GetByID")
}

// =============================================================================
// SEARCH RECIPES BY INGREDIENTS TESTS
// =============================================================================

func TestRecipeService_SearchRecipesByIngredients_Success(t *testing.T) {
	setup := setupRecipeServiceTest()
	ingredientIDs := []int{1, 2}
	expectedRecipes := []models.Recipe{
		testdata.NewRecipeBuilder().WithID(1).Build(),
	}

	setup.ingredientRepo.On("IngredientExists", 1).Return(true, nil)
	setup.ingredientRepo.On("IngredientExists", 2).Return(true, nil)
	setup.recipeRepo.On("SearchRecipesByIngredients", ingredientIDs).Return(expectedRecipes, nil)

	result, err := setup.service.SearchRecipesByIngredients(ingredientIDs)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	setup.ingredientRepo.AssertExpectations(t)
	setup.recipeRepo.AssertExpectations(t)
}

func TestRecipeService_SearchRecipesByIngredients_EmptyInput(t *testing.T) {
	setup := setupRecipeServiceTest()

	result, err := setup.service.SearchRecipesByIngredients([]int{})

	assert.NoError(t, err)
	assert.Empty(t, result)
	setup.ingredientRepo.AssertNotCalled(t, "IngredientExists")
	setup.recipeRepo.AssertNotCalled(t, "SearchRecipesByIngredients")
}

func TestRecipeService_SearchRecipesByIngredients_InvalidIngredientID(t *testing.T) {
	setup := setupRecipeServiceTest()
	ingredientIDs := []int{0, 1}

	result, err := setup.service.SearchRecipesByIngredients(ingredientIDs)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrIngredientNotFound, err)
	setup.ingredientRepo.AssertNotCalled(t, "IngredientExists")
}

func TestRecipeService_SearchRecipesByIngredients_IngredientNotFound(t *testing.T) {
	setup := setupRecipeServiceTest()
	ingredientIDs := []int{1, 999}

	setup.ingredientRepo.On("IngredientExists", 1).Return(true, nil)
	setup.ingredientRepo.On("IngredientExists", 999).Return(false, nil)

	result, err := setup.service.SearchRecipesByIngredients(ingredientIDs)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrIngredientNotFound, err)
	setup.ingredientRepo.AssertExpectations(t)
	setup.recipeRepo.AssertNotCalled(t, "SearchRecipesByIngredients")
}

// =============================================================================
// SEARCH RECIPES BY INGREDIENTS WITH INGREDIENTS TESTS
// =============================================================================

func TestRecipeService_SearchRecipesByIngredientsWithIngredients_Success(t *testing.T) {
	setup := setupRecipeServiceTest()
	ingredientIDs := []int{1, 2}
	expectedRecipes := []models.RecipeWithIngredients{
		testdata.NewRecipeWithIngredientsBuilder().Build(),
	}

	setup.ingredientRepo.On("IngredientExists", 1).Return(true, nil)
	setup.ingredientRepo.On("IngredientExists", 2).Return(true, nil)
	setup.recipeRepo.On("SearchRecipesByIngredientsWithIngredients", ingredientIDs).
		Return(expectedRecipes, nil)

	result, err := setup.service.SearchRecipesByIngredientsWithIngredients(ingredientIDs)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	setup.ingredientRepo.AssertExpectations(t)
	setup.recipeRepo.AssertExpectations(t)
}

func TestRecipeService_SearchRecipesByIngredientsWithIngredients_EmptyInput(t *testing.T) {
	setup := setupRecipeServiceTest()

	result, err := setup.service.SearchRecipesByIngredientsWithIngredients([]int{})

	assert.NoError(t, err)
	assert.Empty(t, result)
	setup.ingredientRepo.AssertNotCalled(t, "IngredientExists")
	setup.recipeRepo.AssertNotCalled(t, "SearchRecipesByIngredientsWithIngredients")
}
