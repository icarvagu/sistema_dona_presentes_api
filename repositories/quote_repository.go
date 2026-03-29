package repositories

import (
	"database/sql"
	apperrors "donapresentes/errors"
	"donapresentes/models"
	"fmt"

	"github.com/lib/pq"
)

type QuoteRepository struct {
	db *sql.DB
}

func NewQuoteRepository(db *sql.DB) *QuoteRepository {
	return &QuoteRepository{db: db}
}

func (r *QuoteRepository) GetAll() ([]models.Quote, error) {
	rows, err := r.db.Query(`SELECT id, COALESCE(quote_number, ''), seller_id, customer_id, COALESCE(responsible_name, ''), quote_valid_until, COALESCE(production_lead_time, ''), total_value, created_at, updated_at FROM quotes ORDER BY id DESC`)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	var quotes []models.Quote
	for rows.Next() {
		var quote models.Quote
		err := rows.Scan(&quote.ID, &quote.QuoteNumber, &quote.SellerID, &quote.CustomerID, &quote.ResponsibleName, &quote.QuoteValidUntil, &quote.ProductionLeadTime, &quote.TotalValue, &quote.CreatedAt, &quote.UpdatedAt)
		if err != nil {
			return nil, err
		}

		quote.Seller, _ = r.GetSeller(quote.SellerID)
		quote.Customer, _ = r.GetCustomer(quote.CustomerID)
		quote.Items, _ = r.GetItems(quote.ID)
		quotes = append(quotes, quote)
	}

	return quotes, nil
}

func (r *QuoteRepository) GetByID(id int) (*models.Quote, error) {
	var quote models.Quote
	err := r.db.QueryRow(`SELECT id, COALESCE(quote_number, ''), seller_id, customer_id, COALESCE(responsible_name, ''), quote_valid_until, COALESCE(production_lead_time, ''), total_value, created_at, updated_at FROM quotes WHERE id=$1`, id).
		Scan(&quote.ID, &quote.QuoteNumber, &quote.SellerID, &quote.CustomerID, &quote.ResponsibleName, &quote.QuoteValidUntil, &quote.ProductionLeadTime, &quote.TotalValue, &quote.CreatedAt, &quote.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, apperrors.NewDatabaseError(err)
	}

	quote.Seller, _ = r.GetSeller(quote.SellerID)
	quote.Customer, _ = r.GetCustomer(quote.CustomerID)
	quote.Items, _ = r.GetItems(quote.ID)

	return &quote, nil
}

func (r *QuoteRepository) Create(input *models.QuoteInput) (*models.Quote, error) {
	totalValue := 0.0
	for _, itemInput := range input.Items {
		totalValue += float64(itemInput.Quantity) * itemInput.UnitPrice
	}

	var quote models.Quote
	err := r.db.QueryRow(
		`INSERT INTO quotes (quote_number, seller_id, customer_id, responsible_name, quote_valid_until, production_lead_time, total_value) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, quote_number, total_value, created_at, updated_at`,
		input.QuoteNumber, input.SellerID, input.CustomerID, input.ResponsibleName, input.QuoteValidUntil, input.ProductionLeadTime, totalValue,
	).Scan(&quote.ID, &quote.QuoteNumber, &quote.TotalValue, &quote.CreatedAt, &quote.UpdatedAt)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}

	if quote.QuoteNumber == "" {
		quote.QuoteNumber = fmt.Sprintf("COT-%06d", quote.ID)
		_, _ = r.db.Exec(`UPDATE quotes SET quote_number=$1 WHERE id=$2`, quote.QuoteNumber, quote.ID)
	}

	quote.SellerID = input.SellerID
	quote.CustomerID = input.CustomerID
	quote.ResponsibleName = input.ResponsibleName
	quote.QuoteValidUntil = input.QuoteValidUntil
	quote.ProductionLeadTime = input.ProductionLeadTime

	for _, itemInput := range input.Items {
		item := models.QuoteItem{
			QuoteID:             quote.ID,
			ProductID:           itemInput.ProductID,
			Quantity:            itemInput.Quantity,
			UnitPrice:           itemInput.UnitPrice,
			TotalPrice:          float64(itemInput.Quantity) * itemInput.UnitPrice,
			PersonalizationType: itemInput.PersonalizationType,
		}
		if err := r.CreateItem(&item); err != nil {
			return nil, apperrors.NewDatabaseError(err)
		}
	}

	return r.GetByID(quote.ID)
}

func (r *QuoteRepository) Update(id int, input *models.QuoteInput) (*models.Quote, error) {
	totalValue := 0.0
	for _, itemInput := range input.Items {
		totalValue += float64(itemInput.Quantity) * itemInput.UnitPrice
	}

	_, err := r.db.Exec(
		`UPDATE quotes SET quote_number=$1, seller_id=$2, customer_id=$3, responsible_name=$4, quote_valid_until=$5, production_lead_time=$6, total_value=$7, updated_at=NOW() WHERE id=$8`,
		input.QuoteNumber, input.SellerID, input.CustomerID, input.ResponsibleName, input.QuoteValidUntil, input.ProductionLeadTime, totalValue, id,
	)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}

	_, _ = r.db.Exec(`DELETE FROM quote_items WHERE quote_id=$1`, id)

	for _, itemInput := range input.Items {
		item := models.QuoteItem{
			QuoteID:             id,
			ProductID:           itemInput.ProductID,
			Quantity:            itemInput.Quantity,
			UnitPrice:           itemInput.UnitPrice,
			TotalPrice:          float64(itemInput.Quantity) * itemInput.UnitPrice,
			PersonalizationType: itemInput.PersonalizationType,
		}
		if err := r.CreateItem(&item); err != nil {
			return nil, apperrors.NewDatabaseError(err)
		}
	}

	if input.QuoteNumber == "" {
		quoteNumber := fmt.Sprintf("COT-%06d", id)
		_, _ = r.db.Exec(`UPDATE quotes SET quote_number=$1 WHERE id=$2`, quoteNumber, id)
	}

	return r.GetByID(id)
}

