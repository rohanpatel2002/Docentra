package middleware

import (
	"context"
	"net/http"
	"strings"

	"ai-document-assistant/pkg/utils"
)

// ContextKey is a custom type to specifically prevent context key collisions in Go.
// Using raw strings for context keys is a known anti-pattern in Go.
type ContextKey string

const UserIDKey ContextKey = "user_id"

// AuthMiddleware is a robust Bouncer that verifies incoming HTTP requests.
// It demands a mathematically valid Bearer JWT before allowing the request to proceed to protected routes.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Extract the secure Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "Missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		// 2. Enforce the strict 'Bearer <token>' format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, `{"error": "Invalid authorization format. Expected 'Bearer <token>'"}`, http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// 3. Cryptographically validate the token via our thoroughly tested utils package
		userID, err := utils.ValidateJWT(tokenString)
		if err != nil {
			// SECURITY: We purposefully send a generic 401 Unauthorized message to the client.
			// Giving specific errors (e.g. "expired" vs "invalid signature") allows attackers to probe our validation logic.
			http.Error(w, `{"error": "Unauthorized or expired token"}`, http.StatusUnauthorized)
			return
		}

		// 4. Securely attach the validated User ID to the HTTP Request Context.
		// This guarantees that any downstream handler natively knows EXACTLY who is making the request locally.
		ctx := context.WithValue(r.Context(), UserIDKey, userID)

		// 5. Allow the request to proceed deeper into the API with the new verified context attached
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
