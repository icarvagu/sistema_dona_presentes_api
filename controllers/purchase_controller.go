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

// purchaseService provides operations for purchase order management, payments, attachments, issues, and notifications.
var purchaseService *services.PurchaseService

// InitPurchaseService initializes the purchase service with the required repository.
func InitPurchaseService() {
	purchaseService = services.NewPurchaseService(repositories.NewPurchaseRepository(config.DB), auditService)
}

func writePurchaseError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	if err == sql.ErrNoRows {
		status = http.StatusNotFound
	}
	middleware.ErrorHandler(w, err, status)
}

func requirePurchasePermission(w http.ResponseWriter, r *http.Request, permissions ...string) bool {
	for _, permission := range permissions {
		if middleware.HasPermission(r, permission) {
			return true
		}
	}
	middleware.ErrorHandler(w, apperrors.NewForbiddenError("Acesso negado para esta etapa do módulo de Compras"), http.StatusForbidden)
	return false
}

func purchaseID(r *http.Request) (int, error) { return strconv.Atoi(mux.Vars(r)["id"]) }

// GetPurchases handles GET /purchases — lists purchases with an optional sample filter.
func GetPurchases(w http.ResponseWriter, r *http.Request) {
	if !requirePurchasePermission(w, r, "compras", "arte_final", "financeiro", "diretoria_financeira", "producao", "vendas") {
		return
	}
	var isSample *bool
	if value := r.URL.Query().Get("sample"); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			writePurchaseError(w, err)
			return
		}
		isSample = &parsed
	}
	items, err := purchaseService.List(isSample)
	if err != nil {
		writePurchaseError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(items)
}

