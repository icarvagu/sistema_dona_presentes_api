package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"donapresentes/errors"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// ErrorHandler escreve um erro formatado na resposta HTTP
func ErrorHandler(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")

	response := errors.ErrorResponse{
		Error: err.Error(),
		Code:  statusCode,
	}

	// Se for AppError, usar o código do erro, senão usar o statusCode passado
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

// LoggingMiddleware registra informações sobre as requisições
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)

		duration := time.Since(startTime)
		log.Printf("[HTTP] method=%s path=%s status=%d duration=%s remote=%s", r.Method, r.URL.Path, rec.statusCode, duration, r.RemoteAddr)
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

// NormalizePathMiddleware remove barras duplicadas do path
func NormalizePathMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Remove barras duplicadas do path
		path := r.URL.Path
		for strings.Contains(path, "//") {
			path = strings.ReplaceAll(path, "//", "/")
		}
		r.URL.Path = path
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware adiciona headers CORS para permitir requisições do frontend
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Permite todas as origens (em produção, substitua por origens específicas)
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin, X-CSRF-Token")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type, Location")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.Header().Set("Access-Control-Allow-Credentials", "false")

		// Responde imediatamente para requisições OPTIONS (preflight)
		// Isso deve acontecer ANTES de qualquer processamento do router
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
