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

type ingredientServiceTestSetup struct {
	service        IngredientService
	ingredientRepo *mocks.MockIngredientRepository
	recipeRepo     *mocks.MockRecipeRepository
}

func setupIngredientServiceTest() *ingredientServiceTestSetup {
	ingredientRepo := new(mocks.MockIngredientRepository)
	recipeRepo := new(mocks.MockRecipeRepository)

	service := NewIngredientService(ingredientRepo, recipeRepo)

	return &ingredientServiceTestSetup{
		service:        service,
		ingredientRepo: ingredientRepo,
		recipeRepo:     recipeRepo,
	}
}

// =============================================================================
// GET ALL INGREDIENTS TESTS
// =============================================================================

func TestIngredientService_GetAllIngredients_Success(t *testing.T) {
	setup := setupIngredientServiceTest()
	expectedIngredients := []models.Ingredient{
		testdata.NewIngredientBuilder().WithID(1).WithName("Tomato").Build(),
		testdata.NewIngredientBuilder().WithID(2).WithName("Basil").Build(),
	}

	setup.ingredientRepo.On("GetAllIngredients", defaultParams).Return(expectedIngredients, 2, nil)

	result, meta, err := setup.service.GetAllIngredients(defaultParams)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Tomato", result[0].Name)
	assert.Equal(t, 2, meta.Total)
	setup.ingredientRepo.AssertExpectations(t)
}

func TestIngredientService_GetAllIngredients_RepositoryError(t *testing.T) {
	setup := setupIngredientServiceTest()
	expectedError := errors.New("database error")

	setup.ingredientRepo.On("GetAllIngredients", defaultParams).Return(nil, 0, expectedError)

	result, _, err := setup.service.GetAllIngredients(defaultParams)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")
	setup.ingredientRepo.AssertExpectations(t)
}

// =============================================================================
// GET INGREDIENT BY ID TESTS
// =============================================================================

func TestIngredientService_GetIngredientByID_Success(t *testing.T) {
	setup := setupIngredientServiceTest()
	expectedIngredient := testdata.NewIngredientBuilder().WithID(1).BuildPtr()

	setup.ingredientRepo.On("GetIngredientByID", 1).Return(expectedIngredient, nil)

	result, err := setup.service.GetIngredientByID(1)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.ID)
	setup.ingredientRepo.AssertExpectations(t)
}

func TestIngredientService_GetIngredientByID_InvalidID(t *testing.T) {
	invalidIDs := []int{0, -1, -999}

	for _, id := range invalidIDs {
		t.Run(fmt.Sprintf("id_%d", id), func(t *testing.T) {
			setup := setupIngredientServiceTest()

			result, err := setup.service.GetIngredientByID(id)

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Equal(t, domain.ErrIngredientNotFound, err)
			setup.ingredientRepo.AssertNotCalled(t, "GetIngredientByID")
		})
	}
}

func TestIngredientService_GetIngredientByID_NotFound(t *testing.T) {
	setup := setupIngredientServiceTest()

	setup.ingredientRepo.On("GetIngredientByID", 999).Return(nil, sql.ErrNoRows)

	result, err := setup.service.GetIngredientByID(999)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrIngredientNotFound, err)
	setup.ingredientRepo.AssertExpectations(t)
}

// =============================================================================
// GET INGREDIENTS BY CATEGORY TESTS
// =============================================================================

func TestIngredientService_GetIngredientsByCategory_Success(t *testing.T) {
	setup := setupIngredientServiceTest()
	category := "Vegetable"
	expectedIngredients := []models.Ingredient{
		testdata.NewIngredientBuilder().WithCategory(category).Build(),
	}

	setup.ingredientRepo.On("GetIngredientsByCategory", category, defaultParams).Return(expectedIngredients, 1, nil)

	result, meta, err := setup.service.GetIngredientsByCategory(category, defaultParams)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, category, *result[0].Category)
	assert.Equal(t, 1, meta.Total)
	setup.ingredientRepo.AssertExpectations(t)
}

func TestIngredientService_GetIngredientsByCategory_EmptyCategory(t *testing.T) {
	setup := setupIngredientServiceTest()
	expectedIngredients := []models.Ingredient{
		testdata.NewIngredientBuilder().Build(),
	}

	setup.ingredientRepo.On("GetAllIngredients", defaultParams).Return(expectedIngredients, 1, nil)

	result, _, err := setup.service.GetIngredientsByCategory("", defaultParams)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	setup.ingredientRepo.AssertExpectations(t)
}

