package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"donapresentes/controllers/config"
	apperrors "donapresentes/errors"
	"donapresentes/middleware"
	"donapresentes/models"
	"donapresentes/repositories"
	"donapresentes/services"

	"github.com/gorilla/mux"
)

// saleService provides CRUD operations for sales (vendas).
var saleService *services.SaleService

// pdfService generates PDF documents for sales (pedidos).
var pdfService *services.PDFService

// InitSaleService initializes the sale and PDF services with the required repositories.
func InitSaleService() {
	saleRepo := repositories.NewSaleRepository(config.DB)
	userRepo := repositories.NewUserRepository(config.DB)
	productRepo := repositories.NewProductRepository(config.DB)
	customerRepo := repositories.NewCustomerRepository(config.DB)
	carrierRepo := repositories.NewCarrierRepository(config.DB)
	saleService = services.NewSaleService(saleRepo, userRepo, productRepo, customerRepo, carrierRepo, auditService)

	var err error
	pdfService, err = services.NewPDFService()
	if err != nil {
		log.Printf("[WARN] PDF service não inicializado (chromedp): %v - endpoint /sales/{id}/pdf não disponível", err)
	}
}

// GetSales handles GET /sales — returns sales for the current seller (or all if the user has the ver_todos permission).
func GetSales(w http.ResponseWriter, r *http.Request) {
	userID, _, _ := middleware.GetUserFromRequest(r)
	var vs []models.Sale
	var err error
	if middleware.HasPermission(r, "vendas:ver_todos") {
		vs, err = saleService.GetAll()
	} else {
		vs, err = saleService.GetBySellerID(userID)
	}
	if err != nil {
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vs)
}

// GetSale handles GET /sales/{id} — returns a single sale by ID (owner or ver_todos permission required).
func GetSale(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidID, http.StatusBadRequest)
		return
	}
	v, err := saleService.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.ErrorHandler(w, apperrors.ErrSaleNotFound, http.StatusNotFound)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}
	userID, _, _ := middleware.GetUserFromRequest(r)
	if v.SellerID != userID && !middleware.HasPermission(r, "vendas:ver_todos") {
		workflowForbidden(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// CreateSale handles POST /sales — creates a new sale.
func CreateSale(w http.ResponseWriter, r *http.Request) {
	var input models.SaleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidJSON, http.StatusBadRequest)
		return
	}
	userID, _, _ := middleware.GetUserFromRequest(r)
	if !middleware.HasPermission(r, "vendas:ver_todos") {
		input.SellerID = userID
	}

	v, err := saleService.Create(&input)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.ErrorHandler(w, appErr, appErr.Code)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(v)
}

// UpdateSale handles PUT /sales/{id} — updates an existing sale (owner or ver_todos permission required).
func UpdateSale(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidID, http.StatusBadRequest)
		return
	}
	var input models.SaleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidJSON, http.StatusBadRequest)
		return
	}
	userID, _, _ := middleware.GetUserFromRequest(r)
	ownerID, ownerErr := salesWorkflowService.SaleOwner(id)
	if ownerErr != nil || (ownerID != userID && !middleware.HasPermission(r, "vendas:ver_todos")) {
		workflowForbidden(w)
		return
	}
	if !middleware.HasPermission(r, "vendas:ver_todos") {
		input.SellerID = userID
	}

	v, err := saleService.Update(id, &input)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.ErrorHandler(w, apperrors.ErrSaleNotFound, http.StatusNotFound)
			return
		}
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.ErrorHandler(w, appErr, appErr.Code)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// GetSalePDF handles GET /sales/{id}/pdf — generates and returns a PDF document for a sale (pedido).
func GetSalePDF(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidID, http.StatusBadRequest)
		return
	}
	sale, err := saleService.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.ErrorHandler(w, apperrors.ErrSaleNotFound, http.StatusNotFound)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}
	userID, _, _ := middleware.GetUserFromRequest(r)
	if sale.SellerID != userID && !middleware.HasPermission(r, "vendas:ver_todos") {
		workflowForbidden(w)
		return
	}
	if pdfService == nil {
		middleware.ErrorHandler(w, &apperrors.AppError{
			Code:    http.StatusServiceUnavailable,
			Message: "Serviço de PDF não disponível",
		}, http.StatusServiceUnavailable)
		return
	}
	pdfBytes, err := pdfService.GenerateOrderPDF(sale)
	if err != nil {
		middleware.ErrorHandler(w, &apperrors.AppError{
			Code:       http.StatusInternalServerError,
			Message:    "Erro ao gerar PDF do pedido",
			LogMessage: err.Error(),
		}, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"pedido_%04d.pdf\"", sale.ID))
	w.Write(pdfBytes)
}

// DeleteSale handles DELETE /sales/{id} — deletes a sale by ID (owner or ver_todos permission required).
func DeleteSale(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidID, http.StatusBadRequest)
		return
	}
	userID, _, _ := middleware.GetUserFromRequest(r)
	ownerID, ownerErr := salesWorkflowService.SaleOwner(id)
	if ownerErr != nil || (ownerID != userID && !middleware.HasPermission(r, "vendas:ver_todos")) {
		workflowForbidden(w)
		return
	}
	if err := saleService.Delete(id); err != nil {
		if err == sql.ErrNoRows {
			middleware.ErrorHandler(w, apperrors.ErrSaleNotFound, http.StatusNotFound)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UpdateSaleLayout handles PUT /sales/{id}/layout — updates the layout URLs for a sale.
func UpdateSaleLayout(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidID, http.StatusBadRequest)
		return
	}
	if !canSeeSale(r, id) {
		workflowForbidden(w)
		return
	}
	var body struct {
		LayoutURLs []string `json:"layout_urls"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.ErrorHandler(w, apperrors.NewValidationError("JSON inválido"), http.StatusBadRequest)
		return
	}
	if body.LayoutURLs == nil {
		body.LayoutURLs = []string{}
	}
	if err := saleService.UpdateLayoutURLs(id, body.LayoutURLs); err != nil {
		if err == sql.ErrNoRows {
			middleware.ErrorHandler(w, apperrors.ErrSaleNotFound, http.StatusNotFound)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}
	sale, err := saleService.GetByID(id)
	if err != nil {
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sale)
}
