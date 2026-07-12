package routes

import (
	"donapresentes/controllers"
	"donapresentes/middleware"

	"github.com/gorilla/mux"
)

// RegisterAuthRoutes registers public authentication endpoints on the given router.
// Routes: POST /auth/login, POST /auth/forgot-password, POST /auth/reset-password.
func RegisterAuthRoutes(r *mux.Router) {
	loginRouter := r.PathPrefix("/auth").Subrouter()
	loginRouter.Use(middleware.RateLimitLoginMiddleware)
	loginRouter.HandleFunc("/login", controllers.Login).Methods("POST", "OPTIONS")

	r.HandleFunc("/auth/forgot-password", controllers.ForgotPassword).Methods("POST", "OPTIONS")
	r.HandleFunc("/auth/reset-password", controllers.ResetPassword).Methods("POST", "OPTIONS")
}

// RegisterAuthProtectedRoutes registers protected authentication endpoints on the given router.
// Routes: GET /auth/me, POST /auth/refresh, POST /auth/logout, PUT /auth/change-password.
func RegisterAuthProtectedRoutes(r *mux.Router) {
	r.HandleFunc("/auth/me", controllers.GetCurrentUser).Methods("GET", "OPTIONS")
	r.HandleFunc("/auth/refresh", controllers.RefreshToken).Methods("POST", "OPTIONS")
	r.HandleFunc("/auth/logout", controllers.Logout).Methods("POST", "OPTIONS")
	r.HandleFunc("/auth/change-password", controllers.ChangePassword).Methods("PUT", "OPTIONS")
}
