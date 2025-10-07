package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewGenerator tests creating a JWT generator
func TestNewGenerator(t *testing.T) {
	// Arrange
	config := &Config{
		Secret:   "test-secret",
		Issuer:   "meal-prep-auth",
		Audience: []string{"meal-prep-api"},
		TTL:      24 * time.Hour,
	}

	// Act
	generator := NewGenerator(config)

	// Assert
	assert.NotNil(t, generator, "Generator should not be nil")
	assert.Equal(t, config, generator.config, "Config should be stored")
}

// TestGenerator_Generate_Success tests successful JWT generation
func TestGenerator_Generate_Success(t *testing.T) {
	// Arrange
	config := &Config{
		Secret:   "test-secret-key",
		Issuer:   "meal-prep-auth",
		Audience: []string{"meal-prep-api"},
		TTL:      24 * time.Hour,
	}
	generator := NewGenerator(config)

	userID := 42
	email := "test@example.com"

	// Act
	token, err := generator.Generate(userID, email)

	// Assert
	require.NoError(t, err, "Should generate token without error")
	assert.NotEmpty(t, token, "Token should not be empty")
	assert.True(t, len(token) > 50, "Token should be reasonable length")
}

// TestGenerator_Generate_ValidClaims tests that generated token has correct claims
func TestGenerator_Generate_ValidClaims(t *testing.T) {
	// Arrange
	config := &Config{
		Secret:   "test-secret-key",
		Issuer:   "meal-prep-auth",
		Audience: []string{"meal-prep-api"},
		TTL:      1 * time.Hour,
	}
	generator := NewGenerator(config)

	userID := 123
	email := "user@example.com"
	beforeGeneration := time.Now()

	// Act
	tokenString, err := generator.Generate(userID, email)
	require.NoError(t, err)

	// Parse token to verify claims
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Secret), nil
	})

	// Assert token is valid
	require.NoError(t, err, "Token should be parseable")
	require.True(t, token.Valid, "Token should be valid")

	// Assert claims
	claims, ok := token.Claims.(*Claims)
	require.True(t, ok, "Claims should be correct type")

	// Custom claims
	assert.Equal(t, userID, claims.UserID, "UserID should match")
	assert.Equal(t, email, claims.Email, "Email should match")

	// Standard claims
	assert.Equal(t, "123", claims.Subject, "Subject should be string user ID")
	assert.Equal(t, "meal-prep-auth", claims.Issuer, "Issuer should match config")
	assert.Equal(t, jwt.ClaimStrings{"meal-prep-api"}, claims.Audience, "Audience should match config")

	// Time-based claims (with tolerance for test execution time)
	assert.NotNil(t, claims.IssuedAt, "IssuedAt should be set")
	assert.NotNil(t, claims.NotBefore, "NotBefore should be set")
	assert.NotNil(t, claims.ExpiresAt, "ExpiresAt should be set")

	// Check IssuedAt is recent (within 2 seconds for slower systems)
	issuedAt := claims.IssuedAt.Time
	assert.WithinDuration(t, beforeGeneration, issuedAt, 2*time.Second, "IssuedAt should be recent")

	// Check NotBefore is same as IssuedAt
	assert.Equal(t, issuedAt, claims.NotBefore.Time, "NotBefore should equal IssuedAt")

	// Check ExpiresAt is TTL after IssuedAt
	expectedExpiry := issuedAt.Add(config.TTL)
	assert.Equal(t, expectedExpiry, claims.ExpiresAt.Time, "ExpiresAt should be IssuedAt + TTL")
}

// TestGenerator_Generate_DifferentUsers tests tokens for different users are unique
func TestGenerator_Generate_DifferentUsers(t *testing.T) {
	// Arrange
	config := &Config{
		Secret:   "test-secret",
		Issuer:   "meal-prep-auth",
		Audience: []string{"meal-prep-api"},
		TTL:      24 * time.Hour,
	}
	generator := NewGenerator(config)

	// Act - Generate tokens for different users
	token1, err1 := generator.Generate(1, "user1@example.com")
	token2, err2 := generator.Generate(2, "user2@example.com")

	// Assert
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.NotEqual(t, token1, token2, "Tokens for different users should be different")
}

