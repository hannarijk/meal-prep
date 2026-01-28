package models

import "time"

type Ingredient struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Category    *string   `json:"category,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type RecipeIngredient struct {
	ID           int        `json:"id"`
	RecipeID     int        `json:"recipe_id"`
	IngredientID int        `json:"ingredient_id"`
	Ingredient   Ingredient `json:"ingredient"`
	Quantity     float64    `json:"quantity"`
	Unit         string     `json:"unit"` // cups, grams, pieces, tablespoons, etc.
	Notes        *string    `json:"notes,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type CreateIngredientRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Category    *string `json:"category,omitempty"`
}

type UpdateIngredientRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}
