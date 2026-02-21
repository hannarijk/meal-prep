package mocks

import (
	"github.com/stretchr/testify/mock"
	"meal-prep/shared/models"
)

type MockRecipeRepository struct {
	mock.Mock
}

func (m *MockRecipeRepository) GetAll(params models.PaginationParams) ([]models.Recipe, int, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]models.Recipe), args.Int(1), args.Error(2)
}

func (m *MockRecipeRepository) GetByID(id int) (*models.Recipe, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Recipe), args.Error(1)
}

func (m *MockRecipeRepository) GetByCategory(categoryID int, params models.PaginationParams) ([]models.Recipe, int, error) {
	args := m.Called(categoryID, params)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]models.Recipe), args.Int(1), args.Error(2)
}

func (m *MockRecipeRepository) GetOwnerID(id int) (int, error) {
	args := m.Called(id)
	return args.Int(0), args.Error(1)
}

func (m *MockRecipeRepository) Create(userID int, req models.CreateRecipeRequest) (*models.Recipe, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Recipe), args.Error(1)
}

func (m *MockRecipeRepository) Update(id int, req models.UpdateRecipeRequest) (*models.Recipe, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Recipe), args.Error(1)
}

func (m *MockRecipeRepository) Delete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRecipeRepository) GetAllWithIngredients(params models.PaginationParams) ([]models.RecipeWithIngredients, int, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]models.RecipeWithIngredients), args.Int(1), args.Error(2)
}

func (m *MockRecipeRepository) GetByIDWithIngredients(id int) (*models.RecipeWithIngredients, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RecipeWithIngredients), args.Error(1)
}

func (m *MockRecipeRepository) GetByCategoryWithIngredients(categoryID int, params models.PaginationParams) ([]models.RecipeWithIngredients, int, error) {
	args := m.Called(categoryID, params)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]models.RecipeWithIngredients), args.Int(1), args.Error(2)
}

func (m *MockRecipeRepository) CreateWithIngredients(userID int, req models.CreateRecipeWithIngredientsRequest) (*models.RecipeWithIngredients, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RecipeWithIngredients), args.Error(1)
}

func (m *MockRecipeRepository) UpdateWithIngredients(id int, req models.UpdateRecipeWithIngredientsRequest) (*models.RecipeWithIngredients, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RecipeWithIngredients), args.Error(1)
}

func (m *MockRecipeRepository) SearchRecipesByIngredients(ingredientIDs []int, params models.PaginationParams) ([]models.Recipe, int, error) {
	args := m.Called(ingredientIDs, params)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]models.Recipe), args.Int(1), args.Error(2)
}

func (m *MockRecipeRepository) SearchRecipesByIngredientsWithIngredients(ingredientIDs []int, params models.PaginationParams) ([]models.RecipeWithIngredients, int, error) {
	args := m.Called(ingredientIDs, params)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]models.RecipeWithIngredients), args.Int(1), args.Error(2)
}
