package routes

import (
	"net/http"

	"donapresentes/controllers"

	"github.com/gorilla/mux"
)

func RegisterProdutosRoutes(r *mux.Router) {
	r.HandleFunc("/produtos", controllers.GetProdutos).Methods(http.MethodGet)
	r.HandleFunc("/produtos/{id}", controllers.GetProduto).Methods(http.MethodGet)
	r.HandleFunc("/produtos", controllers.CreateProduto).Methods(http.MethodPost)
	r.HandleFunc("/produtos/{id}", controllers.UpdateProduto).Methods(http.MethodPut)
	r.HandleFunc("/produtos/{id}", controllers.DeleteProduto).Methods(http.MethodDelete)
}
