package middleware

import (
	"net/http"

	"github.com/gorilla/mux"
)

// ApplyCORSFallbackHandlers ensures CORS headers are also present on mux 404/405 responses.
func ApplyCORSFallbackHandlers(router *mux.Router) {
	router.NotFoundHandler = CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}))

	router.MethodNotAllowedHandler = CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}))
}
