package routes

import (
	"net/http"

	"donapresentes/controllers"

	"github.com/gorilla/mux"
)

func RegisterProductsRoutes(r *mux.Router) {
	r.HandleFunc("/products", controllers.GetProducts).Methods(http.MethodGet)
	r.HandleFunc("/products/{id}", controllers.GetProduct).Methods(http.MethodGet)
	r.HandleFunc("/products", controllers.CreateProduct).Methods(http.MethodPost)
	r.HandleFunc("/products/{id}", controllers.UpdateProduct).Methods(http.MethodPut)
	r.HandleFunc("/products/{id}", controllers.DeleteProduct).Methods(http.MethodDelete)
}
