package repository

import (
	"database/sql"
	"github.com/lib/pq"
	"log"
	"meal-prep/shared/database"
	"meal-prep/shared/models"
	"time"
)

type RecommendationRepository interface {
	GetUserPreferences(userID int) (*models.UserPreferences, error)
	UpdateUserPreferences(userID int, categories []int) (*models.UserPreferences, error)

	// Cooking history methods
	LogCooking(userID, recipeID int, rating *int) error
	GetUserCookingHistory(userID int, limit int) ([]models.CookingHistory, error)
	GetLastCookedTimes(userID int) (map[int]time.Time, error)

	// Recommendation queries
	GetRecipesWithTimeDecayScore(userID int, limit int) ([]models.RecipeWithScore, error)
	GetRecipesByPreferences(userID int, limit int) ([]models.RecipeWithScore, error)
	GetHybridRecommendations(userID int, limit int) ([]models.RecipeWithScore, error)

	// Analytics
	LogRecommendation(userID, recipeID int, algorithm string) error
}

type recommendationRepository struct {
	db *database.DB
}

func NewRecommendationRepository(db *database.DB) RecommendationRepository {
	log.Println("INFO: Creating new recommendation repository")
	return &recommendationRepository{db: db}
}

// Helper functions for PostgreSQL array conversion
func intSliceToInt64Array(slice []int) pq.Int64Array {
	array := make(pq.Int64Array, len(slice))
	for i, v := range slice {
		array[i] = int64(v)
	}
	return array
}

func int64ArrayToIntSlice(array pq.Int64Array) []int {
	slice := make([]int, len(array))
	for i, v := range array {
		slice[i] = int(v)
	}
	return slice
}

func (r *recommendationRepository) GetUserPreferences(userID int) (*models.UserPreferences, error) {
	log.Printf("INFO: Getting preferences for user %d", userID)

	var prefs models.UserPreferences
	var categoriesArray pq.Int64Array

	err := r.db.QueryRow(`
		SELECT id, user_id, preferred_categories, created_at, updated_at
		FROM recommendations.user_preferences WHERE user_id = $1`,
		userID).Scan(&prefs.ID, &prefs.UserID, &categoriesArray,
		&prefs.CreatedAt, &prefs.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("INFO: No preferences found for user %d", userID)
		} else {
			log.Printf("ERROR: Failed to get preferences for user %d: %v", userID, err)
		}
		return nil, err
	}

	// Convert pq.Int64Array to []int
	prefs.PreferredCategories = int64ArrayToIntSlice(categoriesArray)

	log.Printf("INFO: Retrieved preferences for user %d: categories=%v", userID, prefs.PreferredCategories)
	return &prefs, nil
}

func (r *recommendationRepository) UpdateUserPreferences(userID int, categories []int) (*models.UserPreferences, error) {
	log.Printf("INFO: Updating preferences for user %d with categories: %v", userID, categories)

	// Convert []int to pq.Int64Array for PostgreSQL
	categoriesArray := intSliceToInt64Array(categories)

	// First, try to insert or update
	result, err := r.db.Exec(`
		INSERT INTO recommendations.user_preferences (user_id, preferred_categories, created_at, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			preferred_categories = $2,
			updated_at = CURRENT_TIMESTAMP`,
		userID, categoriesArray)

	if err != nil {
		log.Printf("ERROR: Failed to update preferences for user %d: %v", userID, err)
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("INFO: Preferences update successful for user %d, rows affected: %d", userID, rowsAffected)

	// Then fetch the updated record
	var prefs models.UserPreferences
	var fetchedArray pq.Int64Array

	err = r.db.QueryRow(`
		SELECT id, user_id, preferred_categories, created_at, updated_at
		FROM recommendations.user_preferences WHERE user_id = $1`,
		userID).Scan(&prefs.ID, &prefs.UserID, &fetchedArray,
		&prefs.CreatedAt, &prefs.UpdatedAt)

	if err != nil {
		log.Printf("ERROR: Failed to fetch updated preferences for user %d: %v", userID, err)
		return nil, err
	}

	// Convert back to []int
	prefs.PreferredCategories = int64ArrayToIntSlice(fetchedArray)

	log.Printf("INFO: Successfully updated and fetched preferences for user %d: %+v", userID, prefs.PreferredCategories)
	return &prefs, nil
}

func (r *recommendationRepository) LogCooking(userID, recipeID int, rating *int) error {
	log.Printf("INFO: Logging cooking for user %d, recipe %d, rating %v", userID, recipeID, rating)

	// First, check if the recipe exists
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM recipe_catalogue.recipes WHERE id = $1)",
		recipeID).Scan(&exists)

	if err != nil {
		log.Printf("ERROR: Failed to check recipe existence for recipe %d: %v", recipeID, err)
		return err
	}

	if !exists {
		log.Printf("ERROR: Recipe %d does not exist", recipeID)
		return sql.ErrNoRows // This will be caught by the service layer
	}

	_, err = r.db.Exec(`
		INSERT INTO recommendations.cooking_history (user_id, recipe_id, cooked_at, rating)
		VALUES ($1, $2, CURRENT_TIMESTAMP, $3)`,
		userID, recipeID, rating)

	if err != nil {
		log.Printf("ERROR: Failed to log cooking for user %d, recipe %d: %v", userID, recipeID, err)
		return err
	}

	log.Printf("INFO: Successfully logged cooking for user %d, recipe %d", userID, recipeID)
	return nil
}

