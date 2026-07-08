package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"donapresentes/controllers/config"
	apperrors "donapresentes/errors"
	"donapresentes/middleware"
	"donapresentes/models"
	"donapresentes/repositories"
	"donapresentes/services"

	"github.com/gorilla/mux"
)

var salesWorkflowService *services.SalesWorkflowService

func InitSalesWorkflowService() {
	salesWorkflowService = services.NewSalesWorkflowService(repositories.NewSalesWorkflowRepository(config.DB))
}

func workflowError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	if err == sql.ErrNoRows {
		status = http.StatusNotFound
	}
	middleware.ErrorHandler(w, &apperrors.AppError{Code: status, Message: err.Error()}, status)
}
func workflowJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func workflowID(r *http.Request, key string) (int, error) { return strconv.Atoi(mux.Vars(r)[key]) }
func workflowInt64ID(r *http.Request, key string) (int64, error) {
	return strconv.ParseInt(mux.Vars(r)[key], 10, 64)
}
func workflowForbidden(w http.ResponseWriter) {
	middleware.ErrorHandler(w, &apperrors.AppError{Code: http.StatusForbidden, Message: "Acesso negado"}, http.StatusForbidden)
}
func canSeeQuote(r *http.Request, id int) bool {
	user, _, _ := middleware.GetUserFromRequest(r)
	owner, err := salesWorkflowService.QuoteOwner(id)
	return err == nil && (owner == user || middleware.HasPermission(r, "orcamentos:ver_todos"))
}
func canSeeSale(r *http.Request, id int) bool {
	user, _, _ := middleware.GetUserFromRequest(r)
	owner, err := salesWorkflowService.SaleOwner(id)
	return err == nil && (owner == user || middleware.HasPermission(r, "vendas:ver_todos"))
}
func canSeeItem(r *http.Request, entity string, id int) bool {
	user, _, _ := middleware.GetUserFromRequest(r)
	owner, err := salesWorkflowService.ItemOwner(entity, id)
	return err == nil && (owner == user || middleware.HasPermission(r, "orcamentos:ver_todos") || middleware.HasPermission(r, "vendas:ver_todos") || middleware.HasPermission(r, "arte_final"))
}

func canManageSalePending(r *http.Request) bool {
	return middleware.HasPermission(r, "compras") || middleware.HasPermission(r, "producao") ||
		middleware.HasPermission(r, "financeiro") || middleware.HasPermission(r, "diretoria_financeira") ||
		middleware.HasPermission(r, "arte_final")
}

func AddQuoteFeedbackEvent(w http.ResponseWriter, r *http.Request) {
	id, err := workflowID(r, "id")
	if err != nil {
		workflowError(w, err)
		return
	}
	if !canSeeQuote(r, id) {
		workflowForbidden(w)
		return
	}
	var input struct {
		ScheduledAt *time.Time `json:"scheduled_at"`
		Observation string     `json:"observation"`
	}
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		workflowError(w, err)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	event, err := salesWorkflowService.AddFeedback(id, user, input.ScheduledAt, input.Observation)
	if err != nil {
		workflowError(w, err)
		return
	}
	workflowJSON(w, http.StatusCreated, event)
}
func GetQuoteFeedbackEvents(w http.ResponseWriter, r *http.Request) {
	id, err := workflowID(r, "id")
	if err != nil {
		workflowError(w, err)
		return
	}
	if !canSeeQuote(r, id) {
		workflowForbidden(w)
		return
	}
	events, err := salesWorkflowService.Feedbacks(id)
	if err != nil {
		workflowError(w, err)
		return
	}
	workflowJSON(w, http.StatusOK, events)
}

