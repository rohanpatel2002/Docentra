package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRateLimitMiddleware(t *testing.T) {
	mw := RateLimit
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	t.Run("Success - Within limit", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()
		mw(nextHandler).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
	t.Run("Failure - Rate limit exceeded", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		// Burst is 10
		for i := 0; i < 10; i++ {
			rr := httptest.NewRecorder()
			mw(nextHandler).ServeHTTP(rr, req)
			assert.Equal(t, http.StatusOK, rr.Code)
		}
		// The 11th request should fail
		rr := httptest.NewRecorder()
		mw(nextHandler).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	})
	t.Run("Success - Different IP has separate limit", func(t *testing.T) {
		req1 := httptest.NewRequest("GET", "/", nil)
		req1.RemoteAddr = "10.0.0.1:1234"
		req2 := httptest.NewRequest("GET", "/", nil)
		req2.RemoteAddr = "10.0.0.2:1234"
		// Exhaust IP1 (burst 10)
		for i := 0; i < 10; i++ {
			mw(nextHandler).ServeHTTP(httptest.NewRecorder(), req1)
		}
		// IP1 should be blocked
		rr1 := httptest.NewRecorder()
		mw(nextHandler).ServeHTTP(rr1, req1)
		assert.Equal(t, http.StatusTooManyRequests, rr1.Code)
		// IP2 should still be fine
		rr2 := httptest.NewRecorder()
		mw(nextHandler).ServeHTTP(rr2, req2)
		assert.Equal(t, http.StatusOK, rr2.Code)
	})
}
func TestSecurityHeaders(t *testing.T) {
	mw := SecurityHeaders
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	t.Run("Success - Headers are set", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()
		mw(nextHandler).ServeHTTP(rr, req)
		assert.Equal(t, "nosniff", rr.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", rr.Header().Get("X-Frame-Options"))
		assert.Equal(t, "1; mode=block", rr.Header().Get("X-XSS-Protection"))
		assert.Equal(t, "strict-origin-when-cross-origin", rr.Header().Get("Referrer-Policy"))
	})
}

func TestCORSMiddleware(t *testing.T) {
	mw := CORS
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("Success - CORS headers are set", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		mw(nextHandler).ServeHTTP(rr, req)

		assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, rr.Header().Get("Access-Control-Allow-Methods"), "GET")
		assert.Contains(t, rr.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	})

	t.Run("Success - Preflight request handling", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/", nil)
		rr := httptest.NewRecorder()

		mw(nextHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
	})
}
