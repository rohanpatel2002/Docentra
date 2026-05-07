package handlers

import (
	"ai-document-assistant/internal/api/middleware"
	"ai-document-assistant/internal/repository"
	"ai-document-assistant/internal/service"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAIService for search tests
type MockAIServiceForSearch struct {
	mock.Mock
}

func (m *MockAIServiceForSearch) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	args := m.Called(ctx, text)
	if args.Get(0) != nil {
		return args.Get(0).([]float32), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockAIServiceForSearch) GetEmbeddings(ctx context.Context, text string) ([]service.ChunkResult, error) {
	args := m.Called(ctx, text)
	return args.Get(0).([]service.ChunkResult), args.Error(1)
}

func TestSearchHandler_Search(t *testing.T) {
	t.Run("Success - Returns search results", func(t *testing.T) {
		mockRepo := new(MockDocumentRepository)
		mockAI := new(MockAIServiceForSearch)
		handler := NewSearchHandler(mockRepo, mockAI)

		expectedResults := []repository.SearchResult{
			{Content: "Chunk 1", DocumentID: 1, OriginalName: "test.pdf", Distance: 0.1},
		}

		mockAI.On("GetEmbedding", mock.Anything, "test query").Return([]float32{0.1, 0.2}, nil)
		mockRepo.On("SearchSimilarChunks", uint(42), mock.Anything, 5).Return(expectedResults, nil)

		reqBody, _ := json.Marshal(SearchRequest{Query: "test query"})
		req := httptest.NewRequest("POST", "/api/search", bytes.NewBuffer(reqBody))

		ctx := context.WithValue(req.Context(), middleware.UserIDKey, uint(42))
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Search(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var actualResults []repository.SearchResult
		json.Unmarshal(w.Body.Bytes(), &actualResults)
		assert.Len(t, actualResults, 1)
		assert.Equal(t, "Chunk 1", actualResults[0].Content)
	})

	t.Run("Failure - Empty Query", func(t *testing.T) {
		mockRepo := new(MockDocumentRepository)
		mockAI := new(MockAIServiceForSearch)
		handler := NewSearchHandler(mockRepo, mockAI)

		reqBody, _ := json.Marshal(SearchRequest{Query: ""})
		req := httptest.NewRequest("POST", "/api/search", bytes.NewBuffer(reqBody))
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, uint(42))
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Search(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Failure - Service Error", func(t *testing.T) {
		mockRepo := new(MockDocumentRepository)
		mockAI := new(MockAIServiceForSearch)
		handler := NewSearchHandler(mockRepo, mockAI)

		mockAI.On("GetEmbedding", mock.Anything, "fail").Return(nil, errors.New("ai error"))

		reqBody, _ := json.Marshal(SearchRequest{Query: "fail"})
		req := httptest.NewRequest("POST", "/api/search", bytes.NewBuffer(reqBody))
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, uint(42))
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.Search(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
