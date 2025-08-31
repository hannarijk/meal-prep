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
	RecipeID int       `json:"recipe_id"`
	CookedAt time.Time `json:"cooked_at"`
	Rating   *int      `json:"rating,omitempty"`
}

type RecommendationRequest struct {
	Limit      int    `json:"limit,omitempty"`
	Algorithm  string `json:"algorithm,omitempty"`
	Categories []int  `json:"categories,omitempty"`
}

type RecipeWithScore struct {
	Recipe              `json:"recipe"`
	RecommendationScore float64    `json:"recommendation_score"`
	LastCookedAt        *time.Time `json:"last_cooked_at,omitempty"`
	DaysSinceCooked     *int       `json:"days_since_cooked,omitempty"`
	Reason              string     `json:"reason"`
}

type RecommendationResponse struct {
	Recipes     []RecipeWithScore `json:"recipes"`
	Algorithm   string            `json:"algorithm"`
	GeneratedAt time.Time         `json:"generated_at"`
	TotalScored int               `json:"total_scored"`
}

type UpdatePreferencesRequest struct {
	PreferredCategories []int `json:"preferred_categories"`
}

type LogCookingRequest struct {
	RecipeID int  `json:"recipe_id"`
	Rating   *int `json:"rating,omitempty"`
}
