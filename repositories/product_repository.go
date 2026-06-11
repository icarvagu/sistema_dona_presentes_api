package repositories

import (
	"database/sql"
	"errors"
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

const productSelectWithSupplier = `
	SELECT p.id, p.product_name, p.internal_code, p.supplier_code, p.supplier_id, 
	       f.id, f.name, f.cnpj, f.state_registration, f.contact_person, f.email, 
	       f.landline_phone, f.mobile_phone, f.responsible_email, f.commercial_address, 
	       f.created_at, f.updated_at, 
	       p.product_group, p.description, p.photos, p.ncm, COALESCE(p.material_origin, ''), 
	       p.stock, p.supplier_stock, p.selling_price, p.kit_type, p.is_composition, 
	       p.moves_stock, p.enabled_for_invoice, p.cost_price, 
	       p.source, p.imported_at, p.last_synced_at, 
	       COALESCE(p.color, ''), COALESCE(p.origin, ''), p.pending_approval,
	       p.created_at, p.updated_at,
	       COALESCE(p.last_cost, 0), p.last_cost_date, COALESCE(p.last_cost_qty1, 0),
	       COALESCE(p.last_cost_qty2, 0), COALESCE(p.last_cost_qty3, 0),
	       COALESCE(p.last_cost_val1, 0), COALESCE(p.last_cost_val2, 0), COALESCE(p.last_cost_val3, 0),
	       COALESCE(p.last_cost_user, '')
	FROM products p 
	JOIN suppliers f ON p.supplier_id = f.id`

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanProductRow(scanner rowScanner, withSupplier bool) (models.Product, *models.Supplier, error) {
	var p models.Product
	var f models.Supplier
	var fotos pq.StringArray
	var importedAt, lastSyncedAt sql.NullTime
	var lastCostDate sql.NullTime
	var source sql.NullString

	if withSupplier {
		err := scanner.Scan(
			&p.ID, &p.ProductName, &p.InternalCode, &p.SupplierCode, &p.SupplierID,
			&f.ID, &f.Name, &f.CNPJ, &f.StateRegistration, &f.ContactPerson, &f.Email,
			&f.LandlinePhone, &f.MobilePhone, &f.ResponsibleEmail, &f.CommercialAddress,
			&f.CreatedAt, &f.UpdatedAt,
			&p.ProductGroup, &p.Description, &fotos, &p.NCM, &p.MaterialOrigin,
			&p.Stock, &p.SupplierStock, &p.SellingPrice, &p.KitType, &p.IsComposition,
			&p.MovesStock, &p.EnabledForInvoice, &p.CostPrice,
			&source, &importedAt, &lastSyncedAt,
			&p.Color, &p.Origin, &p.PendingApproval,
			&p.CreatedAt, &p.UpdatedAt,
			&p.LastCost, &lastCostDate, &p.LastCostQty1, &p.LastCostQty2, &p.LastCostQty3,
			&p.LastCostVal1, &p.LastCostVal2, &p.LastCostVal3, &p.LastCostUser,
		)
		if err != nil {
			return models.Product{}, nil, err
		}
	} else {
		err := scanner.Scan(
			&p.ID, &p.ProductName, &p.InternalCode, &p.SupplierCode, &p.SupplierID,
			&p.ProductGroup, &p.Description, &fotos, &p.NCM, &p.MaterialOrigin,
			&p.Stock, &p.SupplierStock, &p.SellingPrice, &p.KitType, &p.IsComposition,
			&p.MovesStock, &p.EnabledForInvoice, &p.CostPrice,
			&source, &importedAt, &lastSyncedAt,
			&p.Color, &p.Origin, &p.PendingApproval,
			&p.CreatedAt, &p.UpdatedAt,
			&p.LastCost, &lastCostDate, &p.LastCostQty1, &p.LastCostQty2, &p.LastCostQty3,
			&p.LastCostVal1, &p.LastCostVal2, &p.LastCostVal3, &p.LastCostUser,
		)
		if err != nil {
			return models.Product{}, nil, err
		}
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
	if lastCostDate.Valid {
		p.LastCostDate = &lastCostDate.Time
	}

	if withSupplier {
		return p, &f, nil
	}
	return p, nil, nil
}

func (r *ProductRepository) GetAll() ([]models.Product, error) {
	rows, err := r.db.Query(productSelectWithSupplier)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	var res []models.Product
	for rows.Next() {
		p, supplier, err := scanProductRow(rows, true)
		if err != nil {
			return nil, err
		}
		p.Supplier = supplier

		if p.IsComposition {
			items, _ := r.GetProductItems(p.ID)
			p.Items = items
		}
		res = append(res, p)
	}
	return res, nil
}

func (r *ProductRepository) GetAllPaginated(page, limit int) ([]models.Product, int, error) {

	offset := (page - 1) * limit

	var total int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM products`).Scan(&total)
	if err != nil {
		return nil, 0, apperrors.NewDatabaseError(err)
	}

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
		p, supplier, err := scanProductRow(rows, true)
		if err != nil {
			return nil, 0, err
		}
		p.Supplier = supplier

		if p.IsComposition {
			items, _ := r.GetProductItems(p.ID)
			p.Items = items
		}
		res = append(res, p)
	}
	return res, total, nil
}

func (r *ProductRepository) GetByID(id int) (*models.Product, error) {
	p, supplier, err := scanProductRow(r.db.QueryRow(productSelectWithSupplier+` WHERE p.id=$1`, id), true)
	if err != nil {
		return nil, err
	}
	p.Supplier = supplier

	if p.IsComposition {
		items, _ := r.GetProductItems(p.ID)
		p.Items = items
	}
	return &p, nil
}

func (r *ProductRepository) GetGroups() ([]string, error) {
	rows, err := r.db.Query(`SELECT DISTINCT product_group FROM products WHERE product_group IS NOT NULL AND product_group != '' ORDER BY product_group`)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()
	var groups []string
	for rows.Next() {
		var g string
		if err := rows.Scan(&g); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

func (r *ProductRepository) GetByGroup(group string) ([]models.Product, error) {
	rows, err := r.db.Query(productSelectWithSupplier+` WHERE LOWER(p.product_group) = LOWER($1)`, group)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()
	var res []models.Product
	for rows.Next() {
		p, supplier, err := scanProductRow(rows, true)
		if err != nil {
			return nil, err
		}
		p.Supplier = supplier
		if p.IsComposition {
			items, _ := r.GetProductItems(p.ID)
			p.Items = items
		}
		res = append(res, p)
	}
	return res, nil
}

func (r *ProductRepository) SearchByFilter(filter string) ([]models.Product, error) {
	filterPattern := "%" + filter + "%"
	rows, err := r.db.Query(productSelectWithSupplier+` 
		WHERE LOWER(p.product_name) LIKE LOWER($1) 
		   OR LOWER(p.internal_code) LIKE LOWER($1) 
		   OR LOWER(p.supplier_code) LIKE LOWER($1)
		   OR LOWER(p.product_group) LIKE LOWER($1)
		   OR LOWER(p.color) LIKE LOWER($1)`,
		filterPattern)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	var res []models.Product
	for rows.Next() {
		p, supplier, err := scanProductRow(rows, true)
		if err != nil {
			return nil, err
		}
		p.Supplier = supplier

		if p.IsComposition {
			items, _ := r.GetProductItems(p.ID)
			p.Items = items
		}
		res = append(res, p)
	}
	return res, nil
}

func (r *ProductRepository) GetByInternalCode(internalCode string) (*models.Product, error) {
	p, _, err := scanProductRow(r.db.QueryRow(
		`SELECT id, product_name, internal_code, supplier_code, supplier_id, product_group, description, photos, ncm, COALESCE(material_origin, ''), stock, supplier_stock, selling_price, kit_type, is_composition, moves_stock, enabled_for_invoice, cost_price, source, imported_at, last_synced_at, COALESCE(color, ''), COALESCE(origin, ''), pending_approval, created_at, updated_at, COALESCE(last_cost, 0), last_cost_date, COALESCE(last_cost_qty1, 0), COALESCE(last_cost_qty2, 0), COALESCE(last_cost_qty3, 0), COALESCE(last_cost_val1, 0), COALESCE(last_cost_val2, 0), COALESCE(last_cost_val3, 0), COALESCE(last_cost_user, '') FROM products WHERE internal_code=$1`,
		internalCode,
	), false)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepository) Create(p *models.Product) (*models.Product, error) {

	isComposition := len(p.Items) > 0
	if p.KitType == "" {
		p.KitType = "none"
	}
	err := r.db.QueryRow(`INSERT INTO products (product_name, internal_code, supplier_code, supplier_id, product_group, description, photos, ncm, material_origin, stock, supplier_stock, selling_price, kit_type, is_composition, moves_stock, enabled_for_invoice, cost_price, source, imported_at, last_synced_at, color, origin, pending_approval) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23) RETURNING id, created_at, updated_at`,
		p.ProductName, p.InternalCode, p.SupplierCode, p.SupplierID, p.ProductGroup, p.Description, pq.Array(p.Photos), p.NCM, p.MaterialOrigin, p.Stock, p.SupplierStock, p.SellingPrice, p.KitType, isComposition, p.MovesStock, p.EnabledForInvoice, p.CostPrice, p.Source, p.ImportedAt, p.LastSyncedAt, p.Color, p.Origin, p.PendingApproval).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		var pqErr *pq.Error
		if ok := errors.As(err, &pqErr); ok && pqErr.Code == "23503" {
			return nil, apperrors.ErrSupplierNotFound
		}
		return nil, err
	}
	p.IsComposition = isComposition
	return p, nil
}

func (r *ProductRepository) Update(id int, p *models.Product) (*models.Product, error) {

	isComposition := len(p.Items) > 0
	if p.KitType == "" {
		p.KitType = "none"
	}
	_, err := r.db.Exec(`UPDATE products SET product_name=$1, internal_code=$2, supplier_code=$3, supplier_id=$4, product_group=$5, description=$6, photos=$7, ncm=$8, material_origin=$9, stock=$10, supplier_stock=$11, selling_price=$12, kit_type=$13, is_composition=$14, moves_stock=$15, enabled_for_invoice=$16, cost_price=$17, color=$18, origin=$19, pending_approval=$20, updated_at=NOW() WHERE id=$21`,
		p.ProductName, p.InternalCode, p.SupplierCode, p.SupplierID, p.ProductGroup, p.Description, pq.Array(p.Photos), p.NCM, p.MaterialOrigin, p.Stock, p.SupplierStock, p.SellingPrice, p.KitType, isComposition, p.MovesStock, p.EnabledForInvoice, p.CostPrice, p.Color, p.Origin, p.PendingApproval, id)
	if err != nil {
		var pqErr *pq.Error
		if ok := errors.As(err, &pqErr); ok && pqErr.Code == "23503" {
			return nil, apperrors.ErrSupplierNotFound
		}
		return nil, err
	}
	p.IsComposition = isComposition
	p.ID = id
	return p, nil
}

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

func (r *ProductRepository) CreateProductItem(item *models.ProductItem, productParentID int) error {
	err := r.db.QueryRow(
		`INSERT INTO product_items (product_parent_id, product_id, quantity) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`,
		productParentID, item.ProductID, item.Quantity).
		Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	return err
}

func (r *ProductRepository) DeleteProductItems(productID int) error {
	_, err := r.db.Exec("DELETE FROM product_items WHERE product_parent_id=$1", productID)
	return err
}

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
		p, supplier, err := scanProductRow(rows, true)
		if err != nil {
			return nil, err
		}
		p.Supplier = supplier

		if p.IsComposition {
			items, _ := r.GetProductItems(p.ID)
			p.Items = items
		}
		res = append(res, p)
	}
	return res, nil
}

func (r *ProductRepository) GetPendingApproval() ([]models.Product, error) {
	rows, err := r.db.Query(productSelectWithSupplier + ` WHERE p.pending_approval = TRUE ORDER BY p.imported_at DESC`)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	var res []models.Product
	for rows.Next() {
		p, supplier, err := scanProductRow(rows, true)
		if err != nil {
			return nil, err
		}
		p.Supplier = supplier
		res = append(res, p)
	}
	return res, nil
}

func (r *ProductRepository) BulkApproveAll(origin string) (int64, error) {
	res, err := r.db.Exec(
		`UPDATE products SET origin=$1, pending_approval=FALSE, updated_at=NOW() WHERE pending_approval=TRUE`,
		origin)
	if err != nil {
		return 0, err
	}
	count, _ := res.RowsAffected()
	return count, nil
}

func (r *ProductRepository) ApproveProduct(id int, origin string) error {
	_, err := r.db.Exec(
		`UPDATE products SET origin=$1, pending_approval=FALSE, updated_at=NOW() WHERE id=$2`,
		origin, id,
	)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

func (r *ProductRepository) UpdateLastCost(id int, p *models.Product) error {
	var lastCostDate interface{}
	if p.LastCostDate != nil {
		lastCostDate = *p.LastCostDate
	}
	_, err := r.db.Exec(
		`UPDATE products SET last_cost=$1, last_cost_date=$2, last_cost_qty1=$3, last_cost_qty2=$4, last_cost_qty3=$5, last_cost_val1=$6, last_cost_val2=$7, last_cost_val3=$8, last_cost_user=$9, updated_at=NOW() WHERE id=$10`,
		p.LastCost, lastCostDate, p.LastCostQty1, p.LastCostQty2, p.LastCostQty3,
		p.LastCostVal1, p.LastCostVal2, p.LastCostVal3, p.LastCostUser, id,
	)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

func (r *ProductRepository) UpdateFromSync(id int, stock int, supplierStock int, supplierCode string, photos []string, costPrice float64, lastSyncedAt time.Time) error {
	_, err := r.db.Exec(
		`UPDATE products
         SET stock=$1,
             supplier_stock=$2,
             supplier_code=CASE WHEN $3 != '' THEN $3 ELSE supplier_code END,
             photos=$4,
             cost_price=$5,
             last_synced_at=$6,
             updated_at=NOW()
         WHERE id=$7`,
		stock,
		supplierStock,
		supplierCode,
		pq.Array(photos),
		costPrice,
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