// GetPurchase handles GET /purchases/{id} — returns a single purchase by ID.
func GetPurchase(w http.ResponseWriter, r *http.Request) {
	if !requirePurchasePermission(w, r, "compras", "arte_final", "financeiro", "diretoria_financeira", "producao", "vendas") {
		return
	}
	id, err := purchaseID(r)
	if err != nil {
		writePurchaseError(w, err)
		return
	}
	item, err := purchaseService.GetByID(id)
	if err != nil {
		writePurchaseError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(item)
}

// GetPurchaseBySale handles GET /purchases/by-sale/{saleId} — returns the purchase associated with a given sale.
func GetPurchaseBySale(w http.ResponseWriter, r *http.Request) {
	if !requirePurchasePermission(w, r, "compras", "arte_final", "financeiro", "diretoria_financeira", "producao", "vendas") {
		return
	}
	saleID, err := strconv.Atoi(mux.Vars(r)["saleId"])
	if err != nil {
		writePurchaseError(w, err)
		return
	}
	item, err := purchaseService.GetBySaleID(saleID)
	if err != nil {
		writePurchaseError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(item)
}

// ReleaseSaleToPurchases handles POST /purchases/release/{saleId} — releases a sale to the purchases module, creating a purchase order.
func ReleaseSaleToPurchases(w http.ResponseWriter, r *http.Request) {
	if !requirePurchasePermission(w, r, "vendas") {
		return
	}
	saleID, err := strconv.Atoi(mux.Vars(r)["saleId"])
	if err != nil {
		writePurchaseError(w, err)
		return
	}
	var input struct {
		IsSample           bool `json:"is_sample"`
		SampleHasEngraving bool `json:"sample_has_engraving"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&input)
	}
	userID, _, _ := middleware.GetUserFromRequest(r)
	if err := salesWorkflowService.EnsureReadyForPurchases(saleID); err != nil {
		writePurchaseError(w, err)
		return
	}
	item, err := purchaseService.ReleaseSale(saleID, userID, input.IsSample, input.SampleHasEngraving)
	if err != nil {
		writePurchaseError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(item)
}

// UpdatePurchase handles PUT /purchases/{id} — updates an existing purchase order.
func UpdatePurchase(w http.ResponseWriter, r *http.Request) {
	if !requirePurchasePermission(w, r, "compras", "producao") {
		return
	}
	id, err := purchaseID(r)
	if err != nil {
		writePurchaseError(w, err)
		return
	}
	var input models.PurchaseUpdateInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		writePurchaseError(w, err)
		return
	}
	userID, _, _ := middleware.GetUserFromRequest(r)
	item, err := purchaseService.Update(id, userID, &input)
	if err != nil {
		writePurchaseError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(item)
}

// ExecutePurchaseAction handles POST /purchases/{id}/actions — executes a workflow action on a purchase (e.g., attach_corel, approve_first_piece).
func ExecutePurchaseAction(w http.ResponseWriter, r *http.Request) {
	id, err := purchaseID(r)
	if err != nil {
		writePurchaseError(w, err)
		return
	}
	var input models.PurchaseActionInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		writePurchaseError(w, err)
		return
	}
	allowed := false
	switch input.Action {
	case "attach_corel":
		allowed = requirePurchasePermission(w, r, "arte_final")
	case "approve_first_piece":
		allowed = requirePurchasePermission(w, r, "vendas")
	default:
		allowed = requirePurchasePermission(w, r, "compras")
	}
	if !allowed {
		return
	}
	userID, _, _ := middleware.GetUserFromRequest(r)
	item, err := purchaseService.ExecuteAction(id, userID, &input)
	if err != nil {
		writePurchaseError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(item)
}

// AddPurchaseAttachment handles POST /purchases/{id}/attachments — adds an attachment to a purchase order.
func AddPurchaseAttachment(w http.ResponseWriter, r *http.Request) {
	if !requirePurchasePermission(w, r, "compras", "arte_final", "financeiro", "producao") {
		return
	}
	id, err := purchaseID(r)
	if err != nil {
		writePurchaseError(w, err)
		return
	}
	var input models.PurchaseAttachment
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		writePurchaseError(w, err)
		return
	}
	userID, _, _ := middleware.GetUserFromRequest(r)
	item, err := purchaseService.AddAttachment(id, userID, &input)
	if err != nil {
		writePurchaseError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(item)
}

// AddPurchasePayment handles POST /purchases/{id}/payments — adds a payment record to a purchase order.
func AddPurchasePayment(w http.ResponseWriter, r *http.Request) {
	if !requirePurchasePermission(w, r, "financeiro") {
		return
	}
	id, err := purchaseID(r)
	if err != nil {
		writePurchaseError(w, err)
		return
	}
	var input models.PurchasePaymentInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		writePurchaseError(w, err)
		return
	}
	userID, _, _ := middleware.GetUserFromRequest(r)
	item, err := purchaseService.AddPayment(id, userID, &input)
	if err != nil {
		writePurchaseError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(item)
}

// ApprovePurchasePayment handles PUT /purchases/payments/{paymentId}/approve — approves a purchase payment (diretoria_financeira only).
func ApprovePurchasePayment(w http.ResponseWriter, r *http.Request) {
	if !requirePurchasePermission(w, r, "diretoria_financeira") {
		return
	}
	id, err := strconv.Atoi(mux.Vars(r)["paymentId"])
	if err != nil {
		writePurchaseError(w, err)
		return
	}
	var input struct {
		ReceiptURL string `json:"receipt_url"`
	}
	_ = json.NewDecoder(r.Body).Decode(&input)
	userID, _, _ := middleware.GetUserFromRequest(r)
	item, err := purchaseService.ApprovePayment(id, userID, input.ReceiptURL)
	if err != nil {
		writePurchaseError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(item)
}

// AddPurchaseIssue handles POST /purchases/{id}/issues — adds an issue entry to a purchase order.
func AddPurchaseIssue(w http.ResponseWriter, r *http.Request) {
	if !requirePurchasePermission(w, r, "producao", "compras") {
		return
	}
	id, err := purchaseID(r)
	if err != nil {
		writePurchaseError(w, err)
		return
	}
	var input models.PurchaseIssueInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		writePurchaseError(w, err)
		return
	}
	userID, _, _ := middleware.GetUserFromRequest(r)
	item, err := purchaseService.AddIssue(id, userID, &input)
	if err != nil {
		writePurchaseError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(item)
}

// UpdatePurchaseIssue handles PUT /purchases/issues/{issueId} — updates an existing purchase issue.
func UpdatePurchaseIssue(w http.ResponseWriter, r *http.Request) {
	if !requirePurchasePermission(w, r, "compras") {
		return
	}
	id, err := strconv.Atoi(mux.Vars(r)["issueId"])
	if err != nil {
		writePurchaseError(w, err)
		return
	}
	var input models.PurchaseIssueUpdateInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		writePurchaseError(w, err)
		return
	}
	userID, _, _ := middleware.GetUserFromRequest(r)
	item, err := purchaseService.UpdateIssue(id, userID, &input)
	if err != nil {
		writePurchaseError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(item)
}

// GetPurchaseFinancial handles GET /purchases/financial — returns a financial summary of all purchases.
func GetPurchaseFinancial(w http.ResponseWriter, r *http.Request) {
	if !requirePurchasePermission(w, r, "financeiro", "diretoria_financeira", "compras") {
		return
	}
	items, err := purchaseService.FinancialSummary()
	if err != nil {
		writePurchaseError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(items)
}

// GetNotifications handles GET /notifications — returns purchase-related notifications for the current user.
func GetNotifications(w http.ResponseWriter, r *http.Request) {
	userID, _, role := middleware.GetUserFromRequest(r)
	permissions := middleware.GetPermissionsFromRequest(r)
	if role == "admin" {
		permissions = []string{"compras", "arte_final", "financeiro", "diretoria_financeira", "producao", "vendas"}
	}
	items, err := purchaseService.ListNotifications(userID, permissions)
	if err != nil {
		writePurchaseError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(items)
}

// ReadNotification handles PUT /notifications/{id}/read — marks a notification as read.
func ReadNotification(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		writePurchaseError(w, err)
		return
	}
	userID, _, _ := middleware.GetUserFromRequest(r)
	if err = purchaseService.MarkNotificationRead(id, userID); err != nil {
		writePurchaseError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
