package models

import "time"

type Category struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateCategoryRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type UpdateCategoryRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type Recipe struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CategoryID  *int      `json:"category_id,omitempty"`
	Category    *Category `json:"category,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateRecipeRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CategoryID  int    `json:"category_id"`
}

type UpdateRecipeRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	CategoryID  int    `json:"category_id,omitempty"`
}

type RecipeWithIngredients struct {
	Recipe      `json:"recipe"`
	Ingredients []RecipeIngredient `json:"ingredients"`
}

type AddRecipeIngredientRequest struct {
	IngredientID int     `json:"ingredient_id"`
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
	Notes        *string `json:"notes,omitempty"`
}

type CreateRecipeWithIngredientsRequest struct {
	Name        string                       `json:"name"`
	Description *string                      `json:"description,omitempty"`
	CategoryID  int                          `json:"category_id"`
	Ingredients []AddRecipeIngredientRequest `json:"ingredients,omitempty"`
}

type UpdateRecipeWithIngredientsRequest struct {
	Name        *string                      `json:"name,omitempty"`
	Description *string                      `json:"description,omitempty"`
	CategoryID  *int                         `json:"category_id,omitempty"`
	Ingredients []AddRecipeIngredientRequest `json:"ingredients,omitempty"`
}
