package domain

import "errors"

var (
	// Recipe-specific (only recipe-catalogue uses these)
	ErrRecipeNotFound     = errors.New("recipe not found")
	ErrRecipeNameRequired = errors.New("recipe name is required")
	ErrCategoryNotFound   = errors.New("category not found")
	ErrInvalidCategory    = errors.New("invalid category ID")
	ErrForbidden          = errors.New("you do not have permission to modify this recipe")

	// Ingredient-specific (only recipe-catalogue uses these)
	ErrIngredientNotFound     = errors.New("ingredient not found")
	ErrIngredientExists       = errors.New("ingredient already exists")
	ErrIngredientNameRequired = errors.New("ingredient with this name already exists")
	ErrCannotDeleteIngredient = errors.New("cannot delete ingredient - it is used in recipes")

	// Recipe-Ingredient relationship (only recipe-catalogue uses these)
	ErrRecipeIngredientAlreadyExists = errors.New("ingredient already added to this recipe")
	ErrInvalidQuantity               = errors.New("quantity must be greater than 0")
	ErrInvalidUnit                   = errors.New("unit is required")
)
