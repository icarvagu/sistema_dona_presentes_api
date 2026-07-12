package services

import (
	"database/sql"
	"fmt"

	apperrors "donapresentes/errors"
	"donapresentes/models"
	"donapresentes/repositories"
)

// SupplierService manages supplier records with CRUD operations.
// Validates name presence and email format for primary and responsible contacts.
type SupplierService struct {
	supplierRepo repositories.SupplierRepositoryInterface
	auditService *AuditService
}

// NewSupplierService creates a SupplierService with the given repository and audit trail.
func NewSupplierService(supplierRepo repositories.SupplierRepositoryInterface, auditService *AuditService) *SupplierService {
	return &SupplierService{
		supplierRepo: supplierRepo,
		auditService: auditService,
	}
}

// GetAll returns all registered suppliers.
func (s *SupplierService) GetAll() ([]models.Supplier, error) {
	return s.supplierRepo.GetAll()
}

// GetByID returns a single supplier by its ID.
func (s *SupplierService) GetByID(id int) (*models.Supplier, error) {
	return s.supplierRepo.GetByID(id)
}

// Create registers a new supplier.
// Name is required. Email and responsible email are validated if provided.
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

	err := s.supplierRepo.Create(supplier)
	if err == nil {
		s.auditService.LogSimple(nil, "supplier_created", "supplier", fmt.Sprintf("name=%s", supplier.Name), "")
	}
	return err
}

// Update modifies an existing supplier by ID.
// Returns ErrSupplierNotFound if the supplier does not exist.
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

	err = s.supplierRepo.Update(id, supplier)
	if err == nil {
		s.auditService.LogSimple(nil, "supplier_updated", "supplier", fmt.Sprintf("id=%d name=%s", id, supplier.Name), "")
	}
	return err
}

// Delete removes a supplier by its ID.
func (s *SupplierService) Delete(id int) error {
	err := s.supplierRepo.Delete(id)
	if err == nil {
		s.auditService.LogSimple(nil, "supplier_deleted", "supplier", fmt.Sprintf("id=%d", id), "")
	}
	return err
}