func CreateItemLayout(w http.ResponseWriter, r *http.Request) {
	id, err := workflowID(r, "itemId")
	if err != nil {
		workflowError(w, err)
		return
	}
	entity := mux.Vars(r)["entity"]
	if !canSeeItem(r, entity, id) {
		workflowForbidden(w)
		return
	}
	var input struct {
		FileURL string `json:"file_url"`
	}
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		workflowError(w, err)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	layout, err := salesWorkflowService.CreateLayout(entity, id, user, input.FileURL)
	if err != nil {
		workflowError(w, err)
		return
	}
	workflowJSON(w, http.StatusCreated, layout)
}
func ApproveItemLayout(w http.ResponseWriter, r *http.Request) {
	if !middleware.HasPermission(r, "arte_final") && !middleware.HasPermission(r, "vendas:aprovar_layout") {
		workflowForbidden(w)
		return
	}
	id, err := workflowInt64ID(r, "id")
	if err != nil {
		workflowError(w, err)
		return
	}
	var input struct {
		Status string `json:"status"`
		Note   string `json:"note"`
	}
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		workflowError(w, err)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	if err = salesWorkflowService.ApproveLayout(id, user, input.Status, input.Note); err != nil {
		workflowError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func ConvertQuoteToSale(w http.ResponseWriter, r *http.Request) {
	id, err := workflowID(r, "id")
	if err != nil {
		workflowError(w, err)
		return
	}
	if !canSeeQuote(r, id) {
		workflowForbidden(w)
		return
	}
	var input models.QuoteConversionInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		workflowError(w, err)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	saleID, err := salesWorkflowService.ConvertQuote(id, user, input)
	if err != nil {
		workflowError(w, err)
		return
	}
	workflowJSON(w, http.StatusOK, map[string]int{"sale_id": saleID})
}

func UpsertSaleFinancialAnalysis(w http.ResponseWriter, r *http.Request) {
	if !middleware.HasPermission(r, "financeiro") && !middleware.HasPermission(r, "diretoria_financeira") {
		workflowForbidden(w)
		return
	}
	saleID, err := workflowID(r, "id")
	if err != nil {
		workflowError(w, err)
		return
	}
	var input models.FinancialAnalysisInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		workflowError(w, err)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	if err = salesWorkflowService.UpsertFinancial(saleID, user, input); err != nil {
		workflowError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func AddSaleReceipt(w http.ResponseWriter, r *http.Request) {
	saleID, err := workflowID(r, "id")
	if err != nil {
		workflowError(w, err)
		return
	}
	if !canSeeSale(r, saleID) && !middleware.HasPermission(r, "financeiro") && !middleware.HasPermission(r, "diretoria_financeira") {
		workflowForbidden(w)
		return
	}
	var input struct {
		FileURL string `json:"file_url"`
	}
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		workflowError(w, err)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	id, err := salesWorkflowService.AddReceipt(saleID, user, input.FileURL)
	if err != nil {
		workflowError(w, err)
		return
	}
	workflowJSON(w, http.StatusCreated, map[string]int64{"id": id})
}
func ValidateSaleReceipt(w http.ResponseWriter, r *http.Request) {
	if !middleware.HasPermission(r, "financeiro") && !middleware.HasPermission(r, "diretoria_financeira") {
		workflowForbidden(w)
		return
	}
	id, err := workflowInt64ID(r, "id")
	if err != nil {
		workflowError(w, err)
		return
	}
	var input struct {
		Status      string `json:"status"`
		Observation string `json:"observation"`
	}
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		workflowError(w, err)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	if err = salesWorkflowService.ValidateReceipt(id, user, input.Status, input.Observation); err != nil {
		workflowError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func AddSalePending(w http.ResponseWriter, r *http.Request) {
	saleID, err := workflowID(r, "id")
	if err != nil {
		workflowError(w, err)
		return
	}
	if !canSeeSale(r, saleID) && !canManageSalePending(r) {
		workflowForbidden(w)
		return
	}
	var input models.SalePendingInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		workflowError(w, err)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	id, err := salesWorkflowService.AddPending(saleID, user, input)
	if err != nil {
		workflowError(w, err)
		return
	}
	workflowJSON(w, http.StatusCreated, map[string]int64{"id": id})
}
func ResolveSalePending(w http.ResponseWriter, r *http.Request) {
	id, err := workflowInt64ID(r, "id")
	if err != nil {
		workflowError(w, err)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	owner, ownerErr := salesWorkflowService.PendingSaleOwner(id)
	if ownerErr != nil || (owner != user && !middleware.HasPermission(r, "vendas:ver_todos") && !canManageSalePending(r)) {
		workflowForbidden(w)
		return
	}
	var input struct {
		Note string `json:"note"`
	}
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		workflowError(w, err)
		return
	}
	if err = salesWorkflowService.ResolvePending(id, user, input.Note); err != nil {
		workflowError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func SetQuoteImportant(w http.ResponseWriter, r *http.Request) {
	id, err := workflowID(r, "id")
	if err != nil {
		workflowError(w, err)
		return
	}
	if !canSeeQuote(r, id) {
		workflowForbidden(w)
		return
	}
	var input struct {
		Important bool `json:"important"`
	}
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		workflowError(w, err)
		return
	}
	if err = salesWorkflowService.SetQuoteImportant(id, input.Important); err != nil {
		workflowError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func GetSaleWorkflow(w http.ResponseWriter, r *http.Request) {
	id, err := workflowID(r, "id")
	if err != nil {
		workflowError(w, err)
		return
	}
	if !canSeeSale(r, id) && !middleware.HasPermission(r, "financeiro") && !middleware.HasPermission(r, "diretoria_financeira") && !middleware.HasPermission(r, "compras") && !middleware.HasPermission(r, "arte_final") && !middleware.HasPermission(r, "producao") {
		workflowForbidden(w)
		return
	}
	data, err := salesWorkflowService.SaleWorkflow(id)
	if err != nil {
		workflowError(w, err)
		return
	}
	workflowJSON(w, http.StatusOK, data)
}

func AddEngravingApproval(w http.ResponseWriter, r *http.Request) {
	itemID, err := workflowID(r, "itemId")
	if err != nil {
		workflowError(w, err)
		return
	}
	if !canSeeItem(r, "sale_item", itemID) {
		workflowForbidden(w)
		return
	}
	var input models.EngravingApprovalInput
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		workflowError(w, err)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	id, err := salesWorkflowService.AddEngravingApproval(itemID, user, input)
	if err != nil {
		workflowError(w, err)
		return
	}
	workflowJSON(w, http.StatusCreated, map[string]int64{"id": id})
}
func RecordEngravingApproval(w http.ResponseWriter, r *http.Request) {
	if !middleware.HasPermission(r, "compras") {
		workflowForbidden(w)
		return
	}
	id, err := workflowInt64ID(r, "id")
	if err != nil {
		workflowError(w, err)
		return
	}
	var input struct {
		Channel string `json:"channel"`
	}
	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		workflowError(w, err)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	if err = salesWorkflowService.RecordEngravingChannel(id, user, input.Channel); err != nil {
		workflowError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func SellerApproveSale(w http.ResponseWriter, r *http.Request) {
	saleID, err := workflowID(r, "id")
	if err != nil {
		workflowError(w, err)
		return
	}
	user, _, _ := middleware.GetUserFromRequest(r)
	owner, err := salesWorkflowService.SaleOwner(saleID)
	if err != nil {
		workflowError(w, err)
		return
	}
	if owner != user {
		workflowForbidden(w)
		return
	}
	if err = salesWorkflowService.ApproveSeller(saleID, user); err != nil {
		workflowError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func GetSalesWorkflowDashboard(w http.ResponseWriter, r *http.Request) {
	user, _, _ := middleware.GetUserFromRequest(r)
	canSeeAll := middleware.HasPermission(r, "vendas:ver_todos") || middleware.HasPermission(r, "arte_final") ||
		middleware.HasPermission(r, "financeiro") || middleware.HasPermission(r, "diretoria_financeira") ||
		middleware.HasPermission(r, "compras") || middleware.HasPermission(r, "producao")
	data, err := salesWorkflowService.Dashboard(user, canSeeAll)
	if err != nil {
		workflowError(w, err)
		return
	}
	workflowJSON(w, http.StatusOK, data)
}
