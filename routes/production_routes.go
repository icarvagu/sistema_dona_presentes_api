package routes

import (
	"donapresentes/controllers"
	"github.com/gorilla/mux"
	"net/http"
)

// RegisterProductionRoutes registers all production-related HTTP endpoints on the given router.
// Routes: GET /production/dashboard, GET /production/orders, GET /production/orders/{id},
// POST /production/orders/{id}/receipts, POST /production/orders/{id}/occurrences,
// PATCH /production/orders/{id}/occurrences/{occurrenceId}/resolve,
// PATCH /production/orders/{id}/transition, PATCH /production/orders/{id}/assignment,
// POST /production/orders/{id}/engraving-events, POST /production/orders/{id}/volumes,
// POST /production/orders/{id}/fiscal, PUT /production/orders/{id}/shipment,
// GET /production/supplies, POST /production/supplies, POST /production/supplies/movements.
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
