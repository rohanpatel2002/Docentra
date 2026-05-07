package middleware

import (
	"ai-document-assistant/pkg/utils"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	// A dummy handler to check if the next handler is reached
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(UserIDKey)
		assert.NotNil(t, userID)
		w.WriteHeader(http.StatusOK)
	})
	t.Run("Success - Valid JWT", func(t *testing.T) {
		token, _ := utils.GenerateJWT(42) // Valid token for User 42
		req := httptest.NewRequest("GET", "/api/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()
		AuthMiddleware(nextHandler).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
	t.Run("Failure - Missing Header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/protected", nil)
		rr := httptest.NewRecorder()
		AuthMiddleware(nextHandler).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Missing authorization header")
	})
	t.Run("Failure - Invalid Format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/protected", nil)
		req.Header.Set("Authorization", "InvalidToken123")
		rr := httptest.NewRecorder()
		AuthMiddleware(nextHandler).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid authorization format")
	})
	t.Run("Failure - Expired or Malformed Token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/protected", nil)
		req.Header.Set("Authorization", "Bearer obvious.fake.token")
		rr := httptest.NewRecorder()
		AuthMiddleware(nextHandler).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}
