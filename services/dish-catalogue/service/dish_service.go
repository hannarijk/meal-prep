package service

import (
	"database/sql"
	"errors"
	"strings"

	"meal-prep/services/dish-catalogue/repository"
	"meal-prep/shared/models"
)

var (
	ErrDishNotFound     = errors.New("dish not found")
	ErrCategoryNotFound = errors.New("category not found")
	ErrDishNameRequired = errors.New("dish name is required")
	ErrInvalidCategory  = errors.New("invalid category ID")
)

type DishService interface {
	GetAllDishes() ([]models.Dish, error)
	GetDishByID(id int) (*models.Dish, error)
	GetDishesByCategory(categoryID int) ([]models.Dish, error)
	CreateDish(req models.CreateDishRequest) (*models.Dish, error)
	UpdateDish(id int, req models.UpdateDishRequest) (*models.Dish, error)
	DeleteDish(id int) error
	GetAllCategories() ([]models.Category, error)
}

type dishService struct {
	dishRepo     repository.DishRepository
	categoryRepo repository.CategoryRepository
}

func NewDishService(dishRepo repository.DishRepository, categoryRepo repository.CategoryRepository) DishService {
	return &dishService{
		dishRepo:     dishRepo,
		categoryRepo: categoryRepo,
	}
}

func (s *dishService) GetAllDishes() ([]models.Dish, error) {
	return s.dishRepo.GetAll()
}

func (s *dishService) GetDishByID(id int) (*models.Dish, error) {
	if id <= 0 {
		return nil, ErrDishNotFound
	}

	dish, err := s.dishRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrDishNotFound
		}
		return nil, err
	}

	return dish, nil
}

func (s *dishService) GetDishesByCategory(categoryID int) ([]models.Dish, error) {
	if categoryID <= 0 {
		return nil, ErrInvalidCategory
	}

	exists, err := s.categoryRepo.Exists(categoryID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrCategoryNotFound
	}

	return s.dishRepo.GetByCategory(categoryID)
}

func (s *dishService) CreateDish(req models.CreateDishRequest) (*models.Dish, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)

	if req.Name == "" {
		return nil, ErrDishNameRequired
	}

	if req.CategoryID <= 0 {
		return nil, ErrInvalidCategory
	}

	exists, err := s.categoryRepo.Exists(req.CategoryID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrCategoryNotFound
	}

	return s.dishRepo.Create(req)
}

func (s *dishService) UpdateDish(id int, req models.UpdateDishRequest) (*models.Dish, error) {
	if id <= 0 {
		return nil, ErrDishNotFound
	}

	// Check if dish exists
	_, err := s.dishRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrDishNotFound
		}
		return nil, err
	}

	// Validate and clean input
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)

	// If category is being updated, validate it exists
	if req.CategoryID > 0 {
		exists, err := s.categoryRepo.Exists(req.CategoryID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrCategoryNotFound
		}
	}

	return s.dishRepo.Update(id, req)
}

func (s *dishService) DeleteDish(id int) error {
	if id <= 0 {
		return ErrDishNotFound
	}

	err := s.dishRepo.Delete(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrDishNotFound
		}
		return err
	}

	return nil
}

func (s *dishService) GetAllCategories() ([]models.Category, error) {
	return s.categoryRepo.GetAll()
}
