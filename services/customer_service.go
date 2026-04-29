package services

import (
	"database/sql"
	apperrors "donapresentes/errors"
	"donapresentes/models"
	"donapresentes/repositories"
)

type CustomerService struct {
	customerRepo *repositories.CustomerRepository
}

func NewCustomerService(customerRepo *repositories.CustomerRepository) *CustomerService {
	return &CustomerService{
		customerRepo: customerRepo,
	}
}

// GetAll returns all customers
func (s *CustomerService) GetAll() ([]models.Customer, error) {
	return s.customerRepo.GetAll()
}

// GetByID returns a customer by ID
func (s *CustomerService) GetByID(id int) (*models.Customer, error) {
	return s.customerRepo.GetByID(id)
}

// Create creates a new customer with validation
func (s *CustomerService) Create(c *models.Customer) error {
	// Validate required fields
	if c.Name == "" {
		return apperrors.NewMissingFieldError("name")
	}
	if c.CustomerType == "" {
		return apperrors.NewMissingFieldError("customer_type")
	}
	if c.Status == "" {
		return apperrors.NewMissingFieldError("status")
	}

	// Validate customer type
	if c.CustomerType != "PF" && c.CustomerType != "PJ" {
		return apperrors.NewInvalidFieldError("customer_type", "must be 'PF' or 'PJ'")
	}

	// Validate status
	if c.Status != "Ativo" && c.Status != "Inativo" {
		return apperrors.NewInvalidFieldError("status", "must be 'Ativo' or 'Inativo'")
	}

	// Validate CPF/CNPJ based on customer type
	if c.CustomerType == "PF" {
		if c.CPF != nil && *c.CPF != "" {
			if err := ValidateCPF(*c.CPF); err != nil {
				return err
			}
		}
	} else if c.CustomerType == "PJ" {
		if c.CNPJ != nil && *c.CNPJ != "" {
			if err := ValidateCNPJ(*c.CNPJ); err != nil {
				return err
			}
		}
	}

	// Validate email if provided
	if c.Email != "" {
		if err := ValidateEmail(c.Email); err != nil {
			return err
		}
	}

	return s.customerRepo.Create(c)
}

// Update updates an existing customer with validation
func (s *CustomerService) Update(id int, c *models.Customer) error {
	// Validate customer exists
	_, err := s.customerRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperrors.ErrCustomerNotFound
		}
		return err
	}

	// Validate required fields
	if c.Name == "" {
		return apperrors.NewMissingFieldError("name")
	}
	if c.CustomerType == "" {
		return apperrors.NewMissingFieldError("customer_type")
	}
	if c.Status == "" {
		return apperrors.NewMissingFieldError("status")
	}

	// Validate customer type
	if c.CustomerType != "PF" && c.CustomerType != "PJ" {
		return apperrors.NewInvalidFieldError("customer_type", "must be 'PF' or 'PJ'")
	}

	// Validate status
	if c.Status != "Ativo" && c.Status != "Inativo" {
		return apperrors.NewInvalidFieldError("status", "must be 'Ativo' or 'Inativo'")
	}

	// Validate CPF/CNPJ based on customer type
	if c.CustomerType == "PF" {
		if c.CPF != nil && *c.CPF != "" {
			if err := ValidateCPF(*c.CPF); err != nil {
				return err
			}
		}
	} else if c.CustomerType == "PJ" {
		if c.CNPJ != nil && *c.CNPJ != "" {
			if err := ValidateCNPJ(*c.CNPJ); err != nil {
				return err
			}
		}
	}

	// Validate email if provided
	if c.Email != "" {
		if err := ValidateEmail(c.Email); err != nil {
			return err
		}
	}

	return s.customerRepo.Update(id, c)
}

// Delete deletes a customer
func (s *CustomerService) Delete(id int) error {
	return s.customerRepo.Delete(id)
}
