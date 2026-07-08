package routes

import (
	"donapresentes/controllers"
	"donapresentes/middleware"

	"github.com/gorilla/mux"
)

func RegisterAuthRoutes(r *mux.Router) {
	loginRouter := r.PathPrefix("/auth").Subrouter()
	loginRouter.Use(middleware.RateLimitLoginMiddleware)
	loginRouter.HandleFunc("/login", controllers.Login).Methods("POST", "OPTIONS")

	r.HandleFunc("/auth/forgot-password", controllers.ForgotPassword).Methods("POST", "OPTIONS")
	r.HandleFunc("/auth/reset-password", controllers.ResetPassword).Methods("POST", "OPTIONS")
}

func RegisterAuthProtectedRoutes(r *mux.Router) {
	r.HandleFunc("/auth/me", controllers.GetCurrentUser).Methods("GET", "OPTIONS")
	r.HandleFunc("/auth/refresh", controllers.RefreshToken).Methods("POST", "OPTIONS")
	r.HandleFunc("/auth/logout", controllers.Logout).Methods("POST", "OPTIONS")
}
