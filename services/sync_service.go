package services

import (
	"database/sql"

	"donapresentes/models"
	"donapresentes/repositories"
)

type SyncService struct {
	provider           ExternalAPIProvider
	productRepository  *repositories.ProductRepository
	fornecedorRepository *repositories.SupplierRepository
}

func NewSyncService(
	provider ExternalAPIProvider,
	productRepo *repositories.ProductRepository,
	fornecedorRepo *repositories.SupplierRepository,
) *SyncService {
	return &SyncService{
		provider:            provider,
		productRepository:   productRepo,
		fornecedorRepository: fornecedorRepo,
	}
}

// Sincronizar executa a sincronização completa
func (s *SyncService) Sincronizar() (*SyncResult, error) {
	result := &SyncResult{}

	// Obter produtos da API externa
	itemsExternos, err := s.provider.GetProdutos()
	if err != nil {
		return result, err
	}

	result.Total = len(itemsExternos)

	// Garantir que o fornecedor XBZ existe
	fornecedorXBZ := s.provider.MapToSupplier(nil)
	fornecedorID, err := s.garantirFornecedor(fornecedorXBZ)
	if err != nil {
		result.Erros++
		return result, err
	}

	// Sincronizar cada produto
	for _, itemExterno := range itemsExternos {
		produtoLocal := s.provider.MapToLocalProduct(itemExterno)
		if produtoLocal == nil {
			result.Erros++
			continue
		}

		// Associar ao fornecedor
		produtoLocal.SupplierID = fornecedorID

		// Verificar se já existe
		existente, err := s.productRepository.GetByCodigoInterno(produtoLocal.InternalCode)
		if err == nil && existente != nil {
			// Atualizar
			produtoLocal.ID = existente.ID
			produtoLocal.CreatedAt = existente.CreatedAt
			_, err := s.productRepository.Update(produtoLocal.ID, produtoLocal)
			if err != nil {
				result.Erros++
				continue
			}
			result.Atualizados++
		} else if err != sql.ErrNoRows {
			// Erro de BD
			result.Erros++
			continue
		} else {
			// Criar novo
			_, err := s.productRepository.Create(produtoLocal)
			if err != nil {
				result.Erros++
				continue
			}
			result.Criados++
		}
	}

	return result, nil
}

// garantirFornecedor verifica se fornecedor existe, se não cria
func (s *SyncService) garantirFornecedor(fornecedor *models.Supplier) (int, error) {
	// Buscar por nome
	existentes, err := s.fornecedorRepository.GetAll()
	if err != nil {
		return 0, err
	}

	for _, f := range existentes {
		if f.Name == fornecedor.Name {
			return f.ID, nil
		}
	}

	// Criar novo
	// FornecedorRepository.Create returns only error; call Create and then fetch by nome
	if err := s.fornecedorRepository.Create(fornecedor); err != nil {
		return 0, err
	}

	// Buscar novamente para obter o ID
	existentes2, err := s.fornecedorRepository.GetAll()
	if err != nil {
		return 0, err
	}
	for _, f := range existentes2 {
		if f.Name == fornecedor.Name {
			return f.ID, nil
		}
	}

	return 0, nil
}
