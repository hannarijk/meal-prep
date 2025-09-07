package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"meal-prep/services/recipe-catalogue/service"
	"meal-prep/shared/middleware"
	"meal-prep/shared/models"
	"net/http"
	"strconv"
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
		writeErrorResponse(w, "Failed to fetch ingredients", http.StatusInternalServerError)
		return
	}

	writeSuccessResponse(w, ingredients, http.StatusOK)
}

func (h *IngredientHandler) GetIngredientByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeErrorResponse(w, "Invalid ingredient ID", http.StatusBadRequest)
		return
	}

	ingredient, err := h.ingredientService.GetIngredientByID(id)
	if err != nil {
		switch err {
		case service.ErrIngredientNotFound:
			writeErrorResponse(w, err.Error(), http.StatusNotFound)
		default:
			writeErrorResponse(w, "Failed to fetch ingredient", http.StatusInternalServerError)
		}
		return
	}

	writeSuccessResponse(w, ingredient, http.StatusOK)
}

func (h *IngredientHandler) CreateIngredient(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	_ = user // We have the authenticated user if needed for audit logs, etc.

	var req models.CreateIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ingredient, err := h.ingredientService.CreateIngredient(req)
	if err != nil {
		switch err {
		case service.ErrIngredientNameRequired:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		case service.ErrIngredientExists:
			writeErrorResponse(w, err.Error(), http.StatusConflict)
		default:
			writeErrorResponse(w, "Failed to create ingredient", http.StatusInternalServerError)
		}
		return
	}

	writeSuccessResponse(w, ingredient, http.StatusCreated)
}

func (h *IngredientHandler) UpdateIngredient(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	_ = user

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeErrorResponse(w, "Invalid ingredient ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ingredient, err := h.ingredientService.UpdateIngredient(id, req)
	if err != nil {
		switch err {
		case service.ErrIngredientNotFound:
			writeErrorResponse(w, err.Error(), http.StatusNotFound)
		case service.ErrIngredientNameRequired:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			writeErrorResponse(w, "Failed to update ingredient", http.StatusInternalServerError)
		}
		return
	}

	writeSuccessResponse(w, ingredient, http.StatusOK)
}

func (h *IngredientHandler) DeleteIngredient(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	_ = user

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeErrorResponse(w, "Invalid ingredient ID", http.StatusBadRequest)
		return
	}

	err = h.ingredientService.DeleteIngredient(id)
	if err != nil {
		switch err {
		case service.ErrIngredientNotFound:
			writeErrorResponse(w, err.Error(), http.StatusNotFound)
		case service.ErrCannotDeleteIngredient:
			writeErrorResponse(w, err.Error(), http.StatusConflict)
		default:
			writeErrorResponse(w, "Failed to delete ingredient", http.StatusInternalServerError)
		}
		return
	}

	writeSuccessResponse(w, map[string]string{"message": "Ingredient deleted successfully"}, http.StatusNoContent)
}

func (h *IngredientHandler) GetRecipesUsingIngredient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeErrorResponse(w, "Invalid ingredient ID", http.StatusBadRequest)
		return
	}

	recipes, err := h.ingredientService.GetRecipesUsingIngredient(id)
	if err != nil {
		switch err {
		case service.ErrIngredientNotFound:
			writeErrorResponse(w, err.Error(), http.StatusNotFound)
		default:
			writeErrorResponse(w, "Failed to fetch recipes", http.StatusInternalServerError)
		}
		return
	}

	writeSuccessResponse(w, recipes, http.StatusOK)
}

