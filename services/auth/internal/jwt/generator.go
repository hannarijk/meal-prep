package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type Generator struct {
	config *Config
}

func NewGenerator(config *Config) *Generator {
	return &Generator{config: config}
}

func (g *Generator) Generate(userID int, email string) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			Issuer:    g.config.Issuer,
			Audience:  jwt.ClaimStrings(g.config.Audience),
			ExpiresAt: jwt.NewNumericDate(now.Add(g.config.TTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(g.config.Secret))
}
