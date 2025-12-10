package routes

import (
	"net/http"

	"donapresentes/controllers"

	"github.com/gorilla/mux"
)

func RegisterCustomersRoutes(r *mux.Router) {
	r.HandleFunc("/customers", controllers.GetCustomers).Methods(http.MethodGet)
	r.HandleFunc("/customers/{id}", controllers.GetCustomer).Methods(http.MethodGet)
	r.HandleFunc("/customers", controllers.CreateCustomer).Methods(http.MethodPost)
	r.HandleFunc("/customers/{id}", controllers.UpdateCustomer).Methods(http.MethodPut)
	r.HandleFunc("/customers/{id}", controllers.DeleteCustomer).Methods(http.MethodDelete)
}
