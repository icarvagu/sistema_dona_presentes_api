package repositories

import (
	"database/sql"
	"donapresentes/models"
	apperrors "donapresentes/errors"
)

type SaleRepository struct {
	db *sql.DB
}

func NewSaleRepository(db *sql.DB) *SaleRepository {
	return &SaleRepository{db: db}
}

// ValidateVenda checks required fields and relationships
func (r *SaleRepository) ValidateSale(v *models.SaleInput) error {
	if v.SellerID == 0 {
		return apperrors.NewMissingFieldError("seller_id")
	}
	if v.PaymentMethod == "" {
		return apperrors.NewMissingFieldError("payment_method")
	}
	if v.Installments <= 0 {
		return apperrors.NewInvalidFieldError("installments", "must be > 0")
	}
	if len(v.Items) == 0 {
		return apperrors.NewValidationError("sale must have at least 1 item")
	}

	// Validate seller exists
	var tmp int
	if err := r.db.QueryRow("SELECT id FROM funcionarios WHERE id=$1", v.SellerID).Scan(&tmp); err != nil {
		if err == sql.ErrNoRows {
			return apperrors.ErrEmployeeNotFound
		}
		return apperrors.NewDatabaseError(err)
	}

	// Validate items
	for i, item := range v.Items {
		if item.ProductID == 0 {
			return apperrors.NewInvalidFieldError("item "+string(rune(i+1)), "product_id is required")
		}
		if item.Quantity <= 0 {
			return apperrors.NewInvalidFieldError("item "+string(rune(i+1)), "quantity must be > 0")
		}
		if item.UnitPrice <= 0 {
			return apperrors.NewInvalidFieldError("item "+string(rune(i+1)), "unit_price must be > 0")
		}

		// Validate product exists
		if err := r.db.QueryRow("SELECT id FROM produtos WHERE id=$1", item.ProductID).Scan(&tmp); err != nil {
			if err == sql.ErrNoRows {
				return apperrors.ErrProductNotFound
			}
			return apperrors.NewDatabaseError(err)
		}
	}

	return nil
}

// GetAll returns all vendas with itens and relacionamentos
func (r *SaleRepository) GetAll() ([]models.Sale, error) {
	rows, err := r.db.Query(`SELECT id, vendedor_id, forma_pagamento, parcelas, prazo_dias, inicio_primeira_parcela, criado_em, atualizado_em FROM vendas`)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()
	var sales []models.Sale
	for rows.Next() {
		var v models.Sale
		err := rows.Scan(&v.ID, &v.SellerID, &v.PaymentMethod, &v.Installments, &v.PaymentTermDays, &v.FirstInstallmentStart, &v.CreatedAt, &v.UpdatedAt)
		if err != nil {
			return nil, err
		}
		// Fetch seller
		seller, _ := r.GetSeller(v.SellerID)
		v.Seller = seller

		// Fetch items
		items, _ := r.GetItems(v.ID)
		v.Items = items

		sales = append(sales, v)
	}
	return sales, nil
}

// GetByID returns a single venda by ID with itens and relacionamentos
func (r *SaleRepository) GetByID(id int) (*models.Sale, error) {
	var v models.Sale
	err := r.db.QueryRow(`SELECT id, vendedor_id, forma_pagamento, parcelas, prazo_dias, inicio_primeira_parcela, criado_em, atualizado_em FROM vendas WHERE id=$1`, id).
		Scan(&v.ID, &v.SellerID, &v.PaymentMethod, &v.Installments, &v.PaymentTermDays, &v.FirstInstallmentStart, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, apperrors.NewDatabaseError(err)
	}

	seller, _ := r.GetSeller(v.SellerID)
	v.Seller = seller

	items, _ := r.GetItems(v.ID)
	v.Items = items

	return &v, nil
}

// Create inserts a new venda with itens
func (r *SaleRepository) Create(input *models.SaleInput) (*models.Sale, error) {
	if err := r.ValidateSale(input); err != nil {
		return nil, err
	}

	var v models.Sale
	err := r.db.QueryRow(
		`INSERT INTO vendas (vendedor_id, forma_pagamento, parcelas, prazo_dias, inicio_primeira_parcela) VALUES ($1,$2,$3,$4,$5) RETURNING id, criado_em, atualizado_em`,
		input.SellerID, input.PaymentMethod, input.Installments, input.PaymentTermDays, input.FirstInstallmentStart).
		Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}

	v.SellerID = input.SellerID
	v.PaymentMethod = input.PaymentMethod
	v.Installments = input.Installments
	v.PaymentTermDays = input.PaymentTermDays
	v.FirstInstallmentStart = input.FirstInstallmentStart

	// Insert items
	for _, itemInput := range input.Items {
		item := models.SaleItem{
			SaleID:    v.ID,
			ProductID: itemInput.ProductID,
			Quantity:  itemInput.Quantity,
			UnitPrice: itemInput.UnitPrice,
			TotalPrice: float64(itemInput.Quantity) * itemInput.UnitPrice,
		}
		_ = r.CreateItem(&item)
	}

	// Fetch complete data
	completeSale, _ := r.GetByID(v.ID)
	return completeSale, nil
}

