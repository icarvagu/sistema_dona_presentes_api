package repositories

import (
	"database/sql"
	apperrors "donapresentes/errors"
	"donapresentes/models"
)

type SaleRepository struct {
	db *sql.DB
}

func NewSaleRepository(db *sql.DB) *SaleRepository {
	return &SaleRepository{db: db}
}

func (r *SaleRepository) GetAll() ([]models.Sale, error) {
	rows, err := r.db.Query(`SELECT id, seller_id, customer_id, payment_method, installments, payment_term_days, first_installment_start, total_value, created_at, updated_at FROM sales`)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()
	var sales []models.Sale
	for rows.Next() {
		var v models.Sale
		err := rows.Scan(&v.ID, &v.SellerID, &v.CustomerID, &v.PaymentMethod, &v.Installments, &v.PaymentTermDays, &v.FirstInstallmentStart, &v.TotalValue, &v.CreatedAt, &v.UpdatedAt)
		if err != nil {
			return nil, err
		}

		seller, _ := r.GetSeller(v.SellerID)
		v.Seller = seller

		customer, _ := r.GetCustomer(v.CustomerID)
		v.Customer = customer

		items, _ := r.GetItems(v.ID)
		v.Items = items

		carriers, _ := r.GetCarriers(v.ID)
		v.Carriers = carriers

		sales = append(sales, v)
	}
	return sales, nil
}

func (r *SaleRepository) GetBySellerID(sellerID int) ([]models.Sale, error) {
	rows, err := r.db.Query(`SELECT id, seller_id, customer_id, payment_method, installments, payment_term_days, first_installment_start, total_value, created_at, updated_at FROM sales WHERE seller_id=$1`, sellerID)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()
	var sales []models.Sale
	for rows.Next() {
		var v models.Sale
		err := rows.Scan(&v.ID, &v.SellerID, &v.CustomerID, &v.PaymentMethod, &v.Installments, &v.PaymentTermDays, &v.FirstInstallmentStart, &v.TotalValue, &v.CreatedAt, &v.UpdatedAt)
		if err != nil {
			return nil, err
		}
		seller, _ := r.GetSeller(v.SellerID)
		v.Seller = seller
		customer, _ := r.GetCustomer(v.CustomerID)
		v.Customer = customer
		items, _ := r.GetItems(v.ID)
		v.Items = items
		carriers, _ := r.GetCarriers(v.ID)
		v.Carriers = carriers
		sales = append(sales, v)
	}
	return sales, nil
}

func (r *SaleRepository) GetByID(id int) (*models.Sale, error) {
	var v models.Sale
	err := r.db.QueryRow(`SELECT id, seller_id, customer_id, payment_method, installments, payment_term_days, first_installment_start, total_value, created_at, updated_at FROM sales WHERE id=$1`, id).
		Scan(&v.ID, &v.SellerID, &v.CustomerID, &v.PaymentMethod, &v.Installments, &v.PaymentTermDays, &v.FirstInstallmentStart, &v.TotalValue, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, apperrors.NewDatabaseError(err)
	}

	seller, _ := r.GetSeller(v.SellerID)
	v.Seller = seller

	customer, _ := r.GetCustomer(v.CustomerID)
	v.Customer = customer

	items, _ := r.GetItems(v.ID)
	v.Items = items

	carriers, _ := r.GetCarriers(v.ID)
	v.Carriers = carriers

	return &v, nil
}

