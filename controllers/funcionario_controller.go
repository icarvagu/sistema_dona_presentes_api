package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"donapresentes/controllers/config"
	"donapresentes/models"

	"github.com/gorilla/mux"
)

// GetFuncionarios retrieves all funcionarios
func GetFuncionarios(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query("SELECT id, nome_completo, cpf, rg, data_nascimento, sexo, situacao, email_contato, endereco_completo, telefones_contato, observacoes, criado_em FROM funcionarios")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var funcionarios []models.Funcionario
	for rows.Next() {
		var f models.Funcionario
		err := rows.Scan(&f.ID, &f.NomeCompleto, &f.CPF, &f.RG, &f.DataNascimento, &f.Sexo, &f.Situacao, &f.EmailContato, &f.EnderecoCompleto, &f.TelefonesContato, &f.Observacoes, &f.CriadoEm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		funcionarios = append(funcionarios, f)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(funcionarios)
}

// GetFuncionario retrieves a single funcionario by ID
func GetFuncionario(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "ID do funcionário inválido", http.StatusBadRequest)
		return
	}

	var f models.Funcionario
	err = config.DB.QueryRow("SELECT id, nome_completo, cpf, rg, data_nascimento, sexo, situacao, email_contato, endereco_completo, telefones_contato, observacoes, criado_em FROM funcionarios WHERE id=$1", id).Scan(&f.ID, &f.NomeCompleto, &f.CPF, &f.RG, &f.DataNascimento, &f.Sexo, &f.Situacao, &f.EmailContato, &f.EnderecoCompleto, &f.TelefonesContato, &f.Observacoes, &f.CriadoEm)
	if err != nil {
		http.Error(w, "Funcionário não encontrado", http.StatusNotFound)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = config.DB.QueryRow(
		"INSERT INTO funcionarios (nome_completo, cpf, rg, data_nascimento, sexo, situacao, email_contato, endereco_completo, telefones_contato, observacoes) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id, criado_em",
		f.NomeCompleto, f.CPF, f.RG, f.DataNascimento, f.Sexo, f.Situacao, f.EmailContato, f.EnderecoCompleto, f.TelefonesContato, f.Observacoes).Scan(&f.ID, &f.CriadoEm)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, "ID do funcionário inválido", http.StatusBadRequest)
		return
	}

	var f models.Funcionario
	err = json.NewDecoder(r.Body).Decode(&f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = config.DB.Exec(
		"UPDATE funcionarios SET nome_completo=$1, cpf=$2, rg=$3, data_nascimento=$4, sexo=$5, situacao=$6, email_contato=$7, endereco_completo=$8, telefones_contato=$9, observacoes=$10 WHERE id=$11",
		f.NomeCompleto, f.CPF, f.RG, f.DataNascimento, f.Sexo, f.Situacao, f.EmailContato, f.EnderecoCompleto, f.TelefonesContato, f.Observacoes, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, "ID do funcionário inválido", http.StatusBadRequest)
		return
	}

	res, err := config.DB.Exec("DELETE FROM funcionarios WHERE id=$1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, "Funcionário não encontrado", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
