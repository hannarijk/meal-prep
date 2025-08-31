package service

import (
	"database/sql"
	"errors"
	"strings"

	"meal-prep/services/recipe-catalogue/repository"
	"meal-prep/shared/models"
)

var (
	ErrRecipeNotFound     = errors.New("recipe not found")
	ErrCategoryNotFound   = errors.New("category not found")
	ErrRecipeNameRequired = errors.New("recipe name is required")
	ErrInvalidCategory    = errors.New("invalid category ID")
)

type RecipeService interface {
	GetAllRecipes() ([]models.Recipe, error)
	GetRecipeByID(id int) (*models.Recipe, error)
	GetRecipesByCategory(categoryID int) ([]models.Recipe, error)
	CreateRecipe(req models.CreateRecipeRequest) (*models.Recipe, error)
	UpdateRecipe(id int, req models.UpdateRecipeRequest) (*models.Recipe, error)
	DeleteRecipe(id int) error
	GetAllCategories() ([]models.Category, error)
}

type recipeService struct {
	recipeRepo   repository.RecipeRepository
	categoryRepo repository.CategoryRepository
}

func NewRecipeService(recipeRepo repository.RecipeRepository, categoryRepo repository.CategoryRepository) RecipeService {
	return &recipeService{
		recipeRepo:   recipeRepo,
		categoryRepo: categoryRepo,
	}
}

func (s *recipeService) GetAllRecipes() ([]models.Recipe, error) {
	return s.recipeRepo.GetAll()
}

func (s *recipeService) GetRecipeByID(id int) (*models.Recipe, error) {
	if id <= 0 {
		return nil, ErrRecipeNotFound
	}

	recipe, err := s.recipeRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecipeNotFound
		}
		return nil, err
	}

	return recipe, nil
}

func (s *recipeService) GetRecipesByCategory(categoryID int) ([]models.Recipe, error) {
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

	return s.recipeRepo.GetByCategory(categoryID)
}

func (s *recipeService) CreateRecipe(req models.CreateRecipeRequest) (*models.Recipe, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)

	if req.Name == "" {
		return nil, ErrRecipeNameRequired
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

	return s.recipeRepo.Create(req)
}

func (s *recipeService) UpdateRecipe(id int, req models.UpdateRecipeRequest) (*models.Recipe, error) {
	if id <= 0 {
		return nil, ErrRecipeNotFound
	}

	// Check if recipe exists
	_, err := s.recipeRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecipeNotFound
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

	return s.recipeRepo.Update(id, req)
}

func (s *recipeService) DeleteRecipe(id int) error {
	if id <= 0 {
		return ErrRecipeNotFound
	}

	err := s.recipeRepo.Delete(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrRecipeNotFound
		}
		return err
	}

	return nil
}

func (s *recipeService) GetAllCategories() ([]models.Category, error) {
	return s.categoryRepo.GetAll()
}
