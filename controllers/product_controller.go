package controllers

import (
	"database/sql"
	"encoding/json"
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

var productService *services.ProductService

func InitProductService() {
	productRepo := repositories.NewProductRepository(config.DB)
	supplierRepo := repositories.NewSupplierRepository(config.DB)
	productService = services.NewProductService(productRepo, supplierRepo)
}

// PaginatedResponse estrutura para resposta paginada
type PaginatedResponse struct {
	Data       []models.Product `json:"data"`
	Total      int              `json:"total"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	TotalPages int              `json:"total_pages"`
}

// List produtos
func GetProducts(w http.ResponseWriter, r *http.Request) {
	// Verifica se há parâmetro only_new na query string
	onlyNew := r.URL.Query().Get("only_new")
	if onlyNew == "true" || onlyNew == "1" {
		ps, err := productService.GetNewlyImported()
		if err != nil {
			middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ps)
		return
	}

	// Verifica se há parâmetro de filtro na query string
	filter := r.URL.Query().Get("filter")

	// Verifica parâmetros de paginação
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	// Se houver filtro, não usa paginação (retorna todos os resultados do filtro)
	if filter != "" {
		ps, err := productService.SearchByFilter(filter)
		if err != nil {
			middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ps)
		return
	}

	// Se houver parâmetros de paginação, usa paginação
	if pageStr != "" || limitStr != "" {
		page := 1
		limit := 10 // default

		if pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}

		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		ps, total, err := productService.GetAllPaginated(page, limit)
		if err != nil {
			middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
			return
		}

		totalPages := (total + limit - 1) / limit // arredonda para cima
		if totalPages == 0 {
			totalPages = 1
		}

		response := PaginatedResponse{
			Data:       ps,
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Sem filtro e sem paginação - retorna todos (comportamento original)
	ps, err := productService.GetAll()
	if err != nil {
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ps)
}

// Get produto by id
func GetProduct(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidID, http.StatusBadRequest)
		return
	}
	p, err := productService.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.ErrorHandler(w, apperrors.ErrProductNotFound, http.StatusNotFound)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// Create produto with validation
func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	_, _, role := middleware.GetUserFromRequest(r)
	if role != "admin" && p.CostPrice > 0 {
		middleware.ErrorHandler(w, apperrors.NewValidationError("apenas admin pode definir preço de custo"), http.StatusForbidden)
		return
	}

	created, err := productService.Create(&p)
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
	json.NewEncoder(w).Encode(created)
}

// Update produto
func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidID, http.StatusBadRequest)
		return
	}
	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	_, _, role := middleware.GetUserFromRequest(r)
	if role != "admin" {
		existing, getErr := productService.GetByID(id)
		if getErr != nil {
			if getErr == sql.ErrNoRows {
				middleware.ErrorHandler(w, apperrors.ErrProductNotFound, http.StatusNotFound)
				return
			}
			middleware.ErrorHandler(w, apperrors.NewDatabaseError(getErr), http.StatusInternalServerError)
			return
		}
		p.CostPrice = existing.CostPrice
	}

	updated, err := productService.Update(id, &p)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.ErrorHandler(w, apperrors.ErrProductNotFound, http.StatusNotFound)
			return
		}
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.ErrorHandler(w, appErr, appErr.Code)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}
	p.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// Delete produto
func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidID, http.StatusBadRequest)
		return
	}
	if err := productService.Delete(id); err != nil {
		if err == sql.ErrNoRows {
			middleware.ErrorHandler(w, apperrors.ErrProductNotFound, http.StatusNotFound)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func GetPriceFormation(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	productID, err := strconv.Atoi(params["id"])
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidID, http.StatusBadRequest)
		return
	}

	pf, err := productService.GetPriceFormation(productID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.ErrorHandler(w, appErr, appErr.Code)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pf)
}

func UpsertPriceFormation(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	productID, err := strconv.Atoi(params["id"])
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidID, http.StatusBadRequest)
		return
	}

	var input models.ProductPriceFormation
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	pf, err := productService.UpsertPriceFormation(productID, &input)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.ErrorHandler(w, appErr, appErr.Code)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pf)
}

func GetProductsFinancialReport(w http.ResponseWriter, r *http.Request) {
	report, err := productService.GetFinancialReport()
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			middleware.ErrorHandler(w, appErr, appErr.Code)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}
