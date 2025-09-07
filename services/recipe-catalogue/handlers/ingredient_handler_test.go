package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"meal-prep/services/recipe-catalogue/handlers/mocks"
	"meal-prep/services/recipe-catalogue/service"
	"meal-prep/services/recipe-catalogue/service/testdata"
	"meal-prep/shared/middleware"
	"meal-prep/shared/models"
	"meal-prep/shared/utils"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type ingredientHandlerTestSetup struct {
	handler           *IngredientHandler
	ingredientService *mocks.MockIngredientService
}

func setupIngredientHandlerTest() *ingredientHandlerTestSetup {
	mockService := new(mocks.MockIngredientService)
	handler := NewIngredientHandler(mockService)

	return &ingredientHandlerTestSetup{
		handler:           handler,
		ingredientService: mockService,
	}
}

// =============================================================================
// GET ALL INGREDIENTS TESTS
// =============================================================================

func TestIngredientHandler_GetAllIngredients_Success(t *testing.T) {
	setup := setupIngredientHandlerTest()
	expectedIngredients := []models.Ingredient{
		testdata.NewIngredientBuilder().WithID(1).WithName("Tomato").Build(),
		testdata.NewIngredientBuilder().WithID(2).WithName("Basil").Build(),
	}

	setup.ingredientService.On("GetAllIngredients").Return(expectedIngredients, nil)

	req := httptest.NewRequest("GET", "/ingredients", nil)
	recorder := httptest.NewRecorder()

	setup.handler.GetAllIngredients(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var response []models.Ingredient
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "Tomato", response[0].Name)

	setup.ingredientService.AssertExpectations(t)
}

func TestIngredientHandler_GetAllIngredients_WithCategoryFilter(t *testing.T) {
	setup := setupIngredientHandlerTest()
	expectedIngredients := []models.Ingredient{
		testdata.NewIngredientBuilder().WithCategory("Vegetable").Build(),
	}

	setup.ingredientService.On("GetIngredientsByCategory", "Vegetable").Return(expectedIngredients, nil)

	req := httptest.NewRequest("GET", "/ingredients?category=Vegetable", nil)
	recorder := httptest.NewRecorder()

	setup.handler.GetAllIngredients(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response []models.Ingredient
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)

	setup.ingredientService.AssertExpectations(t)
}

func TestIngredientHandler_GetAllIngredients_WithSearchQuery(t *testing.T) {
	setup := setupIngredientHandlerTest()
	expectedIngredients := []models.Ingredient{
		testdata.NewIngredientBuilder().WithName("Cherry Tomato").Build(),
	}

	setup.ingredientService.On("SearchIngredients", "tomato").Return(expectedIngredients, nil)

	req := httptest.NewRequest("GET", "/ingredients?search=tomato", nil)
	recorder := httptest.NewRecorder()

	setup.handler.GetAllIngredients(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response []models.Ingredient
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)

	setup.ingredientService.AssertExpectations(t)
}