// TestGenerator_Generate_SameUserSameSecond tests token consistency
// NOTE: JWTs have SECOND-level precision, so tokens generated within the same
// second for the same user will be identical. This is expected behavior.
func TestGenerator_Generate_SameUserSameSecond(t *testing.T) {
	// Arrange
	config := &Config{
		Secret:   "test-secret",
		Issuer:   "meal-prep-auth",
		Audience: []string{"meal-prep-api"},
		TTL:      24 * time.Hour,
	}
	generator := NewGenerator(config)

	userID := 42
	email := "test@example.com"

	// Act - Generate same user token in different seconds
	token1, err1 := generator.Generate(userID, email)
	require.NoError(t, err1)

	// Sleep FULL second to ensure different timestamp
	// JWT timestamps are second-precision, not millisecond
	time.Sleep(1100 * time.Millisecond)

	token2, err2 := generator.Generate(userID, email)
	require.NoError(t, err2)

	// Assert - Tokens generated in different seconds should be different
	assert.NotEqual(t, token1, token2, "Tokens generated in different seconds should have different timestamps")

	// Verify both tokens are valid and have correct user info
	for i, token := range []string{token1, token2} {
		parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.Secret), nil
		})
		require.NoError(t, err, "Token %d should be valid", i+1)

		claims := parsed.Claims.(*Claims)
		assert.Equal(t, userID, claims.UserID, "Token %d should have correct user ID", i+1)
		assert.Equal(t, email, claims.Email, "Token %d should have correct email", i+1)
	}
}

// TestGenerator_Generate_TableDriven tests various scenarios using table-driven tests
func TestGenerator_Generate_TableDriven(t *testing.T) {
	config := &Config{
		Secret:   "test-secret",
		Issuer:   "meal-prep-auth",
		Audience: []string{"meal-prep-api"},
		TTL:      24 * time.Hour,
	}
	generator := NewGenerator(config)

	tests := []struct {
		name        string
		userID      int
		email       string
		expectError bool
		description string
	}{
		{
			name:        "valid_user",
			userID:      1,
			email:       "user@example.com",
			expectError: false,
			description: "Standard valid user should generate token",
		},
		{
			name:        "large_user_id",
			userID:      999999,
			email:       "biguser@example.com",
			expectError: false,
			description: "Large user ID should work",
		},
		{
			name:        "zero_user_id",
			userID:      0,
			email:       "zero@example.com",
			expectError: false,
			description: "Zero user ID is technically valid",
		},
		{
			name:        "long_email",
			userID:      5,
			email:       "very.long.email.address.for.testing@subdomain.example.com",
			expectError: false,
			description: "Long email should work",
		},
		{
			name:        "special_chars_email",
			userID:      6,
			email:       "user+tag@example.com",
			expectError: false,
			description: "Email with special chars should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			token, err := generator.Generate(tt.userID, tt.email)

			// Assert
			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Empty(t, token, "Token should be empty on error")
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotEmpty(t, token, "Token should not be empty")

				// Verify token can be parsed
				parsedToken, parseErr := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte(config.Secret), nil
				})
				assert.NoError(t, parseErr, "Token should be parseable")
				assert.True(t, parsedToken.Valid, "Token should be valid")

				claims := parsedToken.Claims.(*Claims)
				assert.Equal(t, tt.userID, claims.UserID, "UserID should match")
				assert.Equal(t, tt.email, claims.Email, "Email should match")
			}
		})
	}
}

// TestGenerator_TokenValidation tests that tokens can be validated
func TestGenerator_TokenValidation(t *testing.T) {
	// Arrange
	config := &Config{
		Secret:   "correct-secret",
		Issuer:   "meal-prep-auth",
		Audience: []string{"meal-prep-api"},
		TTL:      1 * time.Hour,
	}
	generator := NewGenerator(config)

	token, err := generator.Generate(1, "test@example.com")
	require.NoError(t, err)

	tests := []struct {
		name        string
		secret      string
		shouldValid bool
		description string
	}{
		{
			name:        "correct_secret",
			secret:      "correct-secret",
			shouldValid: true,
			description: "Token should validate with correct secret",
		},
		{
			name:        "wrong_secret",
			secret:      "wrong-secret",
			shouldValid: false,
			description: "Token should fail with wrong secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act - Try to parse token with different secrets
			parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(tt.secret), nil
			})

			// Assert
			if tt.shouldValid {
				assert.NoError(t, err, tt.description)
				assert.True(t, parsedToken.Valid, tt.description)
			} else {
				assert.Error(t, err, tt.description)
			}
		})
	}
}

