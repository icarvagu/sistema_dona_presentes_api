package repositories

import (
	"database/sql"
	"fmt"
	"time"

	apperrors "donapresentes/errors"
	"donapresentes/models"

	"github.com/lib/pq"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Constante SQL para SELECT de produtos com supplier (usado em múltiplos métodos)
const productSelectWithSupplier = `
	SELECT p.id, p.product_name, p.internal_code, p.supplier_id, 
	       f.id, f.name, f.cnpj, f.state_registration, f.contact_person, f.email, 
	       f.landline_phone, f.mobile_phone, f.responsible_email, f.commercial_address, 
	       f.created_at, f.updated_at, 
	       p.product_group, p.description, p.photos, p.ncm, COALESCE(p.material_origin, ''), 
	       p.stock, p.selling_price, p.kit_type, p.is_composition, 
	       p.moves_stock, p.enabled_for_invoice, p.cost_price, 
	       p.source, p.imported_at, p.last_synced_at, 
	       p.created_at, p.updated_at 
	FROM products p 
	JOIN suppliers f ON p.supplier_id = f.id`

// GetAll returns all products with nested supplier (without pagination - kept for backward compatibility)
func (r *ProductRepository) GetAll() ([]models.Product, error) {
	rows, err := r.db.Query(productSelectWithSupplier)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	var res []models.Product
	for rows.Next() {
		var p models.Product
		var f models.Supplier
		var fotos pq.StringArray
		var importedAt, lastSyncedAt sql.NullTime
		var source sql.NullString
		err := rows.Scan(&p.ID, &p.ProductName, &p.InternalCode, &p.SupplierID, &f.ID, &f.Name, &f.CNPJ, &f.StateRegistration, &f.ContactPerson, &f.Email, &f.LandlinePhone, &f.MobilePhone, &f.ResponsibleEmail, &f.CommercialAddress, &f.CreatedAt, &f.UpdatedAt, &p.ProductGroup, &p.Description, &fotos, &p.NCM, &p.MaterialOrigin, &p.Stock, &p.SellingPrice, &p.KitType, &p.IsComposition, &p.MovesStock, &p.EnabledForInvoice, &p.CostPrice, &source, &importedAt, &lastSyncedAt, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if fotos == nil {
			p.Photos = []string{}
		} else {
			p.Photos = []string(fotos)
		}
		if source.Valid {
			p.Source = source.String
		} else {
			p.Source = "manual"
		}
		if importedAt.Valid {
			p.ImportedAt = &importedAt.Time
		}
		if lastSyncedAt.Valid {
			p.LastSyncedAt = &lastSyncedAt.Time
		}
		p.Supplier = &f
		// Buscar items se for composição
		if p.IsComposition {
			items, _ := r.GetProductItems(p.ID)
			p.Items = items
		}
		res = append(res, p)
	}
	return res, nil
}

// GetAllPaginated returns paginated products with nested supplier
func (r *ProductRepository) GetAllPaginated(page, limit int) ([]models.Product, int, error) {
	// Calcular offset
	offset := (page - 1) * limit

	// Query para contar total de registros
	var total int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM products`).Scan(&total)
	if err != nil {
		return nil, 0, apperrors.NewDatabaseError(err)
	}

	// Query paginada
	rows, err := r.db.Query(productSelectWithSupplier+` 
		ORDER BY p.id 
		LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return nil, 0, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	var res []models.Product
	for rows.Next() {
		var p models.Product
		var f models.Supplier
		var fotos pq.StringArray
		var importedAt, lastSyncedAt sql.NullTime
		var source sql.NullString
		err := rows.Scan(&p.ID, &p.ProductName, &p.InternalCode, &p.SupplierID, &f.ID, &f.Name, &f.CNPJ, &f.StateRegistration, &f.ContactPerson, &f.Email, &f.LandlinePhone, &f.MobilePhone, &f.ResponsibleEmail, &f.CommercialAddress, &f.CreatedAt, &f.UpdatedAt, &p.ProductGroup, &p.Description, &fotos, &p.NCM, &p.MaterialOrigin, &p.Stock, &p.SellingPrice, &p.KitType, &p.IsComposition, &p.MovesStock, &p.EnabledForInvoice, &p.CostPrice, &source, &importedAt, &lastSyncedAt, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		if fotos == nil {
			p.Photos = []string{}
		} else {
			p.Photos = []string(fotos)
		}
		if source.Valid {
			p.Source = source.String
		} else {
			p.Source = "manual"
		}
		if importedAt.Valid {
			p.ImportedAt = &importedAt.Time
		}
		if lastSyncedAt.Valid {
			p.LastSyncedAt = &lastSyncedAt.Time
		}
		p.Supplier = &f
		// Buscar items se for composição
		if p.IsComposition {
			items, _ := r.GetProductItems(p.ID)
			p.Items = items
		}
		res = append(res, p)
	}
	return res, total, nil
}

// GetByID returns one product by id with nested supplier
func (r *ProductRepository) GetByID(id int) (*models.Product, error) {
	var p models.Product
	var f models.Supplier
	var fotos pq.StringArray
	var importedAt, lastSyncedAt sql.NullTime
	var source sql.NullString
	err := r.db.QueryRow(productSelectWithSupplier+` WHERE p.id=$1`, id).
		Scan(&p.ID, &p.ProductName, &p.InternalCode, &p.SupplierID, &f.ID, &f.Name, &f.CNPJ, &f.StateRegistration, &f.ContactPerson, &f.Email, &f.LandlinePhone, &f.MobilePhone, &f.ResponsibleEmail, &f.CommercialAddress, &f.CreatedAt, &f.UpdatedAt, &p.ProductGroup, &p.Description, &fotos, &p.NCM, &p.MaterialOrigin, &p.Stock, &p.SellingPrice, &p.KitType, &p.IsComposition, &p.MovesStock, &p.EnabledForInvoice, &p.CostPrice, &source, &importedAt, &lastSyncedAt, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if fotos == nil {
		p.Photos = []string{}
	} else {
		p.Photos = []string(fotos)
	}
	if source.Valid {
		p.Source = source.String
	} else {
		p.Source = "manual"
	}
	if importedAt.Valid {
		p.ImportedAt = &importedAt.Time
	}
	if lastSyncedAt.Valid {
		p.LastSyncedAt = &lastSyncedAt.Time
	}
	p.Supplier = &f
	// Buscar items se for composição
	if p.IsComposition {
		items, _ := r.GetProductItems(p.ID)
		p.Items = items
	}
	return &p, nil
}

// SearchByFilter searches products by product_name, internal_code, or product_group
func (r *ProductRepository) SearchByFilter(filter string) ([]models.Product, error) {
	filterPattern := "%" + filter + "%"
	rows, err := r.db.Query(productSelectWithSupplier+` 
		WHERE LOWER(p.product_name) LIKE LOWER($1) 
		   OR LOWER(p.internal_code) LIKE LOWER($1) 
		   OR LOWER(p.product_group) LIKE LOWER($1)`,
		filterPattern)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	var res []models.Product
	for rows.Next() {
		var p models.Product
		var f models.Supplier
		var fotos pq.StringArray
		var importedAt, lastSyncedAt sql.NullTime
		var source sql.NullString
		err := rows.Scan(&p.ID, &p.ProductName, &p.InternalCode, &p.SupplierID, &f.ID, &f.Name, &f.CNPJ, &f.StateRegistration, &f.ContactPerson, &f.Email, &f.LandlinePhone, &f.MobilePhone, &f.ResponsibleEmail, &f.CommercialAddress, &f.CreatedAt, &f.UpdatedAt, &p.ProductGroup, &p.Description, &fotos, &p.NCM, &p.MaterialOrigin, &p.Stock, &p.SellingPrice, &p.KitType, &p.IsComposition, &p.MovesStock, &p.EnabledForInvoice, &p.CostPrice, &source, &importedAt, &lastSyncedAt, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if fotos == nil {
			p.Photos = []string{}
		} else {
			p.Photos = []string(fotos)
		}
		if source.Valid {
			p.Source = source.String
		} else {
			p.Source = "manual"
		}
		if importedAt.Valid {
			p.ImportedAt = &importedAt.Time
		}
		if lastSyncedAt.Valid {
			p.LastSyncedAt = &lastSyncedAt.Time
		}
		p.Supplier = &f
		// Buscar items se for composição
		if p.IsComposition {
			items, _ := r.GetProductItems(p.ID)
			p.Items = items
		}
		res = append(res, p)
	}
	return res, nil
}

// GetByCodigoInterno returns a product by internal code (no supplier nested)
func (r *ProductRepository) GetByCodigoInterno(codigoInterno string) (*models.Product, error) {
	var p models.Product
	var fotos pq.StringArray
	var importedAt, lastSyncedAt sql.NullTime
	var source sql.NullString
	err := r.db.QueryRow(
		`SELECT id, product_name, internal_code, supplier_id, product_group, description, photos, ncm, COALESCE(material_origin, ''), stock, selling_price, kit_type, is_composition, moves_stock, enabled_for_invoice, cost_price, source, imported_at, last_synced_at, created_at, updated_at FROM products WHERE internal_code=$1`,
		codigoInterno,
	).Scan(&p.ID, &p.ProductName, &p.InternalCode, &p.SupplierID, &p.ProductGroup, &p.Description, &fotos, &p.NCM, &p.MaterialOrigin, &p.Stock, &p.SellingPrice, &p.KitType, &p.IsComposition, &p.MovesStock, &p.EnabledForInvoice, &p.CostPrice, &source, &importedAt, &lastSyncedAt, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if fotos == nil {
		p.Photos = []string{}
	} else {
		p.Photos = []string(fotos)
	}
	if source.Valid {
		p.Source = source.String
	} else {
		p.Source = "manual"
	}
	if importedAt.Valid {
		p.ImportedAt = &importedAt.Time
	}
	if lastSyncedAt.Valid {
		p.LastSyncedAt = &lastSyncedAt.Time
	}
	return &p, nil
}

// Create inserts a new product. Validates supplier existence.
func (r *ProductRepository) Create(p *models.Product) (*models.Product, error) {
	// check supplier exists
	var tmp int
	if err := r.db.QueryRow("SELECT id FROM suppliers WHERE id=$1", p.SupplierID).Scan(&tmp); err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrSupplierNotFound
		}
		return nil, apperrors.NewDatabaseError(err)
	}

	// Determinar is_composition baseado na quantidade de items
	isComposition := len(p.Items) > 0
	if p.KitType == "" {
		p.KitType = "none"
	}
	err := r.db.QueryRow(`INSERT INTO products (product_name, internal_code, supplier_id, product_group, description, photos, ncm, material_origin, stock, selling_price, kit_type, is_composition, moves_stock, enabled_for_invoice, cost_price, source, imported_at, last_synced_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18) RETURNING id, created_at, updated_at`,
		p.ProductName, p.InternalCode, p.SupplierID, p.ProductGroup, p.Description, pq.Array(p.Photos), p.NCM, p.MaterialOrigin, p.Stock, p.SellingPrice, p.KitType, isComposition, p.MovesStock, p.EnabledForInvoice, p.CostPrice, p.Source, p.ImportedAt, p.LastSyncedAt).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	p.IsComposition = isComposition
	return p, nil
}

// Update updates an existing product
func (r *ProductRepository) Update(id int, p *models.Product) (*models.Product, error) {
	// ensure supplier exists
	var tmp int
	if err := r.db.QueryRow("SELECT id FROM suppliers WHERE id=$1", p.SupplierID).Scan(&tmp); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("supplier not found")
		}
		return nil, err
	}

	// Determinar is_composition baseado na quantidade de items
	isComposition := len(p.Items) > 0
	if p.KitType == "" {
		p.KitType = "none"
	}
	_, err := r.db.Exec(`UPDATE products SET product_name=$1, internal_code=$2, supplier_id=$3, product_group=$4, description=$5, photos=$6, ncm=$7, material_origin=$8, stock=$9, selling_price=$10, kit_type=$11, is_composition=$12, moves_stock=$13, enabled_for_invoice=$14, cost_price=$15, updated_at=NOW() WHERE id=$16`,
		p.ProductName, p.InternalCode, p.SupplierID, p.ProductGroup, p.Description, pq.Array(p.Photos), p.NCM, p.MaterialOrigin, p.Stock, p.SellingPrice, p.KitType, isComposition, p.MovesStock, p.EnabledForInvoice, p.CostPrice, id)
	if err != nil {
		return nil, err
	}
	p.IsComposition = isComposition
	p.ID = id
	return p, nil
}

// Delete removes a product
func (r *ProductRepository) Delete(id int) error {
	res, err := r.db.Exec("DELETE FROM products WHERE id=$1", id)
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

// GetProductItems returns all items for a product (composition)
func (r *ProductRepository) GetProductItems(productID int) ([]models.ProductItem, error) {
	rows, err := r.db.Query(`
		SELECT pi.id, pi.product_parent_id, pi.product_id, pi.quantity, pi.created_at, pi.updated_at,
		       p.id, p.product_name, p.internal_code, p.supplier_id, p.product_group, 
		       p.description, p.photos, p.ncm, p.material_origin, p.stock, p.selling_price, p.is_composition, p.created_at, p.updated_at
		FROM product_items pi
		JOIN products p ON pi.product_id = p.id
		WHERE pi.product_parent_id = $1
		ORDER BY pi.id`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.ProductItem
	for rows.Next() {
		var item models.ProductItem
		var product models.Product
		var fotos pq.StringArray
		var productGroup, description, ncm, materialOrigin sql.NullString

		err := rows.Scan(
			&item.ID, &item.ProductParentID, &item.ProductID, &item.Quantity, &item.CreatedAt, &item.UpdatedAt,
			&product.ID, &product.ProductName, &product.InternalCode, &product.SupplierID,
			&productGroup, &description, &fotos, &ncm,
			&materialOrigin, &product.Stock, &product.SellingPrice, &product.IsComposition, &product.CreatedAt, &product.UpdatedAt)
		if err != nil {
			return nil, err
		}

		// Converter campos nullable
		if productGroup.Valid {
			product.ProductGroup = productGroup.String
		}
		if description.Valid {
			product.Description = description.String
		}
		if ncm.Valid {
			product.NCM = ncm.String
		}
		if materialOrigin.Valid {
			product.MaterialOrigin = materialOrigin.String
		}

		// Converter fotos
		if fotos == nil {
			product.Photos = []string{}
		} else {
			product.Photos = []string(fotos)
		}

		item.Product = &product
		items = append(items, item)
	}
	return items, nil
}

// CreateProductItem creates a new product item (for composition)
func (r *ProductRepository) CreateProductItem(item *models.ProductItem, productParentID int) error {
	err := r.db.QueryRow(
		`INSERT INTO product_items (product_parent_id, product_id, quantity) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`,
		productParentID, item.ProductID, item.Quantity).
		Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	return err
}

// DeleteProductItems deletes all items for a product
func (r *ProductRepository) DeleteProductItems(productID int) error {
	_, err := r.db.Exec("DELETE FROM product_items WHERE product_parent_id=$1", productID)
	return err
}

// GetNewlyImported returns products imported in the last 7 days (regardless of source)
func (r *ProductRepository) GetNewlyImported() ([]models.Product, error) {
	rows, err := r.db.Query(productSelectWithSupplier + ` 
		WHERE p.imported_at IS NOT NULL 
		  AND p.imported_at >= NOW() - INTERVAL '7 days'
		ORDER BY p.imported_at DESC`)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	var res []models.Product
	for rows.Next() {
		var p models.Product
		var f models.Supplier
		var fotos pq.StringArray
		var importedAt, lastSyncedAt sql.NullTime
		var source sql.NullString
		err := rows.Scan(&p.ID, &p.ProductName, &p.InternalCode, &p.SupplierID,
			&f.ID, &f.Name, &f.CNPJ, &f.StateRegistration, &f.ContactPerson, &f.Email,
			&f.LandlinePhone, &f.MobilePhone, &f.ResponsibleEmail, &f.CommercialAddress,
			&f.CreatedAt, &f.UpdatedAt,
			&p.ProductGroup, &p.Description, &fotos, &p.NCM, &p.MaterialOrigin,
			&p.Stock, &p.SellingPrice, &p.KitType, &p.IsComposition,
			&p.MovesStock, &p.EnabledForInvoice, &p.CostPrice,
			&source, &importedAt, &lastSyncedAt,
			&p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if fotos == nil {
			p.Photos = []string{}
		} else {
			p.Photos = []string(fotos)
		}
		if source.Valid {
			p.Source = source.String
		} else {
			p.Source = "manual"
		}
		if importedAt.Valid {
			p.ImportedAt = &importedAt.Time
		}
		if lastSyncedAt.Valid {
			p.LastSyncedAt = &lastSyncedAt.Time
		}
		p.Supplier = &f
		// Buscar items se for composição
		if p.IsComposition {
			items, _ := r.GetProductItems(p.ID)
			p.Items = items
		}
		res = append(res, p)
	}
	return res, nil
}

// UpdateFromSync atualiza apenas campos vindos da sincronização externa (estoque, fotos, preço, last_synced_at)
func (r *ProductRepository) UpdateFromSync(id int, stock int, photos []string, sellingPrice float64, lastSyncedAt time.Time) error {
	_, err := r.db.Exec(
		`UPDATE products
         SET stock=$1,
             photos=$2,
             selling_price=$3,
             last_synced_at=$4,
             updated_at=NOW()
         WHERE id=$5`,
		stock,
		pq.Array(photos),
		sellingPrice,
		lastSyncedAt,
		id,
	)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

func (r *ProductRepository) GetFinancialReport() ([]models.FinancialReportItem, error) {
	rows, err := r.db.Query(`
		SELECT p.id, p.product_name, p.internal_code, p.kit_type,
		       p.cost_price, p.selling_price,
		       (p.selling_price - p.cost_price) AS margin_value,
		       CASE
		         WHEN p.selling_price > 0 THEN ((p.selling_price - p.cost_price) / p.selling_price) * 100
		         ELSE 0
		       END AS margin_percent,
		       p.moves_stock, p.enabled_for_invoice
		FROM products p
		ORDER BY p.product_name ASC`)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	var items []models.FinancialReportItem
	for rows.Next() {
		var item models.FinancialReportItem
		if err := rows.Scan(
			&item.ProductID,
			&item.ProductName,
			&item.InternalCode,
			&item.KitType,
			&item.CostPrice,
			&item.SellingPrice,
			&item.MarginValue,
			&item.MarginPercent,
			&item.MovesStock,
			&item.EnabledForInvoice,
		); err != nil {
			return nil, apperrors.NewDatabaseError(err)
		}
		items = append(items, item)
	}

	return items, nil
}
