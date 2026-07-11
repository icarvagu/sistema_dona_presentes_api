package services

import (
	"database/sql"
	"fmt"

	apperrors "donapresentes/errors"
	"donapresentes/models"
	"donapresentes/repositories"
)

type SaleService struct {
	saleRepo      *repositories.SaleRepository
	userRepo      *repositories.UserRepository
	productRepo   *repositories.ProductRepository
	customerRepo  *repositories.CustomerRepository
	carrierRepo   *repositories.CarrierRepository
	auditService  *AuditService
}

func NewSaleService(saleRepo *repositories.SaleRepository, userRepo *repositories.UserRepository, productRepo *repositories.ProductRepository, customerRepo *repositories.CustomerRepository, carrierRepo *repositories.CarrierRepository, auditService *AuditService) *SaleService {
	return &SaleService{
		saleRepo:     saleRepo,
		userRepo:     userRepo,
		productRepo:  productRepo,
		customerRepo: customerRepo,
		carrierRepo:  carrierRepo,
		auditService: auditService,
	}
}

func (s *SaleService) GetAll() ([]models.Sale, error) {
	return s.saleRepo.GetAll()
}

func (s *SaleService) GetBySellerID(sellerID int) ([]models.Sale, error) {
	return s.saleRepo.GetBySellerID(sellerID)
}

func (s *SaleService) GetByID(id int) (*models.Sale, error) {
	return s.saleRepo.GetByID(id)
}

func (s *SaleService) ValidateSale(input *models.SaleInput) error {

	if input.SellerID == 0 {
		return apperrors.NewMissingFieldError("seller_id")
	}
	if input.CustomerID == 0 {
		return apperrors.NewMissingFieldError("customer_id")
	}
	if input.PaymentMethod == "" {
		return apperrors.NewMissingFieldError("payment_method")
	}
	if input.Installments <= 0 {
		return apperrors.NewInvalidFieldError("installments", "must be > 0")
	}
	if len(input.Items) == 0 {
		return apperrors.NewInvalidFieldError("items", "sale must have at least 1 item")
	}

	_, err := s.userRepo.GetByID(input.SellerID)
	if err != nil {
		if err == sql.ErrNoRows || apperrors.IsNotFound(err) {
			return apperrors.NewNotFoundError("Vendedor não encontrado")
		}
		return apperrors.NewDatabaseError(err)
	}

	_, err = s.customerRepo.GetByID(input.CustomerID)
	if err != nil {
		if err == sql.ErrNoRows || apperrors.IsNotFound(err) {
			return apperrors.NewNotFoundError("Cliente não encontrado")
		}
		return apperrors.NewDatabaseError(err)
	}

	for _, item := range input.Items {
		if item.ProductID == 0 {
			return apperrors.NewInvalidFieldError("item", "product_id is required")
		}
		if item.Quantity <= 0 {
			return apperrors.NewInvalidFieldError("item", "quantity must be > 0")
		}
		if item.UnitPrice <= 0 {
			return apperrors.NewInvalidFieldError("item", "unit_price must be > 0")
		}

		_, err := s.productRepo.GetByID(item.ProductID)
		if err != nil {
			if err == sql.ErrNoRows {
				return apperrors.ErrProductNotFound
			}
			return apperrors.NewDatabaseError(err)
		}
	}

	if len(input.CarrierIDs) > 0 {
		for _, carrierID := range input.CarrierIDs {
			if carrierID <= 0 {
				return apperrors.NewInvalidFieldError("carrier_ids", "carrier_id must be > 0")
			}
			_, err := s.carrierRepo.GetByID(carrierID)
			if err != nil {
				if err == sql.ErrNoRows {
					return apperrors.NewNotFoundError("Carrier not found")
				}
				return apperrors.NewDatabaseError(err)
			}
		}
	}

	return nil
}

func (s *SaleService) Create(input *models.SaleInput) (*models.Sale, error) {

	if err := s.ValidateSale(input); err != nil {
		return nil, err
	}

	sale, err := s.saleRepo.Create(input)
	if err == nil {
		s.auditService.LogSimple(&input.SellerID, "sale_created", "sale", fmt.Sprintf("id=%d customer_id=%d", sale.ID, input.CustomerID), "")
	}
	return sale, err
}

func (s *SaleService) Update(id int, input *models.SaleInput) (*models.Sale, error) {

	_, err := s.saleRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrSaleNotFound
		}
		return nil, err
	}

	if err := s.ValidateSale(input); err != nil {
		return nil, err
	}

	sale, err := s.saleRepo.Update(id, input)
	if err == nil {
		s.auditService.LogSimple(&sale.SellerID, "sale_updated", "sale", fmt.Sprintf("id=%d", id), "")
	}
	return sale, err
}

func (s *SaleService) Delete(id int) error {
	err := s.saleRepo.Delete(id)
	if err == nil {
		s.auditService.LogSimple(nil, "sale_deleted", "sale", fmt.Sprintf("id=%d", id), "")
	}
	return err
}

func (s *SaleService) UpdateLayoutURLs(id int, urls []string) error {
	return s.saleRepo.UpdateLayoutURLs(id, urls)
}
