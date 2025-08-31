package service

import (
	"database/sql"
	"errors"
	"time"

	"meal-prep/services/recommendations/repository"
	"meal-prep/shared/models"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrRecipeNotFound    = errors.New("recipe not found")
	ErrInvalidAlgorithm  = errors.New("invalid recommendation algorithm")
	ErrInvalidRating     = errors.New("rating must be between 1 and 5")
	ErrInvalidLimit      = errors.New("limit must be between 1 and 50")
	ErrPreferencesNotSet = errors.New("user preferences not configured")
)

const (
	AlgorithmPreference = "preference"
	AlgorithmTimeDecay  = "time_decay"
	AlgorithmHybrid     = "hybrid"
	DefaultLimit        = 10
	MaxLimit            = 50
)

type RecommendationService interface {
	// Core recommendation methods
	GetRecommendations(userID int, req models.RecommendationRequest) (*models.RecommendationResponse, error)

	// User preferences
	GetUserPreferences(userID int) (*models.UserPreferences, error)
	UpdateUserPreferences(userID int, req models.UpdatePreferencesRequest) (*models.UserPreferences, error)

	// Cooking history
	LogCooking(userID int, req models.LogCookingRequest) error
	GetCookingHistory(userID int, limit int) ([]models.CookingHistory, error)
}

type recommendationService struct {
	repo repository.RecommendationRepository
}

func NewRecommendationService(repo repository.RecommendationRepository) RecommendationService {
	return &recommendationService{repo: repo}
}

func (s *recommendationService) GetRecommendations(userID int, req models.RecommendationRequest) (*models.RecommendationResponse, error) {
	if userID <= 0 {
		return nil, ErrUserNotFound
	}

	// Validate and set defaults
	algorithm := s.validateAlgorithm(req.Algorithm)
	limit := s.validateLimit(req.Limit)

	var recipes []models.RecipeWithScore
	var err error

	// Execute appropriate algorithm
	switch algorithm {
	case AlgorithmPreference:
		recipes, err = s.repo.GetRecipesByPreferences(userID, limit)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, ErrPreferencesNotSet
			}
			return nil, err
		}
	case AlgorithmTimeDecay:
		recipes, err = s.repo.GetRecipesWithTimeDecayScore(userID, limit)
	case AlgorithmHybrid:
		recipes, err = s.repo.GetHybridRecommendations(userID, limit)
	default:
		return nil, ErrInvalidAlgorithm
	}

	if err != nil {
		return nil, err
	}

	// Log recommendations for analytics (async in production)
	go s.logRecommendations(userID, recipes, algorithm)

	return &models.RecommendationResponse{
		Recipes:     recipes,
		Algorithm:   algorithm,
		GeneratedAt: time.Now(),
		TotalScored: len(recipes),
	}, nil
}

func (s *recommendationService) GetUserPreferences(userID int) (*models.UserPreferences, error) {
	if userID <= 0 {
		return nil, ErrUserNotFound
	}

	prefs, err := s.repo.GetUserPreferences(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return empty preferences if none exist
			return &models.UserPreferences{
				UserID:              userID,
				PreferredCategories: []int{},
				CreatedAt:           time.Now(),
				UpdatedAt:           time.Now(),
			}, nil
		}
		return nil, err
	}

	return prefs, nil
}

func (s *recommendationService) UpdateUserPreferences(userID int, req models.UpdatePreferencesRequest) (*models.UserPreferences, error) {
	if userID <= 0 {
		return nil, ErrUserNotFound
	}

	// Validate categories (should be positive integers)
	validCategories := make([]int, 0, len(req.PreferredCategories))
	for _, categoryID := range req.PreferredCategories {
		if categoryID > 0 {
			validCategories = append(validCategories, categoryID)
		}
	}

	return s.repo.UpdateUserPreferences(userID, validCategories)
}

func (s *recommendationService) LogCooking(userID int, req models.LogCookingRequest) error {
	if userID <= 0 {
		return ErrUserNotFound
	}

	if req.RecipeID <= 0 {
		return ErrRecipeNotFound
	}

	if req.Rating != nil && (*req.Rating < 1 || *req.Rating > 5) {
		return ErrInvalidRating
	}

	err := s.repo.LogCooking(userID, req.RecipeID, req.Rating)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrRecipeNotFound
		}
		return err
	}

	return nil
}

func (s *recommendationService) GetCookingHistory(userID int, limit int) ([]models.CookingHistory, error) {
	if userID <= 0 {
		return nil, ErrUserNotFound
	}

	limit = s.validateLimit(limit)
	return s.repo.GetUserCookingHistory(userID, limit)
}

// Helper methods
func (s *recommendationService) validateAlgorithm(algorithm string) string {
	switch algorithm {
	case AlgorithmPreference, AlgorithmTimeDecay, AlgorithmHybrid:
		return algorithm
	default:
		// Default to hybrid as it's the most sophisticated
		return AlgorithmHybrid
	}
}

func (s *recommendationService) validateLimit(limit int) int {
	if limit <= 0 {
		return DefaultLimit
	}
	if limit > MaxLimit {
		return MaxLimit
	}
	return limit
}

func (s *recommendationService) logRecommendations(userID int, recipes []models.RecipeWithScore, algorithm string) {
	// In production, this would be in a goroutine or async queue
	for _, recipe := range recipes {
		s.repo.LogRecommendation(userID, recipe.ID, algorithm)
	}
}
