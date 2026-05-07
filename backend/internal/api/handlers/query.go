package handlers

import (
	"ai-document-assistant/internal/api/middleware"
	"ai-document-assistant/internal/repository"
	"ai-document-assistant/internal/service"
	"encoding/json"
	"net/http"

	pgvector "github.com/pgvector/pgvector-go"
)

type QueryHandler struct {
	repo      repository.DocumentRepository
	aiService service.AIService
}

func NewQueryHandler(repo repository.DocumentRepository, ai service.AIService) *QueryHandler {
	return &QueryHandler{
		repo:      repo,
		aiService: ai,
	}
}

type QueryRequest struct {
	Query string `json:"query"`
}
type QueryResponse struct {
	Query   string                    `json:"query"`
	Context []repository.SearchResult `json:"context"`
}

// Query handles semantic retrieval to find relevant document
func (h *QueryHandler) Query(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Query == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}
	// Vectorize the query text
	vector, err := h.aiService.GetEmbedding(r.Context(), req.Query)
	if err != nil {
		http.Error(w, "Failed to vectorize query", http.StatusInternalServerError)
		return
	}
	// Native Sql vector similarity search
	results, err := h.repo.SearchSimilarChunks(userID, pgvector.NewVector(vector), 5)
	if err != nil {
		http.Error(w, "Semantic search failed", http.StatusInternalServerError)
		return
	}
	resp := QueryResponse{
		Query:   req.Query,
		Context: results,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
