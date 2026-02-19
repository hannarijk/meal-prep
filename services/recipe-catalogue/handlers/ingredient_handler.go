package handlers

import (
	"encoding/json"
	"meal-prep/services/recipe-catalogue/domain"
	"meal-prep/services/recipe-catalogue/service"
	"meal-prep/shared/middleware"
	"meal-prep/shared/models"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type IngredientHandler struct {
	ingredientService service.IngredientService
}

func NewIngredientHandler(ingredientService service.IngredientService) *IngredientHandler {
	return &IngredientHandler{ingredientService: ingredientService}
}

func (h *IngredientHandler) GetAllIngredients(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	searchQuery := r.URL.Query().Get("search")

	var ingredients []models.Ingredient
	var err error

	if searchQuery != "" {
		ingredients, err = h.ingredientService.SearchIngredients(searchQuery)
	} else if category != "" {
		ingredients, err = h.ingredientService.GetIngredientsByCategory(category)
	} else {
		ingredients, err = h.ingredientService.GetAllIngredients()
	}

	if err != nil {
		models.WriteErrorResponse(w, "Failed to fetch ingredients", http.StatusInternalServerError)
		return
	}

	models.WriteSuccessResponse(w, ingredients, http.StatusOK)
}

func (h *IngredientHandler) GetIngredientByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		models.WriteErrorResponse(w, "Invalid ingredient ID", http.StatusBadRequest)
		return
	}

	ingredient, err := h.ingredientService.GetIngredientByID(id)
	if err != nil {
		switch err {
		case domain.ErrIngredientNotFound:
			models.WriteErrorResponse(w, err.Error(), http.StatusNotFound)
		default:
			models.WriteErrorResponse(w, "Failed to fetch ingredient", http.StatusInternalServerError)
		}
		return
	}

	models.WriteSuccessResponse(w, ingredient, http.StatusOK)
}

