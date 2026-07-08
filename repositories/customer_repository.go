package repositories

import (
	"database/sql"
	apperrors "donapresentes/errors"
	"donapresentes/models"
)

type CustomerRepository struct {
	db *sql.DB
}

func NewCustomerRepository(db *sql.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) GetAll() ([]models.Customer, error) {
	rows, err := r.db.Query(`SELECT id, customer_type, status, name, cnpj, cpf, email, business_phone, mobile_phone, website, notes,
		COALESCE(trade_name,''), COALESCE(company_name,''), COALESCE(state_registration,''), COALESCE(city_registration,''),
		COALESCE(responsible,''), COALESCE(contact_financial_name,''), COALESCE(contact_financial_email,''), COALESCE(contact_financial_phone,''),
		COALESCE(contact_nf_name,''), COALESCE(contact_nf_email,''), COALESCE(contact_nf_phone,''),
		COALESCE(contact_commercial_name,''), COALESCE(contact_commercial_email,''), COALESCE(contact_commercial_phone,''),
		created_at, updated_at FROM customers`)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var c models.Customer
		err := rows.Scan(&c.ID, &c.CustomerType, &c.Status, &c.Name, &c.CNPJ, &c.CPF, &c.Email, &c.BusinessPhone, &c.MobilePhone, &c.Website, &c.Notes,
			&c.TradeName, &c.CompanyName, &c.StateRegistration, &c.CityRegistration,
			&c.Responsible, &c.ContactFinancialName, &c.ContactFinancialEmail, &c.ContactFinancialPhone,
			&c.ContactNFName, &c.ContactNFEmail, &c.ContactNFPhone,
			&c.ContactCommercialName, &c.ContactCommercialEmail, &c.ContactCommercialPhone,
			&c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, apperrors.NewDatabaseError(err)
		}
		customers = append(customers, c)
	}
	return customers, nil
}

func (r *CustomerRepository) GetByID(id int) (*models.Customer, error) {
	var c models.Customer
	err := r.db.QueryRow(`SELECT id, customer_type, status, name, cnpj, cpf, email, business_phone, mobile_phone, website, notes,
		COALESCE(trade_name,''), COALESCE(company_name,''), COALESCE(state_registration,''), COALESCE(city_registration,''),
		COALESCE(responsible,''), COALESCE(contact_financial_name,''), COALESCE(contact_financial_email,''), COALESCE(contact_financial_phone,''),
		COALESCE(contact_nf_name,''), COALESCE(contact_nf_email,''), COALESCE(contact_nf_phone,''),
		COALESCE(contact_commercial_name,''), COALESCE(contact_commercial_email,''), COALESCE(contact_commercial_phone,''),
		created_at, updated_at FROM customers WHERE id=$1`, id).
		Scan(&c.ID, &c.CustomerType, &c.Status, &c.Name, &c.CNPJ, &c.CPF, &c.Email, &c.BusinessPhone, &c.MobilePhone, &c.Website, &c.Notes,
			&c.TradeName, &c.CompanyName, &c.StateRegistration, &c.CityRegistration,
			&c.Responsible, &c.ContactFinancialName, &c.ContactFinancialEmail, &c.ContactFinancialPhone,
			&c.ContactNFName, &c.ContactNFEmail, &c.ContactNFPhone,
			&c.ContactCommercialName, &c.ContactCommercialEmail, &c.ContactCommercialPhone,
			&c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CustomerRepository) Create(c *models.Customer) error {
	err := r.db.QueryRow(
		`INSERT INTO customers (customer_type, status, name, cnpj, cpf, email, business_phone, mobile_phone, website, notes,
			trade_name, company_name, state_registration, city_registration, responsible,
			contact_financial_name, contact_financial_email, contact_financial_phone,
			contact_nf_name, contact_nf_email, contact_nf_phone,
			contact_commercial_name, contact_commercial_email, contact_commercial_phone)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24)
		 RETURNING id, created_at, updated_at`,
		c.CustomerType, c.Status, c.Name, c.CNPJ, c.CPF, c.Email, c.BusinessPhone, c.MobilePhone, c.Website, c.Notes,
		c.TradeName, c.CompanyName, c.StateRegistration, c.CityRegistration, c.Responsible,
		c.ContactFinancialName, c.ContactFinancialEmail, c.ContactFinancialPhone,
		c.ContactNFName, c.ContactNFEmail, c.ContactNFPhone,
		c.ContactCommercialName, c.ContactCommercialEmail, c.ContactCommercialPhone).
		Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	return err
}

func (r *CustomerRepository) Update(id int, c *models.Customer) error {
	_, err := r.db.Exec(
		`UPDATE customers SET customer_type=$1, status=$2, name=$3, cnpj=$4, cpf=$5, email=$6, business_phone=$7, mobile_phone=$8, website=$9, notes=$10,
			trade_name=$11, company_name=$12, state_registration=$13, city_registration=$14, responsible=$15,
			contact_financial_name=$16, contact_financial_email=$17, contact_financial_phone=$18,
			contact_nf_name=$19, contact_nf_email=$20, contact_nf_phone=$21,
			contact_commercial_name=$22, contact_commercial_email=$23, contact_commercial_phone=$24,
			updated_at=NOW() WHERE id=$25`,
		c.CustomerType, c.Status, c.Name, c.CNPJ, c.CPF, c.Email, c.BusinessPhone, c.MobilePhone, c.Website, c.Notes,
		c.TradeName, c.CompanyName, c.StateRegistration, c.CityRegistration, c.Responsible,
		c.ContactFinancialName, c.ContactFinancialEmail, c.ContactFinancialPhone,
		c.ContactNFName, c.ContactNFEmail, c.ContactNFPhone,
		c.ContactCommercialName, c.ContactCommercialEmail, c.ContactCommercialPhone, id)
	return err
}

func (r *CustomerRepository) Delete(id int) error {
	res, err := r.db.Exec("DELETE FROM customers WHERE id=$1", id)
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
