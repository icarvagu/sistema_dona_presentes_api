package routes

import (
	"donapresentes/controllers"
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterDashboardRoutes(r *mux.Router) {
	r.HandleFunc("/dashboard/seller/{id}", controllers.GetSellerDashboard).Methods(http.MethodGet)
}
