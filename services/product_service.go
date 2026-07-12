package services

import (
	"database/sql"
	"fmt"

	apperrors "donapresentes/errors"
	"donapresentes/models"
	"donapresentes/repositories"
)

// ProductService manages product catalog operations including CRUD, approval workflow,
// kit composition validation, financial reports, and sync from external suppliers.
type ProductService struct {
	productRepo  *repositories.ProductRepository
	supplierRepo *repositories.SupplierRepository
	auditService *AuditService
}

// NewProductService creates a ProductService with the required repositories and audit trail.
func NewProductService(productRepo *repositories.ProductRepository, supplierRepo *repositories.SupplierRepository, auditService *AuditService) *ProductService {
	return &ProductService{
		productRepo:  productRepo,
		supplierRepo: supplierRepo,
		auditService: auditService,
	}
}

// GetAll returns all products in the catalog.
func (s *ProductService) GetAll() ([]models.Product, error) {
	return s.productRepo.GetAll()
}

// GetAllPaginated returns a page of products with the total count.
func (s *ProductService) GetAllPaginated(page, limit int) ([]models.Product, int, error) {
	return s.productRepo.GetAllPaginated(page, limit)
}

// GetByID returns a single product by its ID.
func (s *ProductService) GetByID(id int) (*models.Product, error) {
	return s.productRepo.GetByID(id)
}

// SearchByFilter searches products by a text filter.
func (s *ProductService) SearchByFilter(filter string) ([]models.Product, error) {
	return s.productRepo.SearchByFilter(filter)
}

// GetGroups returns all distinct product groups.
func (s *ProductService) GetGroups() ([]string, error) {
	return s.productRepo.GetGroups()
}

// GetPendingApproval returns all products awaiting tax analysis approval.
func (s *ProductService) GetPendingApproval() ([]models.Product, error) {
	return s.productRepo.GetPendingApproval()
}

// ApproveProduct approves a single product for use, identified by its origin source.
func (s *ProductService) ApproveProduct(id int, origin string) error {
	err := s.productRepo.ApproveProduct(id, origin)
	if err == nil {
		s.auditService.LogSimple(nil, "product_approved", "product", fmt.Sprintf("id=%d origin=%s", id, origin), "")
	}
	return err
}

// BulkApproveAll approves all pending products from a given origin source.
func (s *ProductService) BulkApproveAll(origin string) (int64, error) {
	count, err := s.productRepo.BulkApproveAll(origin)
	if err == nil {
		s.auditService.LogSimple(nil, "product_bulk_approved", "product", fmt.Sprintf("count=%d origin=%s", count, origin), "")
	}
	return count, err
}

// UpdateLastCost updates only the last cost fields on a product.
func (s *ProductService) UpdateLastCost(id int, p *models.Product) error {
	return s.productRepo.UpdateLastCost(id, p)
}

// GetByGroup returns all products belonging to a specific group.
func (s *ProductService) GetByGroup(group string) ([]models.Product, error) {
	return s.productRepo.GetByGroup(group)
}

// validateRequiredFields ensures product_name and internal_code are present.
func (s *ProductService) validateRequiredFields(p *models.Product) error {
	if p.ProductName == "" {
		return apperrors.NewMissingFieldError("product_name")
	}
	if p.InternalCode == "" {
		return apperrors.NewMissingFieldError("internal_code")
	}
	return nil
}

// validateSupplier ensures the referenced supplier exists.
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

// validateItems enforces kit composition rules:
// kit_type must be "none", "internal_composition", or "supplier_ready";
// internal_composition kits require at least one item; supplier_ready kits
// cannot have composition items; "none" products must not have items.
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

// Create registers a new product and its kit items.
// All new products are marked as pending tax approval.
func (s *ProductService) Create(p *models.Product) (*models.Product, error) {
	// Todo produto precisa passar pela análise tributária antes de ficar disponível.
	p.PendingApproval = true

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

	finalProduct, err := s.productRepo.GetByID(created.ID)
	if err == nil {
		s.auditService.LogSimple(nil, "product_created", "product", fmt.Sprintf("id=%d name=%s", finalProduct.ID, finalProduct.ProductName), "")
	}
	return finalProduct, err
}

// Update modifies an existing product and replaces its kit items.
// The existing pending approval status is preserved.
func (s *ProductService) Update(id int, p *models.Product) (*models.Product, error) {

	existing, err := s.productRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrProductNotFound
		}
		return nil, err
	}
	// A edição cadastral não substitui a autorização feita pela fila de aprovação.
	p.PendingApproval = existing.PendingApproval

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

	result, err := s.productRepo.GetByID(id)
	if err == nil {
		s.auditService.LogSimple(nil, "product_updated", "product", fmt.Sprintf("id=%d name=%s", id, result.ProductName), "")
	}
	return result, err
}

// Delete removes a product by its ID.
func (s *ProductService) Delete(id int) error {
	err := s.productRepo.Delete(id)
	if err == nil {
		s.auditService.LogSimple(nil, "product_deleted", "product", fmt.Sprintf("id=%d", id), "")
	}
	return err
}

// GetNewlyImported returns products that were recently imported from external sources.
func (s *ProductService) GetNewlyImported() ([]models.Product, error) {
	return s.productRepo.GetNewlyImported()
}

// GetFinancialReport returns the financial report for all products.
func (s *ProductService) GetFinancialReport() ([]models.FinancialReportItem, error) {
	return s.productRepo.GetFinancialReport()
}
