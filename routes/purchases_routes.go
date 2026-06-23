package routes

import (
	"net/http"

	"donapresentes/controllers"

	"github.com/gorilla/mux"
)

func RegisterPurchasesRoutes(r *mux.Router) {
	r.HandleFunc("/purchases", controllers.GetPurchases).Methods(http.MethodGet)
	r.HandleFunc("/purchases/financial", controllers.GetPurchaseFinancial).Methods(http.MethodGet)
	r.HandleFunc("/purchases/by-sale/{saleId}", controllers.GetPurchaseBySale).Methods(http.MethodGet)
	r.HandleFunc("/purchases/{id}", controllers.GetPurchase).Methods(http.MethodGet)
	r.HandleFunc("/purchases/release/{saleId}", controllers.ReleaseSaleToPurchases).Methods(http.MethodPost)
	r.HandleFunc("/purchases/{id}", controllers.UpdatePurchase).Methods(http.MethodPut)
	r.HandleFunc("/purchases/{id}/actions", controllers.ExecutePurchaseAction).Methods(http.MethodPost)
	r.HandleFunc("/purchases/{id}/attachments", controllers.AddPurchaseAttachment).Methods(http.MethodPost)
	r.HandleFunc("/purchases/{id}/payments", controllers.AddPurchasePayment).Methods(http.MethodPost)
	r.HandleFunc("/purchases/payments/{paymentId}/approve", controllers.ApprovePurchasePayment).Methods(http.MethodPut)
	r.HandleFunc("/purchases/{id}/issues", controllers.AddPurchaseIssue).Methods(http.MethodPost)
	r.HandleFunc("/purchases/issues/{issueId}", controllers.UpdatePurchaseIssue).Methods(http.MethodPut)
	r.HandleFunc("/notifications", controllers.GetNotifications).Methods(http.MethodGet)
	r.HandleFunc("/notifications/{id}/read", controllers.ReadNotification).Methods(http.MethodPut)
}
