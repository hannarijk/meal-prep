package service

import (
	"errors"
	"meal-prep/services/auth/repository"
	"meal-prep/shared/models"
	"meal-prep/shared/utils"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists         = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrWeakPassword       = errors.New("password must be at least 6 characters")
)

type AuthService interface {
	Register(email, password string) (*models.AuthResponse, error)
	Login(email, password string) (*models.AuthResponse, error)
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

func (s *authService) Register(email, password string) (*models.AuthResponse, error) {
	if len(password) < 6 {
		return nil, ErrWeakPassword
	}

	exists, err := s.userRepo.EmailExists(email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.Create(email, string(hashedPassword))
	if err != nil {
		return nil, err
	}

	token, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *authService) Login(email, password string) (*models.AuthResponse, error) {
	user, passwordHash, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}
