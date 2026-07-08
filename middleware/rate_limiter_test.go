package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterAllow(t *testing.T) {
	rl := NewRateLimiter(3, 1*time.Minute)

	tests := []struct {
		name    string
		key     string
		allowed bool
	}{
		{"primeira requisição", "user1", true},
		{"segunda requisição", "user1", true},
		{"terceira requisição", "user1", true},
		{"quarta requisição (deve falhar)", "user1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, _ := rl.Allow(tt.key)
			if allowed != tt.allowed {
				t.Fatalf("esperado allowed=%v, got %v", tt.allowed, allowed)
			}
		})
	}
}

func TestRateLimiterDifferentKeys(t *testing.T) {
	rl := NewRateLimiter(2, 1*time.Minute)

	allowed1, _ := rl.Allow("user1")
	if !allowed1 {
		t.Fatalf("user1 deveria ser permitido")
	}

	allowed1, _ = rl.Allow("user1")
	if !allowed1 {
		t.Fatalf("user1 ainda deveria ser permitido")
	}

	allowed2, _ := rl.Allow("user2")
	if !allowed2 {
		t.Fatalf("user2 deveria ser permitido (chave diferente)")
	}

	allowed2, _ = rl.Allow("user2")
	if !allowed2 {
		t.Fatalf("user2 ainda deveria ser permitido (chave diferente)")
	}

	allowed1, retryAfter := rl.Allow("user1")
	if allowed1 {
		t.Fatalf("user1 deveria estar bloqueado")
	}
	if retryAfter <= 0 {
		t.Fatalf("retryAfter deve ser positivo, got %v", retryAfter)
	}
}

func TestRateLimiterResetAfterWindow(t *testing.T) {
	rl := NewRateLimiter(1, 100*time.Millisecond)

	allowed, _ := rl.Allow("user1")
	if !allowed {
		t.Fatalf("primeira requisição deveria ser permitida")
	}

	allowed, _ = rl.Allow("user1")
	if allowed {
		t.Fatalf("segunda requisição deveria ser bloqueada")
	}

	time.Sleep(150 * time.Millisecond)

	allowed, _ = rl.Allow("user1")
	if !allowed {
		t.Fatalf("após janela expirar, requisição deveria ser permitida")
	}
}

func TestRateLimiterRetryAfterHeader(t *testing.T) {
	rl := NewRateLimiter(1, 10*time.Minute)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.RemoteAddr = "test-ip"
		key := r.RemoteAddr + ":" + r.Method + ":" + r.URL.Path
		allowed, retryAfter := rl.Allow(key)
		if !allowed {
			writeRateLimitResponse(w, retryAfter)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	req1 := httptest.NewRequest("GET", "/", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("primeira request esperava 200, got %d", rec1.Code)
	}

	req2 := httptest.NewRequest("GET", "/", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("segunda request esperava 429, got %d", rec2.Code)
	}

	retryAfter := rec2.Header().Get("Retry-After")
	if retryAfter == "" {
		t.Fatalf("esperava header Retry-After na resposta 429")
	}
}
