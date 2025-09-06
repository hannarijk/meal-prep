package service

import (
	"database/sql"
	"errors"
	"strings"

	"meal-prep/services/recipe-catalogue/repository"
	"meal-prep/shared/models"
)

var (
	ErrIngredientNotFound     = errors.New("ingredient not found")
	ErrIngredientNameRequired = errors.New("ingredient name is required")
	ErrIngredientExists       = errors.New("ingredient with this name already exists")
	ErrInvalidQuantity        = errors.New("quantity must be greater than 0")
	ErrInvalidUnit            = errors.New("unit is required")
	ErrCannotDeleteIngredient = errors.New("cannot delete ingredient - it is used in recipes")
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
	GenerateShoppingList(req models.ShoppingListRequest) ([]models.ShoppingListItem, error)
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
		return nil, ErrIngredientNotFound
	}

	ingredient, err := s.ingredientRepo.GetIngredientByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrIngredientNotFound
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
		return nil, ErrIngredientNameRequired
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
		return nil, ErrIngredientNotFound
	}

	// Check if ingredient exists
	_, err := s.ingredientRepo.GetIngredientByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrIngredientNotFound
		}
		return nil, err
	}

	// Validate and clean input
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, ErrIngredientNameRequired
		}
		req.Name = &name
	}

	if req.Description != nil {
		desc := strings.TrimSpace(*req.Description)
		req.Description = &desc
	}

	if req.Category != nil {
		cat := strings.TrimSpace(*req.Category)
		req.Category = &cat
	}

	return s.ingredientRepo.UpdateIngredient(id, req)
}

func (s *ingredientService) DeleteIngredient(id int) error {
	if id <= 0 {
		return ErrIngredientNotFound
	}

	// Check if ingredient exists
	_, err := s.ingredientRepo.GetIngredientByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrIngredientNotFound
		}
		return err
	}

	// Check if ingredient is used in any recipes
	recipes, err := s.ingredientRepo.GetRecipesUsingIngredient(id)
	if err != nil {
		return err
	}

	if len(recipes) > 0 {
		return ErrCannotDeleteIngredient
	}

	err = s.ingredientRepo.DeleteIngredient(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrIngredientNotFound
		}
		return err
	}

	return nil
}

func (s *ingredientService) GetRecipeIngredients(recipeID int) ([]models.RecipeIngredient, error) {
	if recipeID <= 0 {
		return nil, ErrRecipeNotFound
	}

	// Verify recipe exists
	_, err := s.recipeRepo.GetByID(recipeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecipeNotFound
		}
		return nil, err
	}

	return s.ingredientRepo.GetRecipeIngredients(recipeID)
}

func (s *ingredientService) AddRecipeIngredient(recipeID int, req models.AddRecipeIngredientRequest) (*models.RecipeIngredient, error) {
	if recipeID <= 0 {
		return nil, ErrRecipeNotFound
	}

	// Validate request
	if err := s.validateRecipeIngredientRequest(req); err != nil {
		return nil, err
	}

	// Verify recipe exists
	_, err := s.recipeRepo.GetByID(recipeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecipeNotFound
		}
		return nil, err
	}

	// Verify ingredient exists
	exists, err := s.ingredientRepo.IngredientExists(req.IngredientID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrIngredientNotFound
	}

	return s.ingredientRepo.AddRecipeIngredient(recipeID, req)
}

func (s *ingredientService) UpdateRecipeIngredient(recipeID, ingredientID int, req models.AddRecipeIngredientRequest) (*models.RecipeIngredient, error) {
	if recipeID <= 0 {
		return nil, ErrRecipeNotFound
	}
	if ingredientID <= 0 {
		return nil, ErrIngredientNotFound
	}

	// Validate request
	if err := s.validateRecipeIngredientRequest(req); err != nil {
		return nil, err
	}

	// Verify recipe exists
	_, err := s.recipeRepo.GetByID(recipeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecipeNotFound
		}
		return nil, err
	}

	// Update uses the ingredientID from the URL, not the request
	req.IngredientID = ingredientID

	return s.ingredientRepo.UpdateRecipeIngredient(recipeID, ingredientID, req)
}

