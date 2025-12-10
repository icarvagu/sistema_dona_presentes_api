package repositories

import (
	"database/sql"
	"donapresentes/models"
	apperrors "donapresentes/errors"
)

type EmployeeRepository struct {
	db *sql.DB
}

func NewEmployeeRepository(db *sql.DB) *EmployeeRepository {
	return &EmployeeRepository{db: db}
}

// GetAll retrieves all employees
func (r *EmployeeRepository) GetAll() ([]models.Employee, error) {
	rows, err := r.db.Query("SELECT id, full_name, cpf, rg, birth_date, gender, status, contact_email, full_address, contact_phone, notes, created_at FROM funcionarios")
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	var employees []models.Employee
	for rows.Next() {
		var e models.Employee
		err := rows.Scan(&e.ID, &e.FullName, &e.CPF, &e.RG, &e.BirthDate, &e.Gender, &e.Status, &e.ContactEmail, &e.FullAddress, &e.ContactPhone, &e.Notes, &e.CreatedAt)
		if err != nil {
			return nil, err
		}
		employees = append(employees, e)
	}
	return employees, nil
}

// GetByID retrieves a single employee by ID
func (r *EmployeeRepository) GetByID(id int) (*models.Employee, error) {
	var e models.Employee
	err := r.db.QueryRow("SELECT id, full_name, cpf, rg, birth_date, gender, status, contact_email, full_address, contact_phone, notes, created_at FROM funcionarios WHERE id=$1", id).
		Scan(&e.ID, &e.FullName, &e.CPF, &e.RG, &e.BirthDate, &e.Gender, &e.Status, &e.ContactEmail, &e.FullAddress, &e.ContactPhone, &e.Notes, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// Create creates a new employee
func (r *EmployeeRepository) Create(e *models.Employee) error {
	err := r.db.QueryRow(
		"INSERT INTO funcionarios (full_name, cpf, rg, birth_date, gender, status, contact_email, full_address, contact_phone, notes) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id, created_at",
		e.FullName, e.CPF, e.RG, e.BirthDate, e.Gender, e.Status, e.ContactEmail, e.FullAddress, e.ContactPhone, e.Notes).
		Scan(&e.ID, &e.CreatedAt)
	return err
}

// Update updates an existing employee
func (r *EmployeeRepository) Update(id int, e *models.Employee) error {
	_, err := r.db.Exec(
		"UPDATE funcionarios SET full_name=$1, cpf=$2, rg=$3, birth_date=$4, gender=$5, status=$6, contact_email=$7, full_address=$8, contact_phone=$9, notes=$10 WHERE id=$11",
		e.FullName, e.CPF, e.RG, e.BirthDate, e.Gender, e.Status, e.ContactEmail, e.FullAddress, e.ContactPhone, e.Notes, id)
	return err
}

// Delete deletes an employee by ID
func (r *EmployeeRepository) Delete(id int) error {
	res, err := r.db.Exec("DELETE FROM funcionarios WHERE id=$1", id)
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
