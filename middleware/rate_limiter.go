package middleware

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type rateLimiterEntry struct {
	count   int
	resetAt time.Time
}

// RateLimiter is an in-memory token-bucket-like rate limiter keyed by an
// arbitrary string (typically IP + method + path).
type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateLimiterEntry
	limit   int
	window  time.Duration
}

// NewRateLimiter creates a RateLimiter with the given request limit and sliding
// window duration, and starts a background goroutine to clean up expired entries.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		entries: make(map[string]*rateLimiterEntry),
		limit:   limit,
		window:  window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, entry := range rl.entries {
			if now.After(entry.resetAt) {
				delete(rl.entries, key)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow(key string) (bool, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.entries[key]

	if !exists || now.After(entry.resetAt) {
		rl.entries[key] = &rateLimiterEntry{
			count:   1,
			resetAt: now.Add(rl.window),
		}
		return true, 0
	}

	entry.count++
	if entry.count > rl.limit {
		retryAfter := time.Until(entry.resetAt)
		return false, retryAfter
	}
	return true, 0
}

// LoginRateLimiter enforces a strict per-IP rate limit on login endpoints.
var LoginRateLimiter = NewRateLimiter(5, 1*time.Minute)

// GlobalRateLimiter enforces a general per-route rate limit for all requests.
var GlobalRateLimiter = NewRateLimiter(120, 1*time.Minute)

func writeRateLimitResponse(w http.ResponseWriter, retryAfter time.Duration) {
	w.Header().Set("Retry-After", retryAfter.String())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": "Muitas requisições. Tente novamente em alguns instantes",
		"code":  429,
	})
}

// RateLimitMiddlewarePerRoute applies the global rate limiter keyed by
// remote address + HTTP method + URL path.
func RateLimitMiddlewarePerRoute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.RemoteAddr + ":" + r.Method + ":" + r.URL.Path
		allowed, retryAfter := GlobalRateLimiter.Allow(key)
		if !allowed {
			writeRateLimitResponse(w, retryAfter)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RateLimitLoginMiddleware applies the login rate limiter keyed by remote
// address, returning 429 Too Many Requests when the limit is exceeded.
func RateLimitLoginMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.RemoteAddr
		allowed, retryAfter := LoginRateLimiter.Allow(key)
		if !allowed {
			writeRateLimitResponse(w, retryAfter)
			return
		}
		next.ServeHTTP(w, r)
	})
}
