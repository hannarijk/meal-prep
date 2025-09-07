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

type recipeHandlerTestSetup struct {
	handler       *RecipeHandler
	recipeService *mocks.MockRecipeService
	router        *mux.Router
}

func setupRecipeHandlerTest() *recipeHandlerTestSetup {
	mockService := new(mocks.MockRecipeService)
	handler := NewRecipeHandler(mockService)
	router := mux.NewRouter()

	return &recipeHandlerTestSetup{
		handler:       handler,
		recipeService: mockService,
		router:        router,
	}
}

// =============================================================================
// GET ALL RECIPES TESTS
// =============================================================================

func TestRecipeHandler_GetAllRecipes_Success(t *testing.T) {
	setup := setupRecipeHandlerTest()
	expectedRecipes := []models.Recipe{
		testdata.NewRecipeBuilder().WithID(1).WithName("Recipe 1").Build(),
		testdata.NewRecipeBuilder().WithID(2).WithName("Recipe 2").Build(),
	}

	setup.recipeService.On("GetAllRecipes").Return(expectedRecipes, nil)

	req := httptest.NewRequest("GET", "/recipes", nil)
	recorder := httptest.NewRecorder()

	setup.handler.GetAllRecipes(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	var response []models.Recipe
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "Recipe 1", response[0].Name)

	setup.recipeService.AssertExpectations(t)
}

func TestRecipeHandler_GetAllRecipes_ServiceError(t *testing.T) {
	setup := setupRecipeHandlerTest()
	expectedError := errors.New("database error")

	setup.recipeService.On("GetAllRecipes").Return(nil, expectedError)

	req := httptest.NewRequest("GET", "/recipes", nil)
	recorder := httptest.NewRecorder()

	setup.handler.GetAllRecipes(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "recipe_catalogue_error", response.Error)
	assert.Equal(t, "Failed to fetch recipes", response.Message)

	setup.recipeService.AssertExpectations(t)
}

// =============================================================================
// GET RECIPE BY ID TESTS
// =============================================================================

