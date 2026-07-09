package services

import (
	"database/sql"
	apperrors "donapresentes/errors"
	"donapresentes/models"
	"donapresentes/repositories"
)

type SupplierService struct {
	supplierRepo repositories.SupplierRepositoryInterface
}

func NewSupplierService(supplierRepo repositories.SupplierRepositoryInterface) *SupplierService {
	return &SupplierService{
		supplierRepo: supplierRepo,
	}
}

func (s *SupplierService) GetAll() ([]models.Supplier, error) {
	return s.supplierRepo.GetAll()
}

func (s *SupplierService) GetByID(id int) (*models.Supplier, error) {
	return s.supplierRepo.GetByID(id)
}

func (s *SupplierService) Create(supplier *models.Supplier) error {

	if supplier.Name == "" {
		return apperrors.NewMissingFieldError("name")
	}

	if supplier.Email != nil && *supplier.Email != "" {
		if err := ValidateEmail(*supplier.Email); err != nil {
			return err
		}
	}

	if supplier.ResponsibleEmail != nil && *supplier.ResponsibleEmail != "" {
		if err := ValidateEmail(*supplier.ResponsibleEmail); err != nil {
			return err
		}
	}

	return s.supplierRepo.Create(supplier)
}

func (s *SupplierService) Update(id int, supplier *models.Supplier) error {

	_, err := s.supplierRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperrors.ErrSupplierNotFound
		}
		return err
	}

	if supplier.Name == "" {
		return apperrors.NewMissingFieldError("name")
	}

	if supplier.Email != nil && *supplier.Email != "" {
		if err := ValidateEmail(*supplier.Email); err != nil {
			return err
		}
	}

	if supplier.ResponsibleEmail != nil && *supplier.ResponsibleEmail != "" {
		if err := ValidateEmail(*supplier.ResponsibleEmail); err != nil {
			return err
		}
	}

	return s.supplierRepo.Update(id, supplier)
}

func (s *SupplierService) Delete(id int) error {
	return s.supplierRepo.Delete(id)
}
