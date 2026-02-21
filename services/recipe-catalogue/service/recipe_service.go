package service

import (
	"database/sql"
	"meal-prep/services/recipe-catalogue/domain"
	"strings"

	"meal-prep/services/recipe-catalogue/repository"
	"meal-prep/shared/models"
)

type RecipeService interface {
	GetAllRecipes() ([]models.Recipe, error)
	GetRecipeByID(id int) (*models.Recipe, error)
	GetRecipesByCategory(categoryID int) ([]models.Recipe, error)
	CreateRecipe(userID int, req models.CreateRecipeRequest) (*models.Recipe, error)
	UpdateRecipe(userID int, id int, req models.UpdateRecipeRequest) (*models.Recipe, error)
	DeleteRecipe(userID int, id int) error
	GetAllCategories() ([]models.Category, error)

	GetAllRecipesWithIngredients() ([]models.RecipeWithIngredients, error)
	GetRecipeByIDWithIngredients(id int) (*models.RecipeWithIngredients, error)
	GetRecipesByCategoryWithIngredients(categoryID int) ([]models.RecipeWithIngredients, error)
	CreateRecipeWithIngredients(userID int, req models.CreateRecipeWithIngredientsRequest) (*models.RecipeWithIngredients, error)
	UpdateRecipeWithIngredients(userID int, id int, req models.UpdateRecipeWithIngredientsRequest) (*models.RecipeWithIngredients, error)

	SearchRecipesByIngredients(ingredientIDs []int) ([]models.Recipe, error)
	SearchRecipesByIngredientsWithIngredients(ingredientIDs []int) ([]models.RecipeWithIngredients, error)
}

type recipeService struct {
	recipeRepo     repository.RecipeRepository
	categoryRepo   repository.CategoryRepository
	ingredientRepo repository.IngredientRepository
}

func NewRecipeService(recipeRepo repository.RecipeRepository, categoryRepo repository.CategoryRepository, ingredientRepo repository.IngredientRepository) RecipeService {
	return &recipeService{
		recipeRepo:     recipeRepo,
		categoryRepo:   categoryRepo,
		ingredientRepo: ingredientRepo,
	}
}

func (s *recipeService) GetAllRecipes() ([]models.Recipe, error) {
	return s.recipeRepo.GetAll()
}

func (s *recipeService) GetRecipeByID(id int) (*models.Recipe, error) {
	if id <= 0 {
		return nil, domain.ErrRecipeNotFound
	}

	recipe, err := s.recipeRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrRecipeNotFound
		}
		return nil, err
	}

	return recipe, nil
}

func (s *recipeService) GetRecipesByCategory(categoryID int) ([]models.Recipe, error) {
	if categoryID <= 0 {
		return nil, domain.ErrInvalidCategory
	}

	exists, err := s.categoryRepo.Exists(categoryID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrCategoryNotFound
	}

	return s.recipeRepo.GetByCategory(categoryID)
}

func (s *recipeService) CreateRecipe(userID int, req models.CreateRecipeRequest) (*models.Recipe, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)

	if req.Name == "" {
		return nil, domain.ErrRecipeNameRequired
	}

	if req.CategoryID <= 0 {
		return nil, domain.ErrInvalidCategory
	}

	exists, err := s.categoryRepo.Exists(req.CategoryID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrCategoryNotFound
	}

	return s.recipeRepo.Create(userID, req)
}

func (s *recipeService) UpdateRecipe(userID int, id int, req models.UpdateRecipeRequest) (*models.Recipe, error) {
	if req.Name == "" {
		return nil, domain.ErrRecipeNameRequired
	}

	if req.CategoryID <= 0 {
		return nil, domain.ErrInvalidCategory
	}

	if id <= 0 {
		return nil, domain.ErrRecipeNotFound
	}

	// GetOwnerID confirms the recipe exists and returns who owns it.
	// This replaces the old GetByID call, saving one round-trip to the DB.
	ownerID, err := s.recipeRepo.GetOwnerID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrRecipeNotFound
		}
		return nil, err
	}

	if ownerID != userID {
		return nil, domain.ErrForbidden
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)

	if req.CategoryID > 0 {
		exists, err := s.categoryRepo.Exists(req.CategoryID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, domain.ErrCategoryNotFound
		}
	}

	return s.recipeRepo.Update(id, req)
}

func (s *recipeService) DeleteRecipe(userID int, id int) error {
	if id <= 0 {
		return domain.ErrRecipeNotFound
	}

	ownerID, err := s.recipeRepo.GetOwnerID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrRecipeNotFound
		}
		return err
	}

	if ownerID != userID {
		return domain.ErrForbidden
	}

	return s.recipeRepo.Delete(id)
}

func (s *recipeService) GetAllCategories() ([]models.Category, error) {
	return s.categoryRepo.GetAll()
}

