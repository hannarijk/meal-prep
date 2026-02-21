package repository

import (
	"database/sql"
	"meal-prep/shared/database"
	"meal-prep/shared/models"

	"github.com/lib/pq"
)

type RecipeRepository interface {
	GetAll() ([]models.Recipe, error)
	GetByID(id int) (*models.Recipe, error)
	GetByCategory(categoryID int) ([]models.Recipe, error)
	GetOwnerID(id int) (int, error)
	Create(userID int, req models.CreateRecipeRequest) (*models.Recipe, error)
	Update(id int, req models.UpdateRecipeRequest) (*models.Recipe, error)
	Delete(id int) error

	GetAllWithIngredients() ([]models.RecipeWithIngredients, error)
	GetByIDWithIngredients(id int) (*models.RecipeWithIngredients, error)
	GetByCategoryWithIngredients(categoryID int) ([]models.RecipeWithIngredients, error)
	CreateWithIngredients(userID int, req models.CreateRecipeWithIngredientsRequest) (*models.RecipeWithIngredients, error)
	UpdateWithIngredients(id int, req models.UpdateRecipeWithIngredientsRequest) (*models.RecipeWithIngredients, error)

	SearchRecipesByIngredients(ingredientIDs []int) ([]models.Recipe, error)
	SearchRecipesByIngredientsWithIngredients(ingredientIDs []int) ([]models.RecipeWithIngredients, error)
}

type recipeRepository struct {
	db             *database.DB
	ingredientRepo IngredientRepository
}

func NewRecipeRepository(db *database.DB) RecipeRepository {
	return &recipeRepository{
		db:             db,
		ingredientRepo: NewIngredientRepository(db),
	}
}