func (r *QuoteRepository) Delete(id int) error {
	res, err := r.db.Exec(`DELETE FROM quotes WHERE id=$1`, id)
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

func (r *QuoteRepository) GetItems(quoteID int) ([]models.QuoteItem, error) {
	rows, err := r.db.Query(`SELECT id, quote_id, product_id, quantity, unit_price, total_price, personalization_type, created_at, updated_at FROM quote_items WHERE quote_id=$1 ORDER BY id ASC`, quoteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.QuoteItem
	for rows.Next() {
		var item models.QuoteItem
		err := rows.Scan(&item.ID, &item.QuoteID, &item.ProductID, &item.Quantity, &item.UnitPrice, &item.TotalPrice, &item.PersonalizationType, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, err
		}
		item.Product, _ = r.GetProductBasic(item.ProductID)
		items = append(items, item)
	}

	return items, nil
}

func (r *QuoteRepository) CreateItem(item *models.QuoteItem) error {
	return r.db.QueryRow(
		`INSERT INTO quote_items (quote_id, product_id, quantity, unit_price, total_price, personalization_type) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id, created_at, updated_at`,
		item.QuoteID, item.ProductID, item.Quantity, item.UnitPrice, item.TotalPrice, item.PersonalizationType,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
}

func (r *QuoteRepository) GetSeller(sellerID int) (*models.User, error) {
	var user models.User
	var rg, gender, contactEmail, fullAddress, contactPhone, notes sql.NullString
	var birthDate sql.NullTime

	err := r.db.QueryRow(
		`SELECT id, username, role, full_name, cpf, rg, birth_date, gender, status, contact_email, full_address, contact_phone, notes, created_at, updated_at FROM users WHERE id=$1`, sellerID,
	).Scan(&user.ID, &user.Username, &user.Role, &user.FullName, &user.CPF, &rg, &birthDate, &gender, &user.Status, &contactEmail, &fullAddress, &contactPhone, &notes, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if rg.Valid {
		user.RG = &rg.String
	}
	if gender.Valid {
		user.Gender = &gender.String
	}
	if contactEmail.Valid {
		user.ContactEmail = &contactEmail.String
	}
	if fullAddress.Valid {
		user.FullAddress = &fullAddress.String
	}
	if contactPhone.Valid {
		user.ContactPhone = &contactPhone.String
	}
	if notes.Valid {
		user.Notes = &notes.String
	}
	if birthDate.Valid {
		user.BirthDate = &birthDate.Time
	}

	return &user, nil
}

func (r *QuoteRepository) GetCustomer(customerID int) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.QueryRow(`SELECT id, customer_type, status, name, cnpj, cpf, email, business_phone, mobile_phone, website, notes, created_at, updated_at FROM customers WHERE id=$1`, customerID).
		Scan(&customer.ID, &customer.CustomerType, &customer.Status, &customer.Name, &customer.CNPJ, &customer.CPF, &customer.Email, &customer.BusinessPhone, &customer.MobilePhone, &customer.Website, &customer.Notes, &customer.CreatedAt, &customer.UpdatedAt)
	if err != nil {
		return nil, err
	}
	addresses, _ := r.GetCustomerAddresses(customerID)
	customer.Addresses = addresses
	return &customer, nil
}

func (r *QuoteRepository) GetCustomerAddresses(customerID int) ([]models.Address, error) {
	rows, err := r.db.Query(`SELECT id, customer_id, address_type, address, created_at, updated_at FROM customer_addresses WHERE customer_id=$1 ORDER BY CASE WHEN address_type='entrega' THEN 0 ELSE 1 END`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addresses []models.Address
	for rows.Next() {
		var address models.Address
		err := rows.Scan(&address.ID, &address.CustomerID, &address.AddressType, &address.AddressLine, &address.CreatedAt, &address.UpdatedAt)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, address)
	}

	return addresses, nil
}

func (r *QuoteRepository) GetProductBasic(productID int) (*models.Product, error) {
	var product models.Product
	var productGroup, description, ncm, materialOrigin sql.NullString
	var photos pq.StringArray
	err := r.db.QueryRow(`SELECT id, product_name, internal_code, supplier_id, product_group, description, photos, ncm, material_origin, stock, created_at, updated_at FROM products WHERE id=$1`, productID).
		Scan(&product.ID, &product.ProductName, &product.InternalCode, &product.SupplierID, &productGroup, &description, &photos, &ncm, &materialOrigin, &product.Stock, &product.CreatedAt, &product.UpdatedAt)
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
	product.Photos = []string(photos)
	return &product, nil
}