func (h *IngredientHandler) CreateIngredient(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromGatewayContext(r.Context())
	if !ok {
		models.WriteErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	_ = user // We have the authenticated user if needed for audit logs, etc.

	var req models.CreateIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.WriteErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ingredient, err := h.ingredientService.CreateIngredient(req)
	if err != nil {
		switch err {
		case domain.ErrIngredientNameRequired:
			models.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		case domain.ErrIngredientExists:
			models.WriteErrorResponse(w, err.Error(), http.StatusConflict)
		default:
			models.WriteErrorResponse(w, "Failed to create ingredient", http.StatusInternalServerError)
		}
		return
	}

	models.WriteSuccessResponse(w, ingredient, http.StatusCreated)
}

func (h *IngredientHandler) UpdateIngredient(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromGatewayContext(r.Context())
	if !ok {
		models.WriteErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	_ = user

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		models.WriteErrorResponse(w, "Invalid ingredient ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.WriteErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ingredient, err := h.ingredientService.UpdateIngredient(id, req)
	if err != nil {
		switch err {
		case domain.ErrIngredientNotFound:
			models.WriteErrorResponse(w, err.Error(), http.StatusNotFound)
		case domain.ErrIngredientNameRequired:
			models.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			models.WriteErrorResponse(w, "Failed to update ingredient", http.StatusInternalServerError)
		}
		return
	}

	models.WriteSuccessResponse(w, ingredient, http.StatusOK)
}

func (h *IngredientHandler) DeleteIngredient(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromGatewayContext(r.Context())
	if !ok {
		models.WriteErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	_ = user

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		models.WriteErrorResponse(w, "Invalid ingredient ID", http.StatusBadRequest)
		return
	}

	err = h.ingredientService.DeleteIngredient(id)
	if err != nil {
		switch err {
		case domain.ErrIngredientNotFound:
			models.WriteErrorResponse(w, err.Error(), http.StatusNotFound)
		case domain.ErrCannotDeleteIngredient:
			models.WriteErrorResponse(w, err.Error(), http.StatusConflict)
		default:
			models.WriteErrorResponse(w, "Failed to delete ingredient", http.StatusInternalServerError)
		}
		return
	}

	models.WriteSuccessResponse(w, map[string]string{"message": "Ingredient deleted successfully"}, http.StatusNoContent)
}

func (h *IngredientHandler) GetRecipesUsingIngredient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		models.WriteErrorResponse(w, "Invalid ingredient ID", http.StatusBadRequest)
		return
	}

	recipes, err := h.ingredientService.GetRecipesUsingIngredient(id)
	if err != nil {
		switch err {
		case domain.ErrIngredientNotFound:
			models.WriteErrorResponse(w, err.Error(), http.StatusNotFound)
		default:
			models.WriteErrorResponse(w, "Failed to fetch recipes", http.StatusInternalServerError)
		}
		return
	}

	models.WriteSuccessResponse(w, recipes, http.StatusOK)
}

func (h *IngredientHandler) GetRecipeIngredients(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recipeID, err := strconv.Atoi(vars["id"])
	if err != nil {
		models.WriteErrorResponse(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	ingredients, err := h.ingredientService.GetRecipeIngredients(recipeID)
	if err != nil {
		switch err {
		case domain.ErrRecipeNotFound:
			models.WriteErrorResponse(w, err.Error(), http.StatusNotFound)
		default:
			models.WriteErrorResponse(w, "Failed to fetch recipe ingredients", http.StatusInternalServerError)
		}
		return
	}

	models.WriteSuccessResponse(w, ingredients, http.StatusOK)
}

func (h *IngredientHandler) AddRecipeIngredient(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromGatewayContext(r.Context())
	if !ok {
		models.WriteErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	_ = user

	vars := mux.Vars(r)
	recipeID, err := strconv.Atoi(vars["id"])
	if err != nil {
		models.WriteErrorResponse(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	var req models.AddRecipeIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.WriteErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ingredient, err := h.ingredientService.AddRecipeIngredient(recipeID, req)
	if err != nil {
		switch err {
		case domain.ErrRecipeNotFound:
			models.WriteErrorResponse(w, err.Error(), http.StatusNotFound)
		case domain.ErrIngredientNotFound:
			models.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		case domain.ErrInvalidQuantity:
			models.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		case domain.ErrInvalidUnit:
			models.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		case domain.ErrRecipeIngredientAlreadyExists:
			models.WriteErrorResponse(w, err.Error(), http.StatusConflict)
		default:
			models.WriteErrorResponse(w, "Failed to add ingredient to recipe", http.StatusInternalServerError)
		}
		return
	}

	models.WriteSuccessResponse(w, ingredient, http.StatusCreated)
}

func (h *IngredientHandler) UpdateRecipeIngredient(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromGatewayContext(r.Context())
	if !ok {
		models.WriteErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	_ = user

	vars := mux.Vars(r)
	recipeID, err := strconv.Atoi(vars["recipeId"])
	if err != nil {
		models.WriteErrorResponse(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	ingredientID, err := strconv.Atoi(vars["ingredientId"])
	if err != nil {
		models.WriteErrorResponse(w, "Invalid ingredient ID", http.StatusBadRequest)
		return
	}

	var req models.AddRecipeIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.WriteErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ingredient, err := h.ingredientService.UpdateRecipeIngredient(recipeID, ingredientID, req)
	if err != nil {
		switch err {
		case domain.ErrRecipeNotFound:
			models.WriteErrorResponse(w, err.Error(), http.StatusNotFound)
		case domain.ErrIngredientNotFound:
			models.WriteErrorResponse(w, err.Error(), http.StatusNotFound)
		case domain.ErrInvalidQuantity:
			models.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		case domain.ErrInvalidUnit:
			models.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			models.WriteErrorResponse(w, "Failed to update recipe ingredient", http.StatusInternalServerError)
		}
		return
	}

	models.WriteSuccessResponse(w, ingredient, http.StatusOK)
}

func (h *IngredientHandler) RemoveRecipeIngredient(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromGatewayContext(r.Context())
	if !ok {
		models.WriteErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	_ = user

	vars := mux.Vars(r)
	recipeID, err := strconv.Atoi(vars["recipeId"])
	if err != nil {
		models.WriteErrorResponse(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	ingredientID, err := strconv.Atoi(vars["ingredientId"])
	if err != nil {
		models.WriteErrorResponse(w, "Invalid ingredient ID", http.StatusBadRequest)
		return
	}

	err = h.ingredientService.RemoveRecipeIngredient(recipeID, ingredientID)
	if err != nil {
		switch err {
		case domain.ErrRecipeNotFound:
			models.WriteErrorResponse(w, err.Error(), http.StatusNotFound)
		case domain.ErrIngredientNotFound:
			models.WriteErrorResponse(w, err.Error(), http.StatusNotFound)
		default:
			models.WriteErrorResponse(w, "Failed to remove ingredient from recipe", http.StatusInternalServerError)
		}
		return
	}

	models.WriteSuccessResponse(w, map[string]string{"message": "Ingredient removed from recipe successfully"}, http.StatusOK)
}

func (h *IngredientHandler) SetRecipeIngredients(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromGatewayContext(r.Context())
	if !ok {
		models.WriteErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	_ = user

	vars := mux.Vars(r)
	recipeID, err := strconv.Atoi(vars["id"])
	if err != nil {
		models.WriteErrorResponse(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	var req []models.AddRecipeIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.WriteErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = h.ingredientService.SetRecipeIngredients(recipeID, req)
	if err != nil {
		switch err {
		case domain.ErrRecipeNotFound:
			models.WriteErrorResponse(w, err.Error(), http.StatusNotFound)
		case domain.ErrIngredientNotFound:
			models.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		case domain.ErrInvalidQuantity:
			models.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		case domain.ErrInvalidUnit:
			models.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			models.WriteErrorResponse(w, "Failed to update recipe ingredients", http.StatusInternalServerError)
		}
		return
	}

	// Get updated ingredients to return
	ingredients, err := h.ingredientService.GetRecipeIngredients(recipeID)
	if err != nil {
		models.WriteErrorResponse(w, "Failed to fetch updated ingredients", http.StatusInternalServerError)
		return
	}

	models.WriteSuccessResponse(w, ingredients, http.StatusOK)
}
