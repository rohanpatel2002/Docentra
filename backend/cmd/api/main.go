package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"ai-document-assistant/internal/api/handlers"
	customMiddleware "ai-document-assistant/internal/api/middleware"
	"ai-document-assistant/internal/repository"
	"ai-document-assistant/pkg/database"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	// Load env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	// Initialize the Database connection
	database.ConnectDB()

	// Initialize dependencies
	userRepo := repository.NewUserRepository(database.DB)
	authHandler := handlers.NewAuthHandler(userRepo)

	// 3. Initializing router
	r := chi.NewRouter()

	// Basic middleware
	r.Use(middleware.Logger)    //Logs every incoming request
	r.Use(middleware.Recoverer) // Prevents the server from crashing

	// Create a basic health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "message": "AI Document Assistant API is running!"}`))
	})

	// Auth Routes
	r.Post("/api/register", authHandler.Register)
	r.Post("/api/login", authHandler.Login)

	// Protected Routes
	r.Group(func(r chi.Router) {
		r.Use(customMiddleware.AuthMiddleware)
		r.Get("/api/protected/me", func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(customMiddleware.UserIDKey)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "success",
				"message": "You securely accessed a protected route using JWT Authentication!",
				"user_id": userID,
			})
		})
	})

	//Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default to 8080 if not specified
	}
	log.Printf("Server starting on port %s...\n", port)
	err = http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
