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

// GetAll returns all products
func (s *ProductService) GetAll() ([]models.Product, error) {
	return s.productRepo.GetAll()
}

// GetAllPaginated returns paginated products
func (s *ProductService) GetAllPaginated(page, limit int) ([]models.Product, int, error) {
	return s.productRepo.GetAllPaginated(page, limit)
}

// GetByID returns a product by ID
func (s *ProductService) GetByID(id int) (*models.Product, error) {
	return s.productRepo.GetByID(id)
}

// SearchByFilter searches products by filter
func (s *ProductService) SearchByFilter(filter string) ([]models.Product, error) {
	return s.productRepo.SearchByFilter(filter)
}

// validateRequiredFields valida campos obrigatórios do produto
func (s *ProductService) validateRequiredFields(p *models.Product) error {
	if p.ProductName == "" {
		return apperrors.NewMissingFieldError("product_name")
	}
	if p.InternalCode == "" {
		return apperrors.NewMissingFieldError("internal_code")
	}
	return nil
}

// validateSupplier valida se o fornecedor existe
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

// validateItems valida os itens de composição do produto
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
		// Verificar se produto componente existe
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

// Create creates a new product with validation
func (s *ProductService) Create(p *models.Product) (*models.Product, error) {
	// Validate required fields
	if err := s.validateRequiredFields(p); err != nil {
		return nil, err
	}

	// Validate supplier exists
	if err := s.validateSupplier(p); err != nil {
		return nil, err
	}

	// Validar items se houver (todos os produtos podem ter items)
	if err := s.validateItems(p); err != nil {
		return nil, err
	}

	// Criar produto (todos os produtos são kits, podem ter composição ou não)
	created, err := s.productRepo.Create(p)
	if err != nil {
		return nil, err
	}

	// Criar items se for composição (qualquer quantidade > 0)
	if len(p.Items) > 0 {
		for _, itemInput := range p.Items {
			item := &models.ProductItem{
				ProductID: itemInput.ProductID,
				Quantity:  itemInput.Quantity,
			}
			if err := s.productRepo.CreateProductItem(item, created.ID); err != nil {
				// Rollback: deletar produto se criação de item falhar
				s.productRepo.Delete(created.ID)
				return nil, apperrors.NewDatabaseError(err)
			}
		}
	}

	// Buscar produto completo com items
	return s.productRepo.GetByID(created.ID)
}

// Update updates an existing product with validation
func (s *ProductService) Update(id int, p *models.Product) (*models.Product, error) {
	// Validate product exists
	_, err := s.productRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrProductNotFound
		}
		return nil, err
	}

	// Validate required fields
	if err := s.validateRequiredFields(p); err != nil {
		return nil, err
	}

	// Validate supplier exists
	if err := s.validateSupplier(p); err != nil {
		return nil, err
	}

	// Validar items se houver (todos os produtos são kits, podem ter composição ou não)
	if err := s.validateItems(p); err != nil {
		return nil, err
	}

	// Deletar items antigos
	if err := s.productRepo.DeleteProductItems(id); err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}

	// Atualizar produto
	_, err = s.productRepo.Update(id, p)
	if err != nil {
		return nil, err
	}

	// Criar novos items se for composição (qualquer quantidade > 0)
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

	// Buscar produto completo com items
	return s.productRepo.GetByID(id)
}

// Delete deletes a product
func (s *ProductService) Delete(id int) error {
	return s.productRepo.Delete(id)
}

// GetNewlyImported returns products imported in the last 7 days from XBZ
func (s *ProductService) GetNewlyImported() ([]models.Product, error) {
	return s.productRepo.GetNewlyImported()
}

func (s *ProductService) GetFinancialReport() ([]models.FinancialReportItem, error) {
	return s.productRepo.GetFinancialReport()
}
