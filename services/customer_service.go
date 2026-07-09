package services

import (
	"database/sql"
	apperrors "donapresentes/errors"
	"donapresentes/models"
	"donapresentes/repositories"
)

type CustomerService struct {
	customerRepo repositories.CustomerRepositoryInterface
}

func NewCustomerService(customerRepo repositories.CustomerRepositoryInterface) *CustomerService {
	return &CustomerService{
		customerRepo: customerRepo,
	}
}

func (s *CustomerService) GetAll() ([]models.Customer, error) {
	return s.customerRepo.GetAll()
}

func (s *CustomerService) GetByID(id int) (*models.Customer, error) {
	return s.customerRepo.GetByID(id)
}

func (s *CustomerService) Create(c *models.Customer) error {

	if c.Name == "" {
		return apperrors.NewMissingFieldError("name")
	}
	if c.CustomerType == "" {
		return apperrors.NewMissingFieldError("customer_type")
	}
	if c.Status == "" {
		return apperrors.NewMissingFieldError("status")
	}

	if c.CustomerType != "PF" && c.CustomerType != "PJ" {
		return apperrors.NewInvalidFieldError("customer_type", "must be 'PF' or 'PJ'")
	}

	if c.Status != "Ativo" && c.Status != "Inativo" {
		return apperrors.NewInvalidFieldError("status", "must be 'Ativo' or 'Inativo'")
	}

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

	if c.Email != "" {
		if err := ValidateEmail(c.Email); err != nil {
			return err
		}
	}

	return s.customerRepo.Create(c)
}

func (s *CustomerService) Update(id int, c *models.Customer) error {

	_, err := s.customerRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperrors.ErrCustomerNotFound
		}
		return err
	}

	if c.Name == "" {
		return apperrors.NewMissingFieldError("name")
	}
	if c.CustomerType == "" {
		return apperrors.NewMissingFieldError("customer_type")
	}
	if c.Status == "" {
		return apperrors.NewMissingFieldError("status")
	}

	if c.CustomerType != "PF" && c.CustomerType != "PJ" {
		return apperrors.NewInvalidFieldError("customer_type", "must be 'PF' or 'PJ'")
	}

	if c.Status != "Ativo" && c.Status != "Inativo" {
		return apperrors.NewInvalidFieldError("status", "must be 'Ativo' or 'Inativo'")
	}

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

	if c.Email != "" {
		if err := ValidateEmail(c.Email); err != nil {
			return err
		}
	}

	return s.customerRepo.Update(id, c)
}

func (s *CustomerService) Delete(id int) error {
	return s.customerRepo.Delete(id)
}
