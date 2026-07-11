package services

import (
	"database/sql"
	"fmt"

	apperrors "donapresentes/errors"
	"donapresentes/models"
	"donapresentes/repositories"
)

type SupplierService struct {
	supplierRepo repositories.SupplierRepositoryInterface
	auditService *AuditService
}

func NewSupplierService(supplierRepo repositories.SupplierRepositoryInterface, auditService *AuditService) *SupplierService {
	return &SupplierService{
		supplierRepo: supplierRepo,
		auditService: auditService,
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

	err := s.supplierRepo.Create(supplier)
	if err == nil {
		s.auditService.LogSimple(nil, "supplier_created", "supplier", fmt.Sprintf("name=%s", supplier.Name), "")
	}
	return err
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

	err = s.supplierRepo.Update(id, supplier)
	if err == nil {
		s.auditService.LogSimple(nil, "supplier_updated", "supplier", fmt.Sprintf("id=%d name=%s", id, supplier.Name), "")
	}
	return err
}

func (s *SupplierService) Delete(id int) error {
	err := s.supplierRepo.Delete(id)
	if err == nil {
		s.auditService.LogSimple(nil, "supplier_deleted", "supplier", fmt.Sprintf("id=%d", id), "")
	}
	return err
}
