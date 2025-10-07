package test

import (
	"meal-prep/services/auth/internal/jwt"
	"time"
)

func GenerateTestJWT(userID int, email string) string {
	config := &jwt.Config{
		Secret:   "test-secret",
		Issuer:   "meal-prep-auth",
		Audience: []string{"meal-prep-api"},
		TTL:      24 * time.Hour,
	}

	generator := jwt.NewGenerator(config)
	token, _ := generator.Generate(userID, email)
	return token
}
