package main

import (
	"net/http"
	"os"

	"meal-prep/services/recommendations/handlers"
	"meal-prep/services/recommendations/repository"
	"meal-prep/services/recommendations/service"
	"meal-prep/shared/database"
	"meal-prep/shared/logging"
	"meal-prep/shared/middleware"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Initialize logging first
	logging.Init("recommendations-service")

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
	recRepo := repository.NewRecommendationRepository(db)
	recService := service.NewRecommendationService(recRepo)
	recHandler := handlers.NewRecommendationHandler(recService)

	// Routes - all require authentication
	router := mux.NewRouter()
	router.Use(middleware.LoggingMiddleware("recommendations-service"))
	router.Use(middleware.ExtractUserFromGatewayHeaders)

	router.HandleFunc("/recommendations", recHandler.GetRecommendations).Methods("GET")
	router.HandleFunc("/preferences", recHandler.GetUserPreferences).Methods("GET")
	router.HandleFunc("/preferences", recHandler.UpdateUserPreferences).Methods("PUT")
	router.HandleFunc("/cooking", recHandler.LogCooking).Methods("POST")
	router.HandleFunc("/cooking/history", recHandler.GetCookingHistory).Methods("GET")

	// Health check without auth
	healthRouter := mux.NewRouter()
	healthRouter.Use(middleware.LoggingMiddleware("recommendations-service"))
	healthRouter.HandleFunc("/health", healthCheck).Methods("GET")

	// Combine routers
	mainRouter := mux.NewRouter()
	mainRouter.PathPrefix("/health").Handler(healthRouter)
	mainRouter.PathPrefix("/").Handler(router)

	port := os.Getenv("RECOMMENDATIONS_PORT")
	if port == "" {
		port = "8003"
	}

	logging.Logger.Info("Starting recommendations service", "port", port)
	if err := http.ListenAndServe(":"+port, mainRouter); err != nil {
		logging.Logger.Error("Server failed to start", "error", err)
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	//logging.WithContext(r.Context()).Debug("Health check requested")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "service": "recommendations"}`))
}
