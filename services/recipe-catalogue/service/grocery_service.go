package service

import (
	"meal-prep/services/recipe-catalogue/repository"
	"meal-prep/shared/models"
)

type GroceryService interface {
	GenerateGroceryList(req models.GroceryListRequest) ([]models.GroceryListItem, error)
}

type groceryService struct {
	ingredientRepo repository.IngredientRepository
	recipeRepo     repository.RecipeRepository
}

func NewGroceryService(ingredientRepo repository.IngredientRepository, recipeRepo repository.RecipeRepository) GroceryService {
	return &groceryService{
		ingredientRepo: ingredientRepo,
		recipeRepo:     recipeRepo,
	}
}

func (s *groceryService) GenerateGroceryList(req models.GroceryListRequest) ([]models.GroceryListItem, error) {
	if len(req.RecipeIDs) == 0 {
		return []models.GroceryListItem{}, nil
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
	aggregatedIngredients := make(map[int]*models.GroceryListItem)

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
				aggregatedIngredients[key] = &models.GroceryListItem{
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
	var groceryList []models.GroceryListItem
	for _, item := range aggregatedIngredients {
		groceryList = append(groceryList, *item)
	}

	return groceryList, nil
}
