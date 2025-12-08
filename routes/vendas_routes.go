package routes

import (
	"net/http"

	"donapresentes/controllers"

	"github.com/gorilla/mux"
)

func RegisterVendasRoutes(r *mux.Router) {
	r.HandleFunc("/vendas", controllers.GetVendas).Methods(http.MethodGet)
	r.HandleFunc("/vendas/{id}", controllers.GetVenda).Methods(http.MethodGet)
	r.HandleFunc("/vendas", controllers.CreateVenda).Methods(http.MethodPost)
	r.HandleFunc("/vendas/{id}", controllers.UpdateVenda).Methods(http.MethodPut)
	r.HandleFunc("/vendas/{id}", controllers.DeleteVenda).Methods(http.MethodDelete)
}
