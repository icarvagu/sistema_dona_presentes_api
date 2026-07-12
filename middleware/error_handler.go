package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"donapresentes/errors"
	"donapresentes/repositories"
)

var errorLogRepo *repositories.ErrorLogRepository

// InitErrorLogRepo injects the error log repository so that middleware can persist
// 5xx errors and panics to the database.
func InitErrorLogRepo(repo *repositories.ErrorLogRepository) {
	errorLogRepo = repo
}

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

func safeErrorMessage(err error) string {
	if appErr, ok := err.(*errors.AppError); ok {
		return appErr.Message
	}
	log.Printf("[ERROR] raw error (sanitized for response): %s", err.Error())
	return "Erro interno do servidor"
}

// ErrorHandler writes a JSON error response to the client, extracting code and
// details from *errors.AppError when available.
func ErrorHandler(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")

	response := errors.ErrorResponse{
		Error: safeErrorMessage(err),
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

// ErrorHandlerWithRequest writes a JSON error response and, for 5xx errors,
// persists an error log record using the configured ErrorLogRepository.
func ErrorHandlerWithRequest(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")

	response := errors.ErrorResponse{
		Error: safeErrorMessage(err),
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

	if errorLogRepo != nil && statusCode >= 500 {
		requestID := GetRequestID(r)
		var userID *int
		if uid, ok := r.Context().Value("user_id").(int); ok && uid > 0 {
			userID = &uid
		}
		errorLogRepo.Insert(&repositories.ErrorLogRecord{
			RequestID:    requestID,
			UserID:       userID,
			ErrorCode:    finalStatusCode,
			ErrorMessage: err.Error(),
			Path:         r.URL.Path,
			Method:       r.Method,
			CreatedAt:    time.Now(),
		})
	}

	w.WriteHeader(finalStatusCode)
	json.NewEncoder(w).Encode(response)
}

// LoggingMiddleware logs every request with its method, path, status code,
// duration, and remote address.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)

		duration := time.Since(startTime)
		log.Printf("[HTTP] method=%s path=%s status=%d duration=%s remote=%s", r.Method, r.URL.Path, rec.statusCode, duration, r.RemoteAddr)
	})
}

// RecoveryMiddleware catches panics in the HTTP handler chain, logs them, and
// returns a 500 Internal Server Error response.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC] %v", err)
				if errorLogRepo != nil {
					requestID := GetRequestID(r)
					var userID *int
					if uid, ok := r.Context().Value("user_id").(int); ok && uid > 0 {
						userID = &uid
					}
					errMsg := ""
					switch v := err.(type) {
					case error:
						errMsg = v.Error()
					default:
						errMsg = fmt.Sprintf("%v", v)
					}
					errorLogRepo.Insert(&repositories.ErrorLogRecord{
						RequestID:    requestID,
						UserID:       userID,
						ErrorCode:    http.StatusInternalServerError,
						ErrorMessage: errMsg,
						Path:         r.URL.Path,
						Method:       r.Method,
						CreatedAt:    time.Now(),
					})
				}
				ErrorHandlerWithRequest(w, r, errors.ErrInternalServer, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// NormalizePathMiddleware collapses consecutive slashes in the URL path to
// prevent double-slash 404 mismatches.
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

// CORSMiddleware sets CORS headers for cross-origin requests, matching the
// request Origin against a configurable allowlist.
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
