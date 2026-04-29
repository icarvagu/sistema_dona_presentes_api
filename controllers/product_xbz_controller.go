package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"

	"donapresentes/controllers/config"
	"donapresentes/repositories"
	"donapresentes/services"
)

var (
	xbzService     *services.XBZService
	syncInProgress bool
	syncMutex      sync.Mutex
)

func InitProductXBZController() {
	cnpj := os.Getenv("XBZ_CNPJ")
	token := os.Getenv("XBZ_TOKEN")
	if cnpj == "" || token == "" {
		log.Printf("[XBZ Sync] XBZ_CNPJ/XBZ_TOKEN não configurados; endpoint de sync ficará indisponível")
		xbzService = nil
		return
	}
	xbzService = services.NewXBZService(cnpj, token)
}

// SyncProductsFromXBZHandler starts async product sync from XBZ API.
func SyncProductsFromXBZHandler(w http.ResponseWriter, r *http.Request) {
	if xbzService == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Integração XBZ não configurada",
			"status":  "disabled",
		})
		return
	}

	// Verificar se já existe uma sincronização em andamento
	syncMutex.Lock()
	if syncInProgress {
		syncMutex.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Sincronização já está em andamento",
			"status":  "in_progress",
		})
		return
	}
	syncInProgress = true
	syncMutex.Unlock()

	// Executar sincronização em goroutine
	go func() {
		defer func() {
			syncMutex.Lock()
			syncInProgress = false
			syncMutex.Unlock()
		}()

		log.Printf("[XBZ Sync] Iniciando sincronização assíncrona")
		productRepo := repositories.NewProductRepository(config.DB)
		supplierRepo := repositories.NewSupplierRepository(config.DB)

		syncService := services.NewSyncService(xbzService, productRepo, supplierRepo)
		result, err := syncService.Synchronize()
		if err != nil {
			log.Printf("[XBZ Sync] Erro na sincronização: %v", err)
			return
		}

		log.Printf("[XBZ Sync] Sincronização concluída - Total: %d, Criados: %d, Atualizados: %d, Erros: %d",
			result.Total, result.Criados, result.Atualizados, result.Erros)
	}()

	// Retornar resposta imediata
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Sincronização iniciada com sucesso",
		"status":  "started",
	})
}
