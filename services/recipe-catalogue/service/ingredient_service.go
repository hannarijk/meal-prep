package service

import (
	"database/sql"
	"meal-prep/services/recipe-catalogue/domain"
	"strings"

	"meal-prep/services/recipe-catalogue/repository"
	"meal-prep/shared/models"
)

type IngredientService interface {
	GetAllIngredients() ([]models.Ingredient, error)
	GetIngredientByID(id int) (*models.Ingredient, error)
	GetIngredientsByCategory(category string) ([]models.Ingredient, error)
	SearchIngredients(query string) ([]models.Ingredient, error)
	CreateIngredient(req models.CreateIngredientRequest) (*models.Ingredient, error)
	UpdateIngredient(id int, req models.UpdateIngredientRequest) (*models.Ingredient, error)
	DeleteIngredient(id int) error

	GetRecipeIngredients(recipeID int) ([]models.RecipeIngredient, error)
	AddRecipeIngredient(recipeID int, req models.AddRecipeIngredientRequest) (*models.RecipeIngredient, error)
	UpdateRecipeIngredient(recipeID, ingredientID int, req models.AddRecipeIngredientRequest) (*models.RecipeIngredient, error)
	RemoveRecipeIngredient(recipeID, ingredientID int) error
	SetRecipeIngredients(recipeID int, ingredients []models.AddRecipeIngredientRequest) error

	GetRecipesUsingIngredient(ingredientID int) ([]models.Recipe, error)
}

type ingredientService struct {
	ingredientRepo repository.IngredientRepository
	recipeRepo     repository.RecipeRepository
}

func NewIngredientService(ingredientRepo repository.IngredientRepository, recipeRepo repository.RecipeRepository) IngredientService {
	return &ingredientService{
		ingredientRepo: ingredientRepo,
		recipeRepo:     recipeRepo,
	}
}

func (s *ingredientService) GetAllIngredients() ([]models.Ingredient, error) {
	return s.ingredientRepo.GetAllIngredients()
}

func (s *ingredientService) GetIngredientByID(id int) (*models.Ingredient, error) {
	if id <= 0 {
		return nil, domain.ErrIngredientNotFound
	}

	ingredient, err := s.ingredientRepo.GetIngredientByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrIngredientNotFound
		}
		return nil, err
	}

	return ingredient, nil
}

func (s *ingredientService) GetIngredientsByCategory(category string) ([]models.Ingredient, error) {
	category = strings.TrimSpace(category)
	if category == "" {
		return s.ingredientRepo.GetAllIngredients()
	}

	return s.ingredientRepo.GetIngredientsByCategory(category)
}

func (s *ingredientService) SearchIngredients(query string) ([]models.Ingredient, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return s.ingredientRepo.GetAllIngredients()
	}

	return s.ingredientRepo.SearchIngredients(query)
}

func (s *ingredientService) CreateIngredient(req models.CreateIngredientRequest) (*models.Ingredient, error) {
	// Validate input
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return nil, domain.ErrIngredientNameRequired
	}

	if req.Description != nil {
		desc := strings.TrimSpace(*req.Description)
		req.Description = &desc
	}

	if req.Category != nil {
		cat := strings.TrimSpace(*req.Category)
		req.Category = &cat
	}

	return s.ingredientRepo.CreateIngredient(req)
}

func (s *ingredientService) UpdateIngredient(id int, req models.UpdateIngredientRequest) (*models.Ingredient, error) {
	if id <= 0 {
		return nil, domain.ErrIngredientNotFound
	}

	// Check if ingredient exists
	_, err := s.ingredientRepo.GetIngredientByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrIngredientNotFound
		}
		return nil, err
	}

	// Validate and clean input
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, domain.ErrIngredientNameRequired
	}
	req.Name = name

	desc := strings.TrimSpace(req.Description)
	req.Description = desc

	cat := strings.TrimSpace(req.Category)
	req.Category = cat

	return s.ingredientRepo.UpdateIngredient(id, req)
}

func (s *ingredientService) DeleteIngredient(id int) error {
	if id <= 0 {
		return domain.ErrIngredientNotFound
	}

	// Check if ingredient exists
	_, err := s.ingredientRepo.GetIngredientByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrIngredientNotFound
		}
		return err
	}

	// Check if ingredient is used in any recipes
	recipes, err := s.ingredientRepo.GetRecipesUsingIngredient(id)
	if err != nil {
		return err
	}

	if len(recipes) > 0 {
		return domain.ErrCannotDeleteIngredient
	}

	err = s.ingredientRepo.DeleteIngredient(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrIngredientNotFound
		}
		return err
	}

	return nil
}

