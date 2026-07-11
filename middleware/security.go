package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	apperrors "donapresentes/errors"
)

var (
	reXSS        = regexp.MustCompile(`(?i)<script|javascript:|onerror=|onload=|onclick=|onmouseover|alert\(|prompt\(|confirm\(`)
	reSQLInj     = regexp.MustCompile(`(?i)('|--|/\*|;|\bunion\b|\bselect\b|\bdrop\b|\bdelete\b|\binsert\b|\bupdate\b|\bcreate\b|\balter\b|\bexec\b|\bxp_\b)`)
	reNoSQLInj   = regexp.MustCompile(`(?i)\$where|\$gt|\$lt|\$ne|\$regex|\$nin|\$in`)
	rePathTraver = regexp.MustCompile(`(?i)\.\./|\.\.\\|/etc/passwd|/proc/self|%00|null|boot\.ini`)
	reBinary     = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F]`)
	reOnlyDigits = regexp.MustCompile(`^\d+$`)
)

func SanitizeInput(input string) string {
	trimmed := strings.TrimSpace(input)
	return reBinary.ReplaceAllString(trimmed, "")
}

func isSuspicious(payload string) bool {
	return reXSS.MatchString(payload) ||
		reSQLInj.MatchString(payload) ||
		reNoSQLInj.MatchString(payload) ||
		rePathTraver.MatchString(payload)
}

type timeoutResponseWriter struct {
	http.ResponseWriter
	timedOut bool
}

func InputValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = SanitizeInput(r.URL.Path)
		r.URL.RawQuery = SanitizeInput(r.URL.RawQuery)

		if strings.Count(r.URL.Path, "/") > 20 {
			ErrorHandler(w, apperrors.NewValidationError("URL inválida"), http.StatusBadRequest)
			return
		}

		for key, values := range r.URL.Query() {
			if len(key) > 256 {
				ErrorHandler(w, apperrors.NewValidationError("Parâmetro inválido"), http.StatusBadRequest)
				return
			}
			for _, v := range values {
				if len(v) > 4096 || isSuspicious(key) || isSuspicious(v) {
					ErrorHandler(w, apperrors.NewValidationError("Conteúdo suspeito detectado"), http.StatusBadRequest)
					return
				}
			}
		}

		for key, values := range r.Header {
			log.Printf("[DEBUG-ALLHEADERS] key=%q values=%q", key, values)
			uk := strings.ToUpper(key)
			if strings.HasPrefix(uk, "X-FORWARDED") || strings.HasPrefix(uk, "X-REAL-") || key == "Cookie" || key == "Authorization" || key == "Content-Type" || key == "Accept" || key == "Origin" || key == "Referer" || key == "User-Agent" || strings.HasPrefix(uk, "SEC-") || key == "Accept-Language" || key == "Accept-Encoding" {
				continue
			}
			for _, v := range values {
				if len(v) > 8192 || isSuspicious(v) {
					ErrorHandler(w, apperrors.NewValidationError("Header inválido"), http.StatusBadRequest)
					return
				}
			}
		}

		if r.ContentLength > 10<<20 {
			ErrorHandler(w, apperrors.NewValidationError("Corpo da requisição muito grande"), http.StatusRequestEntityTooLarge)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

		next.ServeHTTP(w, r)
	})
}

func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			trw := &timeoutResponseWriter{ResponseWriter: w}
			done := make(chan bool, 1)

			go func() {
				next.ServeHTTP(trw, r)
				done <- true
			}()

			select {
			case <-done:
			case <-time.After(timeout):
				trw.timedOut = true
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusGatewayTimeout)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "A requisição excedeu o tempo limite",
					"code":  504,
				})
			}
		})
	}
}
