package repositories

import (
	"database/sql"
	"donapresentes/models"
	"fmt"
)

type VendaRepository struct {
	db *sql.DB
}

func NewVendaRepository(db *sql.DB) *VendaRepository {
	return &VendaRepository{db: db}
}

// ValidateVenda checks required fields and relationships
func (r *VendaRepository) ValidateVenda(v *models.VendaInput) error {
	if v.VendedorID == 0 {
		return fmt.Errorf("vendedor_id é obrigatório")
	}
	if v.FormaPagamento == "" {
		return fmt.Errorf("forma_pagamento é obrigatória")
	}
	if v.Parcelas <= 0 {
		return fmt.Errorf("parcelas deve ser maior que 0")
	}
	if len(v.Itens) == 0 {
		return fmt.Errorf("venda deve ter pelo menos 1 item")
	}

	// Validate vendedor exists
	var tmp int
	if err := r.db.QueryRow("SELECT id FROM funcionarios WHERE id=$1", v.VendedorID).Scan(&tmp); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("vendedor não encontrado")
		}
		return err
	}

	// Validate itens
	for i, item := range v.Itens {
		if item.ProdutoID == 0 {
			return fmt.Errorf("item %d: produto_id é obrigatório", i+1)
		}
		if item.Quantidade <= 0 {
			return fmt.Errorf("item %d: quantidade deve ser maior que 0", i+1)
		}
		if item.ValorUnitario <= 0 {
			return fmt.Errorf("item %d: valor_unitario deve ser maior que 0", i+1)
		}

		// Validate produto exists
		if err := r.db.QueryRow("SELECT id FROM produtos WHERE id=$1", item.ProdutoID).Scan(&tmp); err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("item %d: produto não encontrado", i+1)
			}
			return err
		}
	}

	return nil
}