func (r *SaleRepository) Create(input *models.SaleInput) (*models.Sale, error) {

	totalValue := 0.0

	for _, itemInput := range input.Items {
		totalValue += float64(itemInput.Quantity) * itemInput.UnitPrice
	}

	var v models.Sale
	err := r.db.QueryRow(
		`INSERT INTO sales (seller_id, customer_id, payment_method, installments, payment_term_days, first_installment_start, total_value) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, total_value, created_at, updated_at`,
		input.SellerID, input.CustomerID, input.PaymentMethod, input.Installments, input.PaymentTermDays, input.FirstInstallmentStart, totalValue).
		Scan(&v.ID, &v.TotalValue, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}

	v.SellerID = input.SellerID
	v.CustomerID = input.CustomerID
	v.PaymentMethod = input.PaymentMethod
	v.Installments = input.Installments
	v.PaymentTermDays = input.PaymentTermDays
	v.FirstInstallmentStart = input.FirstInstallmentStart

	for _, itemInput := range input.Items {
		item := models.SaleItem{
			SaleID:     v.ID,
			ProductID:  itemInput.ProductID,
			Quantity:   itemInput.Quantity,
			UnitPrice:  itemInput.UnitPrice,
			TotalPrice: float64(itemInput.Quantity) * itemInput.UnitPrice,
		}
		_ = r.CreateItem(&item)
	}

	if len(input.CarrierIDs) > 0 {
		for _, carrierID := range input.CarrierIDs {
			_ = r.CreateCarrierLink(v.ID, carrierID)
		}
	}

	completeSale, _ := r.GetByID(v.ID)
	return completeSale, nil
}

func (r *SaleRepository) Update(id int, input *models.SaleInput) (*models.Sale, error) {

	totalValue := 0.0

	for _, itemInput := range input.Items {
		totalValue += float64(itemInput.Quantity) * itemInput.UnitPrice
	}

	_, err := r.db.Exec(
		`UPDATE sales SET seller_id=$1, customer_id=$2, payment_method=$3, installments=$4, payment_term_days=$5, first_installment_start=$6, total_value=$7, updated_at=NOW() WHERE id=$8`,
		input.SellerID, input.CustomerID, input.PaymentMethod, input.Installments, input.PaymentTermDays, input.FirstInstallmentStart, totalValue, id)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}

	r.db.Exec("DELETE FROM sale_items WHERE sale_id=$1", id)

	for _, itemInput := range input.Items {
		item := models.SaleItem{
			SaleID:     id,
			ProductID:  itemInput.ProductID,
			Quantity:   itemInput.Quantity,
			UnitPrice:  itemInput.UnitPrice,
			TotalPrice: float64(itemInput.Quantity) * itemInput.UnitPrice,
		}
		_ = r.CreateItem(&item)
	}

	r.DeleteCarrierLinks(id)

	if len(input.CarrierIDs) > 0 {
		for _, carrierID := range input.CarrierIDs {
			_ = r.CreateCarrierLink(id, carrierID)
		}
	}

	return r.GetByID(id)
}

