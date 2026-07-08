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

type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateLimiterEntry
	limit   int
	window  time.Duration
}

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

var LoginRateLimiter = NewRateLimiter(5, 1*time.Minute)
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
