package mocks

import (
	"meal-prep/shared/models"

	"github.com/stretchr/testify/mock"
)

// MockIngredientService mocks the ingredient service interface
type MockIngredientService struct {
	mock.Mock
}

// =============================================================================
// BASIC INGREDIENT OPERATIONS
// =============================================================================

func (m *MockIngredientService) GetAllIngredients(params models.PaginationParams) ([]models.Ingredient, models.PaginationMeta, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(models.PaginationMeta), args.Error(2)
	}
	return args.Get(0).([]models.Ingredient), args.Get(1).(models.PaginationMeta), args.Error(2)
}

func (m *MockIngredientService) GetIngredientByID(id int) (*models.Ingredient, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Ingredient), args.Error(1)
}

func (m *MockIngredientService) GetIngredientsByCategory(category string, params models.PaginationParams) ([]models.Ingredient, models.PaginationMeta, error) {
	args := m.Called(category, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(models.PaginationMeta), args.Error(2)
	}
	return args.Get(0).([]models.Ingredient), args.Get(1).(models.PaginationMeta), args.Error(2)
}

func (m *MockIngredientService) SearchIngredients(query string, params models.PaginationParams) ([]models.Ingredient, models.PaginationMeta, error) {
	args := m.Called(query, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(models.PaginationMeta), args.Error(2)
	}
	return args.Get(0).([]models.Ingredient), args.Get(1).(models.PaginationMeta), args.Error(2)
}

func (m *MockIngredientService) CreateIngredient(req models.CreateIngredientRequest) (*models.Ingredient, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Ingredient), args.Error(1)
}

func (m *MockIngredientService) UpdateIngredient(id int, req models.UpdateIngredientRequest) (*models.Ingredient, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Ingredient), args.Error(1)
}

func (m *MockIngredientService) DeleteIngredient(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

// =============================================================================
// RECIPE-INGREDIENT RELATIONSHIP OPERATIONS
// =============================================================================

func (m *MockIngredientService) GetRecipeIngredients(recipeID int) ([]models.RecipeIngredient, error) {
	args := m.Called(recipeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.RecipeIngredient), args.Error(1)
}

func (m *MockIngredientService) AddRecipeIngredient(recipeID int, req models.AddRecipeIngredientRequest) (*models.RecipeIngredient, error) {
	args := m.Called(recipeID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RecipeIngredient), args.Error(1)
}

func (m *MockIngredientService) UpdateRecipeIngredient(recipeID, ingredientID int, req models.AddRecipeIngredientRequest) (*models.RecipeIngredient, error) {
	args := m.Called(recipeID, ingredientID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RecipeIngredient), args.Error(1)
}

func (m *MockIngredientService) RemoveRecipeIngredient(recipeID, ingredientID int) error {
	args := m.Called(recipeID, ingredientID)
	return args.Error(0)
}

func (m *MockIngredientService) SetRecipeIngredients(recipeID int, ingredients []models.AddRecipeIngredientRequest) error {
	args := m.Called(recipeID, ingredients)
	return args.Error(0)
}

// =============================================================================
// COMPLEX OPERATIONS
// =============================================================================

func (m *MockIngredientService) GetRecipesUsingIngredient(ingredientID int, params models.PaginationParams) ([]models.Recipe, models.PaginationMeta, error) {
	args := m.Called(ingredientID, params)
	if args.Get(0) == nil {
		return nil, args.Get(1).(models.PaginationMeta), args.Error(2)
	}
	return args.Get(0).([]models.Recipe), args.Get(1).(models.PaginationMeta), args.Error(2)
}
