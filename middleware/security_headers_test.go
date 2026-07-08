package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeadersAreSet(t *testing.T) {
	handler := SecurityHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	expectedHeaders := map[string]string{
		"X-Frame-Options":     "DENY",
		"X-Content-Type-Options": "nosniff",
		"Referrer-Policy":     "strict-origin-when-cross-origin",
		"Permissions-Policy":  "geolocation=(), microphone=(), camera=()",
		"X-XSS-Protection":    "0",
		"Access-Control-Allow-Origin": "",
	}

	for header, expected := range expectedHeaders {
		actual := rec.Header().Get(header)
		if actual != expected {
			t.Fatalf("header '%s': esperado '%s', got '%s'", header, expected, actual)
		}
	}
}

func TestSecurityHeadersDontAffectBody(t *testing.T) {
	handler := SecurityHeadersMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != "ok" {
		t.Fatalf("body foi alterado pelos headers: esperado 'ok', got '%s'", rec.Body.String())
	}
}