func (r *recommendationRepository) GetUserCookingHistory(userID int, limit int) ([]models.CookingHistory, error) {
	log.Printf("INFO: Getting cooking history for user %d, limit %d", userID, limit)

	query := `
		SELECT id, user_id, recipe_id, cooked_at, rating
		FROM recommendations.cooking_history 
		WHERE user_id = $1
		ORDER BY cooked_at DESC
		LIMIT $2`

	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		log.Printf("ERROR: Failed to query cooking history for user %d: %v", userID, err)
		return nil, err
	}
	defer rows.Close()

	var history []models.CookingHistory
	for rows.Next() {
		var h models.CookingHistory
		err := rows.Scan(&h.ID, &h.UserID, &h.RecipeID, &h.CookedAt, &h.Rating)
		if err != nil {
			log.Printf("ERROR: Failed to scan cooking history row for user %d: %v", userID, err)
			return nil, err
		}
		history = append(history, h)
	}

	log.Printf("INFO: Retrieved %d cooking history records for user %d", len(history), userID)
	return history, nil
}

func (r *recommendationRepository) GetLastCookedTimes(userID int) (map[int]time.Time, error) {
	log.Printf("INFO: Getting last cooked times for user %d", userID)

	query := `
		SELECT DISTINCT ON (recipe_id) recipe_id, cooked_at
		FROM recommendations.cooking_history 
		WHERE user_id = $1
		ORDER BY recipe_id, cooked_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		log.Printf("ERROR: Failed to query last cooked times for user %d: %v", userID, err)
		return nil, err
	}
	defer rows.Close()

	lastCooked := make(map[int]time.Time)
	for rows.Next() {
		var recipeID int
		var cookedAt time.Time
		if err := rows.Scan(&recipeID, &cookedAt); err != nil {
			log.Printf("ERROR: Failed to scan last cooked time for user %d: %v", userID, err)
			return nil, err
		}
		lastCooked[recipeID] = cookedAt
	}

	log.Printf("INFO: Retrieved last cooked times for %d recipes for user %d", len(lastCooked), userID)
	return lastCooked, nil
}

func (r *recommendationRepository) GetRecipesWithTimeDecayScore(userID int, limit int) ([]models.RecipeWithScore, error) {
	log.Printf("INFO: Getting time decay recommendations for user %d, limit %d", userID, limit)

	query := `
		WITH recipe_last_cooked AS (
			SELECT DISTINCT ON (ch.recipe_id) 
				ch.recipe_id,
				ch.cooked_at,
				EXTRACT(days FROM CURRENT_TIMESTAMP - ch.cooked_at) as days_since
			FROM recommendations.cooking_history ch
			WHERE ch.user_id = $1
			ORDER BY ch.recipe_id, ch.cooked_at DESC
		),
		scored_recipes AS (
			SELECT 
				d.id, d.name, d.description, d.category_id, d.created_at, d.updated_at,
				c.id as cat_id, c.name as cat_name, c.description as cat_desc,
				dlc.cooked_at,
				dlc.days_since,
				CASE 
					WHEN dlc.days_since IS NULL THEN 0.5
					WHEN dlc.days_since < 7 THEN 0.1
					WHEN dlc.days_since < 30 THEN 0.7
					WHEN dlc.days_since < 90 THEN 1.0
					ELSE 1.2
				END as time_score
			FROM recipe_catalogue.recipes d
			LEFT JOIN recipe_catalogue.categories c ON d.category_id = c.id
			LEFT JOIN recipe_last_cooked dlc ON d.id = dlc.recipe_id
		)
		SELECT 
			id, name, description, category_id, created_at, updated_at,
			cat_id, cat_name, cat_desc, cooked_at, days_since, time_score
		FROM scored_recipes
		ORDER BY time_score DESC, RANDOM()
		LIMIT $2`

	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		log.Printf("ERROR: Failed to query time decay recommendations for user %d: %v", userID, err)
		return nil, err
	}
	defer rows.Close()

	recipes, err := r.scanScoredRecipes(rows, "time_decay")
	if err != nil {
		log.Printf("ERROR: Failed to scan time decay recipes for user %d: %v", userID, err)
		return nil, err
	}

	log.Printf("INFO: Generated %d time decay recommendations for user %d", len(recipes), userID)
	return recipes, nil
}

func (r *recommendationRepository) GetRecipesByPreferences(userID int, limit int) ([]models.RecipeWithScore, error) {
	log.Printf("INFO: Getting preference-based recommendations for user %d, limit %d", userID, limit)

	query := `
		SELECT 
			d.id, d.name, d.description, d.category_id, d.created_at, d.updated_at,
			c.id, c.name, c.description,
			NULL::timestamp as cooked_at,
			NULL::numeric as days_since,
			1.0 as preference_score
		FROM recipe_catalogue.recipes d
		LEFT JOIN recipe_catalogue.categories c ON d.category_id = c.id
		WHERE d.category_id IN (
			SELECT unnest(preferred_categories) 
			FROM recommendations.user_preferences 
			WHERE user_id = $1
		)
		ORDER BY RANDOM()
		LIMIT $2`

	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		log.Printf("ERROR: Failed to query preference recommendations for user %d: %v", userID, err)
		return nil, err
	}
	defer rows.Close()

	recipes, err := r.scanScoredRecipes(rows, "preference")
	if err != nil {
		log.Printf("ERROR: Failed to scan preference recipes for user %d: %v", userID, err)
		return nil, err
	}

	// If no recipes found (user has no preferences set), get random recipes instead
	if len(recipes) == 0 {
		log.Printf("INFO: No preference-based recipes found for user %d, falling back to random recipes", userID)
		return r.getRandomRecipes(limit)
	}

	log.Printf("INFO: Generated %d preference-based recommendations for user %d", len(recipes), userID)
	return recipes, nil
}

func (r *recommendationRepository) GetHybridRecommendations(userID int, limit int) ([]models.RecipeWithScore, error) {
	log.Printf("INFO: Getting hybrid recommendations for user %d, limit %d", userID, limit)

	query := `
		WITH user_pref_categories AS (
			SELECT unnest(preferred_categories) as category_id
			FROM recommendations.user_preferences 
			WHERE user_id = $1
		),
		recipe_last_cooked AS (
			SELECT DISTINCT ON (ch.recipe_id) 
				ch.recipe_id,
				ch.cooked_at,
				EXTRACT(days FROM CURRENT_TIMESTAMP - ch.cooked_at) as days_since
			FROM recommendations.cooking_history ch
			WHERE ch.user_id = $1
			ORDER BY ch.recipe_id, ch.cooked_at DESC
		),
		scored_recipes AS (
			SELECT 
				d.id, d.name, d.description, d.category_id, d.created_at, d.updated_at,
				c.id as cat_id, c.name as cat_name, c.description as cat_desc,
				dlc.cooked_at,
				dlc.days_since,
				CASE 
					WHEN dlc.days_since IS NULL THEN 0.5
					WHEN dlc.days_since < 7 THEN 0.1
					WHEN dlc.days_since < 30 THEN 0.7
					WHEN dlc.days_since < 90 THEN 1.0
					ELSE 1.2
				END as time_score,
				CASE 
					WHEN EXISTS(SELECT 1 FROM user_pref_categories) 
						AND d.category_id IN (SELECT category_id FROM user_pref_categories) THEN 1.0
					WHEN NOT EXISTS(SELECT 1 FROM user_pref_categories) THEN 0.7
					ELSE 0.3
				END as preference_score
			FROM recipe_catalogue.recipes d
			LEFT JOIN recipe_catalogue.categories c ON d.category_id = c.id
			LEFT JOIN recipe_last_cooked dlc ON d.id = dlc.recipe_id
		)
		SELECT 
			id, name, description, category_id, created_at, updated_at,
			cat_id, cat_name, cat_desc, cooked_at, days_since,
			(time_score * 0.6 + preference_score * 0.4) as final_score
		FROM scored_recipes
		ORDER BY final_score DESC, RANDOM()
		LIMIT $2`

	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		log.Printf("ERROR: Failed to query hybrid recommendations for user %d: %v", userID, err)
		return nil, err
	}
	defer rows.Close()

	recipes, err := r.scanScoredRecipes(rows, "hybrid")
	if err != nil {
		log.Printf("ERROR: Failed to scan hybrid recipes for user %d: %v", userID, err)
		return nil, err
	}

	log.Printf("INFO: Generated %d hybrid recommendations for user %d", len(recipes), userID)
	return recipes, nil
}

func (r *recommendationRepository) LogRecommendation(userID, recipeID int, algorithm string) error {
	_, err := r.db.Exec(`
		INSERT INTO recommendations.recommendation_history (user_id, recipe_id, recommended_at, algorithm_used)
		VALUES ($1, $2, CURRENT_TIMESTAMP, $3)`,
		userID, recipeID, algorithm)

	if err != nil {
		log.Printf("ERROR: Failed to log recommendation for user %d, recipe %d, algorithm %s: %v", userID, recipeID, algorithm, err)
		return err
	}

	// Only log successful analytics logs at debug level to avoid spam
	return nil
}

// Helper method for random recipes (fallback when no preferences/history)
func (r *recommendationRepository) getRandomRecipes(limit int) ([]models.RecipeWithScore, error) {
	log.Printf("INFO: Getting random recipes, limit %d", limit)

	query := `
		SELECT 
			d.id, d.name, d.description, d.category_id, d.created_at, d.updated_at,
			c.id, c.name, c.description,
			NULL::timestamp as cooked_at,
			NULL::numeric as days_since,
			0.5 as random_score
		FROM recipe_catalogue.recipes d
		LEFT JOIN recipe_catalogue.categories c ON d.category_id = c.id
		ORDER BY RANDOM()
		LIMIT $1`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		log.Printf("ERROR: Failed to query random recipes: %v", err)
		return nil, err
	}
	defer rows.Close()

	recipes, err := r.scanScoredRecipes(rows, "random")
	if err != nil {
		log.Printf("ERROR: Failed to scan random recipes: %v", err)
		return nil, err
	}

	log.Printf("INFO: Generated %d random recommendations", len(recipes))
	return recipes, nil
}

func (r *recommendationRepository) scanScoredRecipes(rows *sql.Rows, algorithm string) ([]models.RecipeWithScore, error) {
	var recipes []models.RecipeWithScore

	for rows.Next() {
		var dws models.RecipeWithScore
		var category models.Category
		var categoryID sql.NullInt64
		var categoryName sql.NullString
		var categoryDesc sql.NullString
		var cookedAt sql.NullTime
		var daysSince sql.NullInt64
		var score float64

		err := rows.Scan(
			&dws.ID, &dws.Name, &dws.Description, &dws.CategoryID,
			&dws.CreatedAt, &dws.UpdatedAt,
			&categoryID, &categoryName, &categoryDesc,
			&cookedAt, &daysSince, &score,
		)
		if err != nil {
			log.Printf("ERROR: Failed to scan scored recipe row for %s: %v", algorithm, err)
			return nil, err
		}

		// Handle category data safely
		if categoryID.Valid && categoryName.Valid {
			category.ID = int(categoryID.Int64)
			category.Name = categoryName.String
			if categoryDesc.Valid {
				category.Description = &categoryDesc.String
			}
			dws.Category = &category
		} else {
			// Create a default category for recipes without categories
			category.ID = 0
			category.Name = "Uncategorized"
			dws.Category = &category
		}

		dws.RecommendationScore = score

		if cookedAt.Valid {
			dws.LastCookedAt = &cookedAt.Time
		}
		if daysSince.Valid {
			days := int(daysSince.Int64)
			dws.DaysSinceCooked = &days
		}

		// Generate reason based on algorithm and data
		dws.Reason = r.generateReason(algorithm, dws.DaysSinceCooked, dws.Category.Name)

		recipes = append(recipes, dws)
	}

	log.Printf("INFO: Successfully scanned %d recipes for %s algorithm", len(recipes), algorithm)
	return recipes, nil
}

func (r *recommendationRepository) generateReason(algorithm string, daysSince *int, categoryName string) string {
	switch algorithm {
	case "time_decay":
		if daysSince == nil {
			return "New recipe to try"
		} else if *daysSince < 7 {
			return "Recently enjoyed"
		} else if *daysSince < 30 {
			return "Time to revisit"
		} else if *daysSince < 90 {
			return "You might be missing this"
		} else {
			return "Long time favorite"
		}
	case "preference":
		return "Based on your preferences for " + categoryName
	case "hybrid":
		if daysSince != nil && *daysSince > 30 {
			return "Perfect time to revisit this " + categoryName + " favorite"
		}
		return "Great match for your " + categoryName + " preference"
	case "random":
		return "Discover something new in " + categoryName
	default:
		return "Recommended for you"
	}
}
