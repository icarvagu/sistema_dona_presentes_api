package routes

import (
	"net/http"

	"donapresentes/controllers"

	"github.com/gorilla/mux"
)

func RegisterSalesRoutes(r *mux.Router) {
	r.HandleFunc("/sales", controllers.GetSales).Methods(http.MethodGet)
	r.HandleFunc("/sales/{id}", controllers.GetSale).Methods(http.MethodGet)
	r.HandleFunc("/sales", controllers.CreateSale).Methods(http.MethodPost)
	r.HandleFunc("/sales/{id}", controllers.UpdateSale).Methods(http.MethodPut)
	r.HandleFunc("/sales/{id}", controllers.DeleteSale).Methods(http.MethodDelete)
}
