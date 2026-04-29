package routes

import (
	"donapresentes/controllers"

	"github.com/gorilla/mux"
)

func RegisterAuthRoutes(r *mux.Router) {
	r.HandleFunc("/auth/login", controllers.Login).Methods("POST", "OPTIONS")
}
