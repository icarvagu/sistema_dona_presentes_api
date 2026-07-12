package services

import (
	"database/sql"
	"fmt"

	apperrors "donapresentes/errors"
	"donapresentes/models"
	"donapresentes/repositories"
)

// CarrierService manages transport carriers (shipping companies) with CRUD operations.
// Validates carrier type ("Pessoa Jurídica" or "Pessoa Física") and email format.
type CarrierService struct {
	carrierRepo  repositories.CarrierRepositoryInterface
	auditService *AuditService
}

// NewCarrierService creates a CarrierService with the given repository and audit trail.
func NewCarrierService(carrierRepo repositories.CarrierRepositoryInterface, auditService *AuditService) *CarrierService {
	return &CarrierService{
		carrierRepo: carrierRepo,
		auditService: auditService,
	}
}

// GetAll returns all registered carriers.
func (s *CarrierService) GetAll() ([]models.Carrier, error) {
	return s.carrierRepo.GetAll()
}

// GetByID returns a single carrier by its ID.
func (s *CarrierService) GetByID(id int) (*models.Carrier, error) {
	return s.carrierRepo.GetByID(id)
}

// Create registers a new carrier.
// Name and carrier_type are required. Carrier type must be "Pessoa Jurídica" or "Pessoa Física".
// Email, if provided, is validated.
func (s *CarrierService) Create(c *models.Carrier) error {

	if c.Name == "" {
		return apperrors.NewMissingFieldError("name")
	}
	if c.CarrierType == "" {
		return apperrors.NewMissingFieldError("carrier_type")
	}

	if c.CarrierType != "Pessoa Jurídica" && c.CarrierType != "Pessoa Física" {
		return apperrors.NewInvalidFieldError("carrier_type", "must be 'Pessoa Jurídica' or 'Pessoa Física'")
	}

	if c.Email != nil && *c.Email != "" {
		if err := ValidateEmail(*c.Email); err != nil {
			return err
		}
	}

	err := s.carrierRepo.Create(c)
	if err == nil {
		s.auditService.LogSimple(nil, "carrier_created", "carrier", fmt.Sprintf("name=%s", c.Name), "")
	}
	return err
}

// Update modifies an existing carrier by ID.
// Returns ErrCarrierNotFound if the carrier does not exist.
func (s *CarrierService) Update(id int, c *models.Carrier) error {

	_, err := s.carrierRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperrors.ErrCarrierNotFound
		}
		return err
	}

	if c.Name == "" {
		return apperrors.NewMissingFieldError("name")
	}
	if c.CarrierType == "" {
		return apperrors.NewMissingFieldError("carrier_type")
	}

	if c.CarrierType != "Pessoa Jurídica" && c.CarrierType != "Pessoa Física" {
		return apperrors.NewInvalidFieldError("carrier_type", "must be 'Pessoa Jurídica' or 'Pessoa Física'")
	}

	if c.Email != nil && *c.Email != "" {
		if err := ValidateEmail(*c.Email); err != nil {
			return err
		}
	}

	err = s.carrierRepo.Update(id, c)
	if err == nil {
		s.auditService.LogSimple(nil, "carrier_updated", "carrier", fmt.Sprintf("id=%d name=%s", id, c.Name), "")
	}
	return err
}

// Delete removes a carrier by its ID.
func (s *CarrierService) Delete(id int) error {
	err := s.carrierRepo.Delete(id)
	if err == nil {
		s.auditService.LogSimple(nil, "carrier_deleted", "carrier", fmt.Sprintf("id=%d", id), "")
	}
	return err
}
