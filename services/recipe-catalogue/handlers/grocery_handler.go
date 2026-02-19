package handlers

import (
	"encoding/json"
	"meal-prep/services/recipe-catalogue/domain"
	"meal-prep/services/recipe-catalogue/service"
	"meal-prep/shared/middleware"
	"meal-prep/shared/models"
	"net/http"
)

type GroceryHandler struct {
	groceryService service.GroceryService
}

func NewGroceryHandler(groceryService service.GroceryService) *GroceryHandler {
	return &GroceryHandler{groceryService: groceryService}
}

func (h *GroceryHandler) GenerateGroceryList(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromGatewayContext(r.Context())
	if !ok {
		models.WriteErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	_ = user

	var req models.GroceryListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.WriteErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req.RecipeIDs) == 0 {
		models.WriteErrorResponse(w, "At least one recipe ID is required", http.StatusBadRequest)
		return
	}

	groceryList, err := h.groceryService.GenerateGroceryList(req)
	if err != nil {
		switch err {
		case domain.ErrRecipeNotFound:
			models.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			models.WriteErrorResponse(w, "Failed to generate grocery list", http.StatusInternalServerError)
		}
		return
	}

	models.WriteSuccessResponse(w, groceryList, http.StatusOK)
}
