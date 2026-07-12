package routes

import (
	"donapresentes/controllers"
	"github.com/gorilla/mux"
	"net/http"
)

// RegisterSalesWorkflowRoutes registers all sales workflow and related HTTP endpoints on the given router.
// Routes: GET /sales-workflow/dashboard, GET /quotes/{id}/feedback-events,
// POST /quotes/{id}/feedback-events, POST /quotes/{id}/convert, PATCH /quotes/{id}/important,
// POST /item-layouts/{entity}/{itemId}, PATCH /item-layouts/{id}/approval,
// PUT /sales/{id}/financial-analysis, GET /sales/{id}/workflow, POST /sales/{id}/receipts,
// PATCH /sales/receipts/{id}/validation, POST /sales/{id}/pending-events,
// PATCH /sales/pending-events/{id}/resolve, POST /sale-items/{itemId}/engraving-approvals,
// PATCH /engraving-approvals/{id}/record, PATCH /sales/{id}/seller-approval.
func RegisterSalesWorkflowRoutes(r *mux.Router) {
	r.HandleFunc("/sales-workflow/dashboard", controllers.GetSalesWorkflowDashboard).Methods(http.MethodGet)
	r.HandleFunc("/quotes/{id}/feedback-events", controllers.GetQuoteFeedbackEvents).Methods(http.MethodGet)
	r.HandleFunc("/quotes/{id}/feedback-events", controllers.AddQuoteFeedbackEvent).Methods(http.MethodPost)
	r.HandleFunc("/quotes/{id}/convert", controllers.ConvertQuoteToSale).Methods(http.MethodPost)
	r.HandleFunc("/quotes/{id}/important", controllers.SetQuoteImportant).Methods(http.MethodPatch)
	r.HandleFunc("/item-layouts/{entity}/{itemId}", controllers.CreateItemLayout).Methods(http.MethodPost)
	r.HandleFunc("/item-layouts/{id}/approval", controllers.ApproveItemLayout).Methods(http.MethodPatch)
	r.HandleFunc("/sales/{id}/financial-analysis", controllers.UpsertSaleFinancialAnalysis).Methods(http.MethodPut)
	r.HandleFunc("/sales/{id}/workflow", controllers.GetSaleWorkflow).Methods(http.MethodGet)
	r.HandleFunc("/sales/{id}/receipts", controllers.AddSaleReceipt).Methods(http.MethodPost)
	r.HandleFunc("/sales/receipts/{id}/validation", controllers.ValidateSaleReceipt).Methods(http.MethodPatch)
	r.HandleFunc("/sales/{id}/pending-events", controllers.AddSalePending).Methods(http.MethodPost)
	r.HandleFunc("/sales/pending-events/{id}/resolve", controllers.ResolveSalePending).Methods(http.MethodPatch)
	r.HandleFunc("/sale-items/{itemId}/engraving-approvals", controllers.AddEngravingApproval).Methods(http.MethodPost)
	r.HandleFunc("/engraving-approvals/{id}/record", controllers.RecordEngravingApproval).Methods(http.MethodPatch)
	r.HandleFunc("/sales/{id}/seller-approval", controllers.SellerApproveSale).Methods(http.MethodPatch)
}
