package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"meal-prep/services/recommendations/service"
	"meal-prep/shared/middleware"
	"meal-prep/shared/models"
)

type RecommendationHandler struct {
	recService service.RecommendationService
}

func NewRecommendationHandler(recService service.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{recService: recService}
}

func (h *RecommendationHandler) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	var req models.RecommendationRequest
	if algorithm := r.URL.Query().Get("algorithm"); algorithm != "" {
		req.Algorithm = algorithm
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = limit
		}
	}

	recommendations, err := h.recService.GetRecommendations(user.UserID, req)
	if err != nil {
		switch err {
		case service.ErrPreferencesNotSet:
			writeErrorResponse(w, "Please set your food preferences first", http.StatusBadRequest)
		case service.ErrInvalidAlgorithm:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		case service.ErrInvalidLimit:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			writeErrorResponse(w, "Failed to generate recommendations", http.StatusInternalServerError)
		}
		return
	}

	writeSuccessResponse(w, recommendations, http.StatusOK)
}

func (h *RecommendationHandler) GetUserPreferences(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	preferences, err := h.recService.GetUserPreferences(user.UserID)
	if err != nil {
		writeErrorResponse(w, "Failed to fetch preferences", http.StatusInternalServerError)
		return
	}

	writeSuccessResponse(w, preferences, http.StatusOK)
}

func (h *RecommendationHandler) UpdateUserPreferences(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		log.Println("ERROR: No user in context")
		writeErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	log.Printf("INFO: User %d updating preferences", user.UserID)

	var req models.UpdatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR: JSON decode failed: %v", err)
		writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("INFO: Preferences request: %+v", req)

	preferences, err := h.recService.UpdateUserPreferences(user.UserID, req)
	if err != nil {
		log.Printf("ERROR: UpdateUserPreferences failed: %v", err)
		writeErrorResponse(w, "Failed to update preferences", http.StatusInternalServerError)
		return
	}

	log.Printf("INFO: Preferences updated successfully: %+v", preferences)
	writeSuccessResponse(w, preferences, http.StatusOK)
}

func (h *RecommendationHandler) LogCooking(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	var req models.LogCookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := h.recService.LogCooking(user.UserID, req)
	if err != nil {
		switch err {
		case service.ErrRecipeNotFound:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		case service.ErrInvalidRating:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			writeErrorResponse(w, "Failed to log cooking", http.StatusInternalServerError)
		}
		return
	}

	writeSuccessResponse(w, map[string]string{"message": "Cooking logged successfully"}, http.StatusCreated)
}

func (h *RecommendationHandler) GetCookingHistory(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	limit := 20 // default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	history, err := h.recService.GetCookingHistory(user.UserID, limit)
	if err != nil {
		writeErrorResponse(w, "Failed to fetch cooking history", http.StatusInternalServerError)
		return
	}

	writeSuccessResponse(w, history, http.StatusOK)
}

func writeErrorResponse(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.ErrorResponse{
		Error:   "recommendations_error",
		Code:    code,
		Message: message,
	})
}

func writeSuccessResponse(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
