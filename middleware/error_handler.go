package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"donapresentes/errors"
)

// ErrorHandler escreve um erro formatado na resposta HTTP
func ErrorHandler(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := errors.ErrorResponse{
		Error: err.Error(),
		Code:  statusCode,
	}

	if appErr, ok := err.(*errors.AppError); ok {
		response.Details = appErr.Details
		w.WriteHeader(appErr.Code)
		if appErr.LogMessage != "" {
			log.Printf("[ERROR] %s: %s", appErr.Message, appErr.LogMessage)
		}
	}

	json.NewEncoder(w).Encode(response)
}

// LoggingMiddleware registra informações sobre as requisições
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		
		log.Printf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		
		next.ServeHTTP(w, r)
		
		duration := time.Since(startTime)
		log.Printf("[%s] %s completed in %v", r.Method, r.URL.Path, duration)
	})
}

// RecoveryMiddleware recupera de panics
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
