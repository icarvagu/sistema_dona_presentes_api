package services

import (
	"donapresentes/models"
)

// ExternalAPIProvider define a interface para integração com APIs externas
type ExternalAPIProvider interface {
	// GetProdutos busca a lista de produtos da API externa
	GetProdutos() ([]interface{}, error)

	// MapToLocalProduct maps the external object to the local Product model
	MapToLocalProduct(interface{}) *models.Product

	// MapToSupplier maps the external provider info to the local Supplier model if needed
	MapToSupplier(interface{}) *models.Supplier
}

// SyncConfig contém configurações para sincronização
type SyncConfig struct {
	Provider ExternalAPIProvider
	Database interface{} // Será injetado pelo controller
}

// SyncResult resultado da sincronização
type SyncResult struct {
	Total       int `json:"total"`
	Criados     int `json:"criados"`
	Atualizados int `json:"atualizados"`
	Erros       int `json:"erros"`
	Deletados   int `json:"deletados"`
}
