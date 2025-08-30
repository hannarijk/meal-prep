package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"meal-prep/services/auth/handlers/mocks"
	"meal-prep/services/auth/service"
	"meal-prep/shared/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthHandler_Register(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name             string
		requestBody      string
		contentType      string
		setupMocks       func(*mocks.MockAuthService)
		expectedStatus   int
		expectedResponse map[string]interface{}
		description      string
	}{
		{
			name:        "successful_registration",
			requestBody: `{"email":"test@example.com","password":"password123"}`,
			contentType: "application/json",
			description: "Valid registration request should return 201 with token",
			setupMocks: func(mockService *mocks.MockAuthService) {
				mockService.On("Register", "test@example.com", "password123").Return(
					&models.AuthResponse{
						Token: "jwt_token_here",
						User: models.User{
							ID:        1,
							Email:     "test@example.com",
							CreatedAt: now,
							UpdatedAt: now,
						},
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedResponse: map[string]interface{}{
				"token": "jwt_token_here",
				"user": map[string]interface{}{
					"id":         float64(1), // JSON numbers become float64
					"email":      "test@example.com",
					"created_at": now.Format(time.RFC3339Nano),
					"updated_at": now.Format(time.RFC3339Nano),
				},
			},
		},
		{
			name:        "invalid_json",
			requestBody: `{"email":"test@example.com","password":}`, // Invalid JSON
			contentType: "application/json",
			description: "Invalid JSON should return 400 Bad Request",
			setupMocks: func(mockService *mocks.MockAuthService) {
				// No service calls expected for invalid JSON
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: map[string]interface{}{
				"error":   "auth_error",
				"code":    float64(400),
				"message": "Invalid JSON",
			},
		},
		{
			name:        "missing_email",
			requestBody: `{"password":"password123"}`,
			contentType: "application/json",
			description: "Missing email should return 400 Bad Request",
			setupMocks: func(mockService *mocks.MockAuthService) {
				// No service calls expected for invalid input
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: map[string]interface{}{
				"error":   "auth_error",
				"code":    float64(400),
				"message": "Email and password are required",
			},
		},
		{
			name:        "missing_password",
			requestBody: `{"email":"test@example.com"}`,
			contentType: "application/json",
			description: "Missing password should return 400 Bad Request",
			setupMocks: func(mockService *mocks.MockAuthService) {
				// No service calls expected for invalid input
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: map[string]interface{}{
				"error":   "auth_error",
				"code":    float64(400),
				"message": "Email and password are required",
			},
		},
		{
			name:        "user_already_exists",
			requestBody: `{"email":"existing@example.com","password":"password123"}`,
			contentType: "application/json",
			description: "Duplicate user should return 409 Conflict",
			setupMocks: func(mockService *mocks.MockAuthService) {
				mockService.On("Register", "existing@example.com", "password123").Return(
					(*models.AuthResponse)(nil), service.ErrUserExists)
			},
			expectedStatus: http.StatusConflict,
			expectedResponse: map[string]interface{}{
				"error":   "auth_error",
				"code":    float64(409),
				"message": service.ErrUserExists.Error(),
			},
		},
		{
			name:        "weak_password",
			requestBody: `{"email":"test@example.com","password":"123"}`,
			contentType: "application/json",
			description: "Weak password should return 400 Bad Request",
			setupMocks: func(mockService *mocks.MockAuthService) {
				mockService.On("Register", "test@example.com", "123").Return(
					(*models.AuthResponse)(nil), service.ErrWeakPassword)
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: map[string]interface{}{
				"error":   "auth_error",
				"code":    float64(400),
				"message": service.ErrWeakPassword.Error(),
			},
		},
		{
			name:        "internal_server_error",
			requestBody: `{"email":"test@example.com","password":"password123"}`,
			contentType: "application/json",
			description: "Unexpected service errors should return 500",
			setupMocks: func(mockService *mocks.MockAuthService) {
				mockService.On("Register", "test@example.com", "password123").Return(
					(*models.AuthResponse)(nil), errors.New("unexpected database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: map[string]interface{}{
				"error":   "auth_error",
				"code":    float64(500),
				"message": "Internal server error",
			},
		},
		{
			name:        "missing_content_type",
			requestBody: `{"email":"test@example.com","password":"password123"}`,
			contentType: "", // No content type
			description: "Missing content type should still work (graceful handling)",
			setupMocks: func(mockService *mocks.MockAuthService) {
				mockService.On("Register", "test@example.com", "password123").Return(
					&models.AuthResponse{
						Token: "jwt_token",
						User:  models.User{ID: 1, Email: "test@example.com"},
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedResponse: map[string]interface{}{
				"token": "jwt_token",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockService := new(mocks.MockAuthService)
			tt.setupMocks(mockService)
			handler := NewAuthHandler(mockService)

			// Create HTTP request
			req := httptest.NewRequest("POST", "/register", bytes.NewBufferString(tt.requestBody))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			// Create response recorder
			recorder := httptest.NewRecorder()

			// Act
			handler.Register(recorder, req)

			// Assert HTTP status
			assert.Equal(t, tt.expectedStatus, recorder.Code)

			// Assert response content type
			assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

			// Parse and verify JSON response
			var response map[string]interface{}
			err := json.NewDecoder(recorder.Body).Decode(&response)
			assert.NoError(t, err)

			// Verify expected response fields
			for key, expectedValue := range tt.expectedResponse {
				if actualValue, exists := response[key]; exists {
					assert.Equal(t, expectedValue, actualValue, "Response field %s mismatch", key)
				} else {
					t.Errorf("Expected response field %s not found in response", key)
				}
			}

			// Verify service mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		setupMocks     func(*mocks.MockAuthService)
		expectedStatus int
		expectToken    bool
		expectError    bool
		description    string
	}{
		{
			name:        "successful_login",
			requestBody: `{"email":"user@example.com","password":"correctpassword"}`,
			description: "Valid login should return 200 with token",
			setupMocks: func(mockService *mocks.MockAuthService) {
				mockService.On("Login", "user@example.com", "correctpassword").Return(
					&models.AuthResponse{
						Token: "valid_jwt_token",
						User:  models.User{ID: 1, Email: "user@example.com"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectToken:    true,
			expectError:    false,
		},
		{
			name:        "invalid_credentials",
			requestBody: `{"email":"user@example.com","password":"wrongpassword"}`,
			description: "Wrong password should return 401 Unauthorized",
			setupMocks: func(mockService *mocks.MockAuthService) {
				mockService.On("Login", "user@example.com", "wrongpassword").Return(
					(*models.AuthResponse)(nil), service.ErrInvalidCredentials)
			},
			expectedStatus: http.StatusUnauthorized,
			expectToken:    false,
			expectError:    true,
		},
		{
			name:        "empty_request_body",
			requestBody: `{}`,
			description: "Empty credentials should return 400 Bad Request",
			setupMocks: func(mockService *mocks.MockAuthService) {
				// No service calls expected
			},
			expectedStatus: http.StatusBadRequest,
			expectToken:    false,
			expectError:    true,
		},
		{
			name:        "malformed_json",
			requestBody: `{"email":"user@example.com","password":`,
			description: "Malformed JSON should return 400 Bad Request",
			setupMocks: func(mockService *mocks.MockAuthService) {
				// No service calls expected for malformed JSON
			},
			expectedStatus: http.StatusBadRequest,
			expectToken:    false,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockService := new(mocks.MockAuthService)
			tt.setupMocks(mockService)
			handler := NewAuthHandler(mockService)

			req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			// Act
			handler.Login(recorder, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, recorder.Code)
			assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

			var response map[string]interface{}
			err := json.NewDecoder(recorder.Body).Decode(&response)
			assert.NoError(t, err)

			if tt.expectToken {
				assert.Contains(t, response, "token")
				assert.NotEmpty(t, response["token"])
				assert.Contains(t, response, "user")
			}

			if tt.expectError {
				assert.Contains(t, response, "error")
				assert.Contains(t, response, "message")
				assert.NotContains(t, response, "token")
			}

			mockService.AssertExpectations(t)
		})
	}
}

// Test HTTP method restrictions
func TestAuthHandler_MethodNotAllowed(t *testing.T) {
	mockService := new(mocks.MockAuthService)
	handler := NewAuthHandler(mockService)

	// Test wrong HTTP methods
	methods := []string{"GET", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run("method_"+method+"_not_allowed", func(t *testing.T) {
			req := httptest.NewRequest(method, "/register", nil)
			recorder := httptest.NewRecorder()

			// Since we're testing just the handler, not the router,
			// we expect the handler to process any method
			// (Method filtering happens at router level)
			handler.Register(recorder, req)

			// The handler should still try to process it
			// (will fail on JSON parsing, but that's expected)
			assert.Equal(t, http.StatusBadRequest, recorder.Code)
		})
	}
}

// Benchmark handler performance
func BenchmarkAuthHandler_Register(b *testing.B) {
	mockService := new(mocks.MockAuthService)
	mockService.On("Register", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(
		&models.AuthResponse{
			Token: "benchmark_token",
			User:  models.User{ID: 1, Email: "bench@example.com"},
		}, nil).Times(b.N)

	handler := NewAuthHandler(mockService)
	requestBody := `{"email":"bench@example.com","password":"password123"}`

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/register", bytes.NewBufferString(requestBody))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		handler.Register(recorder, req)

		if recorder.Code != http.StatusCreated {
			b.Fatalf("Expected status %d, got %d", http.StatusCreated, recorder.Code)
		}
	}
}