func (s *ingredientService) RemoveRecipeIngredient(recipeID, ingredientID int) error {
	if recipeID <= 0 {
		return ErrRecipeNotFound
	}
	if ingredientID <= 0 {
		return ErrIngredientNotFound
	}

	// Verify recipe exists
	_, err := s.recipeRepo.GetByID(recipeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrRecipeNotFound
		}
		return err
	}

	err = s.ingredientRepo.RemoveRecipeIngredient(recipeID, ingredientID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrIngredientNotFound
		}
		return err
	}

	return nil
}

func (s *ingredientService) SetRecipeIngredients(recipeID int, ingredients []models.AddRecipeIngredientRequest) error {
	if recipeID <= 0 {
		return ErrRecipeNotFound
	}

	// Verify recipe exists
	_, err := s.recipeRepo.GetByID(recipeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrRecipeNotFound
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
			return ErrIngredientNotFound
		}
	}

	return s.ingredientRepo.SetRecipeIngredients(recipeID, ingredients)
}

func (s *ingredientService) GetRecipesUsingIngredient(ingredientID int) ([]models.Recipe, error) {
	if ingredientID <= 0 {
		return nil, ErrIngredientNotFound
	}

	// Verify ingredient exists
	_, err := s.ingredientRepo.GetIngredientByID(ingredientID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrIngredientNotFound
		}
		return nil, err
	}

	return s.ingredientRepo.GetRecipesUsingIngredient(ingredientID)
}

func (s *ingredientService) GenerateShoppingList(req models.ShoppingListRequest) ([]models.ShoppingListItem, error) {
	if len(req.RecipeIDs) == 0 {
		return []models.ShoppingListItem{}, nil
	}

	// Get ingredients for all recipes
	ingredientsMap, err := s.ingredientRepo.GetIngredientsForRecipes(req.RecipeIDs)
	if err != nil {
		return nil, err
	}

	// Get recipe names for reference
	recipeNamesMap := make(map[int]string)
	for _, recipeID := range req.RecipeIDs {
		recipe, err := s.recipeRepo.GetByID(recipeID)
		if err == nil {
			recipeNamesMap[recipeID] = recipe.Name
		}
	}

	// Aggregate ingredients by ingredient ID
	aggregatedIngredients := make(map[int]*models.ShoppingListItem)

	for recipeID, ingredients := range ingredientsMap {
		recipeName := recipeNamesMap[recipeID]

		for _, ingredient := range ingredients {
			key := ingredient.IngredientID

			if existing, exists := aggregatedIngredients[key]; exists {
				// Same ingredient from another recipe - add quantities if same unit
				if existing.Unit == ingredient.Unit {
					existing.TotalQuantity += ingredient.Quantity
				} else {
					// Different units - note in recipes but don't add quantities
					existing.TotalQuantity = -1 // Flag for manual calculation
				}
				existing.Recipes = append(existing.Recipes, recipeName)
			} else {
				// New ingredient
				aggregatedIngredients[key] = &models.ShoppingListItem{
					IngredientID:  ingredient.IngredientID,
					Ingredient:    ingredient.Ingredient,
					TotalQuantity: ingredient.Quantity,
					Unit:          ingredient.Unit,
					Recipes:       []string{recipeName},
				}
			}
		}
	}

	// Convert map to slice
	var shoppingList []models.ShoppingListItem
	for _, item := range aggregatedIngredients {
		shoppingList = append(shoppingList, *item)
	}

	return shoppingList, nil
}

func (s *ingredientService) validateRecipeIngredientRequest(req models.AddRecipeIngredientRequest) error {
	if req.IngredientID <= 0 {
		return ErrIngredientNotFound
	}

	if req.Quantity <= 0 {
		return ErrInvalidQuantity
	}

	req.Unit = strings.TrimSpace(req.Unit)
	if req.Unit == "" {
		return ErrInvalidUnit
	}

	if req.Notes != nil {
		notes := strings.TrimSpace(*req.Notes)
		req.Notes = &notes
	}

	return nil
}