func (s *recipeService) GetAllRecipesWithIngredients() ([]models.RecipeWithIngredients, error) {
	return s.recipeRepo.GetAllWithIngredients()
}

func (s *recipeService) GetRecipeByIDWithIngredients(id int) (*models.RecipeWithIngredients, error) {
	if id <= 0 {
		return nil, domain.ErrRecipeNotFound
	}

	recipe, err := s.recipeRepo.GetByIDWithIngredients(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrRecipeNotFound
		}
		return nil, err
	}

	return recipe, nil
}

func (s *recipeService) GetRecipesByCategoryWithIngredients(categoryID int) ([]models.RecipeWithIngredients, error) {
	if categoryID <= 0 {
		return nil, domain.ErrInvalidCategory
	}

	exists, err := s.categoryRepo.Exists(categoryID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrCategoryNotFound
	}

	return s.recipeRepo.GetByCategoryWithIngredients(categoryID)
}

func (s *recipeService) CreateRecipeWithIngredients(userID int, req models.CreateRecipeWithIngredientsRequest) (*models.RecipeWithIngredients, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Description != nil {
		desc := strings.TrimSpace(*req.Description)
		req.Description = &desc
	}

	if req.Name == "" {
		return nil, domain.ErrRecipeNameRequired
	}

	if req.CategoryID <= 0 {
		return nil, domain.ErrInvalidCategory
	}

	exists, err := s.categoryRepo.Exists(req.CategoryID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrCategoryNotFound
	}

	if len(req.Ingredients) > 0 {
		for _, ingredient := range req.Ingredients {
			if ingredient.IngredientID <= 0 {
				return nil, domain.ErrIngredientNotFound
			}

			if ingredient.Quantity <= 0 {
				return nil, domain.ErrInvalidQuantity
			}

			ingredient.Unit = strings.TrimSpace(ingredient.Unit)
			if ingredient.Unit == "" {
				return nil, domain.ErrInvalidUnit
			}

			exists, err := s.ingredientRepo.IngredientExists(ingredient.IngredientID)
			if err != nil {
				return nil, err
			}
			if !exists {
				return nil, domain.ErrIngredientNotFound
			}
		}
	}

	return s.recipeRepo.CreateWithIngredients(userID, req)
}

func (s *recipeService) UpdateRecipeWithIngredients(userID int, id int, req models.UpdateRecipeWithIngredientsRequest) (*models.RecipeWithIngredients, error) {
	if id <= 0 {
		return nil, domain.ErrRecipeNotFound
	}

	ownerID, err := s.recipeRepo.GetOwnerID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrRecipeNotFound
		}
		return nil, err
	}

	if ownerID != userID {
		return nil, domain.ErrForbidden
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, domain.ErrRecipeNameRequired
	}
	req.Name = name

	req.Description = strings.TrimSpace(req.Description)

	exists, err := s.categoryRepo.Exists(req.CategoryID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrCategoryNotFound
	}

	if req.Ingredients != nil {
		for _, ingredient := range req.Ingredients {
			if ingredient.IngredientID <= 0 {
				return nil, domain.ErrIngredientNotFound
			}

			if ingredient.Quantity <= 0 {
				return nil, domain.ErrInvalidQuantity
			}

			ingredient.Unit = strings.TrimSpace(ingredient.Unit)
			if ingredient.Unit == "" {
				return nil, domain.ErrInvalidUnit
			}

			exists, err := s.ingredientRepo.IngredientExists(ingredient.IngredientID)
			if err != nil {
				return nil, err
			}
			if !exists {
				return nil, domain.ErrIngredientNotFound
			}
		}
	}

	return s.recipeRepo.UpdateWithIngredients(id, req)
}

func (s *recipeService) SearchRecipesByIngredients(ingredientIDs []int) ([]models.Recipe, error) {
	if len(ingredientIDs) == 0 {
		return []models.Recipe{}, nil
	}

	for _, ingredientID := range ingredientIDs {
		if ingredientID <= 0 {
			return nil, domain.ErrIngredientNotFound
		}

		exists, err := s.ingredientRepo.IngredientExists(ingredientID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, domain.ErrIngredientNotFound
		}
	}

	return s.recipeRepo.SearchRecipesByIngredients(ingredientIDs)
}

func (s *recipeService) SearchRecipesByIngredientsWithIngredients(ingredientIDs []int) ([]models.RecipeWithIngredients, error) {
	if len(ingredientIDs) == 0 {
		return []models.RecipeWithIngredients{}, nil
	}

	for _, ingredientID := range ingredientIDs {
		if ingredientID <= 0 {
			return nil, domain.ErrIngredientNotFound
		}

		exists, err := s.ingredientRepo.IngredientExists(ingredientID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, domain.ErrIngredientNotFound
		}
	}

	return s.recipeRepo.SearchRecipesByIngredientsWithIngredients(ingredientIDs)
}