// Update updates an existing venda and its itens
func (r *SaleRepository) Update(id int, input *models.SaleInput) (*models.Sale, error) {
	if err := r.ValidateSale(input); err != nil {
		return nil, err
	}

	_, err := r.db.Exec(
		`UPDATE vendas SET vendedor_id=$1, forma_pagamento=$2, parcelas=$3, prazo_dias=$4, inicio_primeira_parcela=$5, atualizado_em=NOW() WHERE id=$6`,
		input.SellerID, input.PaymentMethod, input.Installments, input.PaymentTermDays, input.FirstInstallmentStart, id)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}

	// Delete existing items
	r.db.Exec("DELETE FROM venda_itens WHERE venda_id=$1", id)

	// Insert new items
	for _, itemInput := range input.Items {
		item := models.SaleItem{
			SaleID:    id,
			ProductID: itemInput.ProductID,
			Quantity:  itemInput.Quantity,
			UnitPrice: itemInput.UnitPrice,
			TotalPrice: float64(itemInput.Quantity) * itemInput.UnitPrice,
		}
		_ = r.CreateItem(&item)
	}

	// Fetch complete data
	return r.GetByID(id)
}

// Delete removes a venda (itens removed by cascade)
func (r *SaleRepository) Delete(id int) error {
	res, err := r.db.Exec("DELETE FROM vendas WHERE id=$1", id)
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

// Helper methods
func (r *SaleRepository) GetSeller(sellerID int) (*models.Employee, error) {
	var f models.Employee
	err := r.db.QueryRow(`SELECT id, nome_completo, cpf, rg, data_nascimento, sexo, situacao, email_contato, endereco_completo, telefones_contato, observacoes, criado_em FROM funcionarios WHERE id=$1`, sellerID).
		Scan(&f.ID, &f.FullName, &f.CPF, &f.RG, &f.BirthDate, &f.Gender, &f.Status, &f.ContactEmail, &f.FullAddress, &f.ContactPhone, &f.Notes, &f.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *SaleRepository) GetItems(saleID int) ([]models.SaleItem, error) {
	rows, err := r.db.Query(`SELECT id, venda_id, produto_id, quantidade, valor_unitario, valor_total, criado_em, atualizado_em FROM venda_itens WHERE venda_id=$1`, saleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.SaleItem
	for rows.Next() {
		var item models.SaleItem
		err := rows.Scan(&item.ID, &item.SaleID, &item.ProductID, &item.Quantity, &item.UnitPrice, &item.TotalPrice, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, err
		}
		// Fetch product
		product, _ := r.GetProductBasic(item.ProductID)
		item.Product = product

		items = append(items, item)
	}
	return items, nil
}

func (r *SaleRepository) CreateItem(item *models.SaleItem) error {
	err := r.db.QueryRow(
		`INSERT INTO venda_itens (venda_id, produto_id, quantidade, valor_unitario, valor_total) VALUES ($1,$2,$3,$4,$5) RETURNING id, criado_em, atualizado_em`,
		item.SaleID, item.ProductID, item.Quantity, item.UnitPrice, item.TotalPrice).
		Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	return err
}

func (r *SaleRepository) GetProductBasic(productID int) (*models.Product, error) {
	var p models.Product
	var productGroup, description, ncm, materialOrigin sql.NullString
	err := r.db.QueryRow(`SELECT id, nome_produto, codigo_interno, codigo_fornecedor, grupo_produto, descricao, ncm, origem_material, estoque, criado_em, atualizado_em FROM produtos WHERE id=$1`, productID).
		Scan(&p.ID, &p.ProductName, &p.InternalCode, &p.SupplierID, &productGroup, &description, &ncm, &materialOrigin, &p.Stock, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if productGroup.Valid {
		p.ProductGroup = productGroup.String
	}
	if description.Valid {
		p.Description = description.String
	}
	if ncm.Valid {
		p.NCM = ncm.String
	}
	if materialOrigin.Valid {
		p.MaterialOrigin = materialOrigin.String
	}
	return &p, nil
}
