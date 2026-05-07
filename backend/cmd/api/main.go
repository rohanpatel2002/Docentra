package main

import (
	"log"
	"net/http"
	"os"

	"ai-document-assistant/internal/api/handlers"
	customMiddleware "ai-document-assistant/internal/api/middleware"
	"ai-document-assistant/internal/repository"
	"ai-document-assistant/internal/service"
	"ai-document-assistant/internal/storage"
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
	// CLI paths
	pythonPath := os.Getenv("PYTHON_PATH")
	if pythonPath == "" {
		pythonPath = "/Users/rohan/Desktop/AI Document Assistant/embedding-service/.venv/bin/python3"
	}
	scriptPath := "/Users/rohan/Desktop/AI Document Assistant/embedding-service/app/main.py"
	// Initialize dependencies
	userRepo := repository.NewUserRepository(database.DB)
	authHandler := handlers.NewAuthHandler(userRepo)
	// Initialize Storage and Document Handler
	localStore, err := storage.NewLocalStorage("uploads")
	if err != nil {
		log.Fatalf("Failed to initialize secure local storage: %v", err)
	}
	docRepo := repository.NewDocumentRepository(database.DB)
	// CLI based Ai Services
	aiSvc := service.NewAIService(pythonPath, scriptPath)
	searchHandler := handlers.NewSearchHandler(docRepo, aiSvc)
	queryHandler := handlers.NewQueryHandler(docRepo, aiSvc)
	docProcessor := service.NewProcessor(docRepo, localStore, aiSvc)
	docHandler := handlers.NewDocumentHandler(docRepo, localStore, docProcessor)
	// Initializing router
	r := chi.NewRouter()

	// Operational & Security Base
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(customMiddleware.CORS)            // Cross-Origin Support
	r.Use(customMiddleware.RateLimit)       // Protection from overuse
	r.Use(customMiddleware.SecurityHeaders) // Browser Security Protection

	// Create a basic health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "message": "AI Document Assistant API is running!"}`))
	})

	// Auth Routes
	r.Post("/api/register", authHandler.Register)
	r.Post("/api/login", authHandler.Login)
	// Protected Routes (Identity Verified)
	r.Group(func(r chi.Router) {
		r.Use(customMiddleware.AuthMiddleware)
		// Document Management
		r.Get("/api/documents", docHandler.GetDocuments)
		r.Post("/api/documents", docHandler.UploadDocument)
		r.Get("/api/documents/{id}", docHandler.GetDocumentStatus)
		r.Delete("/api/documents/{id}", docHandler.DeleteDocument)

		r.Post("/api/search", searchHandler.Search)
		r.Post("/api/query", queryHandler.Query)
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
