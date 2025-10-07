package service

import (
	"errors"
	"meal-prep/services/auth/service/mocks"
	"meal-prep/shared/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_MissingJWTConfig(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)

	assert.Panics(t, func() {
		NewAuthService(mockRepo)
	}, "Should panic when JWT config missing")
}

// Specifically test that JWT generator works
func TestAuthService_TokenGeneration(t *testing.T) {
	setupTestJWTEnv(t) // ← FIX: Added missing JWT setup

	// Arrange
	mockRepo := new(mocks.MockUserRepository)
	mockRepo.On("EmailExists", "test@example.com").Return(false, nil)
	mockRepo.On("Create", "test@example.com", mock.AnythingOfType("string")).Return(
		&models.User{
			ID:        1,
			Email:     "test@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil)

	service := NewAuthService(mockRepo)

	// Act
	result, err := service.Register("test@example.com", "password")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Token)
	assert.Contains(t, result.Token, ".")

	mockRepo.AssertExpectations(t)
}

func TestAuthService_TokenUniqueness(t *testing.T) {
	setupTestJWTEnv(t)

	// Arrange
	mockRepo := new(mocks.MockUserRepository)
	mockRepo.On("EmailExists", "user1@example.com").Return(false, nil)
	mockRepo.On("Create", "user1@example.com", mock.AnythingOfType("string")).Return(
		&models.User{ID: 1, Email: "user1@example.com", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil)

	mockRepo.On("EmailExists", "user2@example.com").Return(false, nil)
	mockRepo.On("Create", "user2@example.com", mock.AnythingOfType("string")).Return(
		&models.User{ID: 2, Email: "user2@example.com", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil)

	service := NewAuthService(mockRepo)

	// Act
	result1, _ := service.Register("user1@example.com", "password")
	result2, _ := service.Register("user2@example.com", "password")

	// Assert
	assert.NotEqual(t, result1.Token, result2.Token)

	mockRepo.AssertExpectations(t)
}

func TestAuthService_Register(t *testing.T) {
	setupTestJWTEnv(t)

	tests := []struct {
		name          string
		email         string
		password      string
		setupMocks    func(*mocks.MockUserRepository)
		expectedError error
		expectSuccess bool
	}{
		{
			name:     "successful_registration",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("EmailExists", "test@example.com").Return(false, nil)
				mockRepo.On("Create", "test@example.com", mock.AnythingOfType("string")).Return(
					&models.User{
						ID:        1,
						Email:     "test@example.com",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil)
			},
			expectedError: nil,
			expectSuccess: true,
		},
		{
			name:     "weak_password",
			email:    "test@example.com",
			password: "123",
			setupMocks: func(mockRepo *mocks.MockUserRepository) {
				// No mock setup needed - validation happens before repo call
			},
			expectedError: ErrWeakPassword,
			expectSuccess: false,
		},
		{
			name:     "user_already_exists",
			email:    "existing@example.com",
			password: "password123",
			setupMocks: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("EmailExists", "existing@example.com").Return(true, nil)
			},
			expectedError: ErrUserExists,
			expectSuccess: false,
		},
		{
			name:     "database_error_during_email_check",
			email:    "test@example.com",
			password: "validpassword123",
			setupMocks: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("EmailExists", "test@example.com").Return(false, errors.New("database connection failed"))
			},
			expectedError: errors.New("database connection failed"),
			expectSuccess: false,
		},
		{
			name:     "database_error_during_user_creation",
			email:    "test@example.com",
			password: "validpassword123",
			setupMocks: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("EmailExists", "test@example.com").Return(false, nil)
				mockRepo.On("Create", "test@example.com", mock.AnythingOfType("string")).Return(
					(*models.User)(nil), errors.New("insert failed"))
			},
			expectedError: errors.New("insert failed"),
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(mocks.MockUserRepository)
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
	setupTestJWTEnv(t)

	// Create test password hash
	testPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)

	tests := []struct {
		name          string
		email         string
		password      string
		setupMocks    func(*mocks.MockUserRepository)
		expectedError error
		expectSuccess bool
	}{
		{
			name:     "successful_login",
			email:    "test@example.com",
			password: testPassword,
			setupMocks: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("GetByEmail", "test@example.com").Return(
					&models.User{
						ID:        1,
						Email:     "test@example.com",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
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
			setupMocks: func(mockRepo *mocks.MockUserRepository) {
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
			setupMocks: func(mockRepo *mocks.MockUserRepository) {
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
			mockRepo := new(mocks.MockUserRepository)
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

func TestAuthService_validateInput(t *testing.T) {
	setupTestJWTEnv(t)

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
			mockRepo := new(mocks.MockUserRepository)
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

// =============================================================================
// TEST HELPERS
// =============================================================================

func setupTestJWTEnv(t *testing.T) {
	t.Helper()
	t.Setenv("JWT_SECRET", "test-secret-for-auth-tests")
	t.Setenv("JWT_ISSUER", "meal-prep-auth")
	t.Setenv("JWT_AUDIENCE", "meal-prep-api")
}
