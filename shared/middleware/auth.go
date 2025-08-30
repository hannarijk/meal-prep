package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"meal-prep/shared/logging"
	"meal-prep/shared/models"
	"meal-prep/shared/utils"
)

type contextKey string

const UserContextKey contextKey = "user"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logging.WithContext(r.Context())

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.Warn("Missing authorization header")
			writeErrorResponse(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			logger.Warn("Invalid authorization header format")
			writeErrorResponse(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		claims, err := utils.ValidateJWT(bearerToken[1])
		if err != nil {
			logger.Warn("Invalid JWT token", "error", err)
			writeErrorResponse(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		logger.Debug("User authenticated successfully", "user_id", claims.UserID, "email", claims.Email)

		// Add user info to request context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		ctx = logging.WithUserID(ctx, claims.UserID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(ctx context.Context) (*utils.Claims, bool) {
	user, ok := ctx.Value(UserContextKey).(*utils.Claims)
	return user, ok
}

func writeErrorResponse(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.ErrorResponse{
		Error:   "authentication_failed",
		Code:    code,
		Message: message,
	})
}
