package controllers

import (
	"encoding/json"
	"net/http"

	"donapresentes/controllers/config"
	apperrors "donapresentes/errors"
	"donapresentes/middleware"
	"donapresentes/models"
	"donapresentes/repositories"
	"donapresentes/services"
)

var authService *services.AuthService

func InitAuthService() {
	userRepo := repositories.NewUserRepository(config.DB)
	authService = services.NewAuthService(userRepo)
}

// Login autentica um usuário e retorna um token JWT
func Login(w http.ResponseWriter, r *http.Request) {
	var input models.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	// Validar campos obrigatórios
	if input.Username == "" || input.Password == "" {
		middleware.ErrorHandler(w, apperrors.NewMissingFieldError("username e password são obrigatórios"), http.StatusBadRequest)
		return
	}

	response, err := authService.Login(&input)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.ErrorHandler(w, appErr, appErr.Code)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCurrentUser retorna informações do usuário atual
func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, _, _ := middleware.GetUserFromRequest(r)
	
	user, err := authService.GetUserByID(userID)
	if err != nil {
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
