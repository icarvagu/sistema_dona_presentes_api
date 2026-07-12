package routes

import (
	"net/http"

	"donapresentes/controllers"

	"github.com/gorilla/mux"
)

// RegisterSalesRoutes registers all sale-related HTTP endpoints on the given router.
// Routes: GET /sales (list), POST /sales (create), GET /sales/{id}, PUT /sales/{id},
// DELETE /sales/{id}, GET /sales/{id}/pdf, PUT /sales/{id}/layout.
func RegisterSalesRoutes(r *mux.Router) {
	r.HandleFunc("/sales", controllers.GetSales).Methods(http.MethodGet)
	r.HandleFunc("/sales/{id}", controllers.GetSale).Methods(http.MethodGet)
	r.HandleFunc("/sales/{id}/pdf", controllers.GetSalePDF).Methods(http.MethodGet)
	r.HandleFunc("/sales/{id}/layout", controllers.UpdateSaleLayout).Methods(http.MethodPut)
	r.HandleFunc("/sales", controllers.CreateSale).Methods(http.MethodPost)
	r.HandleFunc("/sales/{id}", controllers.UpdateSale).Methods(http.MethodPut)
	r.HandleFunc("/sales/{id}", controllers.DeleteSale).Methods(http.MethodDelete)
}
