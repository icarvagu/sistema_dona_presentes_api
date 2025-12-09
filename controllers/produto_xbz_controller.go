package controllers

import (
	"encoding/json"
	"net/http"

	"donapresentes/controllers/config"
	"donapresentes/middleware"
	"donapresentes/repositories"
	"donapresentes/services"
)

var xbzService *services.XBZService

func InitProdutoXBZController() {
	// Valores padrão - considere usar variáveis de ambiente
	xbzService = services.NewXBZService("36168035000181", "X142AA979C")
}

// SincronizarProdutos sincroniza produtos da API XBZ com o banco de dados
func SincronizarProdutos(w http.ResponseWriter, r *http.Request) {
	produtoRepo := repositories.NewProdutoRepository(config.DB)
	fornecedorRepo := repositories.NewFornecedorRepository(config.DB)
	
	syncService := services.NewSyncService(xbzService, produtoRepo, fornecedorRepo)
	result, err := syncService.Sincronizar()
	if err != nil {
		middleware.ErrorHandler(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

