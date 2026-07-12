package routes

import (
	"net/http"

	"donapresentes/controllers"

	"github.com/gorilla/mux"
)

// RegisterQuotesRoutes registers all quote-related HTTP endpoints on the given router.
// Routes: GET /quotes (list), POST /quotes (create), GET /quotes/{id}, PUT /quotes/{id},
// DELETE /quotes/{id}, GET /quotes/{id}/pdf, PATCH /quotes/{id}/feedback.
func RegisterQuotesRoutes(r *mux.Router) {
	r.HandleFunc("/quotes", controllers.GetQuotes).Methods(http.MethodGet)
	r.HandleFunc("/quotes/{id}", controllers.GetQuote).Methods(http.MethodGet)
	r.HandleFunc("/quotes/{id}/pdf", controllers.GetQuotePDF).Methods(http.MethodGet)
	r.HandleFunc("/quotes", controllers.CreateQuote).Methods(http.MethodPost)
	r.HandleFunc("/quotes/{id}", controllers.UpdateQuote).Methods(http.MethodPut)
	r.HandleFunc("/quotes/{id}/feedback", controllers.UpdateQuoteFeedback).Methods(http.MethodPatch)
	r.HandleFunc("/quotes/{id}", controllers.DeleteQuote).Methods(http.MethodDelete)
}