func (r *recipeRepository) GetAll() ([]models.Recipe, error) {
	query := `
		SELECT d.id, d.user_id, d.name, d.description, d.category_id, d.created_at, d.updated_at,
		       c.id, c.name, c.description
		FROM recipe_catalogue.recipes d
		LEFT JOIN recipe_catalogue.categories c ON d.category_id = c.id
		ORDER BY d.name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recipes := make([]models.Recipe, 0)
	for rows.Next() {
		recipe, err := r.scanRecipeWithCategory(rows)
		if err != nil {
			return nil, err
		}
		recipes = append(recipes, *recipe)
	}

	return recipes, nil
}

func (r *recipeRepository) GetByID(id int) (*models.Recipe, error) {
	query := `
		SELECT d.id, d.user_id, d.name, d.description, d.category_id, d.created_at, d.updated_at,
		       c.id, c.name, c.description
		FROM recipe_catalogue.recipes d
		LEFT JOIN recipe_catalogue.categories c ON d.category_id = c.id
		WHERE d.id = $1`

	return r.scanRecipeWithCategory(r.db.QueryRow(query, id))
}

func (r *recipeRepository) GetByCategory(categoryID int) ([]models.Recipe, error) {
	query := `
		SELECT d.id, d.user_id, d.name, d.description, d.category_id, d.created_at, d.updated_at,
		       c.id, c.name, c.description
		FROM recipe_catalogue.recipes d
		LEFT JOIN recipe_catalogue.categories c ON d.category_id = c.id
		WHERE d.category_id = $1
		ORDER BY d.name`

	rows, err := r.db.Query(query, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recipes := make([]models.Recipe, 0)
	for rows.Next() {
		recipe, err := r.scanRecipeWithCategory(rows)
		if err != nil {
			return nil, err
		}
		recipes = append(recipes, *recipe)
	}

	return recipes, nil
}

// GetOwnerID returns the user_id of the recipe owner, or sql.ErrNoRows if not found.
// Used by the service layer to enforce ownership before allowing mutations.
func (r *recipeRepository) GetOwnerID(id int) (int, error) {
	var userID int
	err := r.db.QueryRow(
		"SELECT user_id FROM recipe_catalogue.recipes WHERE id = $1", id,
	).Scan(&userID)
	return userID, err
}

func (r *recipeRepository) Create(userID int, req models.CreateRecipeRequest) (*models.Recipe, error) {
	var recipe models.Recipe
	err := r.db.QueryRow(`
		INSERT INTO recipe_catalogue.recipes (name, description, category_id, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, user_id, name, description, category_id, created_at, updated_at`,
		req.Name, req.Description, req.CategoryID, userID).Scan(
		&recipe.ID, &recipe.UserID, &recipe.Name, &recipe.Description, &recipe.CategoryID,
		&recipe.CreatedAt, &recipe.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &recipe, nil
}

func (r *recipeRepository) Update(id int, req models.UpdateRecipeRequest) (*models.Recipe, error) {
	var recipe models.Recipe
	err := r.db.QueryRow(`
        UPDATE recipe_catalogue.recipes
        SET name = $2,
            description = $3,
            category_id = $4,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $1
		RETURNING id, user_id, name, description, category_id, created_at, updated_at`,
		id, req.Name, req.Description, req.CategoryID).Scan(
		&recipe.ID, &recipe.UserID, &recipe.Name, &recipe.Description, &recipe.CategoryID,
		&recipe.CreatedAt, &recipe.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &recipe, nil
}

func (r *recipeRepository) Delete(id int) error {
	result, err := r.db.Exec("DELETE FROM recipe_catalogue.recipes WHERE id = $1", id)
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

func (r *recipeRepository) GetAllWithIngredients() ([]models.RecipeWithIngredients, error) {
	recipes, err := r.GetAll()
	if err != nil {
		return nil, err
	}

	return r.attachIngredientsToRecipes(recipes)
}

func (r *recipeRepository) GetByIDWithIngredients(id int) (*models.RecipeWithIngredients, error) {
	recipe, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	ingredients, err := r.ingredientRepo.GetRecipeIngredients(id)
	if err != nil {
		return nil, err
	}

	return &models.RecipeWithIngredients{
		Recipe:      *recipe,
		Ingredients: ingredients,
	}, nil
}

func (r *recipeRepository) GetByCategoryWithIngredients(categoryID int) ([]models.RecipeWithIngredients, error) {
	recipes, err := r.GetByCategory(categoryID)
	if err != nil {
		return nil, err
	}

	return r.attachIngredientsToRecipes(recipes)
}

func (r *recipeRepository) CreateWithIngredients(userID int, req models.CreateRecipeWithIngredientsRequest) (*models.RecipeWithIngredients, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var recipe models.Recipe
	err = tx.QueryRow(`
		INSERT INTO recipe_catalogue.recipes (name, description, category_id, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, user_id, name, description, category_id, created_at, updated_at`,
		req.Name, req.Description, req.CategoryID, userID).Scan(
		&recipe.ID, &recipe.UserID, &recipe.Name, &recipe.Description, &recipe.CategoryID,
		&recipe.CreatedAt, &recipe.UpdatedAt)
	if err != nil {
		return nil, err
	}

	var ingredients []models.RecipeIngredient
	if len(req.Ingredients) > 0 {
		for _, ingredient := range req.Ingredients {
			_, err = tx.Exec(`
				INSERT INTO recipe_catalogue.recipe_ingredients (recipe_id, ingredient_id, quantity, unit, notes, created_at)
				VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)`,
				recipe.ID, ingredient.IngredientID, ingredient.Quantity, ingredient.Unit, ingredient.Notes)
			if err != nil {
				return nil, err
			}
		}

		if err = tx.Commit(); err != nil {
			return nil, err
		}

		ingredients, err = r.ingredientRepo.GetRecipeIngredients(recipe.ID)
		if err != nil {
			return nil, err
		}
	} else {
		if err = tx.Commit(); err != nil {
			return nil, err
		}
	}

	if recipe.CategoryID != nil {
		fullRecipe, err := r.GetByID(recipe.ID)
		if err != nil {
			return nil, err
		}
		recipe.Category = fullRecipe.Category
	}

	return &models.RecipeWithIngredients{
		Recipe:      recipe,
		Ingredients: ingredients,
	}, nil
}

func (r *recipeRepository) UpdateWithIngredients(id int, req models.UpdateRecipeWithIngredientsRequest) (*models.RecipeWithIngredients, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var recipe models.Recipe
	err = tx.QueryRow(`
		UPDATE recipe_catalogue.recipes
		SET name = COALESCE(NULLIF($2, ''), name),
		    description = COALESCE($3, description),
		    category_id = COALESCE(NULLIF($4, 0), category_id),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, user_id, name, description, category_id, created_at, updated_at`,
		id, req.Name, req.Description, req.CategoryID).Scan(
		&recipe.ID, &recipe.UserID, &recipe.Name, &recipe.Description, &recipe.CategoryID,
		&recipe.CreatedAt, &recipe.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if req.Ingredients != nil {
		_, err = tx.Exec("DELETE FROM recipe_catalogue.recipe_ingredients WHERE recipe_id = $1", id)
		if err != nil {
			return nil, err
		}

		for _, ingredient := range req.Ingredients {
			_, err = tx.Exec(`
				INSERT INTO recipe_catalogue.recipe_ingredients (recipe_id, ingredient_id, quantity, unit, notes, created_at)
				VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)`,
				id, ingredient.IngredientID, ingredient.Quantity, ingredient.Unit, ingredient.Notes)
			if err != nil {
				return nil, err
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	fullRecipe, err := r.GetByIDWithIngredients(id)
	if err != nil {
		return nil, err
	}

	return fullRecipe, nil
}

func (r *recipeRepository) SearchRecipesByIngredients(ingredientIDs []int) ([]models.Recipe, error) {
	if len(ingredientIDs) == 0 {
		return []models.Recipe{}, nil
	}

	query := `
        SELECT DISTINCT r.id, r.user_id, r.name, r.description, r.category_id, r.created_at, r.updated_at,
                        c.id, c.name, c.description
        FROM recipe_catalogue.recipes r
        LEFT JOIN recipe_catalogue.categories c ON r.category_id = c.id
        JOIN recipe_catalogue.recipe_ingredients ri ON r.id = ri.recipe_id
        WHERE ri.ingredient_id = ANY($1)
        ORDER BY r.name`

	rows, err := r.db.Query(query, pq.Array(ingredientIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recipes := make([]models.Recipe, 0)
	for rows.Next() {
		recipe, err := r.scanRecipeWithCategory(rows)
		if err != nil {
			return nil, err
		}
		recipes = append(recipes, *recipe)
	}

	return recipes, nil
}

func (r *recipeRepository) SearchRecipesByIngredientsWithIngredients(ingredientIDs []int) ([]models.RecipeWithIngredients, error) {
	recipes, err := r.SearchRecipesByIngredients(ingredientIDs)
	if err != nil {
		return nil, err
	}

	return r.attachIngredientsToRecipes(recipes)
}

// scanRecipeWithCategory reads a recipe + its left-joined category from any scanner
// (sql.Row or sql.Rows). Centralising scan logic here prevents drift between queries.
func (r *recipeRepository) scanRecipeWithCategory(scanner interface {
	Scan(...interface{}) error
}) (*models.Recipe, error) {
	recipe := models.Recipe{}
	category := models.Category{}
	var categoryDesc sql.NullString

	err := scanner.Scan(
		&recipe.ID, &recipe.UserID, &recipe.Name, &recipe.Description, &recipe.CategoryID,
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

	return &recipe, nil
}

func (r *recipeRepository) attachIngredientsToRecipes(recipes []models.Recipe) ([]models.RecipeWithIngredients, error) {
	if len(recipes) == 0 {
		return []models.RecipeWithIngredients{}, nil
	}

	recipeIDs := make([]int, len(recipes))
	for i, recipe := range recipes {
		recipeIDs[i] = recipe.ID
	}

	ingredientsMap, err := r.ingredientRepo.GetIngredientsForRecipes(recipeIDs)
	if err != nil {
		return nil, err
	}

	var result []models.RecipeWithIngredients
	for _, recipe := range recipes {
		ingredients := ingredientsMap[recipe.ID]
		if ingredients == nil {
			ingredients = []models.RecipeIngredient{}
		}

		result = append(result, models.RecipeWithIngredients{
			Recipe:      recipe,
			Ingredients: ingredients,
		})
	}

	return result, nil
}