package routes

import (
	"net/http"

	"donapresentes/controllers"

	"github.com/gorilla/mux"
)

// RegisterProductsXBZRoutes registers all XBZ product synchronization HTTP endpoints on the given router.
// Routes: POST /products-xbz/sync.
func RegisterProductsXBZRoutes(router *mux.Router) {
	controllers.InitProductXBZController()

	router.HandleFunc("/products-xbz/sync", controllers.SyncProductsFromXBZHandler).Methods(http.MethodPost)
}
