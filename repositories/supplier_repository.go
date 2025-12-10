package repositories

import (
	"database/sql"
	"donapresentes/models"
	apperrors "donapresentes/errors"
)

type SupplierRepository struct {
	db *sql.DB
}

func NewSupplierRepository(db *sql.DB) *SupplierRepository {
	return &SupplierRepository{db: db}
}

// GetAll retrieves all suppliers
func (r *SupplierRepository) GetAll() ([]models.Supplier, error) {
	rows, err := r.db.Query("SELECT id, nome_fantasia_ou_razao_social, cnpj, inscricao_estadual, responsavel_atendimento, email_geral, telefone_fixo, celular, email_responsavel, endereco_comercial, criado_em, atualizado_em FROM fornecedores")
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

// GetByID retrieves a single supplier by ID
func (r *SupplierRepository) GetByID(id int) (*models.Supplier, error) {
	var s models.Supplier
	err := r.db.QueryRow("SELECT id, nome_fantasia_ou_razao_social, cnpj, inscricao_estadual, responsavel_atendimento, email_geral, telefone_fixo, celular, email_responsavel, endereco_comercial, criado_em, atualizado_em FROM fornecedores WHERE id=$1", id).Scan(&s.ID, &s.Name, &s.CNPJ, &s.StateRegistration, &s.ContactPerson, &s.Email, &s.LandlinePhone, &s.MobilePhone, &s.ResponsibleEmail, &s.CommercialAddress, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Create creates a new supplier
func (r *SupplierRepository) Create(s *models.Supplier) error {
	err := r.db.QueryRow(
		"INSERT INTO fornecedores (nome_fantasia_ou_razao_social, cnpj, inscricao_estadual, responsavel_atendimento, email_geral, telefone_fixo, celular, email_responsavel, endereco_comercial) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id, criado_em, atualizado_em",
		s.Name, s.CNPJ, s.StateRegistration, s.ContactPerson, s.Email, s.LandlinePhone, s.MobilePhone, s.ResponsibleEmail, s.CommercialAddress).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
	return err
}

// Update updates an existing supplier
func (r *SupplierRepository) Update(id int, s *models.Supplier) error {
	_, err := r.db.Exec(
		"UPDATE fornecedores SET nome_fantasia_ou_razao_social=$1, cnpj=$2, inscricao_estadual=$3, responsavel_atendimento=$4, email_geral=$5, telefone_fixo=$6, celular=$7, email_responsavel=$8, endereco_comercial=$9, atualizado_em=NOW() WHERE id=$10",
		s.Name, s.CNPJ, s.StateRegistration, s.ContactPerson, s.Email, s.LandlinePhone, s.MobilePhone, s.ResponsibleEmail, s.CommercialAddress, id)
	return err
}

// Delete deletes a supplier by ID
func (r *SupplierRepository) Delete(id int) error {
	res, err := r.db.Exec("DELETE FROM fornecedores WHERE id=$1", id)
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
