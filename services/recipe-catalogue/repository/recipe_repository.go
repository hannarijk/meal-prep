package repository

import (
	"database/sql"
	"meal-prep/shared/database"
	"meal-prep/shared/models"
)

type RecipeRepository interface {
	GetAll() ([]models.Recipe, error)
	GetByID(id int) (*models.Recipe, error)
	GetByCategory(categoryID int) ([]models.Recipe, error)
	Create(req models.CreateRecipeRequest) (*models.Recipe, error)
	Update(id int, req models.UpdateRecipeRequest) (*models.Recipe, error)
	Delete(id int) error
}

type recipeRepository struct {
	db *database.DB
}

func NewRecipeRepository(db *database.DB) RecipeRepository {
	return &recipeRepository{db: db}
}

func (r *recipeRepository) GetAll() ([]models.Recipe, error) {
	query := `
		SELECT d.id, d.name, d.description, d.category_id, d.created_at, d.updated_at,
		       c.id, c.name, c.description
		FROM recipe_catalogue.recipes d
		LEFT JOIN recipe_catalogue.categories c ON d.category_id = c.id
		ORDER BY d.name`

	rows, err := r.db.Query(query)
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

func (r *recipeRepository) GetByID(id int) (*models.Recipe, error) {
	query := `
		SELECT d.id, d.name, d.description, d.category_id, d.created_at, d.updated_at,
		       c.id, c.name, c.description
		FROM recipe_catalogue.recipes d
		LEFT JOIN recipe_catalogue.categories c ON d.category_id = c.id
		WHERE d.id = $1`

	recipe := models.Recipe{}
	category := models.Category{}
	var categoryDesc sql.NullString

	err := r.db.QueryRow(query, id).Scan(
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

	return &recipe, nil
}

func (r *recipeRepository) GetByCategory(categoryID int) ([]models.Recipe, error) {
	query := `
		SELECT d.id, d.name, d.description, d.category_id, d.created_at, d.updated_at,
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

func (r *recipeRepository) Create(req models.CreateRecipeRequest) (*models.Recipe, error) {
	var recipe models.Recipe
	err := r.db.QueryRow(`
		INSERT INTO recipe_catalogue.recipes (name, description, category_id, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, name, description, category_id, created_at, updated_at`,
		req.Name, req.Description, req.CategoryID).Scan(
		&recipe.ID, &recipe.Name, &recipe.Description, &recipe.CategoryID,
		&recipe.CreatedAt, &recipe.UpdatedAt)

	return &recipe, err
}

func (r *recipeRepository) Update(id int, req models.UpdateRecipeRequest) (*models.Recipe, error) {
	var recipe models.Recipe
	err := r.db.QueryRow(`
		UPDATE recipe_catalogue.recipes 
		SET name = COALESCE(NULLIF($2, ''), name),
		    description = COALESCE(NULLIF($3, ''), description),
		    category_id = COALESCE(NULLIF($4, 0), category_id),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, name, description, category_id, created_at, updated_at`,
		id, req.Name, req.Description, req.CategoryID).Scan(
		&recipe.ID, &recipe.Name, &recipe.Description, &recipe.CategoryID,
		&recipe.CreatedAt, &recipe.UpdatedAt)

	return &recipe, err
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
