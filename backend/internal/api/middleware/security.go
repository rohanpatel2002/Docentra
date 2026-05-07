package middleware

import "net/http"

// Adds critical security headers to every response
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent browsers from guessing Content Type
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// Prevent the site from being rendered
		w.Header().Set("X-Frame-Options", "DENY")
		// Enable browser XSS filtering
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		// Ensure cross origin requests dont leak referrer info
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		next.ServeHTTP(w, r)
	})
}