// GetAll returns all vendas with itens and relacionamentos
func (r *VendaRepository) GetAll() ([]models.Venda, error) {
	rows, err := r.db.Query(`SELECT id, vendedor_id, forma_pagamento, parcelas, prazo_dias, inicio_primeira_parcela, criado_em, atualizado_em FROM vendas`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vendas []models.Venda
	for rows.Next() {
		var v models.Venda
		err := rows.Scan(&v.ID, &v.VendedorID, &v.FormaPagamento, &v.Parcelas, &v.PrazoDias, &v.InicioPrimeiraParcela, &v.CriadoEm, &v.AtualizadoEm)
		if err != nil {
			return nil, err
		}

		// Fetch vendedor
		vendedor, _ := r.GetVendedor(v.VendedorID)
		v.Vendedor = vendedor

		// Fetch itens
		itens, _ := r.GetItens(v.ID)
		v.Itens = itens

		vendas = append(vendas, v)
	}
	return vendas, nil
}

// GetByID returns a single venda by ID with itens and relacionamentos
func (r *VendaRepository) GetByID(id int) (*models.Venda, error) {
	var v models.Venda
	err := r.db.QueryRow(`SELECT id, vendedor_id, forma_pagamento, parcelas, prazo_dias, inicio_primeira_parcela, criado_em, atualizado_em FROM vendas WHERE id=$1`, id).
		Scan(&v.ID, &v.VendedorID, &v.FormaPagamento, &v.Parcelas, &v.PrazoDias, &v.InicioPrimeiraParcela, &v.CriadoEm, &v.AtualizadoEm)
	if err != nil {
		return nil, err
	}

	vendedor, _ := r.GetVendedor(v.VendedorID)
	v.Vendedor = vendedor

	itens, _ := r.GetItens(v.ID)
	v.Itens = itens

	return &v, nil
}

// Create inserts a new venda with itens
func (r *VendaRepository) Create(input *models.VendaInput) (*models.Venda, error) {
	if err := r.ValidateVenda(input); err != nil {
		return nil, err
	}

	var v models.Venda
	err := r.db.QueryRow(
		`INSERT INTO vendas (vendedor_id, forma_pagamento, parcelas, prazo_dias, inicio_primeira_parcela) VALUES ($1,$2,$3,$4,$5) RETURNING id, criado_em, atualizado_em`,
		input.VendedorID, input.FormaPagamento, input.Parcelas, input.PrazoDias, input.InicioPrimeiraParcela).
		Scan(&v.ID, &v.CriadoEm, &v.AtualizadoEm)
	if err != nil {
		return nil, err
	}

	v.VendedorID = input.VendedorID
	v.FormaPagamento = input.FormaPagamento
	v.Parcelas = input.Parcelas
	v.PrazoDias = input.PrazoDias
	v.InicioPrimeiraParcela = input.InicioPrimeiraParcela

	// Insert itens
	for _, itemInput := range input.Itens {
		item := models.VendaItem{
			VendaID:       v.ID,
			ProdutoID:     itemInput.ProdutoID,
			Quantidade:    itemInput.Quantidade,
			ValorUnitario: itemInput.ValorUnitario,
			ValorTotal:    float64(itemInput.Quantidade) * itemInput.ValorUnitario,
		}
		_ = r.CreateItem(&item)
	}

	// Fetch complete data
	completeVenda, _ := r.GetByID(v.ID)
	return completeVenda, nil
}

// Update updates an existing venda and its itens
func (r *VendaRepository) Update(id int, input *models.VendaInput) (*models.Venda, error) {
	if err := r.ValidateVenda(input); err != nil {
		return nil, err
	}

	_, err := r.db.Exec(
		`UPDATE vendas SET vendedor_id=$1, forma_pagamento=$2, parcelas=$3, prazo_dias=$4, inicio_primeira_parcela=$5, atualizado_em=NOW() WHERE id=$6`,
		input.VendedorID, input.FormaPagamento, input.Parcelas, input.PrazoDias, input.InicioPrimeiraParcela, id)
	if err != nil {
		return nil, err
	}

	// Delete existing itens
	r.db.Exec("DELETE FROM venda_itens WHERE venda_id=$1", id)

	// Insert new itens
	for _, itemInput := range input.Itens {
		item := models.VendaItem{
			VendaID:       id,
			ProdutoID:     itemInput.ProdutoID,
			Quantidade:    itemInput.Quantidade,
			ValorUnitario: itemInput.ValorUnitario,
			ValorTotal:    float64(itemInput.Quantidade) * itemInput.ValorUnitario,
		}
		_ = r.CreateItem(&item)
	}

	// Fetch complete data
	return r.GetByID(id)
}

// Delete removes a venda (itens removed by cascade)
func (r *VendaRepository) Delete(id int) error {
	res, err := r.db.Exec("DELETE FROM vendas WHERE id=$1", id)
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

// Helper methods
func (r *VendaRepository) GetVendedor(vendedorID int) (*models.Funcionario, error) {
	var f models.Funcionario
	err := r.db.QueryRow(`SELECT id, nome_completo,cpf FROM funcionarios WHERE id=$1`, vendedorID).
		Scan(&f.ID, &f.NomeCompleto, f.CPF)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *VendaRepository) GetItens(vendaID int) ([]models.VendaItem, error) {
	rows, err := r.db.Query(`SELECT id, venda_id, produto_id, quantidade, valor_unitario, valor_total, criado_em, atualizado_em FROM venda_itens WHERE venda_id=$1`, vendaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var itens []models.VendaItem
	for rows.Next() {
		var item models.VendaItem
		err := rows.Scan(&item.ID, &item.VendaID, &item.ProdutoID, &item.Quantidade, &item.ValorUnitario, &item.ValorTotal, &item.CriadoEm, &item.AtualizadoEm)
		if err != nil {
			return nil, err
		}

		// Fetch produto completo
		produto, _ := r.GetProdutoBasico(item.ProdutoID)
		item.Produto = produto

		itens = append(itens, item)
	}
	return itens, nil
}

func (r *VendaRepository) CreateItem(item *models.VendaItem) error {
	err := r.db.QueryRow(
		`INSERT INTO venda_itens (venda_id, produto_id, quantidade, valor_unitario, valor_total) VALUES ($1,$2,$3,$4,$5) RETURNING id, criado_em, atualizado_em`,
		item.VendaID, item.ProdutoID, item.Quantidade, item.ValorUnitario, item.ValorTotal).
		Scan(&item.ID, &item.CriadoEm, &item.AtualizadoEm)
	return err
}

func (r *VendaRepository) GetProdutoBasico(produtoID int) (*models.Produto, error) {
	var p models.Produto
	err := r.db.QueryRow(`SELECT id, nome_produto FROM produtos WHERE id=$1`, produtoID).
		Scan(&p.ID, &p.NomeProduto)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
