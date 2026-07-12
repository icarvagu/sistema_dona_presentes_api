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

// supplierService provides CRUD operations for suppliers (fornecedores).
var supplierService *services.SupplierService

// InitSupplierService initializes the supplier service with the required repository.
func InitSupplierService() {
	supplierRepo := repositories.NewSupplierRepository(config.DB)
	supplierService = services.NewSupplierService(supplierRepo, auditService)
}

// GetSuppliers handles GET /suppliers — returns the list of all suppliers.
func GetSuppliers(w http.ResponseWriter, r *http.Request) {
	suppliers, err := supplierService.GetAll()
	if err != nil {
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suppliers)
}

// GetSupplier handles GET /suppliers/{id} — returns a single supplier by ID.
func GetSupplier(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidID, http.StatusBadRequest)
		return
	}

	s, err := supplierService.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.ErrorHandler(w, apperrors.ErrSupplierNotFound, http.StatusNotFound)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

// CreateSupplier handles POST /suppliers — creates a new supplier.
func CreateSupplier(w http.ResponseWriter, r *http.Request) {
	var s models.Supplier
	err := json.NewDecoder(r.Body).Decode(&s)
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	err = supplierService.Create(&s)
	if err != nil {
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s)
}

// UpdateSupplier handles PUT /suppliers/{id} — updates an existing supplier.
func UpdateSupplier(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidID, http.StatusBadRequest)
		return
	}

	var s models.Supplier
	err = json.NewDecoder(r.Body).Decode(&s)
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	err = supplierService.Update(id, &s)
	if err != nil {
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	s.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

// DeleteSupplier handles DELETE /suppliers/{id} — deletes a supplier by ID.
func DeleteSupplier(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		middleware.ErrorHandler(w, apperrors.ErrInvalidID, http.StatusBadRequest)
		return
	}

	err = supplierService.Delete(id)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.ErrorHandler(w, apperrors.ErrSupplierNotFound, http.StatusNotFound)
			return
		}
		middleware.ErrorHandler(w, apperrors.NewDatabaseError(err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
