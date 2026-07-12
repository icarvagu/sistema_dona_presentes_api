package routes

import (
	"net/http"

	"donapresentes/controllers"

	"github.com/gorilla/mux"
)

// RegisterCustomersRoutes registers all customer-related HTTP endpoints on the given router.
// Routes: GET /customers (list), POST /customers (create), GET /customers/{id},
// PUT /customers/{id}, DELETE /customers/{id}.
func RegisterCustomersRoutes(r *mux.Router) {
	r.HandleFunc("/customers", controllers.GetCustomers).Methods(http.MethodGet)
	r.HandleFunc("/customers/{id}", controllers.GetCustomer).Methods(http.MethodGet)
	r.HandleFunc("/customers", controllers.CreateCustomer).Methods(http.MethodPost)
	r.HandleFunc("/customers/{id}", controllers.UpdateCustomer).Methods(http.MethodPut)
	r.HandleFunc("/customers/{id}", controllers.DeleteCustomer).Methods(http.MethodDelete)
}
