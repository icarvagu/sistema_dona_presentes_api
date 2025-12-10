package repositories

import (
	"database/sql"
	"donapresentes/models"
	apperrors "donapresentes/errors"
	"fmt"

	"github.com/lib/pq"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// GetAll returns all products with nested supplier
func (r *ProductRepository) GetAll() ([]models.Product, error) {
	rows, err := r.db.Query(`SELECT p.id, p.nome_produto, p.codigo_interno, p.codigo_fornecedor, f.id, f.nome_fantasia_ou_razao_social, f.cnpj, f.inscricao_estadual, f.responsavel_atendimento, f.email_geral, f.telefone_fixo, f.celular, f.email_responsavel, f.endereco_comercial, f.criado_em, f.atualizado_em, p.grupo_produto, p.descricao, p.fotos, p.ncm, p.origem_material, p.estoque, p.criado_em, p.atualizado_em FROM produtos p JOIN fornecedores f ON p.codigo_fornecedor = f.id`)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	var res []models.Product
	for rows.Next() {
		var p models.Product
		var f models.Supplier
		var fotos pq.StringArray
		err := rows.Scan(&p.ID, &p.ProductName, &p.InternalCode, &p.SupplierID, &f.ID, &f.Name, &f.CNPJ, &f.StateRegistration, &f.ContactPerson, &f.Email, &f.LandlinePhone, &f.MobilePhone, &f.ResponsibleEmail, &f.CommercialAddress, &f.CreatedAt, &f.UpdatedAt, &p.ProductGroup, &p.Description, &fotos, &p.NCM, &p.MaterialOrigin, &p.Stock, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		p.Photos = []string(fotos)
		p.Supplier = &f
		res = append(res, p)
	}
	return res, nil
}

// GetByID returns one product by id with nested supplier
func (r *ProductRepository) GetByID(id int) (*models.Product, error) {
	var p models.Product
	var f models.Supplier
	var fotos pq.StringArray
	err := r.db.QueryRow(`SELECT p.id, p.nome_produto, p.codigo_interno, p.codigo_fornecedor, f.id, f.nome_fantasia_ou_razao_social, f.cnpj, f.inscricao_estadual, f.responsavel_atendimento, f.email_geral, f.telefone_fixo, f.celular, f.email_responsavel, f.endereco_comercial, f.criado_em, f.atualizado_em, p.grupo_produto, p.descricao, p.fotos, p.ncm, p.origem_material, p.estoque, p.criado_em, p.atualizado_em FROM produtos p JOIN fornecedores f ON p.codigo_fornecedor = f.id WHERE p.id=$1`, id).
		Scan(&p.ID, &p.ProductName, &p.InternalCode, &p.SupplierID, &f.ID, &f.Name, &f.CNPJ, &f.StateRegistration, &f.ContactPerson, &f.Email, &f.LandlinePhone, &f.MobilePhone, &f.ResponsibleEmail, &f.CommercialAddress, &f.CreatedAt, &f.UpdatedAt, &p.ProductGroup, &p.Description, &fotos, &p.NCM, &p.MaterialOrigin, &p.Stock, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	p.Photos = []string(fotos)
	p.Supplier = &f
	return &p, nil
}

// GetByCodigoInterno returns a product by internal code (no supplier nested)
func (r *ProductRepository) GetByCodigoInterno(codigoInterno string) (*models.Product, error) {
	var p models.Product
	var fotos pq.StringArray
	err := r.db.QueryRow(
		`SELECT id, nome_produto, codigo_interno, codigo_fornecedor, grupo_produto, descricao, fotos, ncm, origem_material, estoque, criado_em, atualizado_em FROM produtos WHERE codigo_interno=$1`,
		codigoInterno,
	).Scan(&p.ID, &p.ProductName, &p.InternalCode, &p.SupplierID, &p.ProductGroup, &p.Description, &fotos, &p.NCM, &p.MaterialOrigin, &p.Stock, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	p.Photos = []string(fotos)
	return &p, nil
}

// Create inserts a new product. Validates supplier existence.
func (r *ProductRepository) Create(p *models.Product) (*models.Product, error) {
	// check supplier exists
	var tmp int
	if err := r.db.QueryRow("SELECT id FROM fornecedores WHERE id=$1", p.SupplierID).Scan(&tmp); err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrSupplierNotFound
		}
		return nil, apperrors.NewDatabaseError(err)
	}

	err := r.db.QueryRow(`INSERT INTO produtos (nome_produto, codigo_interno, codigo_fornecedor, grupo_produto, descricao, fotos, ncm, origem_material, estoque) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id, criado_em, atualizado_em`,
		p.ProductName, p.InternalCode, p.SupplierID, p.ProductGroup, p.Description, pq.Array(p.Photos), p.NCM, p.MaterialOrigin, p.Stock).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Update updates an existing product
func (r *ProductRepository) Update(id int, p *models.Product) (*models.Product, error) {
	// ensure supplier exists
	var tmp int
	if err := r.db.QueryRow("SELECT id FROM fornecedores WHERE id=$1", p.SupplierID).Scan(&tmp); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("supplier not found")
		}
		return nil, err
	}

	_, err := r.db.Exec(`UPDATE produtos SET nome_produto=$1, codigo_interno=$2, codigo_fornecedor=$3, grupo_produto=$4, descricao=$5, fotos=$6, ncm=$7, origem_material=$8, estoque=$9, atualizado_em=NOW() WHERE id=$10`,
		p.ProductName, p.InternalCode, p.SupplierID, p.ProductGroup, p.Description, pq.Array(p.Photos), p.NCM, p.MaterialOrigin, p.Stock, id)
	if err != nil {
		return nil, err
	}
	p.ID = id
	return p, nil
}

// Delete removes a product
func (r *ProductRepository) Delete(id int) error {
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
