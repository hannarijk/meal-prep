package jwt

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Secret   string
	Issuer   string
	Audience []string
	TTL      time.Duration
}

func LoadConfig() (*Config, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET not set")
	}

	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		issuer = "meal-prep-auth"
	}

	audience := os.Getenv("JWT_AUDIENCE")
	if audience == "" {
		audience = "meal-prep-api"
	}

	return &Config{
		Secret:   secret,
		Issuer:   issuer,
		Audience: []string{audience},
		TTL:      24 * time.Hour,
	}, nil
}
