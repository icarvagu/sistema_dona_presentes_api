package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"donapresentes/controllers/config"
	"donapresentes/models"
	"donapresentes/repositories"

	"github.com/gorilla/mux"
)

var fornecedorRepo *repositories.FornecedorRepository

func InitFornecedorRepository() {
	fornecedorRepo = repositories.NewFornecedorRepository(config.DB)
}

// GetFornecedores retrieves all fornecedores
func GetFornecedores(w http.ResponseWriter, r *http.Request) {
	fornecedores, err := fornecedorRepo.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fornecedores)
}

// GetFornecedor retrieves a single fornecedor by ID
func GetFornecedor(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "ID do fornecedor inválido", http.StatusBadRequest)
		return
	}

	f, err := fornecedorRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Fornecedor não encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(f)
}

// CreateFornecedor creates a new fornecedor
func CreateFornecedor(w http.ResponseWriter, r *http.Request) {
	var f models.Fornecedor
	err := json.NewDecoder(r.Body).Decode(&f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = fornecedorRepo.Create(&f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(f)
}

// UpdateFornecedor updates an existing fornecedor
func UpdateFornecedor(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "ID do fornecedor inválido", http.StatusBadRequest)
		return
	}

	var f models.Fornecedor
	err = json.NewDecoder(r.Body).Decode(&f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = fornecedorRepo.Update(id, &f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	f.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(f)
}

// DeleteFornecedor deletes a fornecedor by ID
func DeleteFornecedor(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "ID do fornecedor inválido", http.StatusBadRequest)
		return
	}

	err = fornecedorRepo.Delete(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Fornecedor não encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
