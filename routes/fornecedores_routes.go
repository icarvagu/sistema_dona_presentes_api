package routes

import (
	"donapresentes/controllers"

	"github.com/gorilla/mux"
)

func RegisterFornecedoresRoutes(r *mux.Router) {
	r.HandleFunc("/fornecedores", controllers.GetFornecedores).Methods("GET")
	r.HandleFunc("/fornecedores/{id}", controllers.GetFornecedor).Methods("GET")
	r.HandleFunc("/fornecedores", controllers.CreateFornecedor).Methods("POST")
	r.HandleFunc("/fornecedores/{id}", controllers.UpdateFornecedor).Methods("PUT")
	r.HandleFunc("/fornecedores/{id}", controllers.DeleteFornecedor).Methods("DELETE")
}
