package routes

import (
	"donapresentes/controllers"

	"github.com/gorilla/mux"
)

func RegisterProdutoXBZRoutes(router *mux.Router) {
	controllers.InitProdutoXBZController()

	// POST /produtos-xbz/sincronizar - sincroniza produtos da API XBZ
	router.HandleFunc("/produtos-xbz/sincronizar", controllers.SincronizarProdutos).Methods("POST")
}
