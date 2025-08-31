package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"meal-prep/services/recipe-catalogue/service"
	"meal-prep/shared/middleware"
	"meal-prep/shared/models"

	"github.com/gorilla/mux"
)

type RecipeHandler struct {
	recipeService service.RecipeService
}

func NewRecipeHandler(recipeService service.RecipeService) *RecipeHandler {
	return &RecipeHandler{recipeService: recipeService}
}

func (h *RecipeHandler) GetAllRecipes(w http.ResponseWriter, r *http.Request) {
	recipes, err := h.recipeService.GetAllRecipes()
	if err != nil {
		writeErrorResponse(w, "Failed to fetch recipes", http.StatusInternalServerError)
		return
	}

	writeSuccessResponse(w, recipes, http.StatusOK)
}

func (h *RecipeHandler) GetRecipeByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeErrorResponse(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	recipe, err := h.recipeService.GetRecipeByID(id)
	if err != nil {
		switch err {
		case service.ErrRecipeNotFound:
			writeErrorResponse(w, err.Error(), http.StatusNotFound)
		default:
			writeErrorResponse(w, "Failed to fetch recipe", http.StatusInternalServerError)
		}
		return
	}

	writeSuccessResponse(w, recipe, http.StatusOK)
}

func (h *RecipeHandler) GetRecipesByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeErrorResponse(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	recipes, err := h.recipeService.GetRecipesByCategory(categoryID)
	if err != nil {
		switch err {
		case service.ErrCategoryNotFound:
			writeErrorResponse(w, err.Error(), http.StatusNotFound)
		case service.ErrInvalidCategory:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			writeErrorResponse(w, "Failed to fetch recipes", http.StatusInternalServerError)
		}
		return
	}

	writeSuccessResponse(w, recipes, http.StatusOK)
}

func (h *RecipeHandler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	_ = user // We have the authenticated user if needed for audit logs, etc.

	var req models.CreateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	recipe, err := h.recipeService.CreateRecipe(req)
	if err != nil {
		switch err {
		case service.ErrRecipeNameRequired:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		case service.ErrCategoryNotFound:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		case service.ErrInvalidCategory:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			writeErrorResponse(w, "Failed to create recipe", http.StatusInternalServerError)
		}
		return
	}

	writeSuccessResponse(w, recipe, http.StatusCreated)
}

func (h *RecipeHandler) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	_ = user

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeErrorResponse(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	recipe, err := h.recipeService.UpdateRecipe(id, req)
	if err != nil {
		switch err {
		case service.ErrRecipeNotFound:
			writeErrorResponse(w, err.Error(), http.StatusNotFound)
		case service.ErrCategoryNotFound:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		case service.ErrInvalidCategory:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			writeErrorResponse(w, "Failed to update recipe", http.StatusInternalServerError)
		}
		return
	}

	writeSuccessResponse(w, recipe, http.StatusOK)
}

func (h *RecipeHandler) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	_ = user

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeErrorResponse(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	err = h.recipeService.DeleteRecipe(id)
	if err != nil {
		switch err {
		case service.ErrRecipeNotFound:
			writeErrorResponse(w, err.Error(), http.StatusNotFound)
		default:
			writeErrorResponse(w, "Failed to delete recipe", http.StatusInternalServerError)
		}
		return
	}

	writeSuccessResponse(w, map[string]string{"message": "Recipe deleted successfully"}, http.StatusOK)
}

func (h *RecipeHandler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.recipeService.GetAllCategories()
	if err != nil {
		writeErrorResponse(w, "Failed to fetch categories", http.StatusInternalServerError)
		return
	}

	writeSuccessResponse(w, categories, http.StatusOK)
}

func writeErrorResponse(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.ErrorResponse{
		Error:   "recipe_catalogue_error",
		Code:    code,
		Message: message,
	})
}

func writeSuccessResponse(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
