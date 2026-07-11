package services

import (
	"database/sql"
	"fmt"
	"time"

	apperrors "donapresentes/errors"
	"donapresentes/models"
	"donapresentes/repositories"
)

type QuoteService struct {
	quoteRepo    *repositories.QuoteRepository
	userRepo     *repositories.UserRepository
	productRepo  *repositories.ProductRepository
	customerRepo *repositories.CustomerRepository
	auditService *AuditService
}

func NewQuoteService(quoteRepo *repositories.QuoteRepository, userRepo *repositories.UserRepository, productRepo *repositories.ProductRepository, customerRepo *repositories.CustomerRepository, auditService *AuditService) *QuoteService {
	return &QuoteService{
		quoteRepo:    quoteRepo,
		userRepo:     userRepo,
		productRepo:  productRepo,
		customerRepo: customerRepo,
		auditService: auditService,
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
	quote, err := s.quoteRepo.Create(input)
	if err == nil {
		s.auditService.LogSimple(&input.SellerID, "quote_created", "quote", fmt.Sprintf("id=%d customer_id=%d", quote.ID, input.CustomerID), "")
	}
	return quote, err
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
	quote, err := s.quoteRepo.Update(id, input)
	if err == nil {
		s.auditService.LogSimple(nil, "quote_updated", "quote", fmt.Sprintf("id=%d", id), "")
	}
	return quote, err
}

func (s *QuoteService) UpdateFeedback(id int, feedbackDatetime *time.Time, feedbackObservation string) (*models.Quote, error) {
	if _, err := s.quoteRepo.GetByID(id); err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrQuoteNotFound
		}
		return nil, err
	}
	if err := s.quoteRepo.UpdateFeedback(id, feedbackDatetime, feedbackObservation); err != nil {
		return nil, err
	}
	updated, err := s.quoteRepo.GetByID(id)
	if err == nil {
		s.auditService.LogSimple(nil, "quote_feedback_updated", "quote", fmt.Sprintf("id=%d", id), "")
	}
	return updated, err
}

func (s *QuoteService) Delete(id int) error {
	err := s.quoteRepo.Delete(id)
	if err == nil {
		s.auditService.LogSimple(nil, "quote_deleted", "quote", fmt.Sprintf("id=%d", id), "")
	}
	return err
}
