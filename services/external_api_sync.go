package services

import (
	"donapresentes/models"
)

// ExternalAPIProvider define a interface para integração com APIs externas
type ExternalAPIProvider interface {
	// GetProdutos busca a lista de produtos da API externa
	GetProdutos() ([]interface{}, error)

	// MapearParaProdutoLocal mapeia o objeto externo para o modelo local Produto
	MapearParaProdutoLocal(interface{}) *models.Produto

	// MapearParaFornecedor mapeia fornecedor da API para modelo local se necessário
	MapearParaFornecedor(interface{}) *models.Fornecedor
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
