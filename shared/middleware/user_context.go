package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"meal-prep/shared/logging"
	"meal-prep/shared/models"

	"github.com/golang-jwt/jwt/v5"
)

// UserContext holds user info from gateway headers
type UserContext struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
}

type userContextKey string

const UserCtxKey userContextKey = "gateway_user"

// JWT Claims structure (must match auth service)
type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// ExtractUserFromGatewayHeaders reads user info from Kong headers
func ExtractUserFromGatewayHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logging.WithContext(r.Context())

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.Warn("No Authorization header")
			writeContextError(w, "Missing user context", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			logger.Warn("Invalid Authorization header format")
			writeContextError(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		// Parse JWT to extract claims (no validation - Kong already validated)
		token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
		if err != nil {
			logger.Warn("Failed to parse JWT", "error", err)
			writeContextError(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			logger.Warn("Invalid JWT claims type")
			writeContextError(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		// Create user context from JWT claims
		user := &UserContext{
			UserID: claims.UserID,
			Email:  claims.Email,
		}

		logger.Debug("User extracted from JWT token", "user_id", user.UserID)

		// Add to request context
		ctx := context.WithValue(r.Context(), UserCtxKey, user)
		ctx = logging.WithUserID(ctx, user.UserID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromGatewayContext retrieves user from context
func GetUserFromGatewayContext(ctx context.Context) (*UserContext, bool) {
	user, ok := ctx.Value(UserCtxKey).(*UserContext)
	return user, ok
}

func writeContextError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.ErrorResponse{
		Error:   "authentication_failed",
		Code:    code,
		Message: message,
	})
}
