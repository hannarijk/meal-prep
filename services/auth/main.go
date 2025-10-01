package main

import (
	"net/http"
	"os"

	"meal-prep/services/auth/handlers"
	"meal-prep/services/auth/repository"
	"meal-prep/services/auth/service"
	"meal-prep/shared/database"
	"meal-prep/shared/logging"
	"meal-prep/shared/middleware"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// TODO: API Gateway Migration Checklist
// When implementing API Gateway:
// 1. Set ENABLE_CORS=false in production environment
// 2. Remove CORS middleware from all services
// 3. Configure CORS at API Gateway level
// 4. Update frontend VITE_API_BASE_URL to point to gateway
// 5. Remove this comment block
func main() {
	// Initialize logging first
	logging.Init("auth-service")

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
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)

	// Routes with logging middleware
	router := mux.NewRouter()

	// CORS - only for local development (API Gateway will handle this in production)
	if os.Getenv("ENABLE_CORS") == "true" {
		router.Use(middleware.CORSMiddleware)
	}

	router.Use(middleware.LoggingMiddleware("auth-service"))

	router.HandleFunc("/register", authHandler.Register).Methods("POST", "OPTIONS")
	router.HandleFunc("/login", authHandler.Login).Methods("POST", "OPTIONS")
	router.HandleFunc("/health", healthCheck).Methods("GET", "OPTIONS")

	port := os.Getenv("AUTH_PORT")
	if port == "" {
		port = "8001"
	}

	logging.Logger.Info("Starting auth service", "port", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		logging.Logger.Error("Server failed to start", "error", err)
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	logging.WithContext(r.Context()).Debug("Health check requested")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "service": "auth"}`))
}
