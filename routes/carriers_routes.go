package routes

import (
	"donapresentes/controllers"

	"github.com/gorilla/mux"
)

// RegisterCarriersRoutes registers all carrier-related HTTP endpoints on the given router.
// Routes: GET /carriers (list), POST /carriers (create), GET /carriers/{id},
// PUT /carriers/{id}, DELETE /carriers/{id}.
func RegisterCarriersRoutes(r *mux.Router) {
	r.HandleFunc("/carriers", controllers.GetCarriers).Methods("GET")
	r.HandleFunc("/carriers/{id}", controllers.GetCarrier).Methods("GET")
	r.HandleFunc("/carriers", controllers.CreateCarrier).Methods("POST")
	r.HandleFunc("/carriers/{id}", controllers.UpdateCarrier).Methods("PUT")
	r.HandleFunc("/carriers/{id}", controllers.DeleteCarrier).Methods("DELETE")
}
