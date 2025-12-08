package repositories

import (
	"database/sql"
	"donapresentes/models"
)

type TransportadoraRepository struct {
	db *sql.DB
}

func NewTransportadoraRepository(db *sql.DB) *TransportadoraRepository {
	return &TransportadoraRepository{db: db}
}

// GetAll retrieves all transportadoras
func (r *TransportadoraRepository) GetAll() ([]models.Transportadora, error) {
	rows, err := r.db.Query("SELECT id, nome_transportadora, tipo_transportadora, email, telefone_fixo, celular, endereco_completo, contato_principal_nome, contato_principal_telefone, site, created_at, updated_at FROM transportadoras")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transportadoras []models.Transportadora
	for rows.Next() {
		var t models.Transportadora
		err := rows.Scan(&t.ID, &t.NomeTransportadora, &t.TipoTransportadora, &t.Email, &t.TelefoneFixo, &t.Celular, &t.EnderecoCompleto, &t.ContatoPrincipalNome, &t.ContatoPrincipalTelefone, &t.Site, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		transportadoras = append(transportadoras, t)
	}
	return transportadoras, nil
}

// GetByID retrieves a single transportadora by ID
func (r *TransportadoraRepository) GetByID(id int) (*models.Transportadora, error) {
	var t models.Transportadora
	err := r.db.QueryRow("SELECT id, nome_transportadora, tipo_transportadora, email, telefone_fixo, celular, endereco_completo, contato_principal_nome, contato_principal_telefone, site, created_at, updated_at FROM transportadoras WHERE id=$1", id).Scan(&t.ID, &t.NomeTransportadora, &t.TipoTransportadora, &t.Email, &t.TelefoneFixo, &t.Celular, &t.EnderecoCompleto, &t.ContatoPrincipalNome, &t.ContatoPrincipalTelefone, &t.Site, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Create creates a new transportadora
func (r *TransportadoraRepository) Create(t *models.Transportadora) error {
	err := r.db.QueryRow(
		"INSERT INTO transportadoras (nome_transportadora, tipo_transportadora, email, telefone_fixo, celular, endereco_completo, contato_principal_nome, contato_principal_telefone, site) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id, created_at, updated_at",
		t.NomeTransportadora, t.TipoTransportadora, t.Email, t.TelefoneFixo, t.Celular, t.EnderecoCompleto, t.ContatoPrincipalNome, t.ContatoPrincipalTelefone, t.Site).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	return err
}

// Update updates an existing transportadora
func (r *TransportadoraRepository) Update(id int, t *models.Transportadora) error {
	_, err := r.db.Exec(
		"UPDATE transportadoras SET nome_transportadora=$1, tipo_transportadora=$2, email=$3, telefone_fixo=$4, celular=$5, endereco_completo=$6, contato_principal_nome=$7, contato_principal_telefone=$8, site=$9, updated_at=NOW() WHERE id=$10",
		t.NomeTransportadora, t.TipoTransportadora, t.Email, t.TelefoneFixo, t.Celular, t.EnderecoCompleto, t.ContatoPrincipalNome, t.ContatoPrincipalTelefone, t.Site, id)
	return err
}

// Delete deletes a transportadora by ID
func (r *TransportadoraRepository) Delete(id int) error {
	res, err := r.db.Exec("DELETE FROM transportadoras WHERE id=$1", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
