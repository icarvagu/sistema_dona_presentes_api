package routes

import (
	"donapresentes/controllers"

	"github.com/gorilla/mux"
)

func RegisterTransportadoraRoutes(r *mux.Router) {
	r.HandleFunc("/transportadoras", controllers.GetTransportadoras).Methods("GET")
	r.HandleFunc("/transportadoras/{id}", controllers.GetTransportadora).Methods("GET")
	r.HandleFunc("/transportadoras", controllers.CreateTransportadora).Methods("POST")
	r.HandleFunc("/transportadoras/{id}", controllers.UpdateTransportadora).Methods("PUT")
	r.HandleFunc("/transportadoras/{id}", controllers.DeleteTransportadora).Methods("DELETE")
}
