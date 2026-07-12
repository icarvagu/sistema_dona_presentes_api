package routes

import (
	"donapresentes/controllers"

	"github.com/gorilla/mux"
)

// RegisterSuppliersRoutes registers all supplier-related HTTP endpoints on the given router.
// Routes: GET /suppliers (list), POST /suppliers (create), GET /suppliers/{id},
// PUT /suppliers/{id}, DELETE /suppliers/{id}.
func RegisterSuppliersRoutes(r *mux.Router) {
	r.HandleFunc("/suppliers", controllers.GetSuppliers).Methods("GET")
	r.HandleFunc("/suppliers/{id}", controllers.GetSupplier).Methods("GET")
	r.HandleFunc("/suppliers", controllers.CreateSupplier).Methods("POST")
	r.HandleFunc("/suppliers/{id}", controllers.UpdateSupplier).Methods("PUT")
	r.HandleFunc("/suppliers/{id}", controllers.DeleteSupplier).Methods("DELETE")
}
