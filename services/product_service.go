package services

import (
	"database/sql"
	apperrors "donapresentes/errors"
	"donapresentes/models"
	"donapresentes/repositories"
)

type ProductService struct {
	productRepo  *repositories.ProductRepository
	supplierRepo *repositories.SupplierRepository
}

func NewProductService(productRepo *repositories.ProductRepository, supplierRepo *repositories.SupplierRepository) *ProductService {
	return &ProductService{
		productRepo:  productRepo,
		supplierRepo: supplierRepo,
	}
}

func (s *ProductService) GetAll() ([]models.Product, error) {
	return s.productRepo.GetAll()
}

func (s *ProductService) GetAllPaginated(page, limit int) ([]models.Product, int, error) {
	return s.productRepo.GetAllPaginated(page, limit)
}

func (s *ProductService) GetByID(id int) (*models.Product, error) {
	return s.productRepo.GetByID(id)
}

func (s *ProductService) SearchByFilter(filter string) ([]models.Product, error) {
	return s.productRepo.SearchByFilter(filter)
}

func (s *ProductService) GetGroups() ([]string, error) {
	return s.productRepo.GetGroups()
}

func (s *ProductService) GetByGroup(group string) ([]models.Product, error) {
	return s.productRepo.GetByGroup(group)
}

func (s *ProductService) validateRequiredFields(p *models.Product) error {
	if p.ProductName == "" {
		return apperrors.NewMissingFieldError("product_name")
	}
	if p.InternalCode == "" {
		return apperrors.NewMissingFieldError("internal_code")
	}
	return nil
}

func (s *ProductService) validateSupplier(p *models.Product) error {
	_, err := s.supplierRepo.GetByID(p.SupplierID)
	if err != nil {
		if err == sql.ErrNoRows {
			return apperrors.ErrSupplierNotFound
		}
		return err
	}
	return nil
}

func (s *ProductService) validateItems(p *models.Product) error {
	if p.KitType == "" {
		p.KitType = "none"
	}

	if p.KitType != "none" && p.KitType != "internal_composition" && p.KitType != "supplier_ready" {
		return apperrors.NewValidationError("kit_type inválido. Use: none, internal_composition ou supplier_ready")
	}

	if p.KitType == "internal_composition" && len(p.Items) == 0 {
		return apperrors.NewValidationError("kits com composição interna precisam ter ao menos 1 item")
	}

	if p.KitType == "supplier_ready" && len(p.Items) > 0 {
		return apperrors.NewValidationError("kit pronto de fornecedor não pode ter composição interna")
	}

	if p.KitType == "none" && len(p.Items) > 0 {
		return apperrors.NewValidationError("produto com composição deve ser do tipo internal_composition")
	}

	if len(p.Items) == 0 {
		return nil
	}

	for _, item := range p.Items {

		_, err := s.productRepo.GetByID(item.ProductID)
		if err != nil {
			if err == sql.ErrNoRows {
				return apperrors.NewNotFoundError("Produto componente não encontrado")
			}
			return err
		}
		if item.Quantity <= 0 {
			return apperrors.NewValidationError("Quantidade do item deve ser maior que zero")
		}
	}
	return nil
}

func (s *ProductService) Create(p *models.Product) (*models.Product, error) {

	if err := s.validateRequiredFields(p); err != nil {
		return nil, err
	}

	if err := s.validateSupplier(p); err != nil {
		return nil, err
	}

	if err := s.validateItems(p); err != nil {
		return nil, err
	}

	created, err := s.productRepo.Create(p)
	if err != nil {
		return nil, err
	}

	if len(p.Items) > 0 {
		for _, itemInput := range p.Items {
			item := &models.ProductItem{
				ProductID: itemInput.ProductID,
				Quantity:  itemInput.Quantity,
			}
			if err := s.productRepo.CreateProductItem(item, created.ID); err != nil {

				s.productRepo.Delete(created.ID)
				return nil, apperrors.NewDatabaseError(err)
			}
		}
	}

	return s.productRepo.GetByID(created.ID)
}

func (s *ProductService) Update(id int, p *models.Product) (*models.Product, error) {

	_, err := s.productRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrProductNotFound
		}
		return nil, err
	}

	if err := s.validateRequiredFields(p); err != nil {
		return nil, err
	}

	if err := s.validateSupplier(p); err != nil {
		return nil, err
	}

	if err := s.validateItems(p); err != nil {
		return nil, err
	}

	if err := s.productRepo.DeleteProductItems(id); err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}

	_, err = s.productRepo.Update(id, p)
	if err != nil {
		return nil, err
	}

	if len(p.Items) > 0 {
		for _, itemInput := range p.Items {
			item := &models.ProductItem{
				ProductID: itemInput.ProductID,
				Quantity:  itemInput.Quantity,
			}
			if err := s.productRepo.CreateProductItem(item, id); err != nil {
				return nil, apperrors.NewDatabaseError(err)
			}
		}
	}

	return s.productRepo.GetByID(id)
}

func (s *ProductService) Delete(id int) error {
	return s.productRepo.Delete(id)
}

func (s *ProductService) GetNewlyImported() ([]models.Product, error) {
	return s.productRepo.GetNewlyImported()
}

func (s *ProductService) GetFinancialReport() ([]models.FinancialReportItem, error) {
	return s.productRepo.GetFinancialReport()
}
