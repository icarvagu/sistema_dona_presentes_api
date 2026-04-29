package services

import (
	"database/sql"
	apperrors "donapresentes/errors"
	"donapresentes/models"
	"donapresentes/repositories"
)

type SupplierService struct {
	supplierRepo *repositories.SupplierRepository
}

func NewSupplierService(supplierRepo *repositories.SupplierRepository) *SupplierService {
	return &SupplierService{
		supplierRepo: supplierRepo,
	}
}

// GetAll returns all suppliers
func (s *SupplierService) GetAll() ([]models.Supplier, error) {
	return s.supplierRepo.GetAll()
}

// GetByID returns a supplier by ID
func (s *SupplierService) GetByID(id int) (*models.Supplier, error) {
	return s.supplierRepo.GetByID(id)
}

// Create creates a new supplier with validation
func (s *SupplierService) Create(supplier *models.Supplier) error {
	// Validate required fields
	if supplier.Name == "" {
		return apperrors.NewMissingFieldError("name")
	}

	// Validate email if provided
	if supplier.Email != nil && *supplier.Email != "" {
		if err := ValidateEmail(*supplier.Email); err != nil {
			return err
		}
	}

	// Validate responsible email if provided
	if supplier.ResponsibleEmail != nil && *supplier.ResponsibleEmail != "" {
		if err := ValidateEmail(*supplier.ResponsibleEmail); err != nil {
			return err
		}
	}

	return s.supplierRepo.Create(supplier)
}

// Update updates an existing supplier with validation
func (s *SupplierService) Update(id int, supplier *models.Supplier) error {
	// Validate supplier exists
	_, err := s.supplierRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperrors.ErrSupplierNotFound
		}
		return err
	}

	// Validate required fields
	if supplier.Name == "" {
		return apperrors.NewMissingFieldError("name")
	}

	// Validate email if provided
	if supplier.Email != nil && *supplier.Email != "" {
		if err := ValidateEmail(*supplier.Email); err != nil {
			return err
		}
	}

	// Validate responsible email if provided
	if supplier.ResponsibleEmail != nil && *supplier.ResponsibleEmail != "" {
		if err := ValidateEmail(*supplier.ResponsibleEmail); err != nil {
			return err
		}
	}

	return s.supplierRepo.Update(id, supplier)
}

// Delete deletes a supplier
func (s *SupplierService) Delete(id int) error {
	return s.supplierRepo.Delete(id)
}
