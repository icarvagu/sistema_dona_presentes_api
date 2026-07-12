package routes

import (
	"donapresentes/controllers"
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterDashboardRoutes registers all dashboard-related HTTP endpoints on the given router.
// Routes: GET /dashboard/seller/{id}.
func RegisterDashboardRoutes(r *mux.Router) {
	r.HandleFunc("/dashboard/seller/{id}", controllers.GetSellerDashboard).Methods(http.MethodGet)
}