func TestIngredientHandler_GetAllIngredients_ServiceError(t *testing.T) {
	setup := setupIngredientHandlerTest()

	setup.ingredientService.On("GetAllIngredients").Return(nil, errors.New("database error"))

	req := httptest.NewRequest("GET", "/ingredients", nil)
	recorder := httptest.NewRecorder()

	setup.handler.GetAllIngredients(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to fetch ingredients", response.Message)

	setup.ingredientService.AssertExpectations(t)
}

// =============================================================================
// GET INGREDIENT BY ID TESTS
// =============================================================================

func TestIngredientHandler_GetIngredientByID_Success(t *testing.T) {
	setup := setupIngredientHandlerTest()
	expectedIngredient := testdata.NewIngredientBuilder().WithID(1).WithName("Tomato").BuildPtr()

	setup.ingredientService.On("GetIngredientByID", 1).Return(expectedIngredient, nil)

	req := httptest.NewRequest("GET", "/ingredients/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	recorder := httptest.NewRecorder()

	setup.handler.GetIngredientByID(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response models.Ingredient
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, 1, response.ID)
	assert.Equal(t, "Tomato", response.Name)

	setup.ingredientService.AssertExpectations(t)
}

func TestIngredientHandler_GetIngredientByID_InvalidID(t *testing.T) {
	setup := setupIngredientHandlerTest()

	req := httptest.NewRequest("GET", "/ingredients/invalid", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "invalid"})
	recorder := httptest.NewRecorder()

	setup.handler.GetIngredientByID(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid ingredient ID", response.Message)

	setup.ingredientService.AssertNotCalled(t, "GetIngredientByID")
}

func TestIngredientHandler_GetIngredientByID_NotFound(t *testing.T) {
	setup := setupIngredientHandlerTest()

	setup.ingredientService.On("GetIngredientByID", 999).Return(nil, service.ErrIngredientNotFound)

	req := httptest.NewRequest("GET", "/ingredients/999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	recorder := httptest.NewRecorder()

	setup.handler.GetIngredientByID(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, service.ErrIngredientNotFound.Error(), response.Message)

	setup.ingredientService.AssertExpectations(t)
}

// =============================================================================
// CREATE INGREDIENT TESTS (PROTECTED ENDPOINT)
// =============================================================================

func TestIngredientHandler_CreateIngredient_Success(t *testing.T) {
	setup := setupIngredientHandlerTest()
	request := testdata.ValidCreateIngredientRequest()
	expectedIngredient := testdata.NewIngredientBuilder().WithName(request.Name).BuildPtr()

	setup.ingredientService.On("CreateIngredient", mock.AnythingOfType("models.CreateIngredientRequest")).
		Return(expectedIngredient, nil)

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/ingredients", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Add authenticated user to context
	claims := &utils.Claims{UserID: 1, Email: "test@example.com"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()

	setup.handler.CreateIngredient(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code)

	var response models.Ingredient
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, request.Name, response.Name)

	setup.ingredientService.AssertExpectations(t)
}

func TestIngredientHandler_CreateIngredient_MissingAuthentication(t *testing.T) {
	setup := setupIngredientHandlerTest()
	request := testdata.ValidCreateIngredientRequest()

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/ingredients", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	setup.handler.CreateIngredient(recorder, req)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Authentication required", response.Message)

	setup.ingredientService.AssertNotCalled(t, "CreateIngredient")
}

func TestIngredientHandler_CreateIngredient_InvalidJSON(t *testing.T) {
	setup := setupIngredientHandlerTest()

	req := httptest.NewRequest("POST", "/ingredients", bytes.NewBufferString(`{"name":"test","invalid":}`))
	req.Header.Set("Content-Type", "application/json")

	claims := &utils.Claims{UserID: 1, Email: "test@example.com"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()

	setup.handler.CreateIngredient(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid JSON", response.Message)

	setup.ingredientService.AssertNotCalled(t, "CreateIngredient")
}

func TestIngredientHandler_CreateIngredient_ValidationErrors(t *testing.T) {
	testCases := []struct {
		name           string
		serviceError   error
		expectedStatus int
	}{
		{
			name:           "ingredient_name_required",
			serviceError:   service.ErrIngredientNameRequired,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "ingredient_exists",
			serviceError:   service.ErrIngredientExists,
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setup := setupIngredientHandlerTest()
			request := testdata.ValidCreateIngredientRequest()

			setup.ingredientService.On("CreateIngredient", mock.AnythingOfType("models.CreateIngredientRequest")).
				Return(nil, tc.serviceError)

			requestBody, _ := json.Marshal(request)
			req := httptest.NewRequest("POST", "/ingredients", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			claims := &utils.Claims{UserID: 1, Email: "test@example.com"}
			ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
			req = req.WithContext(ctx)

			recorder := httptest.NewRecorder()

			setup.handler.CreateIngredient(recorder, req)

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			var response models.ErrorResponse
			err := json.NewDecoder(recorder.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, tc.serviceError.Error(), response.Message)

			setup.ingredientService.AssertExpectations(t)
		})
	}
}

// =============================================================================
// UPDATE INGREDIENT TESTS (PROTECTED ENDPOINT)
// =============================================================================

func TestIngredientHandler_UpdateIngredient_Success(t *testing.T) {
	setup := setupIngredientHandlerTest()
	request := models.UpdateIngredientRequest{
		Name:        "Updated Ingredient",
		Description: "Updated description",
		Category:    "Updated Category",
	}
	expectedIngredient := testdata.NewIngredientBuilder().WithName("Updated Ingredient").BuildPtr()

	setup.ingredientService.On("UpdateIngredient", 1, mock.AnythingOfType("models.UpdateIngredientRequest")).
		Return(expectedIngredient, nil)

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("PUT", "/ingredients/1", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	claims := &utils.Claims{UserID: 1, Email: "test@example.com"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()

	setup.handler.UpdateIngredient(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response models.Ingredient
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Ingredient", response.Name)

	setup.ingredientService.AssertExpectations(t)
}

func TestIngredientHandler_UpdateIngredient_InvalidID(t *testing.T) {
	setup := setupIngredientHandlerTest()
	request := models.UpdateIngredientRequest{Name: "Updated"}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("PUT", "/ingredients/invalid", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "invalid"})

	claims := &utils.Claims{UserID: 1, Email: "test@example.com"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()

	setup.handler.UpdateIngredient(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid ingredient ID", response.Message)

	setup.ingredientService.AssertNotCalled(t, "UpdateIngredient")
}

// =============================================================================
// DELETE INGREDIENT TESTS (PROTECTED ENDPOINT)
// =============================================================================

func TestIngredientHandler_DeleteIngredient_Success(t *testing.T) {
	setup := setupIngredientHandlerTest()

	setup.ingredientService.On("DeleteIngredient", 1).Return(nil)

	req := httptest.NewRequest("DELETE", "/ingredients/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	claims := &utils.Claims{UserID: 1, Email: "test@example.com"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()

	setup.handler.DeleteIngredient(recorder, req)

	assert.Equal(t, http.StatusNoContent, recorder.Code)

	var response map[string]string
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Ingredient deleted successfully", response["message"])

	setup.ingredientService.AssertExpectations(t)
}

func TestIngredientHandler_DeleteIngredient_CannotDelete(t *testing.T) {
	setup := setupIngredientHandlerTest()

	setup.ingredientService.On("DeleteIngredient", 1).Return(service.ErrCannotDeleteIngredient)

	req := httptest.NewRequest("DELETE", "/ingredients/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	claims := &utils.Claims{UserID: 1, Email: "test@example.com"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()

	setup.handler.DeleteIngredient(recorder, req)

	assert.Equal(t, http.StatusConflict, recorder.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, service.ErrCannotDeleteIngredient.Error(), response.Message)

	setup.ingredientService.AssertExpectations(t)
}

// =============================================================================
// GET RECIPES USING INGREDIENT TESTS
// =============================================================================

func TestIngredientHandler_GetRecipesUsingIngredient_Success(t *testing.T) {
	setup := setupIngredientHandlerTest()
	expectedRecipes := []models.Recipe{
		testdata.NewRecipeBuilder().WithID(1).WithName("Tomato Sauce").Build(),
	}

	setup.ingredientService.On("GetRecipesUsingIngredient", 1).Return(expectedRecipes, nil)

	req := httptest.NewRequest("GET", "/ingredients/1/recipes", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	recorder := httptest.NewRecorder()

	setup.handler.GetRecipesUsingIngredient(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response []models.Recipe
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, "Tomato Sauce", response[0].Name)

	setup.ingredientService.AssertExpectations(t)
}

// =============================================================================
// RECIPE INGREDIENTS TESTS
// =============================================================================

func TestIngredientHandler_GetRecipeIngredients_Success(t *testing.T) {
	setup := setupIngredientHandlerTest()
	expectedIngredients := []models.RecipeIngredient{
		testdata.NewRecipeIngredientBuilder().WithRecipeID(1).Build(),
	}

	setup.ingredientService.On("GetRecipeIngredients", 1).Return(expectedIngredients, nil)

	req := httptest.NewRequest("GET", "/recipes/1/ingredients", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	recorder := httptest.NewRecorder()

	setup.handler.GetRecipeIngredients(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response []models.RecipeIngredient
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)

	setup.ingredientService.AssertExpectations(t)
}

func TestIngredientHandler_AddRecipeIngredient_Success(t *testing.T) {
	setup := setupIngredientHandlerTest()
	request := testdata.ValidAddRecipeIngredientRequest()
	expectedRecipeIngredient := testdata.NewRecipeIngredientBuilder().BuildPtr()

	setup.ingredientService.On("AddRecipeIngredient", 1, mock.AnythingOfType("models.AddRecipeIngredientRequest")).
		Return(expectedRecipeIngredient, nil)

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/recipes/1/ingredients", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	claims := &utils.Claims{UserID: 1, Email: "test@example.com"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()

	setup.handler.AddRecipeIngredient(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code)

	var response models.RecipeIngredient
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)

	setup.ingredientService.AssertExpectations(t)
}

func TestIngredientHandler_AddRecipeIngredient_ValidationErrors(t *testing.T) {
	testCases := []struct {
		name           string
		serviceError   error
		expectedStatus int
	}{
		{
			name:           "recipe_not_found",
			serviceError:   service.ErrRecipeNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "ingredient_not_found",
			serviceError:   service.ErrIngredientNotFound,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid_quantity",
			serviceError:   service.ErrInvalidQuantity,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid_unit",
			serviceError:   service.ErrInvalidUnit,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setup := setupIngredientHandlerTest()
			request := testdata.ValidAddRecipeIngredientRequest()

			setup.ingredientService.On("AddRecipeIngredient", 1, mock.AnythingOfType("models.AddRecipeIngredientRequest")).
				Return(nil, tc.serviceError)

			requestBody, _ := json.Marshal(request)
			req := httptest.NewRequest("POST", "/recipes/1/ingredients", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")
			req = mux.SetURLVars(req, map[string]string{"id": "1"})

			claims := &utils.Claims{UserID: 1, Email: "test@example.com"}
			ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
			req = req.WithContext(ctx)

			recorder := httptest.NewRecorder()

			setup.handler.AddRecipeIngredient(recorder, req)

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			var response models.ErrorResponse
			err := json.NewDecoder(recorder.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, tc.serviceError.Error(), response.Message)

			setup.ingredientService.AssertExpectations(t)
		})
	}
}

// =============================================================================
// GENERATE SHOPPING LIST TESTS (PROTECTED ENDPOINT)
// =============================================================================

func TestIngredientHandler_GenerateShoppingList_Success(t *testing.T) {
	setup := setupIngredientHandlerTest()
	request := models.ShoppingListRequest{RecipeIDs: []int{1, 2}}
	expectedShoppingList := []models.ShoppingListItem{
		{
			IngredientID:  1,
			Ingredient:    testdata.ValidIngredient(),
			TotalQuantity: 150.0,
			Unit:          "grams",
			Recipes:       []string{"Recipe 1", "Recipe 2"},
		},
	}

	setup.ingredientService.On("GenerateShoppingList", mock.AnythingOfType("models.ShoppingListRequest")).
		Return(expectedShoppingList, nil)

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/shopping-list", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	claims := &utils.Claims{UserID: 1, Email: "test@example.com"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()

	setup.handler.GenerateShoppingList(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response []models.ShoppingListItem
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, 150.0, response[0].TotalQuantity)

	setup.ingredientService.AssertExpectations(t)
}

func TestIngredientHandler_GenerateShoppingList_EmptyRecipeList(t *testing.T) {
	setup := setupIngredientHandlerTest()
	request := models.ShoppingListRequest{RecipeIDs: []int{}}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/shopping-list", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	claims := &utils.Claims{UserID: 1, Email: "test@example.com"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()

	setup.handler.GenerateShoppingList(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "At least one recipe ID is required", response.Message)

	setup.ingredientService.AssertNotCalled(t, "GenerateShoppingList")
}