func TestIngredientService_GetIngredientsByCategory_WhitespaceCategory(t *testing.T) {
	setup := setupIngredientServiceTest()
	expectedIngredients := []models.Ingredient{
		testdata.NewIngredientBuilder().Build(),
	}

	setup.ingredientRepo.On("GetAllIngredients", defaultParams).Return(expectedIngredients, 1, nil)

	result, _, err := setup.service.GetIngredientsByCategory("   ", defaultParams)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	setup.ingredientRepo.AssertExpectations(t)
}

// =============================================================================
// SEARCH INGREDIENTS TESTS
// =============================================================================

func TestIngredientService_SearchIngredients_Success(t *testing.T) {
	setup := setupIngredientServiceTest()
	query := "tomato"
	expectedIngredients := []models.Ingredient{
		testdata.NewIngredientBuilder().WithName("Cherry Tomato").Build(),
	}

	setup.ingredientRepo.On("SearchIngredients", query, defaultParams).Return(expectedIngredients, 1, nil)

	result, meta, err := setup.service.SearchIngredients(query, defaultParams)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, meta.Total)
	setup.ingredientRepo.AssertExpectations(t)
}

func TestIngredientService_SearchIngredients_EmptyQuery(t *testing.T) {
	setup := setupIngredientServiceTest()
	expectedIngredients := []models.Ingredient{
		testdata.NewIngredientBuilder().Build(),
	}

	setup.ingredientRepo.On("GetAllIngredients", defaultParams).Return(expectedIngredients, 1, nil)

	result, _, err := setup.service.SearchIngredients("", defaultParams)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	setup.ingredientRepo.AssertExpectations(t)
}

// =============================================================================
// CREATE INGREDIENT TESTS
// =============================================================================

func TestIngredientService_CreateIngredient_Success(t *testing.T) {
	setup := setupIngredientServiceTest()
	request := testdata.NewCreateIngredientRequestBuilder().WithName("New Ingredient").Build()
	expectedIngredient := testdata.NewIngredientBuilder().WithName("New Ingredient").BuildPtr()

	setup.ingredientRepo.On("CreateIngredient", mock.AnythingOfType("models.CreateIngredientRequest")).
		Return(expectedIngredient, nil)

	result, err := setup.service.CreateIngredient(request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "New Ingredient", result.Name)
	setup.ingredientRepo.AssertExpectations(t)
}

func TestIngredientService_CreateIngredient_ValidationErrors(t *testing.T) {
	testCases := []struct {
		name          string
		request       models.CreateIngredientRequest
		expectedError error
	}{
		{
			name:          "empty_name",
			request:       testdata.NewCreateIngredientRequestBuilder().WithName("").Build(),
			expectedError: domain.ErrIngredientNameRequired,
		},
		{
			name:          "whitespace_name",
			request:       testdata.NewCreateIngredientRequestBuilder().WithName("   ").Build(),
			expectedError: domain.ErrIngredientNameRequired,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setup := setupIngredientServiceTest()

			result, err := setup.service.CreateIngredient(tc.request)

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Equal(t, tc.expectedError, err)
			setup.ingredientRepo.AssertNotCalled(t, "CreateIngredient")
		})
	}
}

func TestIngredientService_CreateIngredient_RepositoryError(t *testing.T) {
	setup := setupIngredientServiceTest()
	request := testdata.NewCreateIngredientRequestBuilder().Build()
	expectedError := errors.New("database error")

	setup.ingredientRepo.On("CreateIngredient", mock.AnythingOfType("models.CreateIngredientRequest")).
		Return(nil, expectedError)

	result, err := setup.service.CreateIngredient(request)

	assert.Error(t, err)
	assert.Nil(t, result)
	setup.ingredientRepo.AssertExpectations(t)
}

// =============================================================================
// UPDATE INGREDIENT TESTS
// =============================================================================

