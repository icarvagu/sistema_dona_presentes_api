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

// RequestIDKey is the context key used to store and retrieve the request ID.
const RequestIDKey ctxKey = "request_id"

var requestLogger *services.RequestLoggerService

// InitRequestLogger injects the request logger service so that middleware can
// persist structured request logs.
func InitRequestLogger(s *services.RequestLoggerService) {
	requestLogger = s
}

func generateRequestID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GetRequestID retrieves the request ID from the request context or the
// X-Request-ID header.
func GetRequestID(r *http.Request) string {
	if id, ok := r.Context().Value(RequestIDKey).(string); ok {
		return id
	}
	if id := r.Header.Get("X-Request-ID"); id != "" {
		return id
	}
	return ""
}

// TracingMiddleware ensures every request has a unique request ID, either from
// the X-Request-ID header or generated on the fly, and propagates it via context
// and the response header.
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

// StructuredLoggingMiddleware logs every request as a structured log entry,
// adjusting the log level based on status code and marking slow requests
// (over 5 seconds). It also writes a request log record when a
// RequestLoggerService is configured.
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

// SlowRequestLogger provides a configurable threshold for detecting slow requests.
type SlowRequestLogger struct {
	Threshold time.Duration
}

// NewSlowRequestLogger creates a SlowRequestLogger with the given duration
// threshold.
func NewSlowRequestLogger(threshold time.Duration) *SlowRequestLogger {
	return &SlowRequestLogger{Threshold: threshold}
}

// DBQueryLog holds a database query string and its execution duration.
type DBQueryLog struct {
	Query    string
	Duration time.Duration
}

// LogDBQuery emits a warning log entry when a database query exceeds 200 ms.
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
