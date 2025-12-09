package repositories

import (
	"database/sql"
	"donapresentes/models"
	apperrors "donapresentes/errors"
	"fmt"

	"github.com/lib/pq"
)

type ProdutoRepository struct {
	db *sql.DB
}

func NewProdutoRepository(db *sql.DB) *ProdutoRepository {
	return &ProdutoRepository{db: db}
}

// GetAll returns all produtos with nested fornecedor
func (r *ProdutoRepository) GetAll() ([]models.Produto, error) {
	rows, err := r.db.Query(`SELECT p.id, p.nome_produto, p.codigo_interno, p.codigo_fornecedor, f.id, f.nome_fantasia_ou_razao_social, f.cnpj, f.inscricao_estadual, f.responsavel_atendimento, f.email_geral, f.telefone_fixo, f.celular, f.email_responsavel, f.endereco_comercial, f.criado_em, f.atualizado_em, p.grupo_produto, p.descricao, p.fotos, p.ncm, p.origem_material, p.estoque, p.criado_em, p.atualizado_em FROM produtos p JOIN fornecedores f ON p.codigo_fornecedor = f.id`)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	var res []models.Produto
	for rows.Next() {
		var p models.Produto
		var f models.Fornecedor
		var fotos pq.StringArray
		err := rows.Scan(&p.ID, &p.NomeProduto, &p.CodigoInterno, &p.CodigoFornecedor, &f.ID, &f.FantasyName, &f.CNPJ, &f.StateRegistration, &f.ContactResponsible, &f.GeneralEmail, &f.LandlinePhone, &f.MobilePhone, &f.ResponsibleEmail, &f.CommercialAddress, &f.CreatedAt, &f.UpdatedAt, &p.GrupoProduto, &p.Descricao, &fotos, &p.NCM, &p.OrigemMaterial, &p.Estoque, &p.CriadoEm, &p.AtualizadoEm)
		if err != nil {
			return nil, err
		}
		p.Fotos = []string(fotos)
		p.Fornecedor = &f
		res = append(res, p)
	}
	return res, nil
}

// GetByID returns one produto by id with nested fornecedor
func (r *ProdutoRepository) GetByID(id int) (*models.Produto, error) {
	var p models.Produto
	var f models.Fornecedor
	var fotos pq.StringArray
	err := r.db.QueryRow(`SELECT p.id, p.nome_produto, p.codigo_interno, p.codigo_fornecedor, f.id, f.nome_fantasia_ou_razao_social, f.cnpj, f.inscricao_estadual, f.responsavel_atendimento, f.email_geral, f.telefone_fixo, f.celular, f.email_responsavel, f.endereco_comercial, f.criado_em, f.atualizado_em, p.grupo_produto, p.descricao, p.fotos, p.ncm, p.origem_material, p.estoque, p.criado_em, p.atualizado_em FROM produtos p JOIN fornecedores f ON p.codigo_fornecedor = f.id WHERE p.id=$1`, id).
		Scan(&p.ID, &p.NomeProduto, &p.CodigoInterno, &p.CodigoFornecedor, &f.ID, &f.FantasyName, &f.CNPJ, &f.StateRegistration, &f.ContactResponsible, &f.GeneralEmail, &f.LandlinePhone, &f.MobilePhone, &f.ResponsibleEmail, &f.CommercialAddress, &f.CreatedAt, &f.UpdatedAt, &p.GrupoProduto, &p.Descricao, &fotos, &p.NCM, &p.OrigemMaterial, &p.Estoque, &p.CriadoEm, &p.AtualizadoEm)
	if err != nil {
		return nil, err
	}
	p.Fotos = []string(fotos)
	p.Fornecedor = &f
	return &p, nil
}

// Create inserts a new produto. Validates fornecedor existence.
func (r *ProdutoRepository) Create(p *models.Produto) error {
	// check fornecedor exists
	var tmp int
	if err := r.db.QueryRow("SELECT id FROM fornecedores WHERE id=$1", p.CodigoFornecedor).Scan(&tmp); err != nil {
		if err == sql.ErrNoRows {
			return apperrors.ErrFornecedorNotFound
		}
		return apperrors.NewDatabaseError(err)
	}

	err := r.db.QueryRow(`INSERT INTO produtos (nome_produto, codigo_interno, codigo_fornecedor, grupo_produto, descricao, fotos, ncm, origem_material, estoque) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id, criado_em, atualizado_em`,
		p.NomeProduto, p.CodigoInterno, p.CodigoFornecedor, p.GrupoProduto, p.Descricao, pq.Array(p.Fotos), p.NCM, p.OrigemMaterial, p.Estoque).
		Scan(&p.ID, &p.CriadoEm, &p.AtualizadoEm)
	return err
}

// Update updates an existing produto
func (r *ProdutoRepository) Update(id int, p *models.Produto) error {
	// ensure fornecedor exists
	var tmp int
	if err := r.db.QueryRow("SELECT id FROM fornecedores WHERE id=$1", p.CodigoFornecedor).Scan(&tmp); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("fornecedor not found")
		}
		return err
	}

	_, err := r.db.Exec(`UPDATE produtos SET nome_produto=$1, codigo_interno=$2, codigo_fornecedor=$3, grupo_produto=$4, descricao=$5, fotos=$6, ncm=$7, origem_material=$8, estoque=$9, atualizado_em=NOW() WHERE id=$10`,
		p.NomeProduto, p.CodigoInterno, p.CodigoFornecedor, p.GrupoProduto, p.Descricao, pq.Array(p.Fotos), p.NCM, p.OrigemMaterial, p.Estoque, id)
	return err
}

// Delete removes a produto
func (r *ProdutoRepository) Delete(id int) error {
	res, err := r.db.Exec("DELETE FROM produtos WHERE id=$1", id)
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