func TestRecipeHandler_GetRecipeByID_Success(t *testing.T) {
	setup := setupRecipeHandlerTest()
	expectedRecipe := testdata.NewRecipeBuilder().WithID(1).WithName("Test Recipe").BuildPtr()

	setup.recipeService.On("GetRecipeByID", 1).Return(expectedRecipe, nil)

	req := httptest.NewRequest("GET", "/recipes/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	recorder := httptest.NewRecorder()

	setup.handler.GetRecipeByID(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response models.Recipe
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, 1, response.ID)
	assert.Equal(t, "Test Recipe", response.Name)

	setup.recipeService.AssertExpectations(t)
}

func TestRecipeHandler_GetRecipeByID_InvalidID(t *testing.T) {
	setup := setupRecipeHandlerTest()

	req := httptest.NewRequest("GET", "/recipes/invalid", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "invalid"})
	recorder := httptest.NewRecorder()

	setup.handler.GetRecipeByID(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid recipe ID", response.Message)

	setup.recipeService.AssertNotCalled(t, "GetRecipeByID")
}

func TestRecipeHandler_GetRecipeByID_NotFound(t *testing.T) {
	setup := setupRecipeHandlerTest()

	setup.recipeService.On("GetRecipeByID", 999).Return(nil, service.ErrRecipeNotFound)

	req := httptest.NewRequest("GET", "/recipes/999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	recorder := httptest.NewRecorder()

	setup.handler.GetRecipeByID(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, service.ErrRecipeNotFound.Error(), response.Message)

	setup.recipeService.AssertExpectations(t)
}

// =============================================================================
// GET RECIPES BY CATEGORY TESTS
// =============================================================================

func TestRecipeHandler_GetRecipesByCategory_Success(t *testing.T) {
	setup := setupRecipeHandlerTest()
	expectedRecipes := []models.Recipe{
		testdata.NewRecipeBuilder().WithID(1).WithCategoryID(1).Build(),
	}

	setup.recipeService.On("GetRecipesByCategory", 1).Return(expectedRecipes, nil)

	req := httptest.NewRequest("GET", "/categories/1/recipes", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	recorder := httptest.NewRecorder()

	setup.handler.GetRecipesByCategory(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response []models.Recipe
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)

	setup.recipeService.AssertExpectations(t)
}

func TestRecipeHandler_GetRecipesByCategory_InvalidCategoryID(t *testing.T) {
	setup := setupRecipeHandlerTest()

	req := httptest.NewRequest("GET", "/categories/invalid/recipes", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "invalid"})
	recorder := httptest.NewRecorder()

	setup.handler.GetRecipesByCategory(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid category ID", response.Message)

	setup.recipeService.AssertNotCalled(t, "GetRecipesByCategory")
}

func TestRecipeHandler_GetRecipesByCategory_CategoryNotFound(t *testing.T) {
	setup := setupRecipeHandlerTest()

	setup.recipeService.On("GetRecipesByCategory", 999).Return(nil, service.ErrCategoryNotFound)

	req := httptest.NewRequest("GET", "/categories/999/recipes", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "999"})
	recorder := httptest.NewRecorder()

	setup.handler.GetRecipesByCategory(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, service.ErrCategoryNotFound.Error(), response.Message)

	setup.recipeService.AssertExpectations(t)
}

// =============================================================================
// CREATE RECIPE TESTS (PROTECTED ENDPOINT)
// =============================================================================

func TestRecipeHandler_CreateRecipe_Success(t *testing.T) {
	setup := setupRecipeHandlerTest()
	request := testdata.ValidCreateRequest()
	expectedRecipe := testdata.NewRecipeBuilder().WithName(request.Name).BuildPtr()

	setup.recipeService.On("CreateRecipe", mock.AnythingOfType("models.CreateRecipeRequest")).
		Return(expectedRecipe, nil)

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/recipes", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Add authenticated user to context
	claims := &utils.Claims{UserID: 1, Email: "test@example.com"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()

	setup.handler.CreateRecipe(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code)

	var response models.Recipe
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, request.Name, response.Name)

	setup.recipeService.AssertExpectations(t)
}

func TestRecipeHandler_CreateRecipe_MissingAuthentication(t *testing.T) {
	setup := setupRecipeHandlerTest()
	request := testdata.ValidCreateRequest()

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/recipes", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	setup.handler.CreateRecipe(recorder, req)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Authentication required", response.Message)

	setup.recipeService.AssertNotCalled(t, "CreateRecipe")
}

func TestRecipeHandler_CreateRecipe_InvalidJSON(t *testing.T) {
	setup := setupRecipeHandlerTest()

	req := httptest.NewRequest("POST", "/recipes", bytes.NewBufferString(`{"name":"test","invalid":}`))
	req.Header.Set("Content-Type", "application/json")

	// Add authenticated user to context
	claims := &utils.Claims{UserID: 1, Email: "test@example.com"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()

	setup.handler.CreateRecipe(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid JSON", response.Message)

	setup.recipeService.AssertNotCalled(t, "CreateRecipe")
}

func TestRecipeHandler_CreateRecipe_ValidationErrors(t *testing.T) {
	testCases := []struct {
		name           string
		serviceError   error
		expectedStatus int
	}{
		{
			name:           "recipe_name_required",
			serviceError:   service.ErrRecipeNameRequired,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "category_not_found",
			serviceError:   service.ErrCategoryNotFound,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid_category",
			serviceError:   service.ErrInvalidCategory,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setup := setupRecipeHandlerTest()
			request := testdata.ValidCreateRequest()

			setup.recipeService.On("CreateRecipe", mock.AnythingOfType("models.CreateRecipeRequest")).
				Return(nil, tc.serviceError)

			requestBody, _ := json.Marshal(request)
			req := httptest.NewRequest("POST", "/recipes", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Add authenticated user to context
			claims := &utils.Claims{UserID: 1, Email: "test@example.com"}
			ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
			req = req.WithContext(ctx)

			recorder := httptest.NewRecorder()

			setup.handler.CreateRecipe(recorder, req)

			assert.Equal(t, tc.expectedStatus, recorder.Code)

			var response models.ErrorResponse
			err := json.NewDecoder(recorder.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, tc.serviceError.Error(), response.Message)

			setup.recipeService.AssertExpectations(t)
		})
	}
}

func TestRecipeHandler_CreateRecipe_InternalServerError(t *testing.T) {
	setup := setupRecipeHandlerTest()
	request := testdata.ValidCreateRequest()

	setup.recipeService.On("CreateRecipe", mock.AnythingOfType("models.CreateRecipeRequest")).
		Return(nil, errors.New("database connection failed"))

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/recipes", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Add authenticated user to context
	claims := &utils.Claims{UserID: 1, Email: "test@example.com"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()

	setup.handler.CreateRecipe(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to create recipe", response.Message)

	setup.recipeService.AssertExpectations(t)
}

// =============================================================================
// UPDATE RECIPE TESTS (PROTECTED ENDPOINT)
// =============================================================================

func TestRecipeHandler_UpdateRecipe_Success(t *testing.T) {
	setup := setupRecipeHandlerTest()
	request := models.UpdateRecipeRequest{
		Name:        "Updated Recipe",
		Description: "Updated description",
		CategoryID:  1,
	}
	expectedRecipe := testdata.NewRecipeBuilder().WithName("Updated Recipe").BuildPtr()

	setup.recipeService.On("UpdateRecipe", 1, mock.AnythingOfType("models.UpdateRecipeRequest")).
		Return(expectedRecipe, nil)

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("PUT", "/recipes/1", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	// Add authenticated user to context
	claims := &utils.Claims{UserID: 1, Email: "test@example.com"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()

	setup.handler.UpdateRecipe(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response models.Recipe
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Recipe", response.Name)

	setup.recipeService.AssertExpectations(t)
}

func TestRecipeHandler_UpdateRecipe_InvalidID(t *testing.T) {
	setup := setupRecipeHandlerTest()
	request := models.UpdateRecipeRequest{Name: "Updated"}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("PUT", "/recipes/invalid", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "invalid"})

	// Add authenticated user to context
	claims := &utils.Claims{UserID: 1, Email: "test@example.com"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()

	setup.handler.UpdateRecipe(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid recipe ID", response.Message)

	setup.recipeService.AssertNotCalled(t, "UpdateRecipe")
}

func TestRecipeHandler_UpdateRecipe_RecipeNotFound(t *testing.T) {
	setup := setupRecipeHandlerTest()
	request := models.UpdateRecipeRequest{Name: "Updated"}

	setup.recipeService.On("UpdateRecipe", 999, mock.AnythingOfType("models.UpdateRecipeRequest")).
		Return(nil, service.ErrRecipeNotFound)

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("PUT", "/recipes/999", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "999"})

	// Add authenticated user to context
	claims := &utils.Claims{UserID: 1, Email: "test@example.com"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()

	setup.handler.UpdateRecipe(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, service.ErrRecipeNotFound.Error(), response.Message)

	setup.recipeService.AssertExpectations(t)
}

// =============================================================================
// DELETE RECIPE TESTS (PROTECTED ENDPOINT)
// =============================================================================

func TestRecipeHandler_DeleteRecipe_Success(t *testing.T) {
	setup := setupRecipeHandlerTest()

	setup.recipeService.On("DeleteRecipe", 1).Return(nil)

	req := httptest.NewRequest("DELETE", "/recipes/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	// Add authenticated user to context
	claims := &utils.Claims{UserID: 1, Email: "test@example.com"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()

	setup.handler.DeleteRecipe(recorder, req)

	assert.Equal(t, http.StatusNoContent, recorder.Code)

	var response map[string]string
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Recipe deleted successfully", response["message"])

	setup.recipeService.AssertExpectations(t)
}

func TestRecipeHandler_DeleteRecipe_InvalidID(t *testing.T) {
	setup := setupRecipeHandlerTest()

	req := httptest.NewRequest("DELETE", "/recipes/invalid", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "invalid"})

	// Add authenticated user to context
	claims := &utils.Claims{UserID: 1, Email: "test@example.com"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()

	setup.handler.DeleteRecipe(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid recipe ID", response.Message)

	setup.recipeService.AssertNotCalled(t, "DeleteRecipe")
}

func TestRecipeHandler_DeleteRecipe_NotFound(t *testing.T) {
	setup := setupRecipeHandlerTest()

	setup.recipeService.On("DeleteRecipe", 999).Return(service.ErrRecipeNotFound)

	req := httptest.NewRequest("DELETE", "/recipes/999", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "999"})

	// Add authenticated user to context
	claims := &utils.Claims{UserID: 1, Email: "test@example.com"}
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, claims)
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()

	setup.handler.DeleteRecipe(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, service.ErrRecipeNotFound.Error(), response.Message)

	setup.recipeService.AssertExpectations(t)
}

// =============================================================================
// GET ALL CATEGORIES TESTS
// =============================================================================

func TestRecipeHandler_GetAllCategories_Success(t *testing.T) {
	setup := setupRecipeHandlerTest()
	expectedCategories := []models.Category{
		testdata.NewCategoryBuilder().WithID(1).WithName("Italian").Build(),
		testdata.NewCategoryBuilder().WithID(2).WithName("Mexican").Build(),
	}

	setup.recipeService.On("GetAllCategories").Return(expectedCategories, nil)

	req := httptest.NewRequest("GET", "/categories", nil)
	recorder := httptest.NewRecorder()

	setup.handler.GetAllCategories(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response []models.Category
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "Italian", response[0].Name)

	setup.recipeService.AssertExpectations(t)
}

func TestRecipeHandler_GetAllCategories_ServiceError(t *testing.T) {
	setup := setupRecipeHandlerTest()

	setup.recipeService.On("GetAllCategories").Return(nil, errors.New("database error"))

	req := httptest.NewRequest("GET", "/categories", nil)
	recorder := httptest.NewRecorder()

	setup.handler.GetAllCategories(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to fetch categories", response.Message)

	setup.recipeService.AssertExpectations(t)
}
