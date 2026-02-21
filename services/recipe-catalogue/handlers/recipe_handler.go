package handlers

import (
	"encoding/json"
	"meal-prep/services/recipe-catalogue/domain"
	"net/http"
	"strconv"
	"strings"

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
	includeIngredients := r.URL.Query().Get("include_ingredients") == "true"

	if includeIngredients {
		recipes, err := h.recipeService.GetAllRecipesWithIngredients()
		if err != nil {
			models.WriteErrorResponse(w, "Failed to fetch recipes with ingredients", http.StatusInternalServerError)
			return
		}
		models.WriteSuccessResponse(w, recipes, http.StatusOK)
		return
	}

	recipes, err := h.recipeService.GetAllRecipes()
	if err != nil {
		models.WriteErrorResponse(w, "Failed to fetch recipes", http.StatusInternalServerError)
		return
	}

	models.WriteSuccessResponse(w, recipes, http.StatusOK)
}

func (h *RecipeHandler) GetRecipeByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		models.WriteErrorResponse(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	includeIngredients := r.URL.Query().Get("include_ingredients") == "true"

	if includeIngredients {
		recipe, err := h.recipeService.GetRecipeByIDWithIngredients(id)
		if err != nil {
			switch err {
			case domain.ErrRecipeNotFound:
				models.WriteErrorResponse(w, err.Error(), http.StatusNotFound)
			default:
				models.WriteErrorResponse(w, "Failed to fetch recipe with ingredients", http.StatusInternalServerError)
			}
			return
		}
		models.WriteSuccessResponse(w, recipe, http.StatusOK)
		return
	}

	recipe, err := h.recipeService.GetRecipeByID(id)
	if err != nil {
		switch err {
		case domain.ErrRecipeNotFound:
			models.WriteErrorResponse(w, err.Error(), http.StatusNotFound)
		default:
			models.WriteErrorResponse(w, "Failed to fetch recipe", http.StatusInternalServerError)
		}
		return
	}

	models.WriteSuccessResponse(w, recipe, http.StatusOK)
}

func (h *RecipeHandler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.recipeService.GetAllCategories()
	if err != nil {
		models.WriteErrorResponse(w, "Failed to fetch categories", http.StatusInternalServerError)
		return
	}

	models.WriteSuccessResponse(w, categories, http.StatusOK)
}

