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

	return result, nil
}

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
