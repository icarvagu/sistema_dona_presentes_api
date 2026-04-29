package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestApplyCORSFallbackHandlers_OnMux404And405(t *testing.T) {
	r := mux.NewRouter()
	r.Use(CORSMiddleware)
	ApplyCORSFallbackHandlers(r)

	r.HandleFunc("/ok", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{name: "not found", method: http.MethodGet, path: "/missing", wantStatus: http.StatusNotFound},
		{name: "method not allowed", method: http.MethodPost, path: "/ok", wantStatus: http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Origin", "https://frontend.example")

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://frontend.example" {
				t.Fatalf("expected Access-Control-Allow-Origin header to be propagated, got %q", got)
			}
		})
	}
}
