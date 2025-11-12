package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"donapresentes/controllers/config"
	"donapresentes/models"

	"github.com/gorilla/mux"
)

// GetTransportadoras lists all transportadoras
func GetTransportadoras(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query("SELECT id, nome_transportadora, tipo_transportadora, email, telefone_fixo, celular, endereco_completo, contato_principal_nome, contato_principal_telefone, site, created_at, updated_at FROM transportadoras")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []models.Transportadora
	for rows.Next() {
		var t models.Transportadora
		err := rows.Scan(&t.ID, &t.NomeTransportadora, &t.TipoTransportadora, &t.Email, &t.TelefoneFixo, &t.Celular, &t.EnderecoCompleto, &t.ContatoPrincipalNome, &t.ContatoPrincipalTelefone, &t.Site, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		items = append(items, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// GetTransportadora retrieves one transportadora by ID
func GetTransportadora(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var t models.Transportadora
	err = config.DB.QueryRow("SELECT id, nome_transportadora, tipo_transportadora, email, telefone_fixo, celular, endereco_completo, contato_principal_nome, contato_principal_telefone, site, created_at, updated_at FROM transportadoras WHERE id=$1", id).Scan(&t.ID, &t.NomeTransportadora, &t.TipoTransportadora, &t.Email, &t.TelefoneFixo, &t.Celular, &t.EnderecoCompleto, &t.ContatoPrincipalNome, &t.ContatoPrincipalTelefone, &t.Site, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

// CreateTransportadora creates a new transportadora
func CreateTransportadora(w http.ResponseWriter, r *http.Request) {
	var t models.Transportadora
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := config.DB.QueryRow(
		"INSERT INTO transportadoras (nome_transportadora, tipo_transportadora, email, telefone_fixo, celular, endereco_completo, contato_principal_nome, contato_principal_telefone, site) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id, created_at, updated_at",
		t.NomeTransportadora, t.TipoTransportadora, t.Email, t.TelefoneFixo, t.Celular, t.EnderecoCompleto, t.ContatoPrincipalNome, t.ContatoPrincipalTelefone, t.Site,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

// UpdateTransportadora updates an existing transportadora
func UpdateTransportadora(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var t models.Transportadora
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = config.DB.Exec(
		"UPDATE transportadoras SET nome_transportadora=$1, tipo_transportadora=$2, email=$3, telefone_fixo=$4, celular=$5, endereco_completo=$6, contato_principal_nome=$7, contato_principal_telefone=$8, site=$9, updated_at=NOW() WHERE id=$10",
		t.NomeTransportadora, t.TipoTransportadora, t.Email, t.TelefoneFixo, t.Celular, t.EnderecoCompleto, t.ContatoPrincipalNome, t.ContatoPrincipalTelefone, t.Site, id,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

// DeleteTransportadora deletes a transportadora by ID
func DeleteTransportadora(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	res, err := config.DB.Exec("DELETE FROM transportadoras WHERE id=$1", id)
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
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
