package models

import "time"

type Category struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type Dish struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CategoryID  *int      `json:"category_id,omitempty"`
	Category    *Category `json:"category,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateDishRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CategoryID  int    `json:"category_id"`
}

type UpdateDishRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	CategoryID  int    `json:"category_id,omitempty"`
}
