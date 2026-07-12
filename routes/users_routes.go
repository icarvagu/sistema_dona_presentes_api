package routes

import (
	"donapresentes/controllers"

	"github.com/gorilla/mux"
)

// RegisterUsersRoutes registers all user-related HTTP endpoints on the given router.
// Routes: GET /users (list), PUT /users/profile, GET /users/{id}, PUT /users/{id}.
func RegisterUsersRoutes(router *mux.Router) {
	controllers.InitUserService()

	router.HandleFunc("/users", controllers.GetAllUsers).Methods("GET", "OPTIONS")
	router.HandleFunc("/users/profile", controllers.UpdateProfile).Methods("PUT", "OPTIONS")
	router.HandleFunc("/users/{id}", controllers.GetUser).Methods("GET", "OPTIONS")
	router.HandleFunc("/users/{id}", controllers.UpdateUser).Methods("PUT", "OPTIONS")
}
