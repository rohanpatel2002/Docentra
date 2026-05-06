package main

import (
	"log"
	"net/http"
	"os"

	"ai-document-assistant/pkg/database"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	// 2. Initialize the Database connection
	database.ConnectDB()

	// 3. Initializing router
	r := chi.NewRouter()
	//Basic middleware
	r.Use(middleware.Logger)    //Logs every incoming request
	r.Use(middleware.Recoverer) // Prevents the server from crashing
	// 4. Create a basic health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "message": "AI Document Assistant API is running!"}`))
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
