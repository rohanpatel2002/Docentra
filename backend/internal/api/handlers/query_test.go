package handlers

import (
	"ai-document-assistant/internal/api/middleware"
	"ai-document-assistant/internal/repository"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestQueryHandler_Query(t *testing.T) {
	t.Run("Success - Returns context chunks for query", func(t *testing.T) {
		mockRepo := new(MockDocumentRepository)
		mockAI := new(MockAIServiceForSearch)
		handler := NewQueryHandler(mockRepo, mockAI)
		expectedResults := []repository.SearchResult{
			{Content: "Chunk A from secure doc", DocumentID: 1, OriginalName: "secret.pdf", Distance: 0.05},
		}
		// Mock Ai vectorization
		mockAI.On("GetEmbedding", mock.Anything, "Where is the data?").Return([]float32{0.1, 0.2}, nil)
		// Mock DB search
		mockRepo.On("SearchSimilarChunks", uint(42), mock.Anything, 5).Return(expectedResults, nil)
		reqBody, _ := json.Marshal(QueryRequest{Query: "Where is the data?"})
		req := httptest.NewRequest("POST", "/api/query", bytes.NewBuffer(reqBody))
		// Simulating AuthMiddleware
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, uint(42))
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()
		handler.Query(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		var resp QueryResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "Where is the data?", resp.Query)
		assert.Len(t, resp.Context, 1)
		assert.Equal(t, "Chunk A from secure doc", resp.Context[0].Content)
		mockAI.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
	t.Run("Failure - Requires valid query", func(t *testing.T) {
		handler := NewQueryHandler(nil, nil)

		reqBody, _ := json.Marshal(QueryRequest{Query: ""})
		req := httptest.NewRequest("POST", "/api/query", bytes.NewBuffer(reqBody))
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, uint(42))
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.Query(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
	t.Run("Failure - Unauthorized access", func(t *testing.T) {
		handler := NewQueryHandler(nil, nil)

		reqBody, _ := json.Marshal(QueryRequest{Query: "Test"})
		req := httptest.NewRequest("POST", "/api/query", bytes.NewBuffer(reqBody))
		// Simulating missing Auth
		rr := httptest.NewRecorder()
		handler.Query(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}
