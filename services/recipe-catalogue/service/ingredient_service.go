package service

import (
	"database/sql"
	"meal-prep/services/recipe-catalogue/domain"
	"strings"

	"meal-prep/services/recipe-catalogue/repository"
	"meal-prep/shared/models"
)

type IngredientService interface {
	GetAllIngredients(params models.PaginationParams) ([]models.Ingredient, models.PaginationMeta, error)
	GetIngredientByID(id int) (*models.Ingredient, error)
	GetIngredientsByCategory(category string, params models.PaginationParams) ([]models.Ingredient, models.PaginationMeta, error)
	SearchIngredients(query string, params models.PaginationParams) ([]models.Ingredient, models.PaginationMeta, error)
	CreateIngredient(req models.CreateIngredientRequest) (*models.Ingredient, error)
	UpdateIngredient(id int, req models.UpdateIngredientRequest) (*models.Ingredient, error)
	DeleteIngredient(id int) error

	GetRecipeIngredients(recipeID int) ([]models.RecipeIngredient, error)
	AddRecipeIngredient(recipeID int, req models.AddRecipeIngredientRequest) (*models.RecipeIngredient, error)
	UpdateRecipeIngredient(recipeID, ingredientID int, req models.AddRecipeIngredientRequest) (*models.RecipeIngredient, error)
	RemoveRecipeIngredient(recipeID, ingredientID int) error
	SetRecipeIngredients(recipeID int, ingredients []models.AddRecipeIngredientRequest) error

	GetRecipesUsingIngredient(ingredientID int, params models.PaginationParams) ([]models.Recipe, models.PaginationMeta, error)
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

func (s *ingredientService) GetAllIngredients(params models.PaginationParams) ([]models.Ingredient, models.PaginationMeta, error) {
	ingredients, total, err := s.ingredientRepo.GetAllIngredients(params)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}
	return ingredients, models.NewPaginationMeta(params, total), nil
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

func (s *ingredientService) GetIngredientsByCategory(category string, params models.PaginationParams) ([]models.Ingredient, models.PaginationMeta, error) {
	category = strings.TrimSpace(category)
	if category == "" {
		return s.GetAllIngredients(params)
	}

	ingredients, total, err := s.ingredientRepo.GetIngredientsByCategory(category, params)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}
	return ingredients, models.NewPaginationMeta(params, total), nil
}

func (s *ingredientService) SearchIngredients(query string, params models.PaginationParams) ([]models.Ingredient, models.PaginationMeta, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return s.GetAllIngredients(params)
	}

	ingredients, total, err := s.ingredientRepo.SearchIngredients(query, params)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}
	return ingredients, models.NewPaginationMeta(params, total), nil
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

	// Check if ingredient is used in any recipes — fetch just 1 to see if any exist
	_, total, err := s.ingredientRepo.GetRecipesUsingIngredient(id, models.PaginationParams{Page: 1, PerPage: 1})
	if err != nil {
		return err
	}

	if total > 0 {
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

func (s *ingredientService) GetRecipesUsingIngredient(ingredientID int, params models.PaginationParams) ([]models.Recipe, models.PaginationMeta, error) {
	if ingredientID <= 0 {
		return nil, models.PaginationMeta{}, domain.ErrIngredientNotFound
	}

	// Verify ingredient exists
	_, err := s.ingredientRepo.GetIngredientByID(ingredientID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.PaginationMeta{}, domain.ErrIngredientNotFound
		}
		return nil, models.PaginationMeta{}, err
	}

	recipes, total, err := s.ingredientRepo.GetRecipesUsingIngredient(ingredientID, params)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}
	return recipes, models.NewPaginationMeta(params, total), nil
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
