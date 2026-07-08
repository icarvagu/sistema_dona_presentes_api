package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"donapresentes/controllers/config"
	apperrors "donapresentes/errors"
	"donapresentes/middleware"
	"donapresentes/models"
	"donapresentes/repositories"
	"donapresentes/services"

	"github.com/gorilla/mux"
)

var userService *services.UserService

func InitUserService() {
	userRepo := repositories.NewUserRepository(config.DB)
	refreshRepo := repositories.NewRefreshTokenRepository(config.DB)
	auditService := services.NewAuditService(config.DB)
	authService := services.NewAuthService(userRepo, refreshRepo, auditService)
	cryptoService := services.NewCryptoService()
	userService = services.NewUserService(userRepo, authService, cryptoService)
}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := userService.GetAll()
	if err != nil {
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidID, http.StatusBadRequest)
		return
	}

	userID, _, _ := middleware.GetUserFromRequest(r)
	if userID != id {

		_, _, role := middleware.GetUserFromRequest(r)
		if role != "admin" {
			middleware.ErrorHandler(w, apperrors.NewNotFoundError("Usuário não encontrado"), http.StatusNotFound)
			return
		}
	}

	user, err := userService.GetByID(id)
	if err != nil {
		if apperrors.IsNotFound(err) {
			middleware.ErrorHandler(w, apperrors.NewNotFoundError("Usuário não encontrado"), http.StatusNotFound)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var input models.UserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	user, err := userService.Create(&input)
	if err != nil {
		if apperrors.IsValidationError(err) {
			middleware.ErrorHandler(w, err, http.StatusBadRequest)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidID, http.StatusBadRequest)
		return
	}

	userID, _, _ := middleware.GetUserFromRequest(r)
	if userID != id {

		_, _, role := middleware.GetUserFromRequest(r)
		if role != "admin" {
			middleware.ErrorHandler(w, apperrors.NewNotFoundError("Usuário não encontrado"), http.StatusNotFound)
			return
		}
	}

	var input models.UserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	user, err := userService.Update(id, &input)
	if err != nil {
		if apperrors.IsNotFound(err) {
			middleware.ErrorHandler(w, apperrors.NewNotFoundError("Usuário não encontrado"), http.StatusNotFound)
			return
		}
		if apperrors.IsValidationError(err) {
			middleware.ErrorHandler(w, err, http.StatusBadRequest)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidID, http.StatusBadRequest)
		return
	}

	err = userService.Delete(id)
	if err != nil {
		if apperrors.IsNotFound(err) {
			middleware.ErrorHandler(w, apperrors.NewNotFoundError("Usuário não encontrado"), http.StatusNotFound)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, _, _ := middleware.GetUserFromRequest(r)

	var input models.UserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	user, err := userService.GetByID(userID)
	if err != nil {
		middleware.ErrorHandler(w, apperrors.NewNotFoundError("Usuário não encontrado"), http.StatusNotFound)
		return
	}
	input.Role = user.Role

	updatedUser, err := userService.Update(userID, &input)
	if err != nil {
		if apperrors.IsNotFound(err) {
			middleware.ErrorHandler(w, apperrors.NewNotFoundError("Usuário não encontrado"), http.StatusNotFound)
			return
		}
		if apperrors.IsValidationError(err) {
			middleware.ErrorHandler(w, err, http.StatusBadRequest)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedUser)
}
