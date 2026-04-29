package repositories

import (
	"database/sql"
	apperrors "donapresentes/errors"
	"donapresentes/models"
)

type SupplierRepository struct {
	db *sql.DB
}

func NewSupplierRepository(db *sql.DB) *SupplierRepository {
	return &SupplierRepository{db: db}
}

func (r *SupplierRepository) GetAll() ([]models.Supplier, error) {
	rows, err := r.db.Query("SELECT id, name, cnpj, state_registration, contact_person, email, landline_phone, mobile_phone, responsible_email, commercial_address, created_at, updated_at FROM suppliers")
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	var suppliers []models.Supplier
	for rows.Next() {
		var s models.Supplier
		err := rows.Scan(&s.ID, &s.Name, &s.CNPJ, &s.StateRegistration, &s.ContactPerson, &s.Email, &s.LandlinePhone, &s.MobilePhone, &s.ResponsibleEmail, &s.CommercialAddress, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, err
		}
		suppliers = append(suppliers, s)
	}
	return suppliers, nil
}

func (r *SupplierRepository) GetByID(id int) (*models.Supplier, error) {
	var s models.Supplier
	err := r.db.QueryRow("SELECT id, name, cnpj, state_registration, contact_person, email, landline_phone, mobile_phone, responsible_email, commercial_address, created_at, updated_at FROM suppliers WHERE id=$1", id).Scan(&s.ID, &s.Name, &s.CNPJ, &s.StateRegistration, &s.ContactPerson, &s.Email, &s.LandlinePhone, &s.MobilePhone, &s.ResponsibleEmail, &s.CommercialAddress, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SupplierRepository) Create(s *models.Supplier) error {
	err := r.db.QueryRow(
		"INSERT INTO suppliers (name, cnpj, state_registration, contact_person, email, landline_phone, mobile_phone, responsible_email, commercial_address) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id, created_at, updated_at",
		s.Name, s.CNPJ, s.StateRegistration, s.ContactPerson, s.Email, s.LandlinePhone, s.MobilePhone, s.ResponsibleEmail, s.CommercialAddress).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
	return err
}

func (r *SupplierRepository) Update(id int, s *models.Supplier) error {
	_, err := r.db.Exec(
		"UPDATE suppliers SET name=$1, cnpj=$2, state_registration=$3, contact_person=$4, email=$5, landline_phone=$6, mobile_phone=$7, responsible_email=$8, commercial_address=$9, updated_at=NOW() WHERE id=$10",
		s.Name, s.CNPJ, s.StateRegistration, s.ContactPerson, s.Email, s.LandlinePhone, s.MobilePhone, s.ResponsibleEmail, s.CommercialAddress, id)
	return err
}

func (r *SupplierRepository) Delete(id int) error {
	res, err := r.db.Exec("DELETE FROM suppliers WHERE id=$1", id)
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
