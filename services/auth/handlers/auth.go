package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"meal-prep/services/auth/service"
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
		writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.Email == "" || req.Password == "" {
		writeErrorResponse(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	response, err := h.authService.Register(req.Email, req.Password)
	if err != nil {
		switch err {
		case service.ErrUserExists:
			writeErrorResponse(w, err.Error(), http.StatusConflict)
		case service.ErrWeakPassword:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			writeErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	writeSuccessResponse(w, response, http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		writeErrorResponse(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	response, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			writeErrorResponse(w, err.Error(), http.StatusUnauthorized)
		default:
			writeErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	writeSuccessResponse(w, response, http.StatusOK)
}

func writeErrorResponse(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.ErrorResponse{
		Error:   "auth_error",
		Code:    code,
		Message: message,
	})
}

func writeSuccessResponse(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
