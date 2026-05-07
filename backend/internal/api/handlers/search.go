package handlers

import (
	"ai-document-assistant/internal/api/middleware"
	"ai-document-assistant/internal/repository"
	"ai-document-assistant/internal/service"
	"encoding/json"
	"net/http"

	pgvector "github.com/pgvector/pgvector-go"
)

type SearchHandler struct {
	repo      repository.DocumentRepository
	aiService service.AIService
}

func NewSearchHandler(repo repository.DocumentRepository, ai service.AIService) *SearchHandler {
	return &SearchHandler{
		repo:      repo,
		aiService: ai,
	}
}

type SearchRequest struct {
	Query string `json:"query"`
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	// Vectorize the text question
	vector, err := h.aiService.GetEmbedding(r.Context(), req.Query)
	if err != nil {
		http.Error(w, "Failed to vectorize query", http.StatusInternalServerError)
		return
	}

	// search via pgvector query in the repository
	results, err := h.repo.SearchSimilarChunks(userID, pgvector.NewVector(vector), 5)
	if err != nil {
		http.Error(w, "Database search failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
