package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"donapresentes/errors"
)

func getAllowedOrigins() []string {
	origins := []string{
		"http://localhost:8081",
		"http://localhost:19006",
		"http://localhost:5173",
		"http://localhost:5174",
		"http://localhost:5175",
		"http://localhost:5176",
	}
	if extra := os.Getenv("CORS_ALLOWED_ORIGINS"); extra != "" {
		for _, o := range strings.Split(extra, ",") {
			if trimmed := strings.TrimSpace(o); trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
	}
	return origins
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func ErrorHandler(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")

	response := errors.ErrorResponse{
		Error: err.Error(),
		Code:  statusCode,
	}

	finalStatusCode := statusCode
	if appErr, ok := err.(*errors.AppError); ok {
		response.Details = appErr.Details
		finalStatusCode = appErr.Code
		if appErr.LogMessage != "" {
			log.Printf("[ERROR] %s: %s", appErr.Message, appErr.LogMessage)
		}
	}

	w.WriteHeader(finalStatusCode)
	json.NewEncoder(w).Encode(response)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)

		duration := time.Since(startTime)
		log.Printf("[HTTP] method=%s path=%s status=%d duration=%s remote=%s", r.Method, r.URL.Path, rec.statusCode, duration, r.RemoteAddr)
	})
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC] %v", err)
				ErrorHandler(w, errors.ErrInternalServer, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func NormalizePathMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		path := r.URL.Path
		for strings.Contains(path, "//") {
			path = strings.ReplaceAll(path, "//", "/")
		}
		r.URL.Path = path
		next.ServeHTTP(w, r)
	})
}

func CORSMiddleware(next http.Handler) http.Handler {
	allowedOrigins := getAllowedOrigins()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				break
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key, X-Requested-With, Accept, Origin, X-CSRF-Token")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type, Location")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
