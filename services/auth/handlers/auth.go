package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"meal-prep/services/auth/service"
	"meal-prep/shared/middleware"
	"meal-prep/shared/models"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.WriteErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Email == "" || req.Password == "" {
		models.WriteErrorResponse(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	response, err := h.authService.Register(req.Email, req.Password)
	if err != nil {
		switch err {
		case service.ErrUserExists:
			models.WriteErrorResponse(w, err.Error(), http.StatusConflict)
		case service.ErrWeakPassword:
			models.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			models.WriteErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	models.WriteSuccessResponse(w, response, http.StatusCreated)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromGatewayContext(r.Context())
	if !ok {
		models.WriteErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	result, err := h.authService.Me(user.UserID)
	if err != nil {
		models.WriteErrorResponse(w, "User not found", http.StatusNotFound)
		return
	}

	models.WriteSuccessResponse(w, result, http.StatusOK)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.WriteErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		models.WriteErrorResponse(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	response, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			models.WriteErrorResponse(w, err.Error(), http.StatusUnauthorized)
		default:
			models.WriteErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	models.WriteSuccessResponse(w, response, http.StatusOK)
}

