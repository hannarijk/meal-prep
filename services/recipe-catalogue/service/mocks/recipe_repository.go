package mocks

import (
	"github.com/stretchr/testify/mock"
	"meal-prep/shared/models"
)

type MockRecipeRepository struct {
	mock.Mock
}

func (m *MockRecipeRepository) GetAll() ([]models.Recipe, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Recipe), args.Error(1)
}

func (m *MockRecipeRepository) GetByID(id int) (*models.Recipe, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Recipe), args.Error(1)
}

func (m *MockRecipeRepository) GetByCategory(categoryID int) ([]models.Recipe, error) {
	args := m.Called(categoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Recipe), args.Error(1)
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

func (m *MockRecipeRepository) GetAllWithIngredients() ([]models.RecipeWithIngredients, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.RecipeWithIngredients), args.Error(1)
}

func (m *MockRecipeRepository) GetByIDWithIngredients(id int) (*models.RecipeWithIngredients, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RecipeWithIngredients), args.Error(1)
}

func (m *MockRecipeRepository) GetByCategoryWithIngredients(categoryID int) ([]models.RecipeWithIngredients, error) {
	args := m.Called(categoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.RecipeWithIngredients), args.Error(1)
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

func (m *MockRecipeRepository) SearchRecipesByIngredients(ingredientIDs []int) ([]models.Recipe, error) {
	args := m.Called(ingredientIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Recipe), args.Error(1)
}

func (m *MockRecipeRepository) SearchRecipesByIngredientsWithIngredients(ingredientIDs []int) ([]models.RecipeWithIngredients, error) {
	args := m.Called(ingredientIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.RecipeWithIngredients), args.Error(1)
}