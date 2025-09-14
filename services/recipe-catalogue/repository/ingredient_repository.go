package repository

import (
	"database/sql"
	"github.com/lib/pq"
	"meal-prep/services/recipe-catalogue/domain"
	"meal-prep/shared/database"
	"meal-prep/shared/models"
)

type IngredientRepository interface {
	GetAllIngredients() ([]models.Ingredient, error)
	GetIngredientByID(id int) (*models.Ingredient, error)
	GetIngredientsByCategory(category string) ([]models.Ingredient, error)
	SearchIngredients(query string) ([]models.Ingredient, error)
	CreateIngredient(req models.CreateIngredientRequest) (*models.Ingredient, error)
	UpdateIngredient(id int, req models.UpdateIngredientRequest) (*models.Ingredient, error)
	DeleteIngredient(id int) error
	IngredientExists(id int) (bool, error)

	GetRecipeIngredients(recipeID int) ([]models.RecipeIngredient, error)
	AddRecipeIngredient(recipeID int, req models.AddRecipeIngredientRequest) (*models.RecipeIngredient, error)
	UpdateRecipeIngredient(recipeID, ingredientID int, req models.AddRecipeIngredientRequest) (*models.RecipeIngredient, error)
	RemoveRecipeIngredient(recipeID, ingredientID int) error
	SetRecipeIngredients(recipeID int, ingredients []models.AddRecipeIngredientRequest) error

	GetIngredientsForRecipes(recipeIDs []int) (map[int][]models.RecipeIngredient, error)
	GetRecipesUsingIngredient(ingredientID int) ([]models.Recipe, error)
}

type ingredientRepository struct {
	db *database.DB
}

func NewIngredientRepository(db *database.DB) IngredientRepository {
	return &ingredientRepository{db: db}
}

