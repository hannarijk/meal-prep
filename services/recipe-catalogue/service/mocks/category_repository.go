package mocks

import (
	"github.com/stretchr/testify/mock"
	"meal-prep/shared/models"
)

type MockCategoryRepository struct {
	mock.Mock
}

func (m *MockCategoryRepository) GetAll() ([]models.Category, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Category), args.Error(1)
}

func (m *MockCategoryRepository) GetByID(id int) (*models.Category, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockCategoryRepository) Exists(id int) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockCategoryRepository) Create(req models.CreateCategoryRequest) (*models.Category, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockCategoryRepository) Update(id int, req models.UpdateCategoryRequest) (*models.Category, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockCategoryRepository) Delete(id int) error {
	//TODO implement me
	panic("implement me")
}
