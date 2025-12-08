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

var produtoRepo *repositories.ProdutoRepository

func InitProdutoRepository() {
	produtoRepo = repositories.NewProdutoRepository(config.DB)
}

// List produtos
func GetProdutos(w http.ResponseWriter, r *http.Request) {
	ps, err := produtoRepo.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ps)
}

// Get produto by id
func GetProduto(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	p, err := produtoRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// Create produto with validation
func CreateProduto(w http.ResponseWriter, r *http.Request) {
	var p models.Produto
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// basic validation
	if p.NomeProduto == "" || p.CodigoInterno == "" || p.CodigoFornecedor == 0 {
		http.Error(w, "nome_produto, codigo_interno and codigo_fornecedor are required", http.StatusBadRequest)
		return
	}
	if p.Estoque < 0 {
		http.Error(w, "estoque cannot be negative", http.StatusBadRequest)
		return
	}

	if err := produtoRepo.Create(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

// Update produto
func UpdateProduto(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var p models.Produto
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if p.NomeProduto == "" || p.CodigoInterno == "" || p.CodigoFornecedor == 0 {
		http.Error(w, "nome_produto, codigo_interno and codigo_fornecedor are required", http.StatusBadRequest)
		return
	}
	if p.Estoque < 0 {
		http.Error(w, "estoque cannot be negative", http.StatusBadRequest)
		return
	}
	if err := produtoRepo.Update(id, &p); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	p.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// Delete produto
func DeleteProduto(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := produtoRepo.Delete(id); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