func TestIngredientService_UpdateIngredient_Success(t *testing.T) {
	setup := setupIngredientServiceTest()
	ingredientID := 1
	request := testdata.NewUpdateIngredientRequestBuilder().WithName("Updated Ingredient").Build()
	existingIngredient := testdata.NewIngredientBuilder().WithID(ingredientID).BuildPtr()
	updatedIngredient := testdata.NewIngredientBuilder().WithID(ingredientID).WithName("Updated Ingredient").BuildPtr()

	setup.ingredientRepo.On("GetIngredientByID", ingredientID).Return(existingIngredient, nil)
	setup.ingredientRepo.On("UpdateIngredient", ingredientID, mock.AnythingOfType("models.UpdateIngredientRequest")).
		Return(updatedIngredient, nil)

	result, err := setup.service.UpdateIngredient(ingredientID, request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Updated Ingredient", result.Name)
	setup.ingredientRepo.AssertExpectations(t)
}

func TestIngredientService_UpdateIngredient_InvalidID(t *testing.T) {
	invalidIDs := []int{0, -1}

	for _, id := range invalidIDs {
		t.Run(fmt.Sprintf("id_%d", id), func(t *testing.T) {
			setup := setupIngredientServiceTest()
			request := testdata.NewUpdateIngredientRequestBuilder().Build()

			result, err := setup.service.UpdateIngredient(id, request)

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Equal(t, domain.ErrIngredientNotFound, err)
			setup.ingredientRepo.AssertNotCalled(t, "GetIngredientByID")
		})
	}
}

func TestIngredientService_UpdateIngredient_NotFound(t *testing.T) {
	setup := setupIngredientServiceTest()
	ingredientID := 999
	request := testdata.NewUpdateIngredientRequestBuilder().Build()

	setup.ingredientRepo.On("GetIngredientByID", ingredientID).Return(nil, sql.ErrNoRows)

	result, err := setup.service.UpdateIngredient(ingredientID, request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrIngredientNotFound, err)
	setup.ingredientRepo.AssertExpectations(t)
}

func TestIngredientService_UpdateIngredient_ValidationErrors(t *testing.T) {
	setup := setupIngredientServiceTest()
	ingredientID := 1
	existingIngredient := testdata.NewIngredientBuilder().WithID(ingredientID).BuildPtr()

	testCases := []struct {
		name          string
		request       models.UpdateIngredientRequest
		expectedError error
	}{
		{
			name: "empty_name",
			request: models.UpdateIngredientRequest{
				Name: "",
			},
			expectedError: domain.ErrIngredientNameRequired,
		},
		{
			name: "whitespace_name",
			request: models.UpdateIngredientRequest{
				Name: "   ",
			},
			expectedError: domain.ErrIngredientNameRequired,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setup.ingredientRepo.On("GetIngredientByID", ingredientID).Return(existingIngredient, nil)

			result, err := setup.service.UpdateIngredient(ingredientID, tc.request)

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Equal(t, tc.expectedError, err)
			setup.ingredientRepo.AssertExpectations(t)
		})
	}
}

// =============================================================================
// DELETE INGREDIENT TESTS
// =============================================================================

func TestIngredientService_DeleteIngredient_Success(t *testing.T) {
	setup := setupIngredientServiceTest()
	ingredientID := 1
	existingIngredient := testdata.NewIngredientBuilder().WithID(ingredientID).BuildPtr()
	checkParams := models.PaginationParams{Page: 1, PerPage: 1}

	setup.ingredientRepo.On("GetIngredientByID", ingredientID).Return(existingIngredient, nil)
	setup.ingredientRepo.On("GetRecipesUsingIngredient", ingredientID, checkParams).Return([]models.Recipe{}, 0, nil)
	setup.ingredientRepo.On("DeleteIngredient", ingredientID).Return(nil)

	err := setup.service.DeleteIngredient(ingredientID)

	assert.NoError(t, err)
	setup.ingredientRepo.AssertExpectations(t)
}

func TestIngredientService_DeleteIngredient_InvalidID(t *testing.T) {
	invalidIDs := []int{0, -1}

	for _, id := range invalidIDs {
		t.Run(fmt.Sprintf("id_%d", id), func(t *testing.T) {
			setup := setupIngredientServiceTest()

			err := setup.service.DeleteIngredient(id)

			assert.Error(t, err)
			assert.Equal(t, domain.ErrIngredientNotFound, err)
			setup.ingredientRepo.AssertNotCalled(t, "GetIngredientByID")
		})
	}
}

func TestIngredientService_DeleteIngredient_NotFound(t *testing.T) {
	setup := setupIngredientServiceTest()
	ingredientID := 999

	setup.ingredientRepo.On("GetIngredientByID", ingredientID).Return(nil, sql.ErrNoRows)

	err := setup.service.DeleteIngredient(ingredientID)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrIngredientNotFound, err)
	setup.ingredientRepo.AssertExpectations(t)
}

