package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"meal-prep/services/dish-catalogue/service"
	"meal-prep/shared/middleware"
	"meal-prep/shared/models"

	"github.com/gorilla/mux"
)

type DishHandler struct {
	dishService service.DishService
}

func NewDishHandler(dishService service.DishService) *DishHandler {
	return &DishHandler{dishService: dishService}
}

func (h *DishHandler) GetAllDishes(w http.ResponseWriter, r *http.Request) {
	dishes, err := h.dishService.GetAllDishes()
	if err != nil {
		writeErrorResponse(w, "Failed to fetch dishes", http.StatusInternalServerError)
		return
	}

	writeSuccessResponse(w, dishes, http.StatusOK)
}

func (h *DishHandler) GetDishByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeErrorResponse(w, "Invalid dish ID", http.StatusBadRequest)
		return
	}

	dish, err := h.dishService.GetDishByID(id)
	if err != nil {
		switch err {
		case service.ErrDishNotFound:
			writeErrorResponse(w, err.Error(), http.StatusNotFound)
		default:
			writeErrorResponse(w, "Failed to fetch dish", http.StatusInternalServerError)
		}
		return
	}

	writeSuccessResponse(w, dish, http.StatusOK)
}

func (h *DishHandler) GetDishesByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeErrorResponse(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	dishes, err := h.dishService.GetDishesByCategory(categoryID)
	if err != nil {
		switch err {
		case service.ErrCategoryNotFound:
			writeErrorResponse(w, err.Error(), http.StatusNotFound)
		case service.ErrInvalidCategory:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			writeErrorResponse(w, "Failed to fetch dishes", http.StatusInternalServerError)
		}
		return
	}

	writeSuccessResponse(w, dishes, http.StatusOK)
}

func (h *DishHandler) CreateDish(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	_ = user // We have the authenticated user if needed for audit logs, etc.

	var req models.CreateDishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	dish, err := h.dishService.CreateDish(req)
	if err != nil {
		switch err {
		case service.ErrDishNameRequired:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		case service.ErrCategoryNotFound:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		case service.ErrInvalidCategory:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			writeErrorResponse(w, "Failed to create dish", http.StatusInternalServerError)
		}
		return
	}

	writeSuccessResponse(w, dish, http.StatusCreated)
}

func (h *DishHandler) UpdateDish(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	_ = user

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeErrorResponse(w, "Invalid dish ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateDishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	dish, err := h.dishService.UpdateDish(id, req)
	if err != nil {
		switch err {
		case service.ErrDishNotFound:
			writeErrorResponse(w, err.Error(), http.StatusNotFound)
		case service.ErrCategoryNotFound:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		case service.ErrInvalidCategory:
			writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		default:
			writeErrorResponse(w, "Failed to update dish", http.StatusInternalServerError)
		}
		return
	}

	writeSuccessResponse(w, dish, http.StatusOK)
}

func (h *DishHandler) DeleteDish(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		writeErrorResponse(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	_ = user

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		writeErrorResponse(w, "Invalid dish ID", http.StatusBadRequest)
		return
	}

	err = h.dishService.DeleteDish(id)
	if err != nil {
		switch err {
		case service.ErrDishNotFound:
			writeErrorResponse(w, err.Error(), http.StatusNotFound)
		default:
			writeErrorResponse(w, "Failed to delete dish", http.StatusInternalServerError)
		}
		return
	}

	writeSuccessResponse(w, map[string]string{"message": "Dish deleted successfully"}, http.StatusOK)
}

func (h *DishHandler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.dishService.GetAllCategories()
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
		Error:   "dish_catalogue_error",
		Code:    code,
		Message: message,
	})
}

func writeSuccessResponse(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