func (r *ingredientRepository) GetAllIngredients() ([]models.Ingredient, error) {
	query := `
		SELECT id, name, description, category, created_at 
		FROM recipe_catalogue.ingredients 
		ORDER BY category, name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ingredients []models.Ingredient
	for rows.Next() {
		ingredient, err := r.scanIngredient(rows)
		if err != nil {
			return nil, err
		}
		ingredients = append(ingredients, *ingredient)
	}

	return ingredients, nil
}

func (r *ingredientRepository) GetIngredientByID(id int) (*models.Ingredient, error) {
	query := `
		SELECT id, name, description, category, created_at 
		FROM recipe_catalogue.ingredients 
		WHERE id = $1`

	row := r.db.QueryRow(query, id)
	return r.scanIngredient(row)
}

func (r *ingredientRepository) GetIngredientsByCategory(category string) ([]models.Ingredient, error) {
	query := `
		SELECT id, name, description, category, created_at 
		FROM recipe_catalogue.ingredients 
		WHERE category = $1 
		ORDER BY name`

	rows, err := r.db.Query(query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ingredients []models.Ingredient
	for rows.Next() {
		ingredient, err := r.scanIngredient(rows)
		if err != nil {
			return nil, err
		}
		ingredients = append(ingredients, *ingredient)
	}

	return ingredients, nil
}

func (r *ingredientRepository) SearchIngredients(query string) ([]models.Ingredient, error) {
	searchQuery := `
		SELECT id, name, description, category, created_at 
		FROM recipe_catalogue.ingredients 
		WHERE name ILIKE $1 OR description ILIKE $1 
		ORDER BY name`

	rows, err := r.db.Query(searchQuery, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ingredients []models.Ingredient
	for rows.Next() {
		ingredient, err := r.scanIngredient(rows)
		if err != nil {
			return nil, err
		}
		ingredients = append(ingredients, *ingredient)
	}

	return ingredients, nil
}

func (r *ingredientRepository) CreateIngredient(req models.CreateIngredientRequest) (*models.Ingredient, error) {
	query := `
		INSERT INTO recipe_catalogue.ingredients (name, description, category, created_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		RETURNING id, name, description, category, created_at`

	row := r.db.QueryRow(query, req.Name, req.Description, req.Category)

	ingredient, err := r.scanIngredient(row)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			if pqErr.Constraint == "ingredients_name_key" {
				return nil, domain.ErrIngredientExists
			}
		}
		return nil, err
	}

	return ingredient, nil
}

func (r *ingredientRepository) UpdateIngredient(id int, req models.UpdateIngredientRequest) (*models.Ingredient, error) {
	query := `
        UPDATE recipe_catalogue.ingredients 
        SET name = $2,
            description = $3,
            category = $4,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $1
		RETURNING id, name, description, category, created_at`

	row := r.db.QueryRow(query, id, req.Name, req.Description, req.Category)
	return r.scanIngredient(row)
}

func (r *ingredientRepository) DeleteIngredient(id int) error {
	result, err := r.db.Exec("DELETE FROM recipe_catalogue.ingredients WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *ingredientRepository) IngredientExists(id int) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM recipe_catalogue.ingredients WHERE id = $1)",
		id).Scan(&exists)
	return exists, err
}

func (r *ingredientRepository) GetRecipeIngredients(recipeID int) ([]models.RecipeIngredient, error) {
	query := `
		SELECT ri.id, ri.recipe_id, ri.ingredient_id, ri.quantity, ri.unit, ri.notes, ri.created_at,
		       i.id, i.name, i.description, i.category, i.created_at
		FROM recipe_catalogue.recipe_ingredients ri
		JOIN recipe_catalogue.ingredients i ON ri.ingredient_id = i.id
		WHERE ri.recipe_id = $1
		ORDER BY ri.id`

	rows, err := r.db.Query(query, recipeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipeIngredients []models.RecipeIngredient
	for rows.Next() {
		recipeIngredient, err := r.scanRecipeIngredient(rows)
		if err != nil {
			return nil, err
		}
		recipeIngredients = append(recipeIngredients, *recipeIngredient)
	}

	return recipeIngredients, nil
}

func (r *ingredientRepository) AddRecipeIngredient(recipeID int, req models.AddRecipeIngredientRequest) (*models.RecipeIngredient, error) {
	query := `
		INSERT INTO recipe_catalogue.recipe_ingredients (recipe_id, ingredient_id, quantity, unit, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
		RETURNING id, recipe_id, ingredient_id, quantity, unit, notes, created_at`

	var ri models.RecipeIngredient
	err := r.db.QueryRow(query, recipeID, req.IngredientID, req.Quantity, req.Unit, req.Notes).Scan(
		&ri.ID, &ri.RecipeID, &ri.IngredientID, &ri.Quantity, &ri.Unit, &ri.Notes, &ri.CreatedAt)

	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			if pgErr.Constraint == "recipe_ingredients_unique_per_recipe" {
				return nil, domain.ErrRecipeIngredientAlreadyExists
			}
		}
		return nil, err
	}

	// Get the ingredient details
	ingredient, err := r.GetIngredientByID(req.IngredientID)
	if err != nil {
		return nil, err
	}

	ri.Ingredient = *ingredient
	return &ri, nil
}

func (r *ingredientRepository) UpdateRecipeIngredient(recipeID, ingredientID int, req models.AddRecipeIngredientRequest) (*models.RecipeIngredient, error) {
	query := `
		UPDATE recipe_catalogue.recipe_ingredients 
		SET quantity = $3, unit = $4, notes = $5
		WHERE recipe_id = $1 AND ingredient_id = $2
		RETURNING id, recipe_id, ingredient_id, quantity, unit, notes, created_at`

	var ri models.RecipeIngredient
	err := r.db.QueryRow(query, recipeID, ingredientID, req.Quantity, req.Unit, req.Notes).Scan(
		&ri.ID, &ri.RecipeID, &ri.IngredientID, &ri.Quantity, &ri.Unit, &ri.Notes, &ri.CreatedAt)

	if err != nil {
		return nil, err
	}

	// Get the ingredient details
	ingredient, err := r.GetIngredientByID(ingredientID)
	if err != nil {
		return nil, err
	}

	ri.Ingredient = *ingredient
	return &ri, nil
}

func (r *ingredientRepository) RemoveRecipeIngredient(recipeID, ingredientID int) error {
	result, err := r.db.Exec("DELETE FROM recipe_catalogue.recipe_ingredients WHERE recipe_id = $1 AND ingredient_id = $2", recipeID, ingredientID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *ingredientRepository) SetRecipeIngredients(recipeID int, ingredients []models.AddRecipeIngredientRequest) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Remove all existing ingredients for this recipe
	_, err = tx.Exec("DELETE FROM recipe_catalogue.recipe_ingredients WHERE recipe_id = $1", recipeID)
	if err != nil {
		return err
	}

	// Add all new ingredients
	for _, ingredient := range ingredients {
		_, err = tx.Exec(`
			INSERT INTO recipe_catalogue.recipe_ingredients (recipe_id, ingredient_id, quantity, unit, notes, created_at)
			VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)`,
			recipeID, ingredient.IngredientID, ingredient.Quantity, ingredient.Unit, ingredient.Notes)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *ingredientRepository) GetIngredientsForRecipes(recipeIDs []int) (map[int][]models.RecipeIngredient, error) {
	if len(recipeIDs) == 0 {
		return make(map[int][]models.RecipeIngredient), nil
	}

	query := `
        SELECT ri.id, ri.recipe_id, ri.ingredient_id, ri.quantity, ri.unit, ri.notes, ri.created_at,
               i.id, i.name, i.description, i.category, i.created_at
        FROM recipe_catalogue.recipe_ingredients ri
        JOIN recipe_catalogue.ingredients i ON ri.ingredient_id = i.id
        WHERE ri.recipe_id = ANY($1)
        ORDER BY ri.recipe_id, ri.id`

	// Convert []int to pq.Array for PostgreSQL
	rows, err := r.db.Query(query, pq.Array(recipeIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// ... rest of the method stays the same
	result := make(map[int][]models.RecipeIngredient)
	for rows.Next() {
		recipeIngredient, err := r.scanRecipeIngredient(rows)
		if err != nil {
			return nil, err
		}

		result[recipeIngredient.RecipeID] = append(result[recipeIngredient.RecipeID], *recipeIngredient)
	}

	return result, nil
}

func (r *ingredientRepository) GetRecipesUsingIngredient(ingredientID int) ([]models.Recipe, error) {
	query := `
		SELECT DISTINCT r.id, r.name, r.description, r.category_id, r.created_at, r.updated_at,
		                c.id, c.name, c.description
		FROM recipe_catalogue.recipes r
		LEFT JOIN recipe_catalogue.categories c ON r.category_id = c.id
		JOIN recipe_catalogue.recipe_ingredients ri ON r.id = ri.recipe_id
		WHERE ri.ingredient_id = $1
		ORDER BY r.name`

	rows, err := r.db.Query(query, ingredientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipes []models.Recipe
	for rows.Next() {
		recipe := models.Recipe{}
		category := models.Category{}
		var categoryDesc sql.NullString

		err := rows.Scan(
			&recipe.ID, &recipe.Name, &recipe.Description, &recipe.CategoryID,
			&recipe.CreatedAt, &recipe.UpdatedAt,
			&category.ID, &category.Name, &categoryDesc,
		)
		if err != nil {
			return nil, err
		}

		if categoryDesc.Valid {
			category.Description = &categoryDesc.String
		}
		recipe.Category = &category

		recipes = append(recipes, recipe)
	}

	return recipes, nil
}

// Helper methods
func (r *ingredientRepository) scanIngredient(scanner interface {
	Scan(...interface{}) error
}) (*models.Ingredient, error) {
	var ingredient models.Ingredient
	var description, category sql.NullString

	err := scanner.Scan(&ingredient.ID, &ingredient.Name, &description, &category, &ingredient.CreatedAt)
	if err != nil {
		return nil, err
	}

	if description.Valid {
		ingredient.Description = &description.String
	}
	if category.Valid {
		ingredient.Category = &category.String
	}

	return &ingredient, nil
}

func (r *ingredientRepository) scanRecipeIngredient(scanner interface {
	Scan(...interface{}) error
}) (*models.RecipeIngredient, error) {
	var ri models.RecipeIngredient
	var ingredient models.Ingredient
	var riNotes, ingredientDesc, ingredientCategory sql.NullString

	err := scanner.Scan(
		&ri.ID, &ri.RecipeID, &ri.IngredientID, &ri.Quantity, &ri.Unit, &riNotes, &ri.CreatedAt,
		&ingredient.ID, &ingredient.Name, &ingredientDesc, &ingredientCategory, &ingredient.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if riNotes.Valid {
		ri.Notes = &riNotes.String
	}
	if ingredientDesc.Valid {
		ingredient.Description = &ingredientDesc.String
	}
	if ingredientCategory.Valid {
		ingredient.Category = &ingredientCategory.String
	}

	ri.Ingredient = ingredient
	return &ri, nil
}