func (h *IngredientHandler) GetRecipeIngredients(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recipeID, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeErrorResponse(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	ingredients, err := h.ingredientService.GetRecipeIngredients(recipeID)
	if err != nil {
		switch err {
		case service.ErrRecipeNotFound:
			writeErrorResponse(w, err.Error(), http.StatusNotFound)
		default:
			writeErrorResponse(w, "Failed to fetch recipe ingredients", http.StatusInternalServerError)
		}
		return
	}

	writeSuccessResponse(w, ingredients, http.StatusOK)
}

func (h *IngredientHandler) AddRecipeIngredient(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	_ = user

	vars := mux.Vars(r)
	recipeID, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeErrorResponse(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	var req models.AddRecipeIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ingredient, err := h.ingredientService.AddRecipeIngredient(recipeID, req)
	if err != nil {
		switch err {
		case service.ErrRecipeNotFound:
			writeErrorResponse(w, err.Error(), http.StatusNotFound)
		case service.ErrIngredientNotFound:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		case service.ErrInvalidQuantity:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		case service.ErrInvalidUnit:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			writeErrorResponse(w, "Failed to add ingredient to recipe", http.StatusInternalServerError)
		}
		return
	}

	writeSuccessResponse(w, ingredient, http.StatusCreated)
}

func (h *IngredientHandler) UpdateRecipeIngredient(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	_ = user

	vars := mux.Vars(r)
	recipeID, err := strconv.Atoi(vars["recipeId"])
	if err != nil {
		writeErrorResponse(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	ingredientID, err := strconv.Atoi(vars["ingredientId"])
	if err != nil {
		writeErrorResponse(w, "Invalid ingredient ID", http.StatusBadRequest)
		return
	}

	var req models.AddRecipeIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ingredient, err := h.ingredientService.UpdateRecipeIngredient(recipeID, ingredientID, req)
	if err != nil {
		switch err {
		case service.ErrRecipeNotFound:
			writeErrorResponse(w, err.Error(), http.StatusNotFound)
		case service.ErrIngredientNotFound:
			writeErrorResponse(w, err.Error(), http.StatusNotFound)
		case service.ErrInvalidQuantity:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		case service.ErrInvalidUnit:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			writeErrorResponse(w, "Failed to update recipe ingredient", http.StatusInternalServerError)
		}
		return
	}

	writeSuccessResponse(w, ingredient, http.StatusOK)
}

func (h *IngredientHandler) RemoveRecipeIngredient(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	_ = user

	vars := mux.Vars(r)
	recipeID, err := strconv.Atoi(vars["recipeId"])
	if err != nil {
		writeErrorResponse(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	ingredientID, err := strconv.Atoi(vars["ingredientId"])
	if err != nil {
		writeErrorResponse(w, "Invalid ingredient ID", http.StatusBadRequest)
		return
	}

	err = h.ingredientService.RemoveRecipeIngredient(recipeID, ingredientID)
	if err != nil {
		switch err {
		case service.ErrRecipeNotFound:
			writeErrorResponse(w, err.Error(), http.StatusNotFound)
		case service.ErrIngredientNotFound:
			writeErrorResponse(w, err.Error(), http.StatusNotFound)
		default:
			writeErrorResponse(w, "Failed to remove ingredient from recipe", http.StatusInternalServerError)
		}
		return
	}

	writeSuccessResponse(w, map[string]string{"message": "Ingredient removed from recipe successfully"}, http.StatusOK)
}

func (h *IngredientHandler) SetRecipeIngredients(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	_ = user

	vars := mux.Vars(r)
	recipeID, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeErrorResponse(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	var req []models.AddRecipeIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = h.ingredientService.SetRecipeIngredients(recipeID, req)
	if err != nil {
		switch err {
		case service.ErrRecipeNotFound:
			writeErrorResponse(w, err.Error(), http.StatusNotFound)
		case service.ErrIngredientNotFound:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		case service.ErrInvalidQuantity:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		case service.ErrInvalidUnit:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			writeErrorResponse(w, "Failed to update recipe ingredients", http.StatusInternalServerError)
		}
		return
	}

	// Get updated ingredients to return
	ingredients, err := h.ingredientService.GetRecipeIngredients(recipeID)
	if err != nil {
		writeErrorResponse(w, "Failed to fetch updated ingredients", http.StatusInternalServerError)
		return
	}

	writeSuccessResponse(w, ingredients, http.StatusOK)
}

func (h *IngredientHandler) GenerateShoppingList(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	_ = user

	var req models.ShoppingListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req.RecipeIDs) == 0 {
		writeErrorResponse(w, "At least one recipe ID is required", http.StatusBadRequest)
		return
	}

	shoppingList, err := h.ingredientService.GenerateShoppingList(req)
	if err != nil {
		switch err {
		case service.ErrRecipeNotFound:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			writeErrorResponse(w, "Failed to generate shopping list", http.StatusInternalServerError)
		}
		return
	}

	writeSuccessResponse(w, shoppingList, http.StatusOK)
}
