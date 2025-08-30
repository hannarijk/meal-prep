package main

import (
	"net/http"
	"os"

	"meal-prep/services/dish-catalogue/handlers"
	"meal-prep/services/dish-catalogue/repository"
	"meal-prep/services/dish-catalogue/service"
	"meal-prep/shared/database"
	"meal-prep/shared/logging"
	"meal-prep/shared/middleware"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Initialize logging first
	logging.Init("dish-catalogue-service")

	if err := godotenv.Load(); err != nil {
		logging.Logger.Debug("No .env file found, using system environment variables")
	}

	// Database connection
	db, err := database.NewPostgresConnection()
	if err != nil {
		logging.Logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	logging.Logger.Info("Database connected successfully")

	// Dependency injection chain
	dishRepo := repository.NewDishRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	dishService := service.NewDishService(dishRepo, categoryRepo)
	dishHandler := handlers.NewDishHandler(dishService)

	// Routes with logging middleware
	router := mux.NewRouter()
	router.Use(middleware.LoggingMiddleware("dish-catalogue-service"))

	// Public routes
	router.HandleFunc("/dishes", dishHandler.GetAllDishes).Methods("GET")
	router.HandleFunc("/dishes/{id}", dishHandler.GetDishByID).Methods("GET")
	router.HandleFunc("/categories", dishHandler.GetAllCategories).Methods("GET")
	router.HandleFunc("/categories/{id}/dishes", dishHandler.GetDishesByCategory).Methods("GET")
	router.HandleFunc("/health", healthCheck).Methods("GET")

	// Protected routes
	protected := router.PathPrefix("").Subrouter()
	protected.Use(middleware.AuthMiddleware)
	protected.HandleFunc("/dishes", dishHandler.CreateDish).Methods("POST")
	protected.HandleFunc("/dishes/{id}", dishHandler.UpdateDish).Methods("PUT")
	protected.HandleFunc("/dishes/{id}", dishHandler.DeleteDish).Methods("DELETE")

	port := os.Getenv("DISH_CATALOGUE_PORT")
	if port == "" {
		port = "8002"
	}

	logging.Logger.Info("Starting dish catalogue service", "port", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		logging.Logger.Error("Server failed to start", "error", err)
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	logging.WithContext(r.Context()).Debug("Health check requested")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "service": "dish-catalogue"}`))
}
