package services

import (
	"database/sql"
	"time"

	"donapresentes/models"
	"donapresentes/repositories"
)

type SyncService struct {
	provider           ExternalAPIProvider
	productRepository  *repositories.ProductRepository
	supplierRepository *repositories.SupplierRepository
}

func NewSyncService(
	provider ExternalAPIProvider,
	productRepo *repositories.ProductRepository,
	supplierRepo *repositories.SupplierRepository,
) *SyncService {
	return &SyncService{
		provider:           provider,
		productRepository:  productRepo,
		supplierRepository: supplierRepo,
	}
}

// Synchronize executes the full synchronization flow.
func (s *SyncService) Synchronize() (*SyncResult, error) {
	result := &SyncResult{}

	// Obter produtos da API externa
	externalItems, err := s.provider.GetProdutos()
	if err != nil {
		return result, err
	}

	result.Total = len(externalItems)

	// Ensure the provider supplier (XBZ) exists.
	supplier := s.provider.MapToSupplier(nil)
	supplierID, err := s.ensureSupplier(supplier)
	if err != nil {
		result.Erros++
		return result, err
	}

	// Synchronize each external product.
	for _, externalItem := range externalItems {
		localProduct := s.provider.MapToLocalProduct(externalItem)
		if localProduct == nil {
			result.Erros++
			continue
		}

		// Attach provider supplier.
		localProduct.SupplierID = supplierID

		// Check if it already exists.
		existing, err := s.productRepository.GetByInternalCode(localProduct.InternalCode)
		if err == nil && existing != nil {
			// Update only sync-managed fields.
			syncTime := time.Now()
			if err := s.productRepository.UpdateFromSync(
				existing.ID,
				localProduct.Stock,
				localProduct.Photos,
				localProduct.SellingPrice,
				syncTime,
			); err != nil {
				result.Erros++
				continue
			}
			result.Atualizados++
		} else if err != nil && err != sql.ErrNoRows {
			// Database error other than "not found".
			result.Erros++
			continue
		} else {
			// Create new product (MapToLocalProduct already sets sync fields).
			_, err := s.productRepository.Create(localProduct)
			if err != nil {
				result.Erros++
				continue
			}
			result.Criados++
		}
	}

	return result, nil
}

// Sincronizar is kept for backward compatibility.
func (s *SyncService) Sincronizar() (*SyncResult, error) {
	return s.Synchronize()
}

// ensureSupplier checks whether supplier exists and creates it when needed.
func (s *SyncService) ensureSupplier(supplier *models.Supplier) (int, error) {
	existingSuppliers, err := s.supplierRepository.GetAll()
	if err != nil {
		return 0, err
	}

	for _, s := range existingSuppliers {
		if s.Name == supplier.Name {
			return s.ID, nil
		}
	}

	if err := s.supplierRepository.Create(supplier); err != nil {
		return 0, err
	}

	createdSuppliers, err := s.supplierRepository.GetAll()
	if err != nil {
		return 0, err
	}
	for _, s := range createdSuppliers {
		if s.Name == supplier.Name {
			return s.ID, nil
		}
	}

	return 0, nil
}

// garantirFornecedor is kept for backward compatibility.
func (s *SyncService) garantirFornecedor(fornecedor *models.Supplier) (int, error) {
	return s.ensureSupplier(fornecedor)
}
