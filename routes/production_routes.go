package routes

import (
	"donapresentes/controllers"
	"github.com/gorilla/mux"
	"net/http"
)

func RegisterProductionRoutes(r *mux.Router) {
	r.HandleFunc("/production/dashboard", controllers.GetProductionDashboard).Methods(http.MethodGet)
	r.HandleFunc("/production/orders", controllers.ListProductionOrders).Methods(http.MethodGet)
	r.HandleFunc("/production/orders/{id}", controllers.GetProductionOrder).Methods(http.MethodGet)
	r.HandleFunc("/production/orders/{id}/receipts", controllers.AddProductionReceipt).Methods(http.MethodPost)
	r.HandleFunc("/production/orders/{id}/occurrences", controllers.AddProductionOccurrence).Methods(http.MethodPost)
	r.HandleFunc("/production/orders/{id}/occurrences/{occurrenceId}/resolve", controllers.ResolveProductionOccurrence).Methods(http.MethodPatch)
	r.HandleFunc("/production/orders/{id}/transition", controllers.TransitionProductionOrder).Methods(http.MethodPatch)
	r.HandleFunc("/production/orders/{id}/assignment", controllers.AssignProductionOrder).Methods(http.MethodPatch)
	r.HandleFunc("/production/orders/{id}/engraving-events", controllers.AddProductionEvent).Methods(http.MethodPost)
	r.HandleFunc("/production/orders/{id}/volumes", controllers.AddProductionVolume).Methods(http.MethodPost)
	r.HandleFunc("/production/orders/{id}/fiscal", controllers.AddProductionFiscal).Methods(http.MethodPost)
	r.HandleFunc("/production/orders/{id}/shipment", controllers.UpsertProductionShipment).Methods(http.MethodPut)
	r.HandleFunc("/production/supplies", controllers.ListProductionSupplies).Methods(http.MethodGet)
	r.HandleFunc("/production/supplies", controllers.CreateProductionSupply).Methods(http.MethodPost)
	r.HandleFunc("/production/supplies/movements", controllers.MoveProductionSupply).Methods(http.MethodPost)
}
