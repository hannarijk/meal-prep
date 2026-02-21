package mocks

import (
	"meal-prep/shared/models"

	"github.com/stretchr/testify/mock"
)

// MockRecipeService mocks the recipe service interface
type MockRecipeService struct {
	mock.Mock
}

// =============================================================================
// BASIC RECIPE OPERATIONS
// =============================================================================

func (m *MockRecipeService) GetAllRecipes() ([]models.Recipe, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Recipe), args.Error(1)
}

func (m *MockRecipeService) GetRecipeByID(id int) (*models.Recipe, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Recipe), args.Error(1)
}

func (m *MockRecipeService) GetRecipesByCategory(categoryID int) ([]models.Recipe, error) {
	args := m.Called(categoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Recipe), args.Error(1)
}

func (m *MockRecipeService) CreateRecipe(userID int, req models.CreateRecipeRequest) (*models.Recipe, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Recipe), args.Error(1)
}

func (m *MockRecipeService) UpdateRecipe(userID int, id int, req models.UpdateRecipeRequest) (*models.Recipe, error) {
	args := m.Called(userID, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Recipe), args.Error(1)
}

func (m *MockRecipeService) DeleteRecipe(userID int, id int) error {
	args := m.Called(userID, id)
	return args.Error(0)
}

func (m *MockRecipeService) GetAllCategories() ([]models.Category, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Category), args.Error(1)
}

// =============================================================================
// RECIPES WITH INGREDIENTS OPERATIONS
// =============================================================================

func (m *MockRecipeService) GetAllRecipesWithIngredients() ([]models.RecipeWithIngredients, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.RecipeWithIngredients), args.Error(1)
}

func (m *MockRecipeService) GetRecipeByIDWithIngredients(id int) (*models.RecipeWithIngredients, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RecipeWithIngredients), args.Error(1)
}

func (m *MockRecipeService) GetRecipesByCategoryWithIngredients(categoryID int) ([]models.RecipeWithIngredients, error) {
	args := m.Called(categoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.RecipeWithIngredients), args.Error(1)
}

func (m *MockRecipeService) CreateRecipeWithIngredients(userID int, req models.CreateRecipeWithIngredientsRequest) (*models.RecipeWithIngredients, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RecipeWithIngredients), args.Error(1)
}

func (m *MockRecipeService) UpdateRecipeWithIngredients(userID int, id int, req models.UpdateRecipeWithIngredientsRequest) (*models.RecipeWithIngredients, error) {
	args := m.Called(userID, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RecipeWithIngredients), args.Error(1)
}

// =============================================================================
// SEARCH OPERATIONS
// =============================================================================

func (m *MockRecipeService) SearchRecipesByIngredients(ingredientIDs []int) ([]models.Recipe, error) {
	args := m.Called(ingredientIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Recipe), args.Error(1)
}

func (m *MockRecipeService) SearchRecipesByIngredientsWithIngredients(ingredientIDs []int) ([]models.RecipeWithIngredients, error) {
	args := m.Called(ingredientIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.RecipeWithIngredients), args.Error(1)
}