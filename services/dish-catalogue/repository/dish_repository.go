package repository

import (
	"database/sql"
	"meal-prep/shared/database"
	"meal-prep/shared/models"
)

type DishRepository interface {
	GetAll() ([]models.Dish, error)
	GetByID(id int) (*models.Dish, error)
	GetByCategory(categoryID int) ([]models.Dish, error)
	Create(req models.CreateDishRequest) (*models.Dish, error)
	Update(id int, req models.UpdateDishRequest) (*models.Dish, error)
	Delete(id int) error
}

type dishRepository struct {
	db *database.DB
}

func NewDishRepository(db *database.DB) DishRepository {
	return &dishRepository{db: db}
}

func (r *dishRepository) GetAll() ([]models.Dish, error) {
	query := `
		SELECT d.id, d.name, d.description, d.category_id, d.created_at, d.updated_at,
		       c.id, c.name, c.description
		FROM dish_catalogue.dishes d
		LEFT JOIN dish_catalogue.categories c ON d.category_id = c.id
		ORDER BY d.name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dishes []models.Dish
	for rows.Next() {
		dish := models.Dish{}
		category := models.Category{}
		var categoryDesc sql.NullString

		err := rows.Scan(
			&dish.ID, &dish.Name, &dish.Description, &dish.CategoryID,
			&dish.CreatedAt, &dish.UpdatedAt,
			&category.ID, &category.Name, &categoryDesc,
		)
		if err != nil {
			return nil, err
		}

		if categoryDesc.Valid {
			category.Description = &categoryDesc.String
		}
		dish.Category = &category

		dishes = append(dishes, dish)
	}

	return dishes, nil
}

func (r *dishRepository) GetByID(id int) (*models.Dish, error) {
	query := `
		SELECT d.id, d.name, d.description, d.category_id, d.created_at, d.updated_at,
		       c.id, c.name, c.description
		FROM dish_catalogue.dishes d
		LEFT JOIN dish_catalogue.categories c ON d.category_id = c.id
		WHERE d.id = $1`

	dish := models.Dish{}
	category := models.Category{}
	var categoryDesc sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&dish.ID, &dish.Name, &dish.Description, &dish.CategoryID,
		&dish.CreatedAt, &dish.UpdatedAt,
		&category.ID, &category.Name, &categoryDesc,
	)

	if err != nil {
		return nil, err
	}

	if categoryDesc.Valid {
		category.Description = &categoryDesc.String
	}
	dish.Category = &category

	return &dish, nil
}

func (r *dishRepository) GetByCategory(categoryID int) ([]models.Dish, error) {
	query := `
		SELECT d.id, d.name, d.description, d.category_id, d.created_at, d.updated_at,
		       c.id, c.name, c.description
		FROM dish_catalogue.dishes d
		LEFT JOIN dish_catalogue.categories c ON d.category_id = c.id
		WHERE d.category_id = $1
		ORDER BY d.name`

	rows, err := r.db.Query(query, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dishes []models.Dish
	for rows.Next() {
		dish := models.Dish{}
		category := models.Category{}
		var categoryDesc sql.NullString

		err := rows.Scan(
			&dish.ID, &dish.Name, &dish.Description, &dish.CategoryID,
			&dish.CreatedAt, &dish.UpdatedAt,
			&category.ID, &category.Name, &categoryDesc,
		)
		if err != nil {
			return nil, err
		}

		if categoryDesc.Valid {
			category.Description = &categoryDesc.String
		}
		dish.Category = &category

		dishes = append(dishes, dish)
	}

	return dishes, nil
}

func (r *dishRepository) Create(req models.CreateDishRequest) (*models.Dish, error) {
	var dish models.Dish
	err := r.db.QueryRow(`
		INSERT INTO dish_catalogue.dishes (name, description, category_id, created_at, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, name, description, category_id, created_at, updated_at`,
		req.Name, req.Description, req.CategoryID).Scan(
		&dish.ID, &dish.Name, &dish.Description, &dish.CategoryID,
		&dish.CreatedAt, &dish.UpdatedAt)

	return &dish, err
}

func (r *dishRepository) Update(id int, req models.UpdateDishRequest) (*models.Dish, error) {
	var dish models.Dish
	err := r.db.QueryRow(`
		UPDATE dish_catalogue.dishes 
		SET name = COALESCE(NULLIF($2, ''), name),
		    description = COALESCE(NULLIF($3, ''), description),
		    category_id = COALESCE(NULLIF($4, 0), category_id),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, name, description, category_id, created_at, updated_at`,
		id, req.Name, req.Description, req.CategoryID).Scan(
		&dish.ID, &dish.Name, &dish.Description, &dish.CategoryID,
		&dish.CreatedAt, &dish.UpdatedAt)

	return &dish, err
}

func (r *dishRepository) Delete(id int) error {
	result, err := r.db.Exec("DELETE FROM dish_catalogue.dishes WHERE id = $1", id)
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
