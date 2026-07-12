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

// xbzService provides integration with the XBZ external supplier API for product synchronization.
// syncInProgress is a flag indicating whether a sync operation is currently running.
// syncMutex guards concurrent access to the sync state.
var (
	xbzService     *services.XBZService
	syncInProgress bool
	syncMutex      sync.Mutex
)

// InitProductXBZController initializes the XBZ product sync controller from environment variables.
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

// SyncProductsFromXBZHandler handles POST /products-xbz/sync — starts an asynchronous sync of products from the XBZ supplier API.
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

	go func() {
		defer func() {
			syncMutex.Lock()
			syncInProgress = false
			syncMutex.Unlock()
		}()

		log.Printf("[XBZ Sync] Iniciando sincronização assíncrona")
		productRepo := repositories.NewProductRepository(config.DB)
		supplierRepo := repositories.NewSupplierRepository(config.DB)

		syncService := services.NewSyncService(xbzService, productRepo, supplierRepo, services.NewAuditService(config.DB))
		result, err := syncService.Synchronize()
		if err != nil {
			log.Printf("[XBZ Sync] Erro na sincronização: %v", err)
			return
		}

		log.Printf("[XBZ Sync] Sincronização concluída - Total: %d, Criados: %d, Atualizados: %d, Erros: %d",
			result.Total, result.Criados, result.Atualizados, result.Erros)
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Sincronização iniciada com sucesso",
		"status":  "started",
	})
}
