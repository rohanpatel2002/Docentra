package middleware

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen int64
}

var (
	mu      sync.Mutex
	clients = make(map[string]*rate.Limiter)
)

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		mu.Lock()
		if _, found := clients[ip]; !found {
			// Allow 5 requests per second with a burst of 10
			clients[ip] = rate.NewLimiter(5, 10)
		}
		limiter := clients[ip]
		mu.Unlock()

		if !limiter.Allow() {
			http.Error(w, "Too many requests. Please slow down your Ai queries", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
