package services

import (
	"database/sql"
	apperrors "donapresentes/errors"
	"donapresentes/models"
	"donapresentes/repositories"
)

type CarrierService struct {
	carrierRepo *repositories.CarrierRepository
}

func NewCarrierService(carrierRepo *repositories.CarrierRepository) *CarrierService {
	return &CarrierService{
		carrierRepo: carrierRepo,
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

	if c.Email != "" {
		if err := ValidateEmail(c.Email); err != nil {
			return err
		}
	}

	return s.carrierRepo.Create(c)
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

	if c.Email != "" {
		if err := ValidateEmail(c.Email); err != nil {
			return err
		}
	}

	return s.carrierRepo.Update(id, c)
}

func (s *CarrierService) Delete(id int) error {
	return s.carrierRepo.Delete(id)
}