func TestIngredientService_DeleteIngredient_UsedInRecipes(t *testing.T) {
	setup := setupIngredientServiceTest()
	ingredientID := 1
	existingIngredient := testdata.NewIngredientBuilder().WithID(ingredientID).BuildPtr()
	checkParams := models.PaginationParams{Page: 1, PerPage: 1}

	setup.ingredientRepo.On("GetIngredientByID", ingredientID).Return(existingIngredient, nil)
	setup.ingredientRepo.On("GetRecipesUsingIngredient", ingredientID, checkParams).Return(nil, 1, nil)

	err := setup.service.DeleteIngredient(ingredientID)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrCannotDeleteIngredient, err)
	setup.ingredientRepo.AssertExpectations(t)
	setup.ingredientRepo.AssertNotCalled(t, "DeleteIngredient")
}

// =============================================================================
// GET RECIPE INGREDIENTS TESTS
// =============================================================================

func TestIngredientService_GetRecipeIngredients_Success(t *testing.T) {
	setup := setupIngredientServiceTest()
	recipeID := 1
	existingRecipe := testdata.NewRecipeBuilder().WithID(recipeID).BuildPtr()
	expectedIngredients := []models.RecipeIngredient{
		testdata.NewRecipeIngredientBuilder().WithRecipeID(recipeID).Build(),
	}

	setup.recipeRepo.On("GetByID", recipeID).Return(existingRecipe, nil)
	setup.ingredientRepo.On("GetRecipeIngredients", recipeID).Return(expectedIngredients, nil)

	result, err := setup.service.GetRecipeIngredients(recipeID)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, recipeID, result[0].RecipeID)
	setup.recipeRepo.AssertExpectations(t)
	setup.ingredientRepo.AssertExpectations(t)
}

func TestIngredientService_GetRecipeIngredients_InvalidRecipeID(t *testing.T) {
	setup := setupIngredientServiceTest()

	result, err := setup.service.GetRecipeIngredients(0)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrRecipeNotFound, err)
	setup.recipeRepo.AssertNotCalled(t, "GetByID")
}

func TestIngredientService_GetRecipeIngredients_RecipeNotFound(t *testing.T) {
	setup := setupIngredientServiceTest()
	recipeID := 999

	setup.recipeRepo.On("GetByID", recipeID).Return(nil, sql.ErrNoRows)

	result, err := setup.service.GetRecipeIngredients(recipeID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrRecipeNotFound, err)
	setup.recipeRepo.AssertExpectations(t)
}

// =============================================================================
// ADD RECIPE INGREDIENT TESTS
// =============================================================================

