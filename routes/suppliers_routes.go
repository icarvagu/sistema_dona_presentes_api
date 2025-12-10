package routes

import (
	"donapresentes/controllers"

	"github.com/gorilla/mux"
)

func RegisterSuppliersRoutes(r *mux.Router) {
	r.HandleFunc("/suppliers", controllers.GetSuppliers).Methods("GET")
	r.HandleFunc("/suppliers/{id}", controllers.GetSupplier).Methods("GET")
	r.HandleFunc("/suppliers", controllers.CreateSupplier).Methods("POST")
	r.HandleFunc("/suppliers/{id}", controllers.UpdateSupplier).Methods("PUT")
	r.HandleFunc("/suppliers/{id}", controllers.DeleteSupplier).Methods("DELETE")
}
