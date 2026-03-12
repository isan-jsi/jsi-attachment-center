package middleware

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// RateLimiter holds per-key token bucket limiters.
type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	rps      rate.Limit
	burst    int
}

// NewRateLimiter creates a RateLimiter with the given requests-per-second and burst size.
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rate.Limit(rps),
		burst:    burst,
	}
}

// getLimiter returns an existing or newly created limiter for the given key.
func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if l, ok := rl.limiters[key]; ok {
		return l
	}
	l := rate.NewLimiter(rl.rps, rl.burst)
	rl.limiters[key] = l
	return l
}

// extractKey returns a rate-limit key derived from the request.
// Priority: X-API-Key header → X-Forwarded-For header → RemoteAddr.
func extractKey(r *http.Request) string {
	if key := r.Header.Get("X-API-Key"); key != "" {
		return "apikey:" + key
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return "ip:" + xff
	}
	return "ip:" + r.RemoteAddr
}

// Middleware returns an HTTP middleware that enforces per-key rate limits.
// Requests that exceed the limit receive HTTP 429 with a Retry-After header.
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := extractKey(r)
			limiter := rl.getLimiter(key)
			if !limiter.Allow() {
				w.Header().Set("Retry-After", "1")
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
