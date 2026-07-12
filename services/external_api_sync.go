package services

import (
	"donapresentes/models"
)

// ExternalAPIProvider defines the interface for external product catalog sync.
// Implementations provide products and map them to local domain models.
type ExternalAPIProvider interface {

	GetProducts() ([]interface{}, error)

	MapToLocalProduct(interface{}) *models.Product

	MapToSupplier(interface{}) *models.Supplier
}

// SyncConfig holds the configuration for syncing with an external API provider.
type SyncConfig struct {
	Provider ExternalAPIProvider
	Database interface{}
}

// SyncResult reports the outcome of a sync operation with totals for
// created, updated, deleted records, and error count.
type SyncResult struct {
	Total       int `json:"total"`
	Criados     int `json:"criados"`
	Atualizados int `json:"atualizados"`
	Erros       int `json:"erros"`
	Deletados   int `json:"deletados"`
}
