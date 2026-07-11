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
	productService = services.NewProductService(productRepo, supplierRepo, auditService)
}

func GetProductGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := productService.GetGroups()
	if err != nil {
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}
	if groups == nil {
		groups = []string{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

func GetProducts(w http.ResponseWriter, r *http.Request) {

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

	filter := r.URL.Query().Get("filter")
	group := r.URL.Query().Get("group")

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	if group != "" {
		ps, err := productService.GetByGroup(group)
		if err != nil {
			middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
			return
		}
		if ps == nil {
			ps = []models.Product{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ps)
		return
	}

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

	if pageStr != "" || limitStr != "" {
		page := 1
		limit := 10

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

		totalPages := (total + limit - 1) / limit
		if totalPages == 0 {
			totalPages = 1
		}

		response := models.PaginatedProductResponse{
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

	ps, err := productService.GetAll()
	if err != nil {
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ps)
}

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

func UpdateProductLastCost(w http.ResponseWriter, r *http.Request) {
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
	if err := productService.UpdateLastCost(id, &p); err != nil {
		if err == sql.ErrNoRows {
			middleware.ErrorHandler(w, apperrors.ErrProductNotFound, http.StatusNotFound)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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

func GetPendingProducts(w http.ResponseWriter, r *http.Request) {
	ps, err := productService.GetPendingApproval()
	if err != nil {
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}
	if ps == nil {
		ps = []models.Product{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ps)
}

func BulkApproveProducts(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Origin string `json:"origin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Origin == "" {
		middleware.ErrorHandler(w, apperrors.ErrInvalidJSON, http.StatusBadRequest)
		return
	}
	count, err := productService.BulkApproveAll(body.Origin)
	if err != nil {
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{"updated": count})
}

func ApproveProduct(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidID, http.StatusBadRequest)
		return
	}

	var body struct {
		Origin string `json:"origin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	if err := productService.ApproveProduct(id, body.Origin); err != nil {
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
