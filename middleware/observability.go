package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"donapresentes/logger"
	"donapresentes/services"
)

type ctxKey string

const RequestIDKey ctxKey = "request_id"

var requestLogger *services.RequestLoggerService

func InitRequestLogger(s *services.RequestLoggerService) {
	requestLogger = s
}

func generateRequestID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func GetRequestID(r *http.Request) string {
	if id, ok := r.Context().Value(RequestIDKey).(string); ok {
		return id
	}
	if id := r.Header.Get("X-Request-ID"); id != "" {
		return id
	}
	return ""
}

func TracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func StructuredLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rec, r.WithContext(r.Context()))

		duration := time.Since(startTime)
		requestID := GetRequestID(r)
		durationMs := int(duration.Milliseconds())
		slow := duration > 5*time.Second

		entry := logger.Entry{
			Level:      "info",
			RequestID:  requestID,
			Method:     r.Method,
			Path:       r.URL.Path,
			Status:     rec.statusCode,
			Duration:   duration.Round(time.Microsecond).String(),
			RemoteAddr: r.RemoteAddr,
			UserAgent:  r.UserAgent(),
			Message:    "request completed",
			Slow:       slow,
		}

		if slow {
			entry.Level = "warn"
		}
		if rec.statusCode >= 500 {
			entry.Level = "error"
		} else if rec.statusCode >= 400 {
			entry.Level = "warn"
		}

		logger.Log(entry)

		if requestLogger != nil {
			var userID *int
			if uid, ok := r.Context().Value("user_id").(int); ok && uid > 0 {
				userID = &uid
			}
			requestLogger.Log(services.RequestLogEntry{
				RequestID:  requestID,
				Method:     r.Method,
				Path:       r.URL.Path,
				Status:     rec.statusCode,
				DurationMs: durationMs,
				RemoteAddr: r.RemoteAddr,
				UserAgent:  r.UserAgent(),
				UserID:     userID,
				Slow:       slow,
			})
		}
	})
}

type SlowRequestLogger struct {
	Threshold time.Duration
}

func NewSlowRequestLogger(threshold time.Duration) *SlowRequestLogger {
	return &SlowRequestLogger{Threshold: threshold}
}

type DBQueryLog struct {
	Query    string
	Duration time.Duration
}

func LogDBQuery(query string, duration time.Duration, requestID string) {
	if duration > 200*time.Millisecond {
		entry := logger.Entry{
			Level:      "warn",
			RequestID:  requestID,
			Message:    "slow database query",
			DBQuery:    query,
			DBDuration: duration.Round(time.Microsecond).String(),
			Slow:       true,
		}
		logger.Log(entry)
	}
}
