package service

import (
	"testing"

	"meal-prep/services/auth/service/mocks"
	"meal-prep/shared/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Register(t *testing.T) {
	// Set JWT_SECRET for all subtests
	t.Setenv("JWT_SECRET", "overridden-for-tests")

	tests := []struct {
		name          string
		email         string
		password      string
		setupMocks    func(*mocks.UserRepository)
		expectedError error
		expectSuccess bool
	}{
		{
			name:     "successful_registration",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func(mockRepo *mocks.UserRepository) {
				mockRepo.On("EmailExists", "test@example.com").Return(false, nil)
				mockRepo.On("Create", "test@example.com", mock.AnythingOfType("string")).Return(
					&models.User{
						ID:    1,
						Email: "test@example.com",
					}, nil)
			},
			expectedError: nil,
			expectSuccess: true,
		},
		{
			name:     "weak_password",
			email:    "test@example.com",
			password: "123",
			setupMocks: func(mockRepo *mocks.UserRepository) {
				// No mock setup needed - validation happens before repo call
			},
			expectedError: ErrWeakPassword,
			expectSuccess: false,
		},
		{
			name:     "user_already_exists",
			email:    "existing@example.com",
			password: "password123",
			setupMocks: func(mockRepo *mocks.UserRepository) {
				mockRepo.On("EmailExists", "existing@example.com").Return(true, nil)
			},
			expectedError: ErrUserExists,
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(mocks.UserRepository)
			tt.setupMocks(mockRepo)
			service := NewAuthService(mockRepo)

			// Act
			result, err := service.Register(tt.email, tt.password)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.expectSuccess {
					assert.NotNil(t, result)
					assert.Equal(t, tt.email, result.User.Email)
					assert.NotEmpty(t, result.Token)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	// Set JWT_SECRET for all subtests
	t.Setenv("JWT_SECRET", "overridden-for-tests")
	// Create test password hash
	testPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)

	tests := []struct {
		name          string
		email         string
		password      string
		setupMocks    func(*mocks.UserRepository)
		expectedError error
		expectSuccess bool
	}{
		{
			name:     "successful_login",
			email:    "test@example.com",
			password: testPassword,
			setupMocks: func(mockRepo *mocks.UserRepository) {
				mockRepo.On("GetByEmail", "test@example.com").Return(
					&models.User{ID: 1, Email: "test@example.com"},
					string(hashedPassword),
					nil)
			},
			expectedError: nil,
			expectSuccess: true,
		},
		{
			name:     "invalid_credentials_wrong_password",
			email:    "test@example.com",
			password: "wrongpassword",
			setupMocks: func(mockRepo *mocks.UserRepository) {
				mockRepo.On("GetByEmail", "test@example.com").Return(
					&models.User{ID: 1, Email: "test@example.com"},
					string(hashedPassword),
					nil)
			},
			expectedError: ErrInvalidCredentials,
			expectSuccess: false,
		},
		{
			name:     "user_not_found",
			email:    "nonexistent@example.com",
			password: testPassword,
			setupMocks: func(mockRepo *mocks.UserRepository) {
				mockRepo.On("GetByEmail", "nonexistent@example.com").Return(
					(*models.User)(nil), "", ErrInvalidCredentials)
			},
			expectedError: ErrInvalidCredentials,
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(mocks.UserRepository)
			tt.setupMocks(mockRepo)
			service := NewAuthService(mockRepo)

			// Act
			result, err := service.Login(tt.email, tt.password)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.expectSuccess {
					assert.NotNil(t, result)
					assert.Equal(t, tt.email, result.User.Email)
					assert.NotEmpty(t, result.Token)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// Test helper functions
func TestAuthService_validateInput(t *testing.T) {
	// Set JWT_SECRET for all subtests
	t.Setenv("JWT_SECRET", "overridden-for-tests")

	tests := []struct {
		name      string
		email     string
		password  string
		wantError error
	}{
		{"valid_input", "test@example.com", "password123", nil},
		{"empty_email", "", "password123", nil}, // Service doesn't validate email format
		{"weak_password", "test@example.com", "123", ErrWeakPassword},
		{"empty_password", "test@example.com", "", ErrWeakPassword},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.UserRepository)
			if tt.wantError == nil && tt.password != "" {
				mockRepo.On("EmailExists", tt.email).Return(false, nil)
				mockRepo.On("Create", tt.email, mock.AnythingOfType("string")).Return(
					&models.User{ID: 1, Email: tt.email}, nil)
			}

			service := NewAuthService(mockRepo)
			_, err := service.Register(tt.email, tt.password)

			if tt.wantError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantError, err)
			}
		})
	}
}
