package services

import (
	"database/sql"
	"fmt"

	apperrors "donapresentes/errors"
	"donapresentes/models"
	"donapresentes/repositories"
)

type CarrierService struct {
	carrierRepo  repositories.CarrierRepositoryInterface
	auditService *AuditService
}

func NewCarrierService(carrierRepo repositories.CarrierRepositoryInterface, auditService *AuditService) *CarrierService {
	return &CarrierService{
		carrierRepo: carrierRepo,
		auditService: auditService,
	}
}

func (s *CarrierService) GetAll() ([]models.Carrier, error) {
	return s.carrierRepo.GetAll()
}

func (s *CarrierService) GetByID(id int) (*models.Carrier, error) {
	return s.carrierRepo.GetByID(id)
}

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

func (s *CarrierService) Delete(id int) error {
	err := s.carrierRepo.Delete(id)
	if err == nil {
		s.auditService.LogSimple(nil, "carrier_deleted", "carrier", fmt.Sprintf("id=%d", id), "")
	}
	return err
}
