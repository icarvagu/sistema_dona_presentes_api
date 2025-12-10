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

func InitProductXBZController() {
	// Valores padrão - considere usar variáveis de ambiente
	xbzService = services.NewXBZService("36168035000181", "X142AA979C")
}

// SyncProductsFromXBZ synchronizes products from the XBZ API into the database
func SyncProductsFromXBZ(w http.ResponseWriter, r *http.Request) {
	productRepo := repositories.NewProductRepository(config.DB)
	supplierRepo := repositories.NewSupplierRepository(config.DB)

	syncService := services.NewSyncService(xbzService, productRepo, supplierRepo)
	result, err := syncService.Sincronizar()
	if err != nil {
		middleware.ErrorHandler(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

