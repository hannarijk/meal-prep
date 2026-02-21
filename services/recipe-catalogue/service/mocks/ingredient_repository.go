package mocks

import (
	"github.com/stretchr/testify/mock"
	"meal-prep/shared/models"
)

// MockIngredientRepository mocks the ingredient repository interface
type MockIngredientRepository struct {
	mock.Mock
}

// =============================================================================
// BASIC INGREDIENT OPERATIONS
// =============================================================================

func (m *MockIngredientRepository) GetAllIngredients(params models.PaginationParams) ([]models.Ingredient, int, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]models.Ingredient), args.Int(1), args.Error(2)
}

func (m *MockIngredientRepository) GetIngredientByID(id int) (*models.Ingredient, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Ingredient), args.Error(1)
}

func (m *MockIngredientRepository) GetIngredientsByCategory(category string, params models.PaginationParams) ([]models.Ingredient, int, error) {
	args := m.Called(category, params)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]models.Ingredient), args.Int(1), args.Error(2)
}

func (m *MockIngredientRepository) SearchIngredients(query string, params models.PaginationParams) ([]models.Ingredient, int, error) {
	args := m.Called(query, params)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]models.Ingredient), args.Int(1), args.Error(2)
}

func (m *MockIngredientRepository) CreateIngredient(req models.CreateIngredientRequest) (*models.Ingredient, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Ingredient), args.Error(1)
}

func (m *MockIngredientRepository) UpdateIngredient(id int, req models.UpdateIngredientRequest) (*models.Ingredient, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Ingredient), args.Error(1)
}

func (m *MockIngredientRepository) DeleteIngredient(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockIngredientRepository) IngredientExists(id int) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

// =============================================================================
// RECIPE-INGREDIENT RELATIONSHIP OPERATIONS
// =============================================================================

func (m *MockIngredientRepository) GetRecipeIngredients(recipeID int) ([]models.RecipeIngredient, error) {
	args := m.Called(recipeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.RecipeIngredient), args.Error(1)
}

func (m *MockIngredientRepository) AddRecipeIngredient(recipeID int, req models.AddRecipeIngredientRequest) (*models.RecipeIngredient, error) {
	args := m.Called(recipeID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RecipeIngredient), args.Error(1)
}

func (m *MockIngredientRepository) UpdateRecipeIngredient(recipeID, ingredientID int, req models.AddRecipeIngredientRequest) (*models.RecipeIngredient, error) {
	args := m.Called(recipeID, ingredientID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RecipeIngredient), args.Error(1)
}

func (m *MockIngredientRepository) RemoveRecipeIngredient(recipeID, ingredientID int) error {
	args := m.Called(recipeID, ingredientID)
	return args.Error(0)
}

func (m *MockIngredientRepository) SetRecipeIngredients(recipeID int, ingredients []models.AddRecipeIngredientRequest) error {
	args := m.Called(recipeID, ingredients)
	return args.Error(0)
}

// =============================================================================
// COMPLEX QUERY OPERATIONS
// =============================================================================

func (m *MockIngredientRepository) GetIngredientsForRecipes(recipeIDs []int) (map[int][]models.RecipeIngredient, error) {
	args := m.Called(recipeIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int][]models.RecipeIngredient), args.Error(1)
}

func (m *MockIngredientRepository) GetRecipesUsingIngredient(ingredientID int, params models.PaginationParams) ([]models.Recipe, int, error) {
	args := m.Called(ingredientID, params)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]models.Recipe), args.Int(1), args.Error(2)
}
