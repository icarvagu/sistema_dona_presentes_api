package services

import (
	"database/sql"
	apperrors "donapresentes/errors"
	"donapresentes/models"
	"donapresentes/repositories"
)

type QuoteService struct {
	quoteRepo    *repositories.QuoteRepository
	userRepo     *repositories.UserRepository
	productRepo  *repositories.ProductRepository
	customerRepo *repositories.CustomerRepository
}

func NewQuoteService(quoteRepo *repositories.QuoteRepository, userRepo *repositories.UserRepository, productRepo *repositories.ProductRepository, customerRepo *repositories.CustomerRepository) *QuoteService {
	return &QuoteService{
		quoteRepo:    quoteRepo,
		userRepo:     userRepo,
		productRepo:  productRepo,
		customerRepo: customerRepo,
	}
}

func (s *QuoteService) GetAll() ([]models.Quote, error) {
	return s.quoteRepo.GetAll()
}

func (s *QuoteService) GetBySellerID(sellerID int) ([]models.Quote, error) {
	return s.quoteRepo.GetBySellerID(sellerID)
}

func (s *QuoteService) GetByID(id int) (*models.Quote, error) {
	return s.quoteRepo.GetByID(id)
}

func (s *QuoteService) ValidateQuote(input *models.QuoteInput) error {
	if input.SellerID == 0 {
		return apperrors.NewMissingFieldError("seller_id")
	}
	if input.CustomerID == 0 {
		return apperrors.NewMissingFieldError("customer_id")
	}
	if input.ResponsibleName == "" {
		return apperrors.NewMissingFieldError("responsible_name")
	}
	if input.QuoteValidUntil == nil {
		return apperrors.NewMissingFieldError("quote_valid_until")
	}
	if len(input.Items) == 0 {
		return apperrors.NewInvalidFieldError("items", "quote must have at least 1 item")
	}

	_, err := s.userRepo.GetByID(input.SellerID)
	if err != nil {
		if err == sql.ErrNoRows || apperrors.IsNotFound(err) {
			return apperrors.NewNotFoundError("Vendedor")
		}
		return apperrors.NewDatabaseError(err)
	}

	_, err = s.customerRepo.GetByID(input.CustomerID)
	if err != nil {
		if err == sql.ErrNoRows || apperrors.IsNotFound(err) {
			return apperrors.NewNotFoundError("Cliente")
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

	return nil
}

func (s *QuoteService) Create(input *models.QuoteInput) (*models.Quote, error) {
	if err := s.ValidateQuote(input); err != nil {
		return nil, err
	}
	return s.quoteRepo.Create(input)
}

func (s *QuoteService) Update(id int, input *models.QuoteInput) (*models.Quote, error) {
	_, err := s.quoteRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrQuoteNotFound
		}
		return nil, err
	}
	if err := s.ValidateQuote(input); err != nil {
		return nil, err
	}
	return s.quoteRepo.Update(id, input)
}

func (s *QuoteService) Delete(id int) error {
	return s.quoteRepo.Delete(id)
}