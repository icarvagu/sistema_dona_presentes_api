package repositories

import (
	"database/sql"
	apperrors "donapresentes/errors"
	"donapresentes/models"
)

// CarrierRepository handles all database operations for the carriers table.
type CarrierRepository struct {
	db *sql.DB
}

// NewCarrierRepository creates a new CarrierRepository with the given database connection.
func NewCarrierRepository(db *sql.DB) *CarrierRepository {
	return &CarrierRepository{db: db}
}

// GetAll returns all carriers from the database.
func (r *CarrierRepository) GetAll() ([]models.Carrier, error) {
	rows, err := r.db.Query("SELECT id, name, cnpj, carrier_type, email, landline_phone, mobile_phone, full_address, contact_name, contact_phone, website, created_at, updated_at FROM carriers")
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()
	var carriers []models.Carrier
	for rows.Next() {
		var t models.Carrier
		err := rows.Scan(&t.ID, &t.Name, &t.CNPJ, &t.CarrierType, &t.Email, &t.LandlinePhone, &t.MobilePhone, &t.FullAddress, &t.ContactName, &t.ContactPhone, &t.Website, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		carriers = append(carriers, t)
	}
	return carriers, nil
}

// GetByID returns a single carrier by its primary key.
func (r *CarrierRepository) GetByID(id int) (*models.Carrier, error) {
	var t models.Carrier
	err := r.db.QueryRow("SELECT id, name, cnpj, carrier_type, email, landline_phone, mobile_phone, full_address, contact_name, contact_phone, website, created_at, updated_at FROM carriers WHERE id=$1", id).Scan(&t.ID, &t.Name, &t.CNPJ, &t.CarrierType, &t.Email, &t.LandlinePhone, &t.MobilePhone, &t.FullAddress, &t.ContactName, &t.ContactPhone, &t.Website, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Create inserts a new carrier record and returns the created carrier with its ID and timestamps.
func (r *CarrierRepository) Create(t *models.Carrier) error {
	err := r.db.QueryRow(
		"INSERT INTO carriers (name, cnpj, carrier_type, email, landline_phone, mobile_phone, full_address, contact_name, contact_phone, website) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id, created_at, updated_at",
		t.Name, t.CNPJ, t.CarrierType, t.Email, t.LandlinePhone, t.MobilePhone, t.FullAddress, t.ContactName, t.ContactPhone, t.Website).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	return err
}

// Update modifies an existing carrier record identified by id with the provided data.
func (r *CarrierRepository) Update(id int, t *models.Carrier) error {
	err := r.db.QueryRow(
		"UPDATE carriers SET name=$1, cnpj=$2, carrier_type=$3, email=$4, landline_phone=$5, mobile_phone=$6, full_address=$7, contact_name=$8, contact_phone=$9, website=$10, updated_at=NOW() WHERE id=$11 RETURNING updated_at",
		t.Name, t.CNPJ, t.CarrierType, t.Email, t.LandlinePhone, t.MobilePhone, t.FullAddress, t.ContactName, t.ContactPhone, t.Website, id).Scan(&t.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return sql.ErrNoRows
		}
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

// Delete removes a carrier record by its primary key.
func (r *CarrierRepository) Delete(id int) error {
	res, err := r.db.Exec("DELETE FROM carriers WHERE id=$1", id)
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
