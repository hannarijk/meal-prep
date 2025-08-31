package models

import "time"

type Category struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
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