func TestIngredientService_AddRecipeIngredient_Success(t *testing.T) {
	setup := setupIngredientServiceTest()
	recipeID := 1
	request := testdata.NewAddRecipeIngredientRequestBuilder().Build()
	existingRecipe := testdata.NewRecipeBuilder().WithID(recipeID).BuildPtr()
	expectedRecipeIngredient := testdata.NewRecipeIngredientBuilder().BuildPtr()

	setup.recipeRepo.On("GetByID", recipeID).Return(existingRecipe, nil)
	setup.ingredientRepo.On("IngredientExists", request.IngredientID).Return(true, nil)
	setup.ingredientRepo.On("AddRecipeIngredient", recipeID, mock.AnythingOfType("models.AddRecipeIngredientRequest")).
		Return(expectedRecipeIngredient, nil)

	result, err := setup.service.AddRecipeIngredient(recipeID, request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	setup.recipeRepo.AssertExpectations(t)
	setup.ingredientRepo.AssertExpectations(t)
}

func TestIngredientService_AddRecipeIngredient_ValidationErrors(t *testing.T) {
	setup := setupIngredientServiceTest()
	recipeID := 1

	testCases := []struct {
		name          string
		request       models.AddRecipeIngredientRequest
		expectedError error
	}{
		{
			name: "invalid_ingredient_id",
			request: models.AddRecipeIngredientRequest{
				IngredientID: 0,
				Quantity:     100.0,
				Unit:         "grams",
			},
			expectedError: domain.ErrIngredientNotFound,
		},
		{
			name: "invalid_quantity",
			request: models.AddRecipeIngredientRequest{
				IngredientID: 1,
				Quantity:     0,
				Unit:         "grams",
			},
			expectedError: domain.ErrInvalidQuantity,
		},
		{
			name: "empty_unit",
			request: models.AddRecipeIngredientRequest{
				IngredientID: 1,
				Quantity:     100.0,
				Unit:         "",
			},
			expectedError: domain.ErrInvalidUnit,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := setup.service.AddRecipeIngredient(recipeID, tc.request)

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestIngredientService_AddRecipeIngredient_InvalidRecipeID(t *testing.T) {
	setup := setupIngredientServiceTest()
	request := testdata.NewAddRecipeIngredientRequestBuilder().Build()

	result, err := setup.service.AddRecipeIngredient(0, request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrRecipeNotFound, err)
	setup.recipeRepo.AssertNotCalled(t, "GetByID")
}

func TestIngredientService_AddRecipeIngredient_IngredientNotFound(t *testing.T) {
	setup := setupIngredientServiceTest()
	recipeID := 1
	request := testdata.NewAddRecipeIngredientRequestBuilder().WithIngredientID(999).Build()
	existingRecipe := testdata.NewRecipeBuilder().WithID(recipeID).BuildPtr()

	setup.recipeRepo.On("GetByID", recipeID).Return(existingRecipe, nil)
	setup.ingredientRepo.On("IngredientExists", 999).Return(false, nil)

	result, err := setup.service.AddRecipeIngredient(recipeID, request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrIngredientNotFound, err)
	setup.recipeRepo.AssertExpectations(t)
	setup.ingredientRepo.AssertExpectations(t)
}

// =============================================================================
// UPDATE RECIPE INGREDIENT TESTS
// =============================================================================

func TestIngredientService_UpdateRecipeIngredient_Success(t *testing.T) {
	setup := setupIngredientServiceTest()
	recipeID := 1
	ingredientID := 1
	request := testdata.NewAddRecipeIngredientRequestBuilder().WithQuantity(200.0).Build()
	existingRecipe := testdata.NewRecipeBuilder().WithID(recipeID).BuildPtr()
	expectedRecipeIngredient := testdata.NewRecipeIngredientBuilder().WithQuantity(200.0).BuildPtr()

	setup.recipeRepo.On("GetByID", recipeID).Return(existingRecipe, nil)
	setup.ingredientRepo.On("UpdateRecipeIngredient", recipeID, ingredientID, mock.AnythingOfType("models.AddRecipeIngredientRequest")).
		Return(expectedRecipeIngredient, nil)

	result, err := setup.service.UpdateRecipeIngredient(recipeID, ingredientID, request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200.0, result.Quantity)
	setup.recipeRepo.AssertExpectations(t)
	setup.ingredientRepo.AssertExpectations(t)
}

func TestIngredientService_UpdateRecipeIngredient_InvalidIDs(t *testing.T) {
	setup := setupIngredientServiceTest()
	request := testdata.NewAddRecipeIngredientRequestBuilder().Build()

	testCases := []struct {
		name         string
		recipeID     int
		ingredientID int
		expectedErr  error
	}{
		{"invalid_recipe_id", 0, 1, domain.ErrRecipeNotFound},
		{"invalid_ingredient_id", 1, 0, domain.ErrIngredientNotFound},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := setup.service.UpdateRecipeIngredient(tc.recipeID, tc.ingredientID, request)

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

// =============================================================================
// REMOVE RECIPE INGREDIENT TESTS
// =============================================================================

func TestIngredientService_RemoveRecipeIngredient_Success(t *testing.T) {
	setup := setupIngredientServiceTest()
	recipeID := 1
	ingredientID := 1
	existingRecipe := testdata.NewRecipeBuilder().WithID(recipeID).BuildPtr()

	setup.recipeRepo.On("GetByID", recipeID).Return(existingRecipe, nil)
	setup.ingredientRepo.On("RemoveRecipeIngredient", recipeID, ingredientID).Return(nil)

	err := setup.service.RemoveRecipeIngredient(recipeID, ingredientID)

	assert.NoError(t, err)
	setup.recipeRepo.AssertExpectations(t)
	setup.ingredientRepo.AssertExpectations(t)
}

func TestIngredientService_RemoveRecipeIngredient_InvalidIDs(t *testing.T) {
	setup := setupIngredientServiceTest()

	testCases := []struct {
		name         string
		recipeID     int
		ingredientID int
		expectedErr  error
	}{
		{"invalid_recipe_id", 0, 1, domain.ErrRecipeNotFound},
		{"invalid_ingredient_id", 1, 0, domain.ErrIngredientNotFound},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := setup.service.RemoveRecipeIngredient(tc.recipeID, tc.ingredientID)

			assert.Error(t, err)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}

func TestIngredientService_RemoveRecipeIngredient_NotFound(t *testing.T) {
	setup := setupIngredientServiceTest()
	recipeID := 1
	ingredientID := 1
	existingRecipe := testdata.NewRecipeBuilder().WithID(recipeID).BuildPtr()

	setup.recipeRepo.On("GetByID", recipeID).Return(existingRecipe, nil)
	setup.ingredientRepo.On("RemoveRecipeIngredient", recipeID, ingredientID).Return(sql.ErrNoRows)

	err := setup.service.RemoveRecipeIngredient(recipeID, ingredientID)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrIngredientNotFound, err)
	setup.recipeRepo.AssertExpectations(t)
	setup.ingredientRepo.AssertExpectations(t)
}

// =============================================================================
// SET RECIPE INGREDIENTS TESTS
// =============================================================================

func TestIngredientService_SetRecipeIngredients_Success(t *testing.T) {
	setup := setupIngredientServiceTest()
	recipeID := 1
	ingredients := []models.AddRecipeIngredientRequest{
		testdata.NewAddRecipeIngredientRequestBuilder().WithIngredientID(1).Build(),
		testdata.NewAddRecipeIngredientRequestBuilder().WithIngredientID(2).Build(),
	}
	existingRecipe := testdata.NewRecipeBuilder().WithID(recipeID).BuildPtr()

	setup.recipeRepo.On("GetByID", recipeID).Return(existingRecipe, nil)
	setup.ingredientRepo.On("IngredientExists", 1).Return(true, nil)
	setup.ingredientRepo.On("IngredientExists", 2).Return(true, nil)
	setup.ingredientRepo.On("SetRecipeIngredients", recipeID, ingredients).Return(nil)

	err := setup.service.SetRecipeIngredients(recipeID, ingredients)

	assert.NoError(t, err)
	setup.recipeRepo.AssertExpectations(t)
	setup.ingredientRepo.AssertExpectations(t)
}

func TestIngredientService_SetRecipeIngredients_InvalidRecipeID(t *testing.T) {
	setup := setupIngredientServiceTest()
	ingredients := []models.AddRecipeIngredientRequest{
		testdata.NewAddRecipeIngredientRequestBuilder().Build(),
	}

	err := setup.service.SetRecipeIngredients(0, ingredients)

	assert.Error(t, err)
	assert.Equal(t, domain.ErrRecipeNotFound, err)
	setup.recipeRepo.AssertNotCalled(t, "GetByID")
}

// =============================================================================
// GET RECIPES USING INGREDIENT TESTS
// =============================================================================

func TestIngredientService_GetRecipesUsingIngredient_Success(t *testing.T) {
	setup := setupIngredientServiceTest()
	ingredientID := 1
	existingIngredient := testdata.NewIngredientBuilder().WithID(ingredientID).BuildPtr()
	expectedRecipes := []models.Recipe{
		testdata.NewRecipeBuilder().WithID(1).Build(),
	}

	setup.ingredientRepo.On("GetIngredientByID", ingredientID).Return(existingIngredient, nil)
	setup.ingredientRepo.On("GetRecipesUsingIngredient", ingredientID, defaultParams).Return(expectedRecipes, 1, nil)

	result, meta, err := setup.service.GetRecipesUsingIngredient(ingredientID, defaultParams)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, meta.Total)
	setup.ingredientRepo.AssertExpectations(t)
}

func TestIngredientService_GetRecipesUsingIngredient_InvalidID(t *testing.T) {
	setup := setupIngredientServiceTest()

	result, _, err := setup.service.GetRecipesUsingIngredient(0, defaultParams)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrIngredientNotFound, err)
	setup.ingredientRepo.AssertNotCalled(t, "GetIngredientByID")
}

func TestIngredientService_GetRecipesUsingIngredient_IngredientNotFound(t *testing.T) {
	setup := setupIngredientServiceTest()
	ingredientID := 999

	setup.ingredientRepo.On("GetIngredientByID", ingredientID).Return(nil, sql.ErrNoRows)

	result, _, err := setup.service.GetRecipesUsingIngredient(ingredientID, defaultParams)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrIngredientNotFound, err)
	setup.ingredientRepo.AssertExpectations(t)
}
