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

	router.Use(middleware.LoggingMiddleware("auth-service"))

	router.HandleFunc("/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/login", authHandler.Login).Methods("POST")
	router.Handle("/auth/me", middleware.ExtractUserFromGatewayHeaders(
		http.HandlerFunc(authHandler.Me),
	)).Methods("GET")
	router.HandleFunc("/health", healthCheck).Methods("GET")

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
	//logging.WithContext(r.Context()).Debug("Health check requested")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "service": "auth"}`))
}