func (r *SaleRepository) Delete(id int) error {
	res, err := r.db.Exec("DELETE FROM sales WHERE id=$1", id)
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

func (r *SaleRepository) GetSeller(sellerID int) (*models.User, error) {
	var u models.User
	var rg, gender, contactEmail, fullAddress, contactPhone, notes sql.NullString
	var birthDate sql.NullTime
	
	err := r.db.QueryRow(
		`SELECT id, username, role, full_name, cpf, rg, birth_date, gender, status, 
		 contact_email, full_address, contact_phone, notes, created_at, updated_at 
		 FROM users WHERE id=$1`, sellerID).
		Scan(&u.ID, &u.Username, &u.Role, &u.FullName, &u.CPF, &rg, &birthDate, 
			&gender, &u.Status, &contactEmail, &fullAddress, &contactPhone, &notes, 
			&u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	

	if rg.Valid {
		u.RG = &rg.String
	}
	if gender.Valid {
		u.Gender = &gender.String
	}
	if contactEmail.Valid {
		u.ContactEmail = &contactEmail.String
	}
	if fullAddress.Valid {
		u.FullAddress = &fullAddress.String
	}
	if contactPhone.Valid {
		u.ContactPhone = &contactPhone.String
	}
	if notes.Valid {
		u.Notes = &notes.String
	}
	if birthDate.Valid {
		u.BirthDate = &birthDate.Time
	}
	
	return &u, nil
}

func (r *SaleRepository) GetCustomer(customerID int) (*models.Customer, error) {
	if customerID == 0 {
		return nil, nil
	}
	var c models.Customer
	err := r.db.QueryRow(`SELECT id, customer_type, status, name, cnpj, cpf, email, business_phone, mobile_phone, website, notes, created_at, updated_at FROM customers WHERE id=$1`, customerID).
		Scan(&c.ID, &c.CustomerType, &c.Status, &c.Name, &c.CNPJ, &c.CPF, &c.Email, &c.BusinessPhone, &c.MobilePhone, &c.Website, &c.Notes, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}

	addresses, _ := r.GetCustomerAddresses(customerID)
	c.Addresses = addresses
	return &c, nil
}

func (r *SaleRepository) GetCustomerAddresses(customerID int) ([]models.Address, error) {
	rows, err := r.db.Query(`SELECT id, customer_id, address_type, address, created_at, updated_at FROM customer_addresses WHERE customer_id=$1 ORDER BY CASE WHEN address_type='entrega' THEN 0 ELSE 1 END`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var addrs []models.Address
	for rows.Next() {
		var a models.Address
		err := rows.Scan(&a.ID, &a.CustomerID, &a.AddressType, &a.AddressLine, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, a)
	}
	return addrs, nil
}

func (r *SaleRepository) GetItems(saleID int) ([]models.SaleItem, error) {
	rows, err := r.db.Query(`SELECT id, sale_id, product_id, quantity, unit_price, total_price, created_at, updated_at FROM sale_items WHERE sale_id=$1`, saleID)
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

		product, _ := r.GetProductBasic(item.ProductID)
		item.Product = product

		items = append(items, item)
	}
	return items, nil
}

func (r *SaleRepository) CreateItem(item *models.SaleItem) error {
	err := r.db.QueryRow(
		`INSERT INTO sale_items (sale_id, product_id, quantity, unit_price, total_price) VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at, updated_at`,
		item.SaleID, item.ProductID, item.Quantity, item.UnitPrice, item.TotalPrice).
		Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	return err
}

func (r *SaleRepository) GetProductBasic(productID int) (*models.Product, error) {
	var p models.Product
	var productGroup, description, ncm, materialOrigin sql.NullString
	err := r.db.QueryRow(`SELECT id, product_name, internal_code, supplier_id, product_group, description, ncm, material_origin, stock, created_at, updated_at FROM products WHERE id=$1`, productID).
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

func (r *SaleRepository) GetCarriers(saleID int) ([]models.Carrier, error) {
	rows, err := r.db.Query(`
		SELECT c.id, c.name, c.carrier_type, c.email, c.landline_phone, c.mobile_phone, 
		       c.full_address, c.contact_name, c.contact_phone, c.website, 
		       c.created_at, c.updated_at
		FROM carriers c
		INNER JOIN sale_carriers sc ON c.id = sc.carrier_id
		WHERE sc.sale_id = $1
		ORDER BY c.name`, saleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var carriers []models.Carrier
	for rows.Next() {
		var c models.Carrier
		err := rows.Scan(&c.ID, &c.Name, &c.CarrierType, &c.Email, &c.LandlinePhone, 
			&c.MobilePhone, &c.FullAddress, &c.ContactName, &c.ContactPhone, &c.Website,
			&c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		carriers = append(carriers, c)
	}
	return carriers, nil
}

func (r *SaleRepository) CreateCarrierLink(saleID, carrierID int) error {
	_, err := r.db.Exec(
		`INSERT INTO sale_carriers (sale_id, carrier_id) VALUES ($1, $2)
		 ON CONFLICT (sale_id, carrier_id) DO NOTHING`,
		saleID, carrierID)
	return err
}

func (r *SaleRepository) DeleteCarrierLinks(saleID int) error {
	_, err := r.db.Exec("DELETE FROM sale_carriers WHERE sale_id=$1", saleID)
	return err
}
