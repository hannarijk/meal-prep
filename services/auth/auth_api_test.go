package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"meal-prep/services/auth/handlers"
	"meal-prep/services/auth/repository"
	"meal-prep/services/auth/service"
	"meal-prep/shared/middleware"
	"meal-prep/test/helpers"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AuthE2ETestSuite struct {
	suite.Suite
	testDB *helpers.TestDatabase
	server *httptest.Server
	client *http.Client
}

func (suite *AuthE2ETestSuite) SetupSuite() {
	suite.T().Setenv("JWT_SECRET", "test-secret-for-e2e-tests")
	suite.T().Setenv("JWT_ISSUER", "meal-prep-auth")
	suite.T().Setenv("JWT_AUDIENCE", "meal-prep-api")

	helpers.SuppressTestLogs()

	// Setup real database
	suite.testDB = helpers.SetupPostgresContainer(suite.T())

	// Create real services with real database
	userRepo := repository.NewUserRepository(suite.testDB.DB)
	authService := service.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)

	// Setup real HTTP server
	router := mux.NewRouter()
	router.Use(middleware.LoggingMiddleware("test-auth-service"))
	router.HandleFunc("/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy"}`))
	}).Methods("GET")

	suite.server = httptest.NewServer(router)
	suite.client = &http.Client{Timeout: 10 * time.Second}
}

func (suite *AuthE2ETestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
	suite.testDB.Cleanup(suite.T())
	helpers.RestoreTestLogs()
}

func (suite *AuthE2ETestSuite) SetupTest() {
	suite.testDB.CleanupTestData(suite.T())
}

func (suite *AuthE2ETestSuite) TestCompleteAuthFlow_RegisterAndLogin() {
	// Test complete user authentication workflow

	baseURL := suite.server.URL

	// Step 1: Register new user
	registerPayload := map[string]string{
		"email":    "e2e-user@example.com",
		"password": "securepassword123",
	}

	registerResp := suite.makeRequest("POST", baseURL+"/register", registerPayload)
	assert.Equal(suite.T(), http.StatusCreated, registerResp.StatusCode)

	var registerResult map[string]interface{}
	err := json.NewDecoder(registerResp.Body).Decode(&registerResult)
	assert.NoError(suite.T(), err)

	// Verify registration response
	assert.Contains(suite.T(), registerResult, "token")
	assert.Contains(suite.T(), registerResult, "user")

	registrationToken := registerResult["token"].(string)
	assert.NotEmpty(suite.T(), registrationToken)

	userInfo := registerResult["user"].(map[string]interface{})
	assert.Equal(suite.T(), "e2e-user@example.com", userInfo["email"])
	assert.NotZero(suite.T(), userInfo["id"])

	// Step 2: Try to register same user again (should fail)
	duplicateResp := suite.makeRequest("POST", baseURL+"/register", registerPayload)
	assert.Equal(suite.T(), http.StatusConflict, duplicateResp.StatusCode)

	var errorResult map[string]interface{}
	json.NewDecoder(duplicateResp.Body).Decode(&errorResult)
	assert.Contains(suite.T(), errorResult["message"], "already exists")

	// Step 3: Login with correct credentials
	loginResp := suite.makeRequest("POST", baseURL+"/login", registerPayload)
	assert.Equal(suite.T(), http.StatusOK, loginResp.StatusCode)

	var loginResult map[string]interface{}
	err = json.NewDecoder(loginResp.Body).Decode(&loginResult)
	assert.NoError(suite.T(), err)

	loginToken := loginResult["token"].(string)
	assert.NotEmpty(suite.T(), loginToken)

	// Step 4: Login with wrong credentials (should fail)
	wrongPasswordPayload := map[string]string{
		"email":    "e2e-user@example.com",
		"password": "wrongpassword",
	}

	wrongLoginResp := suite.makeRequest("POST", baseURL+"/login", wrongPasswordPayload)
	assert.Equal(suite.T(), http.StatusUnauthorized, wrongLoginResp.StatusCode)
}

func (suite *AuthE2ETestSuite) TestRegistrationValidation_E2E() {
	baseURL := suite.server.URL

	testCases := []struct {
		name           string
		payload        map[string]string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "weak_password",
			payload:        map[string]string{"email": "test@example.com", "password": "123"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "password must be at least 6 characters",
		},
		{
			name:           "empty_email",
			payload:        map[string]string{"email": "", "password": "password123"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email and password are required",
		},
		{
			name:           "missing_password",
			payload:        map[string]string{"email": "test@example.com"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Email and password are required",
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			resp := suite.makeRequest("POST", baseURL+"/register", tc.payload)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			var errorResult map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&errorResult)
			assert.Contains(t, errorResult["message"].(string), tc.expectedError)
		})
	}
}

func (suite *AuthE2ETestSuite) TestHealthCheck() {
	resp := suite.makeRequest("GET", suite.server.URL+"/health", nil)
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	body := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(suite.T(), "healthy", body["status"])
}

// Helper method to make HTTP requests
func (suite *AuthE2ETestSuite) makeRequest(method, url string, payload interface{}) *http.Response {
	var body *bytes.Buffer

	if payload != nil {
		jsonPayload, _ := json.Marshal(payload)
		body = bytes.NewBuffer(jsonPayload)
	} else {
		body = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, url, body)
	suite.Require().NoError(err)

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := suite.client.Do(req)
	suite.Require().NoError(err)

	return resp
}

func TestAuthE2E(t *testing.T) {
	suite.Run(t, new(AuthE2ETestSuite))
}
