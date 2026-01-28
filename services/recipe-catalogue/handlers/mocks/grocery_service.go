package mocks

import (
	"meal-prep/shared/models"

	"github.com/stretchr/testify/mock"
)

type MockGroceryService struct {
	mock.Mock
}

func (m *MockGroceryService) GenerateGroceryList(req models.GroceryListRequest) ([]models.GroceryListItem, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GroceryListItem), args.Error(1)
}
