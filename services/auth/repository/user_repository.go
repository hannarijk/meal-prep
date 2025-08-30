package repository

import (
	"meal-prep/shared/database"
	"meal-prep/shared/models"
)

type UserRepository interface {
	Create(email, passwordHash string) (*models.User, error)
	GetByEmail(email string) (*models.User, string, error)
	EmailExists(email string) (bool, error)
}

type userRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(email, passwordHash string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(`
		INSERT INTO auth.users (email, password_hash, created_at, updated_at) 
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) 
		RETURNING id, email, created_at, updated_at`,
		email, passwordHash).Scan(
		&user.ID, &user.Email, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &user, err
}

func (r *userRepository) GetByEmail(email string) (*models.User, string, error) {
	var user models.User
	var passwordHash string
	err := r.db.QueryRow(`
		SELECT id, email, password_hash, created_at, updated_at 
		FROM auth.users WHERE email = $1`, email).Scan(
		&user.ID, &user.Email, &passwordHash, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, "", err
	}
	return &user, passwordHash, nil
}

func (r *userRepository) EmailExists(email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM auth.users WHERE email = $1)", email).Scan(&exists)
	return exists, err
}
