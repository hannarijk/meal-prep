package handlers

import (
	"bytes"
	"encoding/json"
	"meal-prep/services/recipe-catalogue/handlers/mocks"
	"meal-prep/services/recipe-catalogue/service/testdata"
	"meal-prep/services/recipe-catalogue/test"
	"meal-prep/shared/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type groceryHandlerTestSetup struct {
	handler        *GroceryHandler
	groceryService *mocks.MockGroceryService
}

func setupGroceryHandlerTest() *groceryHandlerTestSetup {
	mockService := new(mocks.MockGroceryService)
	handler := NewGroceryHandler(mockService)

	return &groceryHandlerTestSetup{
		handler:        handler,
		groceryService: mockService,
	}
}

// =============================================================================
// GENERATE GROCERY LIST TESTS (PROTECTED ENDPOINT)
// =============================================================================

func TestIngredientHandler_GenerateGroceryList_Success(t *testing.T) {
	setup := setupGroceryHandlerTest()
	request := models.GroceryListRequest{RecipeIDs: []int{1, 2}}
	expectedGroceryList := []models.GroceryListItem{
		{
			IngredientID:  1,
			Ingredient:    testdata.ValidIngredient(),
			TotalQuantity: 150.0,
			Unit:          "grams",
			Recipes:       []string{"Recipe 1", "Recipe 2"},
		},
	}

	setup.groceryService.On("GenerateGroceryList", mock.AnythingOfType("models.GroceryListRequest")).
		Return(expectedGroceryList, nil)

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/grocery-list", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	// Add authenticated user to context
	req = test.AddAuthContext(req, 1, "test@example.com")

	recorder := httptest.NewRecorder()

	setup.handler.GenerateGroceryList(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response []models.GroceryListItem
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, 150.0, response[0].TotalQuantity)

	setup.groceryService.AssertExpectations(t)
}

func TestIngredientHandler_GenerateGroceryList_EmptyRecipeList(t *testing.T) {
	setup := setupGroceryHandlerTest()
	request := models.GroceryListRequest{RecipeIDs: []int{}}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/grocery-list", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	// Add authenticated user to context
	req = test.AddAuthContext(req, 1, "test@example.com")

	recorder := httptest.NewRecorder()

	setup.handler.GenerateGroceryList(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response models.ErrorResponse
	err := json.NewDecoder(recorder.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "At least one recipe ID is required", response.Message)

	setup.groceryService.AssertNotCalled(t, "GenerateGroceryList")
}
