package services

import (
	"database/sql"

	"donapresentes/models"
	"donapresentes/repositories"
)

type SyncService struct {
	provider           ExternalAPIProvider
	produtoRepository  *repositories.ProdutoRepository
	fornecedorRepository *repositories.FornecedorRepository
}

func NewSyncService(
	provider ExternalAPIProvider,
	produtoRepo *repositories.ProdutoRepository,
	fornecedorRepo *repositories.FornecedorRepository,
) *SyncService {
	return &SyncService{
		provider:            provider,
		produtoRepository:   produtoRepo,
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
	fornecedorXBZ := s.provider.MapearParaFornecedor(nil)
	fornecedorID, err := s.garantirFornecedor(fornecedorXBZ)
	if err != nil {
		result.Erros++
		return result, err
	}

	// Sincronizar cada produto
	for _, itemExterno := range itemsExternos {
		produtoLocal := s.provider.MapearParaProdutoLocal(itemExterno)
		if produtoLocal == nil {
			result.Erros++
			continue
		}

		// Associar ao fornecedor
		produtoLocal.CodigoFornecedor = fornecedorID

		// Verificar se já existe
		existente, err := s.produtoRepository.GetByCodigoInterno(produtoLocal.CodigoInterno)
		if err == nil && existente != nil {
			// Atualizar
			produtoLocal.ID = existente.ID
			produtoLocal.CriadoEm = existente.CriadoEm
			_, err := s.produtoRepository.Update(produtoLocal.ID, produtoLocal)
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
			_, err := s.produtoRepository.Create(produtoLocal)
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
func (s *SyncService) garantirFornecedor(fornecedor *models.Fornecedor) (int, error) {
	// Buscar por nome
	existentes, err := s.fornecedorRepository.GetAll()
	if err != nil {
		return 0, err
	}

	for _, f := range existentes {
		if f.FantasyName == fornecedor.FantasyName {
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
		if f.FantasyName == fornecedor.FantasyName {
			return f.ID, nil
		}
	}

	return 0, nil
}
