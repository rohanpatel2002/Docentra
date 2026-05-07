package middleware

import (
	"ai-document-assistant/pkg/utils"
	"context"
	"net/http"
	"strings"
)

type ContextKey string

const UserIDKey ContextKey = "user_id"

// Verifies incoming HTTP requests
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "Missing authorization header"}`, http.StatusUnauthorized)
			return
		}
		// Enforce the Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, `{"error": "Invalid authorization format. Expected 'Bearer <token>'"}`, http.StatusUnauthorized)
			return
		}
		tokenString := parts[1]
		// Cryptographically validate the token
		userID, err := utils.ValidateJWT(tokenString)
		if err != nil {
			http.Error(w, `{"error": "Unauthorized or expired token"}`, http.StatusUnauthorized)
			return
		}
		// Attach the validated User ID to the HTTP Request Context
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		// Allow the request to proceed deeper into the API
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}