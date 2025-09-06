package repository

import (
	"database/sql"
	"meal-prep/shared/database"
	"meal-prep/shared/models"
)

type CategoryRepository interface {
	GetAll() ([]models.Category, error)
	GetByID(id int) (*models.Category, error)
	Exists(id int) (bool, error)
	Create(req models.CreateCategoryRequest) (*models.Category, error)
	Update(id int, req models.UpdateCategoryRequest) (*models.Category, error)
	Delete(id int) error
}

type categoryRepository struct {
	db *database.DB
}

func NewCategoryRepository(db *database.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) GetAll() ([]models.Category, error) {
	query := "SELECT id, name, description, created_at, updated_at FROM recipe_catalogue.categories ORDER BY name"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		err := rows.Scan(&category.ID, &category.Name, &category.Description, &category.CreatedAt, &category.UpdatedAt)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (r *categoryRepository) GetByID(id int) (*models.Category, error) {
	var category models.Category
	err := r.db.QueryRow(
		"SELECT id, name, description, created_at, updated_at FROM recipe_catalogue.categories WHERE id = $1",
		id).Scan(&category.ID, &category.Name, &category.Description, &category.CreatedAt, &category.UpdatedAt)

	if err != nil {
		return nil, err // Return nil on error (including sql.ErrNoRows)
	}

	return &category, nil // Return category only on success
}

func (r *categoryRepository) Exists(id int) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM recipe_catalogue.categories WHERE id = $1)",
		id).Scan(&exists)
	return exists, err
}

func (r *categoryRepository) Create(req models.CreateCategoryRequest) (*models.Category, error) {
	var category models.Category
	err := r.db.QueryRow(`
		INSERT INTO recipe_catalogue.categories (name, description, created_at, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, name, description, created_at, updated_at`,
		req.Name, req.Description).Scan(
		&category.ID, &category.Name, &category.Description, &category.CreatedAt, &category.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &category, nil
}

func (r *categoryRepository) Update(id int, req models.UpdateCategoryRequest) (*models.Category, error) {
	var category models.Category
	err := r.db.QueryRow(`
		UPDATE recipe_catalogue.categories 
		SET name = COALESCE($2, name),
		    description = CASE 
		        WHEN $3::text IS NOT NULL THEN $3
		        ELSE description
		    END,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, name, description, created_at, updated_at`,
		id, req.Name, req.Description).Scan(
		&category.ID, &category.Name, &category.Description, &category.CreatedAt, &category.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return &category, nil
}

func (r *categoryRepository) Delete(id int) error {
	result, err := r.db.Exec("DELETE FROM recipe_catalogue.categories WHERE id = $1", id)
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