func (h *RecipeHandler) GetRecipesByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["id"])
	if err != nil {
		models.WriteErrorResponse(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	includeIngredients := r.URL.Query().Get("include_ingredients") == "true"

	if includeIngredients {
		recipes, err := h.recipeService.GetRecipesByCategoryWithIngredients(categoryID)
		if err != nil {
			switch err {
			case domain.ErrCategoryNotFound:
				models.WriteErrorResponse(w, err.Error(), http.StatusNotFound)
			case domain.ErrInvalidCategory:
				models.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
			default:
				models.WriteErrorResponse(w, "Failed to fetch recipes with ingredients", http.StatusInternalServerError)
			}
			return
		}
		models.WriteSuccessResponse(w, recipes, http.StatusOK)
		return
	}

	recipes, err := h.recipeService.GetRecipesByCategory(categoryID)
	if err != nil {
		switch err {
		case domain.ErrCategoryNotFound:
			models.WriteErrorResponse(w, err.Error(), http.StatusNotFound)
		case domain.ErrInvalidCategory:
			models.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			models.WriteErrorResponse(w, "Failed to fetch recipes", http.StatusInternalServerError)
		}
		return
	}

	models.WriteSuccessResponse(w, recipes, http.StatusOK)
}

func (h *RecipeHandler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromGatewayContext(r.Context())
	if !ok {
		models.WriteErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	var req models.CreateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.WriteErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	recipe, err := h.recipeService.CreateRecipe(user.UserID, req)
	if err != nil {
		switch err {
		case domain.ErrRecipeNameRequired:
			models.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		case domain.ErrCategoryNotFound:
			models.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		case domain.ErrInvalidCategory:
			models.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			models.WriteErrorResponse(w, "Failed to create recipe", http.StatusInternalServerError)
		}
		return
	}

	models.WriteSuccessResponse(w, recipe, http.StatusCreated)
}

func (h *RecipeHandler) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromGatewayContext(r.Context())
	if !ok {
		models.WriteErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		models.WriteErrorResponse(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateRecipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.WriteErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	recipe, err := h.recipeService.UpdateRecipe(user.UserID, id, req)
	if err != nil {
		switch err {
		case domain.ErrRecipeNameRequired:
			models.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		case domain.ErrRecipeNotFound:
			models.WriteErrorResponse(w, err.Error(), http.StatusNotFound)
		case domain.ErrForbidden:
			models.WriteErrorResponse(w, err.Error(), http.StatusForbidden)
		case domain.ErrCategoryNotFound:
			models.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		case domain.ErrInvalidCategory:
			models.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			models.WriteErrorResponse(w, "Failed to update recipe", http.StatusInternalServerError)
		}
		return
	}

	models.WriteSuccessResponse(w, recipe, http.StatusOK)
}

func (h *RecipeHandler) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromGatewayContext(r.Context())
	if !ok {
		models.WriteErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		models.WriteErrorResponse(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	err = h.recipeService.DeleteRecipe(user.UserID, id)
	if err != nil {
		switch err {
		case domain.ErrRecipeNotFound:
			models.WriteErrorResponse(w, err.Error(), http.StatusNotFound)
		case domain.ErrForbidden:
			models.WriteErrorResponse(w, err.Error(), http.StatusForbidden)
		default:
			models.WriteErrorResponse(w, "Failed to delete recipe", http.StatusInternalServerError)
		}
		return
	}

	models.WriteSuccessResponse(w, map[string]string{"message": "Recipe deleted successfully"}, http.StatusNoContent)
}

func (h *RecipeHandler) SearchRecipesByIngredients(w http.ResponseWriter, r *http.Request) {
	// Parse ingredient IDs from query parameters
	ingredientIDsParam := r.URL.Query().Get("ingredient_ids")
	if ingredientIDsParam == "" {
		models.WriteErrorResponse(w, "ingredient_ids query parameter is required", http.StatusBadRequest)
		return
	}

	// Parse comma-separated ingredient IDs
	ingredientIDs, err := h.parseIngredientIDs(ingredientIDsParam)
	if err != nil {
		models.WriteErrorResponse(w, "Invalid ingredient IDs format. Use comma-separated integers", http.StatusBadRequest)
		return
	}

	if len(ingredientIDs) == 0 {
		models.WriteErrorResponse(w, "At least one ingredient ID is required", http.StatusBadRequest)
		return
	}

	// Check if client wants recipes with ingredients included
	includeIngredients := r.URL.Query().Get("include_ingredients") == "true"

	if includeIngredients {
		recipes, err := h.recipeService.SearchRecipesByIngredientsWithIngredients(ingredientIDs)
		if err != nil {
			models.WriteErrorResponse(w, "Failed to search recipes with ingredients", http.StatusInternalServerError)
			return
		}
		models.WriteSuccessResponse(w, recipes, http.StatusOK)
		return
	}

	recipes, err := h.recipeService.SearchRecipesByIngredients(ingredientIDs)
	if err != nil {
		models.WriteErrorResponse(w, "Failed to search recipes", http.StatusInternalServerError)
		return
	}

	models.WriteSuccessResponse(w, recipes, http.StatusOK)
}

func (h *RecipeHandler) parseIngredientIDs(ingredientIDsParam string) ([]int, error) {
	if ingredientIDsParam == "" {
		return nil, nil
	}

	idStrings := strings.Split(ingredientIDsParam, ",")
	ingredientIDs := make([]int, 0, len(idStrings))

	for _, idStr := range idStrings {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}

		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			return nil, err
		}
		ingredientIDs = append(ingredientIDs, id)
	}

	return ingredientIDs, nil
}

