package routes

import (
	"donapresentes/controllers"

	"github.com/gorilla/mux"
)

func RegisterFuncionariosRoutes(r *mux.Router) {
	r.HandleFunc("/funcionarios", controllers.GetFuncionarios).Methods("GET")
	r.HandleFunc("/funcionarios/{id}", controllers.GetFuncionario).Methods("GET")
	r.HandleFunc("/funcionarios", controllers.CreateFuncionario).Methods("POST")
	r.HandleFunc("/funcionarios/{id}", controllers.UpdateFuncionario).Methods("PUT")
	r.HandleFunc("/funcionarios/{id}", controllers.DeleteFuncionario).Methods("DELETE")
}
