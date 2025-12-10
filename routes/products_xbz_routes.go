package routes

import (
	"donapresentes/controllers"

	"github.com/gorilla/mux"
)

func RegisterProductsXBZRoutes(router *mux.Router) {
	controllers.InitProductXBZController()

	// POST /products-xbz/sync - sincroniza produtos da API XBZ
	router.HandleFunc("/products-xbz/sync", controllers.SyncProductsFromXBZ).Methods("POST")
}