// TestGenerator_ExpiredToken tests expired token detection
func TestGenerator_ExpiredToken(t *testing.T) {
	// Arrange - Config with very short TTL
	config := &Config{
		Secret:   "test-secret",
		Issuer:   "meal-prep-auth",
		Audience: []string{"meal-prep-api"},
		TTL:      1 * time.Second, // Short expiry
	}
	generator := NewGenerator(config)

	// Act - Generate token
	token, err := generator.Generate(1, "test@example.com")
	require.NoError(t, err)

	// Wait for token to expire (need >1 second because of JWT second precision)
	time.Sleep(1500 * time.Millisecond)

	// Try to parse expired token
	parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Secret), nil
	})

	// Assert - Should fail due to expiration
	assert.Error(t, err, "Expired token should fail validation")
	if parsedToken != nil {
		assert.False(t, parsedToken.Valid, "Expired token should not be valid")
	}
	assert.Contains(t, err.Error(), "expired", "Error should mention expiration")
}

// TestGenerator_SigningMethodValidation tests signing method is correct
func TestGenerator_SigningMethodValidation(t *testing.T) {
	// Arrange
	config := &Config{
		Secret:   "test-secret",
		Issuer:   "meal-prep-auth",
		Audience: []string{"meal-prep-api"},
		TTL:      1 * time.Hour,
	}
	generator := NewGenerator(config)

	// Act
	token, err := generator.Generate(1, "test@example.com")
	require.NoError(t, err)

	// Parse to check signing method
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		if token.Method.Alg() != "HS256" {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(config.Secret), nil
	})

	// Assert
	require.NoError(t, err, "Should parse with HS256 validation")
	assert.True(t, parsedToken.Valid, "Token should be valid")
	assert.Equal(t, "HS256", parsedToken.Method.Alg(), "Should use HS256 algorithm")
}

// TestLoadConfig tests configuration loading
func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name          string
		envVars       map[string]string
		expectError   bool
		expectedValue func(*Config) bool
		description   string
	}{
		{
			name: "all_env_vars_set",
			envVars: map[string]string{
				"JWT_SECRET":   "my-secret",
				"JWT_ISSUER":   "my-issuer",
				"JWT_AUDIENCE": "my-audience",
			},
			expectError: false,
			expectedValue: func(c *Config) bool {
				return c.Secret == "my-secret" && c.Issuer == "my-issuer" && c.Audience[0] == "my-audience"
			},
			description: "Should load all config from env vars",
		},
		{
			name: "missing_secret",
			envVars: map[string]string{
				"JWT_ISSUER":   "my-issuer",
				"JWT_AUDIENCE": "my-audience",
			},
			expectError: true,
			description: "Should error when JWT_SECRET is missing",
		},
		{
			name: "default_issuer",
			envVars: map[string]string{
				"JWT_SECRET": "my-secret",
			},
			expectError: false,
			expectedValue: func(c *Config) bool {
				return c.Issuer == "meal-prep-auth"
			},
			description: "Should use default issuer if not set",
		},
		{
			name: "default_audience",
			envVars: map[string]string{
				"JWT_SECRET": "my-secret",
			},
			expectError: false,
			expectedValue: func(c *Config) bool {
				return c.Audience[0] == "meal-prep-api"
			},
			description: "Should use default audience if not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - Set environment variables
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			// Act
			config, err := LoadConfig()

			// Assert
			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, config, "Config should be nil on error")
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, config, "Config should not be nil")
				assert.Equal(t, 24*time.Hour, config.TTL, "TTL should be 24 hours")

				if tt.expectedValue != nil {
					result := tt.expectedValue(config)
					assert.True(t, result, tt.description)
				}
			}
		})
	}
}

// BenchmarkGenerator_Generate benchmarks token generation performance
func BenchmarkGenerator_Generate(b *testing.B) {
	config := &Config{
		Secret:   "benchmark-secret",
		Issuer:   "meal-prep-auth",
		Audience: []string{"meal-prep-api"},
		TTL:      24 * time.Hour,
	}
	generator := NewGenerator(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generator.Generate(i, "bench@example.com")
	}
}

// BenchmarkGenerator_GenerateAndParse benchmarks full generation and parsing
func BenchmarkGenerator_GenerateAndParse(b *testing.B) {
	config := &Config{
		Secret:   "benchmark-secret",
		Issuer:   "meal-prep-auth",
		Audience: []string{"meal-prep-api"},
		TTL:      24 * time.Hour,
	}
	generator := NewGenerator(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		token, _ := generator.Generate(i, "bench@example.com")
		_, _ = jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.Secret), nil
		})
	}
}
