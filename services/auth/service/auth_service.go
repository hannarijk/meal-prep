package service

import (
	"errors"
	"meal-prep/services/auth/repository"
	"meal-prep/shared/models"

	"golang.org/x/crypto/bcrypt"

	"meal-prep/services/auth/internal/jwt"
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
	userRepo     repository.UserRepository
	jwtGenerator *jwt.Generator
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	jwtConfig, err := jwt.LoadConfig()
	if err != nil {
		panic("Failed to load JWT config: " + err.Error())
	}

	return &authService{
		userRepo:     userRepo,
		jwtGenerator: jwt.NewGenerator(jwtConfig),
	}
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

	token, err := s.jwtGenerator.Generate(user.ID, user.Email)
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

	token, err := s.jwtGenerator.Generate(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}
