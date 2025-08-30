package models

import "time"

type UserPreferences struct {
	ID                  int       `json:"id"`
	UserID              int       `json:"user_id"`
	PreferredCategories []int     `json:"preferred_categories"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type CookingHistory struct {
	ID       int       `json:"id"`
	UserID   int       `json:"user_id"`
	DishID   int       `json:"dish_id"`
	CookedAt time.Time `json:"cooked_at"`
	Rating   *int      `json:"rating,omitempty"`
}

type RecommendationRequest struct {
	Limit      int    `json:"limit,omitempty"`
	Algorithm  string `json:"algorithm,omitempty"` // "preference", "time_decay", "hybrid"
	Categories []int  `json:"categories,omitempty"`
}

type DishWithScore struct {
	Dish                `json:"dish"`
	RecommendationScore float64    `json:"recommendation_score"`
	LastCookedAt        *time.Time `json:"last_cooked_at,omitempty"`
	DaysSinceCooked     *int       `json:"days_since_cooked,omitempty"`
	Reason              string     `json:"reason"`
}

type RecommendationResponse struct {
	Dishes      []DishWithScore `json:"dishes"`
	Algorithm   string          `json:"algorithm"`
	GeneratedAt time.Time       `json:"generated_at"`
	TotalScored int             `json:"total_scored"`
}

type UpdatePreferencesRequest struct {
	PreferredCategories []int `json:"preferred_categories"`
}

type LogCookingRequest struct {
	DishID int  `json:"dish_id"`
	Rating *int `json:"rating,omitempty"`
}
