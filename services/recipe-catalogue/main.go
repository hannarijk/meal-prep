package main

import (
	"net/http"
	"os"

	"meal-prep/services/recipe-catalogue/handlers"
	"meal-prep/services/recipe-catalogue/repository"
	"meal-prep/services/recipe-catalogue/service"
	"meal-prep/shared/database"
	"meal-prep/shared/logging"
	"meal-prep/shared/middleware"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Initialize logging first
	logging.Init("recipe-catalogue-service")

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
	recipeRepo := repository.NewRecipeRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	recipeService := service.NewRecipeService(recipeRepo, categoryRepo)
	recipeHandler := handlers.NewRecipeHandler(recipeService)

	// Routes with logging middleware
	router := mux.NewRouter()
	router.Use(middleware.LoggingMiddleware("recipe-catalogue-service"))

	// Public routes
	router.HandleFunc("/recipes", recipeHandler.GetAllRecipes).Methods("GET")
	router.HandleFunc("/recipes/{id}", recipeHandler.GetRecipeByID).Methods("GET")
	router.HandleFunc("/categories", recipeHandler.GetAllCategories).Methods("GET")
	router.HandleFunc("/categories/{id}/recipes", recipeHandler.GetRecipesByCategory).Methods("GET")
	router.HandleFunc("/health", healthCheck).Methods("GET")

	// Protected routes
	protected := router.PathPrefix("").Subrouter()
	protected.Use(middleware.AuthMiddleware)
	protected.HandleFunc("/recipes", recipeHandler.CreateRecipe).Methods("POST")
	protected.HandleFunc("/recipes/{id}", recipeHandler.UpdateRecipe).Methods("PUT")
	protected.HandleFunc("/recipes/{id}", recipeHandler.DeleteRecipe).Methods("DELETE")

	port := os.Getenv("RECIPE_CATALOGUE_PORT")
	if port == "" {
		port = "8002"
	}

	logging.Logger.Info("Starting recipe catalogue service", "port", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		logging.Logger.Error("Server failed to start", "error", err)
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	logging.WithContext(r.Context()).Debug("Health check requested")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "service": "recipe-catalogue"}`))
}
