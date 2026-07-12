package services

import (
	"database/sql"
	"fmt"

	apperrors "donapresentes/errors"
	"donapresentes/models"
	"donapresentes/repositories"
)

// SaleService manages sales order operations including CRUD, validation,
// and layout URL updates. It enforces that all referenced sellers, customers,
// products, and carriers exist before creating or updating a sale.
type SaleService struct {
	saleRepo     *repositories.SaleRepository
	userRepo     *repositories.UserRepository
	productRepo  *repositories.ProductRepository
	customerRepo *repositories.CustomerRepository
	carrierRepo  *repositories.CarrierRepository
	auditService *AuditService
}

// NewSaleService creates a SaleService with the required repositories and audit trail.
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

// GetAll returns all sales orders.
func (s *SaleService) GetAll() ([]models.Sale, error) {
	return s.saleRepo.GetAll()
}

// GetBySellerID returns all sales orders for a given seller.
func (s *SaleService) GetBySellerID(sellerID int) ([]models.Sale, error) {
	return s.saleRepo.GetBySellerID(sellerID)
}

// GetByID returns a single sale by its ID.
func (s *SaleService) GetByID(id int) (*models.Sale, error) {
	return s.saleRepo.GetByID(id)
}

// ValidateSale enforces that a sale has a seller, customer, payment method,
// at least one installment, and at least one item. All referenced entities
// (seller, customer, products, carriers) must exist. Each item must have
// a valid product, positive quantity, and positive unit price.
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

// Create registers a new sale after validation.
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

// Update modifies an existing sale by ID after validation.
// Returns ErrSaleNotFound if the sale does not exist.
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

// Delete removes a sale by its ID.
func (s *SaleService) Delete(id int) error {
	err := s.saleRepo.Delete(id)
	if err == nil {
		s.auditService.LogSimple(nil, "sale_deleted", "sale", fmt.Sprintf("id=%d", id), "")
	}
	return err
}

// UpdateLayoutURLs updates the layout image URLs associated with a sale.
func (s *SaleService) UpdateLayoutURLs(id int, urls []string) error {
	return s.saleRepo.UpdateLayoutURLs(id, urls)
}
