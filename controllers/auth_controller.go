package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"

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
	refreshRepo := repositories.NewRefreshTokenRepository(config.DB)
	auditService := services.NewAuditService(config.DB)
	authService = services.NewAuthService(userRepo, refreshRepo, auditService)
}

func setTokenCookie(w http.ResponseWriter, token string) {
	secure := os.Getenv("ENABLE_HTTPS") == "true"
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400,
	})
}

func clearTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   os.Getenv("ENABLE_HTTPS") == "true",
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

func getClientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return fwd
	}
	return r.RemoteAddr
}

func Login(w http.ResponseWriter, r *http.Request) {
	var input models.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	if input.Username == "" || input.Password == "" {
		middleware.ErrorHandler(w, apperrors.NewMissingFieldError("username e password são obrigatórios"), http.StatusBadRequest)
		return
	}

	ipAddress := getClientIP(r)
	response, err := authService.Login(&input, ipAddress)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.ErrorHandler(w, appErr, appErr.Code)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	setTokenCookie(w, response.Token)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	var input models.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	if input.RefreshToken == "" {
		middleware.ErrorHandler(w, apperrors.NewMissingFieldError("refresh_token é obrigatório"), http.StatusBadRequest)
		return
	}

	ipAddress := getClientIP(r)
	response, err := authService.RefreshAccessToken(input.RefreshToken, ipAddress)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.ErrorHandler(w, appErr, appErr.Code)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	setTokenCookie(w, response.Token)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	userID, _, _ := middleware.GetUserFromRequest(r)

	if userID > 0 {
		ipAddress := getClientIP(r)
		_ = authService.Logout(userID, ipAddress)
	}

	clearTokenCookie(w)

	w.WriteHeader(http.StatusNoContent)
}

func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, _, _ := middleware.GetUserFromRequest(r)

	user, err := authService.GetUserByID(userID)
	if err != nil {
		if err == sql.ErrNoRows || apperrors.IsNotFound(err) {
			middleware.ErrorHandler(w, apperrors.NewUnauthorizedError("Token inválido ou usuário não encontrado"), http.StatusUnauthorized)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var input models.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	if input.Username == "" {
		middleware.ErrorHandler(w, apperrors.NewMissingFieldError("username é obrigatório"), http.StatusBadRequest)
		return
	}

	token, err := authService.InitiatePasswordReset(input.Username)
	if err != nil {
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func ResetPassword(w http.ResponseWriter, r *http.Request) {
	var input models.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	if input.Token == "" || input.NewPassword == "" {
		middleware.ErrorHandler(w, apperrors.NewMissingFieldError("token e new_password são obrigatórios"), http.StatusBadRequest)
		return
	}

	err := authService.ResetPassword(input.Token, input.NewPassword)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.ErrorHandler(w, appErr, appErr.Code)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Senha alterada com sucesso"})
}
