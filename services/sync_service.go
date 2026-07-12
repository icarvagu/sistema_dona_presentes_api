package services

import (
	"database/sql"
	"fmt"
	"time"

	"donapresentes/models"
	"donapresentes/repositories"
)

// SyncService synchronizes the local product catalog with an external API provider.
// It creates or updates products and ensures the corresponding supplier exists.
type SyncService struct {
	provider           ExternalAPIProvider
	productRepository  *repositories.ProductRepository
	supplierRepository *repositories.SupplierRepository
	auditService       *AuditService
}

// NewSyncService creates a SyncService with the given external API provider,
// product and supplier repositories, and audit trail.
func NewSyncService(
	provider ExternalAPIProvider,
	productRepo *repositories.ProductRepository,
	supplierRepo *repositories.SupplierRepository,
	auditService *AuditService,
) *SyncService {
	return &SyncService{
		provider:           provider,
		productRepository:  productRepo,
		supplierRepository: supplierRepo,
		auditService:       auditService,
	}
}

// Synchronize fetches products from the external provider, ensures the supplier
// exists, and creates or updates local products matched by internal code.
// Returns a SyncResult with counts of created, updated, and errored items.
func (s *SyncService) Synchronize() (*SyncResult, error) {
	result := &SyncResult{}

	externalItems, err := s.provider.GetProducts()
	if err != nil {
		return result, err
	}

	result.Total = len(externalItems)

	supplier := s.provider.MapToSupplier(nil)
	supplierID, err := s.ensureSupplier(supplier)
	if err != nil {
		result.Erros++
		return result, err
	}

	for _, externalItem := range externalItems {
		localProduct := s.provider.MapToLocalProduct(externalItem)
		if localProduct == nil {
			result.Erros++
			continue
		}

		localProduct.SupplierID = supplierID

		existing, err := s.productRepository.GetByInternalCode(localProduct.InternalCode)
		if err == nil && existing != nil {

			syncTime := time.Now()
			if err := s.productRepository.UpdateFromSync(
				existing.ID,
				localProduct.Stock,
				localProduct.SupplierStock,
				localProduct.SupplierCode,
				localProduct.Photos,
				localProduct.CostPrice,
				syncTime,
			); err != nil {
				result.Erros++
				continue
			}
			result.Atualizados++
		} else if err != nil && err != sql.ErrNoRows {

			result.Erros++
			continue
		} else {

			localProduct.PendingApproval = true
			_, err := s.productRepository.Create(localProduct)
			if err != nil {
				result.Erros++
				continue
			}
			result.Criados++
		}
	}

	s.auditService.LogSimple(nil, "sync_completed", "sync", fmt.Sprintf("total=%d created=%d updated=%d errors=%d", result.Total, result.Criados, result.Atualizados, result.Erros), "")
	return result, nil
}

// ensureSupplier finds or creates the supplier mapped from the external provider.
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
