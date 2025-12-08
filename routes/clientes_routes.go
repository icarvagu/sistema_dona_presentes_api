package routes

import (
	"net/http"

	"donapresentes/controllers"

	"github.com/gorilla/mux"
)

func RegisterClientesRoutes(r *mux.Router) {
	r.HandleFunc("/clientes", controllers.GetClientes).Methods(http.MethodGet)
	r.HandleFunc("/clientes/{id}", controllers.GetCliente).Methods(http.MethodGet)
	r.HandleFunc("/clientes", controllers.CreateCliente).Methods(http.MethodPost)
	r.HandleFunc("/clientes/{id}", controllers.UpdateCliente).Methods(http.MethodPut)
	r.HandleFunc("/clientes/{id}", controllers.DeleteCliente).Methods(http.MethodDelete)
}
