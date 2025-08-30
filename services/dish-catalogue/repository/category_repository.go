package repository

import (
	"meal-prep/shared/database"
	"meal-prep/shared/models"
)

type CategoryRepository interface {
	GetAll() ([]models.Category, error)
	GetByID(id int) (*models.Category, error)
	Exists(id int) (bool, error)
}

type categoryRepository struct {
	db *database.DB
}

func NewCategoryRepository(db *database.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) GetAll() ([]models.Category, error) {
	query := "SELECT id, name, description, created_at FROM dish_catalogue.categories ORDER BY name"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		err := rows.Scan(&category.ID, &category.Name, &category.Description, &category.CreatedAt)
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
		"SELECT id, name, description, created_at FROM dish_catalogue.categories WHERE id = $1",
		id).Scan(&category.ID, &category.Name, &category.Description, &category.CreatedAt)

	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) Exists(id int) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM dish_catalogue.categories WHERE id = $1)",
		id).Scan(&exists)
	return exists, err
}
