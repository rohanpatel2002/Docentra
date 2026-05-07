package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmbeddingClient_GetEmbedding(t *testing.T) {
	t.Run("Success Request", func(t *testing.T) {
		// Mock server to simulate Python FastAPI
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/embed", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"embedding": [0.1, 0.2, 0.3]}`))
		}))
		defer server.Close()
		client := NewEmbeddingClient(server.URL)
		embedding, err := client.GetEmbedding(context.Background(), "Praalak tech sol test text")
		assert.NoError(t, err)
		assert.Equal(t, []float64{0.1, 0.2, 0.3}, embedding)
	})
	t.Run("Failure - Empty Text", func(t *testing.T) {
		client := NewEmbeddingClient("http://localhost:8000")
		_, err := client.GetEmbedding(context.Background(), "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot embed empty text")
	})
}
