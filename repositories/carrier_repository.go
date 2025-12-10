package repositories

import (
	"database/sql"
	"donapresentes/models"
	apperrors "donapresentes/errors"
)

type CarrierRepository struct {
	db *sql.DB
}

func NewCarrierRepository(db *sql.DB) *CarrierRepository {
	return &CarrierRepository{db: db}
}

// GetAll retrieves all transportadoras
func (r *CarrierRepository) GetAll() ([]models.Carrier, error) {
	rows, err := r.db.Query("SELECT id, nome_transportadora, tipo_transportadora, email, telefone_fixo, celular, endereco_completo, contato_principal_nome, contato_principal_telefone, site, created_at, updated_at FROM transportadoras")
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()
	var carriers []models.Carrier
	for rows.Next() {
		var t models.Carrier
		err := rows.Scan(&t.ID, &t.Name, &t.CarrierType, &t.Email, &t.LandlinePhone, &t.MobilePhone, &t.FullAddress, &t.ContactName, &t.ContactPhone, &t.Website, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		carriers = append(carriers, t)
	}
	return carriers, nil
}

// GetByID retrieves a single transportadora by ID
func (r *CarrierRepository) GetByID(id int) (*models.Carrier, error) {
	var t models.Carrier
	err := r.db.QueryRow("SELECT id, nome_transportadora, tipo_transportadora, email, telefone_fixo, celular, endereco_completo, contato_principal_nome, contato_principal_telefone, site, created_at, updated_at FROM transportadoras WHERE id=$1", id).Scan(&t.ID, &t.Name, &t.CarrierType, &t.Email, &t.LandlinePhone, &t.MobilePhone, &t.FullAddress, &t.ContactName, &t.ContactPhone, &t.Website, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Create creates a new transportadora
func (r *CarrierRepository) Create(t *models.Carrier) error {
	err := r.db.QueryRow(
		"INSERT INTO transportadoras (nome_transportadora, tipo_transportadora, email, telefone_fixo, celular, endereco_completo, contato_principal_nome, contato_principal_telefone, site) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id, created_at, updated_at",
		t.Name, t.CarrierType, t.Email, t.LandlinePhone, t.MobilePhone, t.FullAddress, t.ContactName, t.ContactPhone, t.Website).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	return err
}

// Update updates an existing transportadora
func (r *CarrierRepository) Update(id int, t *models.Carrier) error {
	_, err := r.db.Exec(
		"UPDATE transportadoras SET nome_transportadora=$1, tipo_transportadora=$2, email=$3, telefone_fixo=$4, celular=$5, endereco_completo=$6, contato_principal_nome=$7, contato_principal_telefone=$8, site=$9, updated_at=NOW() WHERE id=$10",
		t.Name, t.CarrierType, t.Email, t.LandlinePhone, t.MobilePhone, t.FullAddress, t.ContactName, t.ContactPhone, t.Website, id)
	return err
}

// Delete deletes a transportadora by ID
func (r *CarrierRepository) Delete(id int) error {
	res, err := r.db.Exec("DELETE FROM transportadoras WHERE id=$1", id)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
