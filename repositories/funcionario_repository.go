package repositories

import (
	"database/sql"
	"donapresentes/models"
	apperrors "donapresentes/errors"
)

type FuncionarioRepository struct {
	db *sql.DB
}

func NewFuncionarioRepository(db *sql.DB) *FuncionarioRepository {
	return &FuncionarioRepository{db: db}
}

// GetAll retrieves all funcionarios
func (r *FuncionarioRepository) GetAll() ([]models.Funcionario, error) {
	rows, err := r.db.Query("SELECT id, nome_completo, cpf, rg, data_nascimento, sexo, situacao, email_contato, endereco_completo, telefones_contato, observacoes, criado_em FROM funcionarios")
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	var funcionarios []models.Funcionario
	for rows.Next() {
		var f models.Funcionario
		err := rows.Scan(&f.ID, &f.NomeCompleto, &f.CPF, &f.RG, &f.DataNascimento, &f.Sexo, &f.Situacao, &f.EmailContato, &f.EnderecoCompleto, &f.TelefonesContato, &f.Observacoes, &f.CriadoEm)
		if err != nil {
			return nil, err
		}
		funcionarios = append(funcionarios, f)
	}
	return funcionarios, nil
}

// GetByID retrieves a single funcionario by ID
func (r *FuncionarioRepository) GetByID(id int) (*models.Funcionario, error) {
	var f models.Funcionario
	err := r.db.QueryRow("SELECT id, nome_completo, cpf, rg, data_nascimento, sexo, situacao, email_contato, endereco_completo, telefones_contato, observacoes, criado_em FROM funcionarios WHERE id=$1", id).Scan(&f.ID, &f.NomeCompleto, &f.CPF, &f.RG, &f.DataNascimento, &f.Sexo, &f.Situacao, &f.EmailContato, &f.EnderecoCompleto, &f.TelefonesContato, &f.Observacoes, &f.CriadoEm)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// Create creates a new funcionario
func (r *FuncionarioRepository) Create(f *models.Funcionario) error {
	err := r.db.QueryRow(
		"INSERT INTO funcionarios (nome_completo, cpf, rg, data_nascimento, sexo, situacao, email_contato, endereco_completo, telefones_contato, observacoes) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id, criado_em",
		f.NomeCompleto, f.CPF, f.RG, f.DataNascimento, f.Sexo, f.Situacao, f.EmailContato, f.EnderecoCompleto, f.TelefonesContato, f.Observacoes).Scan(&f.ID, &f.CriadoEm)
	return err
}

// Update updates an existing funcionario
func (r *FuncionarioRepository) Update(id int, f *models.Funcionario) error {
	_, err := r.db.Exec(
		"UPDATE funcionarios SET nome_completo=$1, cpf=$2, rg=$3, data_nascimento=$4, sexo=$5, situacao=$6, email_contato=$7, endereco_completo=$8, telefones_contato=$9, observacoes=$10 WHERE id=$11",
		f.NomeCompleto, f.CPF, f.RG, f.DataNascimento, f.Sexo, f.Situacao, f.EmailContato, f.EnderecoCompleto, f.TelefonesContato, f.Observacoes, id)
	return err
}

// Delete deletes a funcionario by ID
func (r *FuncionarioRepository) Delete(id int) error {
	res, err := r.db.Exec("DELETE FROM funcionarios WHERE id=$1", id)
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
