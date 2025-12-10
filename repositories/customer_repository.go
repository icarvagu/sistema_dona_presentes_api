package repositories

import (
	"database/sql"
	"donapresentes/models"
	apperrors "donapresentes/errors"
	"regexp"
)

type CustomerRepository struct {
	db *sql.DB
}

func NewCustomerRepository(db *sql.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

// Validation helpers
func ValidateEmail(email string) error {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	if !regexp.MustCompile(pattern).MatchString(email) {
		return apperrors.ErrInvalidEmail
	}
	return nil
}

func ValidateCPF(cpf string) error {
	cpf = regexp.MustCompile(`\D`).ReplaceAllString(cpf, "")
	if len(cpf) != 11 {
		return apperrors.ErrInvalidCPF
	}
	return nil
}

func ValidateCNPJ(cnpj string) error {
	cnpj = regexp.MustCompile(`\D`).ReplaceAllString(cnpj, "")
	if len(cnpj) != 14 {
		return apperrors.ErrInvalidCNPJ
	}
	return nil
}

// GetAll returns all customers
func (r *CustomerRepository) GetAll() ([]models.Customer, error) {
	rows, err := r.db.Query(`SELECT id, customer_type, status, name, cnpj, cpf, email, business_phone, mobile_phone, website, notes, created_at, updated_at FROM clientes`)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var c models.Customer
		err := rows.Scan(&c.ID, &c.CustomerType, &c.Status, &c.Name, &c.CNPJ, &c.CPF, &c.Email, &c.BusinessPhone, &c.MobilePhone, &c.Website, &c.Notes, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, apperrors.NewDatabaseError(err)
		}
		customers = append(customers, c)
	}
	return customers, nil
}

// GetByID returns a single customer by ID
func (r *CustomerRepository) GetByID(id int) (*models.Customer, error) {
	var c models.Customer
	err := r.db.QueryRow(`SELECT id, customer_type, status, name, cnpj, cpf, email, business_phone, mobile_phone, website, notes, created_at, updated_at FROM clientes WHERE id=$1`, id).
		Scan(&c.ID, &c.CustomerType, &c.Status, &c.Name, &c.CNPJ, &c.CPF, &c.Email, &c.BusinessPhone, &c.MobilePhone, &c.Website, &c.Notes, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Create inserts a new customer
func (r *CustomerRepository) Create(c *models.Customer) error {
	err := r.db.QueryRow(
		`INSERT INTO clientes (customer_type, status, name, cnpj, cpf, email, business_phone, mobile_phone, website, notes) 
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) 
		 RETURNING id, created_at, updated_at`,
		c.CustomerType, c.Status, c.Name, c.CNPJ, c.CPF, c.Email, c.BusinessPhone, c.MobilePhone, c.Website, c.Notes).
		Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	return err
}

// Update updates an existing customer
func (r *CustomerRepository) Update(id int, c *models.Customer) error {
	_, err := r.db.Exec(
		`UPDATE clientes SET customer_type=$1, status=$2, name=$3, cnpj=$4, cpf=$5, email=$6, business_phone=$7, mobile_phone=$8, website=$9, notes=$10, updated_at=NOW() WHERE id=$11`,
		c.CustomerType, c.Status, c.Name, c.CNPJ, c.CPF, c.Email, c.BusinessPhone, c.MobilePhone, c.Website, c.Notes, id)
	return err
}

// Delete deletes a customer by ID
func (r *CustomerRepository) Delete(id int) error {
	res, err := r.db.Exec("DELETE FROM clientes WHERE id=$1", id)
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