func (s *ingredientService) GetRecipeIngredients(recipeID int) ([]models.RecipeIngredient, error) {
	if recipeID <= 0 {
		return nil, domain.ErrRecipeNotFound
	}

	// Verify recipe exists
	_, err := s.recipeRepo.GetByID(recipeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrRecipeNotFound
		}
		return nil, err
	}

	return s.ingredientRepo.GetRecipeIngredients(recipeID)
}

func (s *ingredientService) AddRecipeIngredient(recipeID int, req models.AddRecipeIngredientRequest) (*models.RecipeIngredient, error) {
	if recipeID <= 0 {
		return nil, domain.ErrRecipeNotFound
	}

	// Validate request
	if err := s.validateRecipeIngredientRequest(req); err != nil {
		return nil, err
	}

	// Verify recipe exists
	_, err := s.recipeRepo.GetByID(recipeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrRecipeNotFound
		}
		return nil, err
	}

	// Verify ingredient exists
	exists, err := s.ingredientRepo.IngredientExists(req.IngredientID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrIngredientNotFound
	}

	return s.ingredientRepo.AddRecipeIngredient(recipeID, req)
}

func (s *ingredientService) UpdateRecipeIngredient(recipeID, ingredientID int, req models.AddRecipeIngredientRequest) (*models.RecipeIngredient, error) {
	if recipeID <= 0 {
		return nil, domain.ErrRecipeNotFound
	}
	if ingredientID <= 0 {
		return nil, domain.ErrIngredientNotFound
	}

	// Validate request
	if err := s.validateRecipeIngredientRequest(req); err != nil {
		return nil, err
	}

	// Verify recipe exists
	_, err := s.recipeRepo.GetByID(recipeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrRecipeNotFound
		}
		return nil, err
	}

	// Update uses the ingredientID from the URL, not the request
	req.IngredientID = ingredientID

	return s.ingredientRepo.UpdateRecipeIngredient(recipeID, ingredientID, req)
}

func (s *ingredientService) RemoveRecipeIngredient(recipeID, ingredientID int) error {
	if recipeID <= 0 {
		return domain.ErrRecipeNotFound
	}
	if ingredientID <= 0 {
		return domain.ErrIngredientNotFound
	}

	// Verify recipe exists
	_, err := s.recipeRepo.GetByID(recipeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrRecipeNotFound
		}
		return err
	}

	err = s.ingredientRepo.RemoveRecipeIngredient(recipeID, ingredientID)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrIngredientNotFound
		}
		return err
	}

	return nil
}

func (s *ingredientService) SetRecipeIngredients(recipeID int, ingredients []models.AddRecipeIngredientRequest) error {
	if recipeID <= 0 {
		return domain.ErrRecipeNotFound
	}

	// Verify recipe exists
	_, err := s.recipeRepo.GetByID(recipeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrRecipeNotFound
		}
		return err
	}

	// Validate all ingredients
	for _, ingredient := range ingredients {
		if err := s.validateRecipeIngredientRequest(ingredient); err != nil {
			return err
		}

		// Verify ingredient exists
		exists, err := s.ingredientRepo.IngredientExists(ingredient.IngredientID)
		if err != nil {
			return err
		}
		if !exists {
			return domain.ErrIngredientNotFound
		}
	}

	return s.ingredientRepo.SetRecipeIngredients(recipeID, ingredients)
}

func (s *ingredientService) GetRecipesUsingIngredient(ingredientID int) ([]models.Recipe, error) {
	if ingredientID <= 0 {
		return nil, domain.ErrIngredientNotFound
	}

	// Verify ingredient exists
	_, err := s.ingredientRepo.GetIngredientByID(ingredientID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrIngredientNotFound
		}
		return nil, err
	}

	return s.ingredientRepo.GetRecipesUsingIngredient(ingredientID)
}

func (s *ingredientService) validateRecipeIngredientRequest(req models.AddRecipeIngredientRequest) error {
	if req.IngredientID <= 0 {
		return domain.ErrIngredientNotFound
	}

	if req.Quantity <= 0 {
		return domain.ErrInvalidQuantity
	}

	req.Unit = strings.TrimSpace(req.Unit)
	if req.Unit == "" {
		return domain.ErrInvalidUnit
	}

	if req.Notes != nil {
		notes := strings.TrimSpace(*req.Notes)
		req.Notes = &notes
	}

	return nil
}
