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
	user, ok := middleware.GetUserFromGatewayContext(r.Context())
	if !ok {
		models.WriteErrorResponse(w, "Authentication required", http.StatusUnauthorized)
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
			models.WriteErrorResponse(w, "Please set your food preferences first", http.StatusBadRequest)
		case service.ErrInvalidAlgorithm:
			models.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		case service.ErrInvalidLimit:
			models.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			models.WriteErrorResponse(w, "Failed to generate recommendations", http.StatusInternalServerError)
		}
		return
	}

	models.WriteSuccessResponse(w, recommendations, http.StatusOK)
}

func (h *RecommendationHandler) GetUserPreferences(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromGatewayContext(r.Context())
	if !ok {
		models.WriteErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	preferences, err := h.recService.GetUserPreferences(user.UserID)
	if err != nil {
		models.WriteErrorResponse(w, "Failed to fetch preferences", http.StatusInternalServerError)
		return
	}

	models.WriteSuccessResponse(w, preferences, http.StatusOK)
}

func (h *RecommendationHandler) UpdateUserPreferences(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromGatewayContext(r.Context())
	if !ok {
		log.Println("ERROR: No user in context")
		models.WriteErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	log.Printf("INFO: User %d updating preferences", user.UserID)

	var req models.UpdatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR: JSON decode failed: %v", err)
		models.WriteErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("INFO: Preferences request: %+v", req)

	preferences, err := h.recService.UpdateUserPreferences(user.UserID, req)
	if err != nil {
		log.Printf("ERROR: UpdateUserPreferences failed: %v", err)
		models.WriteErrorResponse(w, "Failed to update preferences", http.StatusInternalServerError)
		return
	}

	log.Printf("INFO: Preferences updated successfully: %+v", preferences)
	models.WriteSuccessResponse(w, preferences, http.StatusOK)
}

func (h *RecommendationHandler) LogCooking(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromGatewayContext(r.Context())
	if !ok {
		models.WriteErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	var req models.LogCookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.WriteErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := h.recService.LogCooking(user.UserID, req)
	if err != nil {
		switch err {
		case service.ErrRecipeNotFound:
			models.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		case service.ErrInvalidRating:
			models.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			models.WriteErrorResponse(w, "Failed to log cooking", http.StatusInternalServerError)
		}
		return
	}

	models.WriteSuccessResponse(w, map[string]string{"message": "Cooking logged successfully"}, http.StatusCreated)
}

func (h *RecommendationHandler) GetCookingHistory(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromGatewayContext(r.Context())
	if !ok {
		models.WriteErrorResponse(w, "Authentication required", http.StatusUnauthorized)
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
		models.WriteErrorResponse(w, "Failed to fetch cooking history", http.StatusInternalServerError)
		return
	}

	models.WriteSuccessResponse(w, history, http.StatusOK)
}

