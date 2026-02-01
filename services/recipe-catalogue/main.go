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
	ingredientRepo := repository.NewIngredientRepository(db)

	recipeService := service.NewRecipeService(recipeRepo, categoryRepo, ingredientRepo)
	ingredientService := service.NewIngredientService(ingredientRepo, recipeRepo)
	groceryService := service.NewGroceryService(ingredientRepo, recipeRepo)

	recipeHandler := handlers.NewRecipeHandler(recipeService)
	ingredientHandler := handlers.NewIngredientHandler(ingredientService)
	groceryHandler := handlers.NewGroceryHandler(groceryService)

	// Routes with logging middleware
	router := mux.NewRouter()
	router.Use(middleware.LoggingMiddleware("recipe-catalogue-service"))

	// Public routes - Recipes
	router.HandleFunc("/recipes", recipeHandler.GetAllRecipes).Methods("GET")
	router.HandleFunc("/recipes/{id:[0-9]+}", recipeHandler.GetRecipeByID).Methods("GET")
	router.HandleFunc("/categories", recipeHandler.GetAllCategories).Methods("GET")
	router.HandleFunc("/categories/{id:[0-9]+}/recipes", recipeHandler.GetRecipesByCategory).Methods("GET")

	// Public routes - Ingredients
	router.HandleFunc("/ingredients", ingredientHandler.GetAllIngredients).Methods("GET")
	router.HandleFunc("/ingredients/{id:[0-9]+}", ingredientHandler.GetIngredientByID).Methods("GET")
	router.HandleFunc("/ingredients/{id:[0-9]+}/recipes", ingredientHandler.GetRecipesUsingIngredient).Methods("GET")

	// Public routes - Recipe ingredients (read-only)
	router.HandleFunc("/recipes/{id:[0-9]+}/ingredients", ingredientHandler.GetRecipeIngredients).Methods("GET")

	// Health check
	router.HandleFunc("/health", healthCheck).Methods("GET")

	// Protected routes - Recipes
	protected := router.PathPrefix("").Subrouter()
	protected.Use(middleware.ExtractUserFromGatewayHeaders)

	// Recipe management
	protected.HandleFunc("/recipes", recipeHandler.CreateRecipe).Methods("POST")
	protected.HandleFunc("/recipes/{id:[0-9]+}", recipeHandler.UpdateRecipe).Methods("PUT")
	protected.HandleFunc("/recipes/{id:[0-9]+}", recipeHandler.DeleteRecipe).Methods("DELETE")

	// Ingredient management
	protected.HandleFunc("/ingredients", ingredientHandler.CreateIngredient).Methods("POST")
	protected.HandleFunc("/ingredients/{id:[0-9]+}", ingredientHandler.UpdateIngredient).Methods("PUT")
	protected.HandleFunc("/ingredients/{id:[0-9]+}", ingredientHandler.DeleteIngredient).Methods("DELETE")

	// Recipe-ingredient relationships
	protected.HandleFunc("/recipes/{id:[0-9]+}/ingredients", ingredientHandler.AddRecipeIngredient).Methods("POST")
	protected.HandleFunc("/recipes/{id:[0-9]+}/ingredients", ingredientHandler.SetRecipeIngredients).Methods("PUT")
	protected.HandleFunc("/recipes/{recipeId:[0-9]+}/ingredients/{ingredientId:[0-9]+}", ingredientHandler.UpdateRecipeIngredient).Methods("PUT")
	protected.HandleFunc("/recipes/{recipeId:[0-9]+}/ingredients/{ingredientId:[0-9]+}", ingredientHandler.RemoveRecipeIngredient).Methods("DELETE")

	// Grocery list generation
	protected.HandleFunc("/grocery-list", groceryHandler.GenerateGroceryList).Methods("POST")

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
	//logging.WithContext(r.Context()).Debug("Health check requested")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "service": "recipe-catalogue"}`))
}
