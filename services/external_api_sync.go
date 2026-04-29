package services

import (
	"donapresentes/models"
)

type ExternalAPIProvider interface {

	GetProducts() ([]interface{}, error)

	MapToLocalProduct(interface{}) *models.Product

	MapToSupplier(interface{}) *models.Supplier
}

type SyncConfig struct {
	Provider ExternalAPIProvider
	Database interface{}
}

type SyncResult struct {
	Total       int `json:"total"`
	Criados     int `json:"criados"`
	Atualizados int `json:"atualizados"`
	Erros       int `json:"erros"`
	Deletados   int `json:"deletados"`
}
