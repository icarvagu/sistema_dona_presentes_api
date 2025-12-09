package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"donapresentes/controllers/config"
	apperrors "donapresentes/errors"
	"donapresentes/middleware"
	"donapresentes/models"
	"donapresentes/repositories"

	"github.com/gorilla/mux"
)

var funcionarioRepo *repositories.FuncionarioRepository

func InitFuncionarioRepository() {
	funcionarioRepo = repositories.NewFuncionarioRepository(config.DB)
}

// GetFuncionarios retrieves all funcionarios
func GetFuncionarios(w http.ResponseWriter, r *http.Request) {
	funcionarios, err := funcionarioRepo.GetAll()
	if err != nil {
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(funcionarios)
}

// GetFuncionario retrieves a single funcionario by ID
func GetFuncionario(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidID, http.StatusBadRequest)
		return
	}

	f, err := funcionarioRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.ErrorHandler(w, apperrors.ErrFuncionarioNotFound, http.StatusNotFound)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(f)
}

// CreateFuncionario creates a new funcionario
func CreateFuncionario(w http.ResponseWriter, r *http.Request) {
	var f models.Funcionario
	err := json.NewDecoder(r.Body).Decode(&f)
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	err = funcionarioRepo.Create(&f)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.ErrorHandler(w, appErr, appErr.Code)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(f)
}

// UpdateFuncionario updates an existing funcionario
func UpdateFuncionario(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidID, http.StatusBadRequest)
		return
	}

	var f models.Funcionario
	err = json.NewDecoder(r.Body).Decode(&f)
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	err = funcionarioRepo.Update(id, &f)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.ErrorHandler(w, appErr, appErr.Code)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	f.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(f)
}

// DeleteFuncionario deletes a funcionario by ID
func DeleteFuncionario(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidID, http.StatusBadRequest)
		return
	}

	err = funcionarioRepo.Delete(id)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.ErrorHandler(w, apperrors.ErrFuncionarioNotFound, http.StatusNotFound)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
