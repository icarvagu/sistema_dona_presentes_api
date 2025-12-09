package repositories

import (
	"database/sql"
	"donapresentes/models"
	apperrors "donapresentes/errors"
)

type FornecedorRepository struct {
	db *sql.DB
}

func NewFornecedorRepository(db *sql.DB) *FornecedorRepository {
	return &FornecedorRepository{db: db}
}

// GetAll retrieves all fornecedores
func (r *FornecedorRepository) GetAll() ([]models.Fornecedor, error) {
	rows, err := r.db.Query("SELECT id, nome_fantasia_ou_razao_social, cnpj, inscricao_estadual, responsavel_atendimento, email_geral, telefone_fixo, celular, email_responsavel, endereco_comercial, criado_em, atualizado_em FROM fornecedores")
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	var fornecedores []models.Fornecedor
	for rows.Next() {
		var f models.Fornecedor
		err := rows.Scan(&f.ID, &f.FantasyName, &f.CNPJ, &f.StateRegistration, &f.ContactResponsible, &f.GeneralEmail, &f.LandlinePhone, &f.MobilePhone, &f.ResponsibleEmail, &f.CommercialAddress, &f.CreatedAt, &f.UpdatedAt)
		if err != nil {
			return nil, err
		}
		fornecedores = append(fornecedores, f)
	}
	return fornecedores, nil
}

// GetByID retrieves a single fornecedor by ID
func (r *FornecedorRepository) GetByID(id int) (*models.Fornecedor, error) {
	var f models.Fornecedor
	err := r.db.QueryRow("SELECT id, nome_fantasia_ou_razao_social, cnpj, inscricao_estadual, responsavel_atendimento, email_geral, telefone_fixo, celular, email_responsavel, endereco_comercial, criado_em, atualizado_em FROM fornecedores WHERE id=$1", id).Scan(&f.ID, &f.FantasyName, &f.CNPJ, &f.StateRegistration, &f.ContactResponsible, &f.GeneralEmail, &f.LandlinePhone, &f.MobilePhone, &f.ResponsibleEmail, &f.CommercialAddress, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// Create creates a new fornecedor and returns its ID and CreatedAt timestamp
func (r *FornecedorRepository) Create(f *models.Fornecedor) error {
	err := r.db.QueryRow(
		"INSERT INTO fornecedores (nome_fantasia_ou_razao_social, cnpj, inscricao_estadual, responsavel_atendimento, email_geral, telefone_fixo, celular, email_responsavel, endereco_comercial) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id, criado_em, atualizado_em",
		f.FantasyName, f.CNPJ, f.StateRegistration, f.ContactResponsible, f.GeneralEmail, f.LandlinePhone, f.MobilePhone, f.ResponsibleEmail, f.CommercialAddress).Scan(&f.ID, &f.CreatedAt, &f.UpdatedAt)
	return err
}

// Update updates an existing fornecedor
func (r *FornecedorRepository) Update(id int, f *models.Fornecedor) error {
	_, err := r.db.Exec(
		"UPDATE fornecedores SET nome_fantasia_ou_razao_social=$1, cnpj=$2, inscricao_estadual=$3, responsavel_atendimento=$4, email_geral=$5, telefone_fixo=$6, celular=$7, email_responsavel=$8, endereco_comercial=$9, atualizado_em=NOW() WHERE id=$10",
		f.FantasyName, f.CNPJ, f.StateRegistration, f.ContactResponsible, f.GeneralEmail, f.LandlinePhone, f.MobilePhone, f.ResponsibleEmail, f.CommercialAddress, id)
	return err
}

// Delete deletes a fornecedor by ID
func (r *FornecedorRepository) Delete(id int) error {
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
