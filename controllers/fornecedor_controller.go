package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"donapresentes/controllers/config"
	"donapresentes/models"

	"github.com/gorilla/mux"
)

// GetFornecedores retrieves all fornecedores
func GetFornecedores(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query("SELECT id, nome_fantasia_ou_razao_social, cnpj, inscricao_estadual, responsavel_atendimento, email_geral, telefone_fixo, celular, email_responsavel, endereco_comercial, criado_em, atualizado_em FROM fornecedores")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var fornecedores []models.Fornecedor
	for rows.Next() {
		var f models.Fornecedor
		err := rows.Scan(&f.ID, &f.FantasyName, &f.CNPJ, &f.StateRegistration, &f.ContactResponsible, &f.GeneralEmail, &f.LandlinePhone, &f.MobilePhone, &f.ResponsibleEmail, &f.CommercialAddress, &f.CreatedAt, &f.UpdatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fornecedores = append(fornecedores, f)
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

	var f models.Fornecedor
	err = config.DB.QueryRow("SELECT id, nome_fantasia_ou_razao_social, cnpj, inscricao_estadual, responsavel_atendimento, email_geral, telefone_fixo, celular, email_responsavel, endereco_comercial, criado_em, atualizado_em FROM fornecedores WHERE id=$1", id).Scan(&f.ID, &f.FantasyName, &f.CNPJ, &f.StateRegistration, &f.ContactResponsible, &f.GeneralEmail, &f.LandlinePhone, &f.MobilePhone, &f.ResponsibleEmail, &f.CommercialAddress, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		http.Error(w, "Fornecedor não encontrado", http.StatusNotFound)
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

	err = config.DB.QueryRow(
		"INSERT INTO fornecedores ( nome_fantasia_ou_razao_social, cnpj, inscricao_estadual, responsavel_atendimento, email_geral, telefone_fixo, celular, email_responsavel, endereco_comercial) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id, criado_em, atualizado_em",
		f.FantasyName, f.CNPJ, f.StateRegistration, f.ContactResponsible, f.GeneralEmail, f.LandlinePhone, f.MobilePhone, f.ResponsibleEmail, f.CommercialAddress).Scan(&f.ID, &f.CreatedAt, &f.UpdatedAt)

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

	_, err = config.DB.Exec(
		"UPDATE fornecedores SET  nome_fantasia_ou_razao_social=$1, cnpj=$2, inscricao_estadual=$3, responsavel_atendimento=$4, email_geral=$5, telefone_fixo=$6, celular=$7, email_responsavel=$8, endereco_comercial=$9, atualizado_em=NOW() WHERE id=$10",
		f.FantasyName, f.CNPJ, f.StateRegistration, f.ContactResponsible, f.GeneralEmail, f.LandlinePhone, f.MobilePhone, f.ResponsibleEmail, f.CommercialAddress, id)

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

	result, err := config.DB.Exec("DELETE FROM fornecedores WHERE id=$1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Fornecedor não encontrado", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
