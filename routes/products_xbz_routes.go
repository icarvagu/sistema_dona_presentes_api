package routes

import (
	"net/http"

	"donapresentes/controllers"

	"github.com/gorilla/mux"
)

func RegisterProductsXBZRoutes(router *mux.Router) {
	controllers.InitProductXBZController()

	router.HandleFunc("/products-xbz/sync", controllers.SyncProductsFromXBZHandler).Methods(http.MethodPost)
}
