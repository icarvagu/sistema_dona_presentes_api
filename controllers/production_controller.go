package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"donapresentes/controllers/config"
	"donapresentes/middleware"
	"donapresentes/models"
	"donapresentes/repositories"
	"donapresentes/services"
	"github.com/gorilla/mux"
)

// productionService provides operations for production order management, receipts, occurrences, and supplies.
var productionService *services.ProductionService

// InitProductionService initializes the production service with the required repository.
func InitProductionService() {
	productionService = services.NewProductionService(repositories.NewProductionRepository(config.DB), auditService)
}

func productionAccess(r *http.Request) models.ProductionAccess {
	uid, _, role := middleware.GetUserFromRequest(r)
	has := func(k string) bool { return middleware.HasPermission(r, k) }
	return models.ProductionAccess{UserID: uid, Admin: role == "admin", Production: has("producao"), Inspection: has("conferencia"), Purchases: has("compras"), Sales: has("vendas"), Finance: has("financeiro"), Board: has("diretoria_financeira"), Logistics: has("logistica"), Driver: has("motorista")}
}
func productionID(r *http.Request) (int64, error)           { return strconv.ParseInt(mux.Vars(r)["id"], 10, 64) }
func decodeProduction(r *http.Request, v interface{}) error { return json.NewDecoder(r.Body).Decode(v) }
func productionFilters(r *http.Request) (models.ProductionFilters, error) {
	f := models.ProductionFilters{Status: r.URL.Query().Get("status"), Search: r.URL.Query().Get("search")}
	if raw := r.URL.Query().Get("owner_id"); raw != "" {
		v, e := strconv.Atoi(raw)
		if e != nil {
			return f, e
		}
		f.OwnerID = &v
	}
	if raw := r.URL.Query().Get("priority"); raw != "" {
		v, e := strconv.Atoi(raw)
		if e != nil {
			return f, e
		}
		f.Priority = &v
	}
	return f, nil
}
// GetProductionDashboard handles GET /production/dashboard — returns dashboard data filtered by status, search, owner, and priority.
func GetProductionDashboard(w http.ResponseWriter, r *http.Request) {
	a := productionAccess(r)
	f, e := productionFilters(r)
	if e != nil {
		workflowError(w, e)
		return
	}
	data, e := productionService.Dashboard(f, a)
	if e != nil {
		workflowError(w, e)
		return
	}
	workflowJSON(w, http.StatusOK, data)
}
// ListProductionOrders handles GET /production/orders — lists production orders with optional filters.
func ListProductionOrders(w http.ResponseWriter, r *http.Request) {
	a := productionAccess(r)
	f, e := productionFilters(r)
	if e != nil {
		workflowError(w, e)
		return
	}
	data, e := productionService.List(f, a)
	if e != nil {
		workflowError(w, e)
		return
	}
	workflowJSON(w, http.StatusOK, data)
}
// AssignProductionOrder handles PATCH /production/orders/{id}/assignment — assigns a production order to a user.
func AssignProductionOrder(w http.ResponseWriter, r *http.Request) {
	id, e := productionID(r)
	if e != nil {
		workflowError(w, e)
		return
	}
	var in models.ProductionAssignmentInput
	if e = decodeProduction(r, &in); e != nil {
		workflowError(w, e)
		return
	}
	if e = productionService.Assign(id, in, productionAccess(r)); e != nil {
		workflowError(w, e)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
// GetProductionOrder handles GET /production/orders/{id} — returns a single production order by ID.
func GetProductionOrder(w http.ResponseWriter, r *http.Request) {
	id, e := productionID(r)
	if e != nil {
		workflowError(w, e)
		return
	}
	data, e := productionService.Get(id, productionAccess(r))
	if e != nil {
		workflowError(w, e)
		return
	}
	workflowJSON(w, http.StatusOK, data)
}
// AddProductionReceipt handles POST /production/orders/{id}/receipts — adds a receipt entry to a production order.
func AddProductionReceipt(w http.ResponseWriter, r *http.Request) {
	id, e := productionID(r)
	if e != nil {
		workflowError(w, e)
		return
	}
	var in models.ProductionReceiptInput
	if e = decodeProduction(r, &in); e != nil {
		workflowError(w, e)
		return
	}
	if in.IdempotencyKey == "" {
		in.IdempotencyKey = r.Header.Get("Idempotency-Key")
	}
	data, e := productionService.Receipt(id, in, productionAccess(r))
	if e != nil {
		workflowError(w, e)
		return
	}
	workflowJSON(w, http.StatusCreated, data)
}
// AddProductionOccurrence handles POST /production/orders/{id}/occurrences — adds an occurrence to a production order.
func AddProductionOccurrence(w http.ResponseWriter, r *http.Request) {
	id, e := productionID(r)
	if e != nil {
		workflowError(w, e)
		return
	}
	var in models.ProductionOccurrenceInput
	if e = decodeProduction(r, &in); e != nil {
		workflowError(w, e)
		return
	}
	data, e := productionService.Occurrence(id, in, productionAccess(r))
	if e != nil {
		workflowError(w, e)
		return
	}
	workflowJSON(w, http.StatusCreated, data)
}
// ResolveProductionOccurrence handles PATCH /production/orders/{id}/occurrences/{occurrenceId}/resolve — resolves an occurrence on a production order.
func ResolveProductionOccurrence(w http.ResponseWriter, r *http.Request) {
	id, e := productionID(r)
	if e != nil {
		workflowError(w, e)
		return
	}
	occ, e := strconv.ParseInt(mux.Vars(r)["occurrenceId"], 10, 64)
	if e != nil {
		workflowError(w, e)
		return
	}
	var in models.ProductionOccurrenceResolutionInput
	if e = decodeProduction(r, &in); e != nil {
		workflowError(w, e)
		return
	}
	if e = productionService.ResolveOccurrence(id, occ, in, productionAccess(r)); e != nil {
		workflowError(w, e)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
// TransitionProductionOrder handles PATCH /production/orders/{id}/transition — transitions the status of a production order.
func TransitionProductionOrder(w http.ResponseWriter, r *http.Request) {
	id, e := productionID(r)
	if e != nil {
		workflowError(w, e)
		return
	}
	var in models.ProductionTransitionInput
	if e = decodeProduction(r, &in); e != nil {
		workflowError(w, e)
		return
	}
	if e = productionService.Transition(id, in, productionAccess(r)); e != nil {
		workflowError(w, e)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
// AddProductionEvent handles POST /production/orders/{id}/engraving-events — adds an engraving event to a production order.
func AddProductionEvent(w http.ResponseWriter, r *http.Request) {
	id, e := productionID(r)
	if e != nil {
		workflowError(w, e)
		return
	}
	var in models.ProductionEventInput
	if e = decodeProduction(r, &in); e != nil {
		workflowError(w, e)
		return
	}
	if in.IdempotencyKey == "" {
		in.IdempotencyKey = r.Header.Get("Idempotency-Key")
	}
	if e = productionService.Event(id, in, productionAccess(r)); e != nil {
		workflowError(w, e)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
// AddProductionVolume handles POST /production/orders/{id}/volumes — adds volume tracking data to a production order.
func AddProductionVolume(w http.ResponseWriter, r *http.Request) {
	id, e := productionID(r)
	if e != nil {
		workflowError(w, e)
		return
	}
	var in models.ProductionVolumeInput
	if e = decodeProduction(r, &in); e != nil {
		workflowError(w, e)
		return
	}
	if e = productionService.Volume(id, in, productionAccess(r)); e != nil {
		workflowError(w, e)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
// AddProductionFiscal handles POST /production/orders/{id}/fiscal — adds fiscal/invoice data to a production order.
func AddProductionFiscal(w http.ResponseWriter, r *http.Request) {
	id, e := productionID(r)
	if e != nil {
		workflowError(w, e)
		return
	}
	var in models.ProductionFiscalInput
	if e = decodeProduction(r, &in); e != nil {
		workflowError(w, e)
		return
	}
	if e = productionService.Fiscal(id, in, productionAccess(r)); e != nil {
		workflowError(w, e)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
// UpsertProductionShipment handles PUT /production/orders/{id}/shipment — creates or updates shipment info for a production order.
func UpsertProductionShipment(w http.ResponseWriter, r *http.Request) {
	id, e := productionID(r)
	if e != nil {
		workflowError(w, e)
		return
	}
	var in models.ProductionShipmentInput
	if e = decodeProduction(r, &in); e != nil {
		workflowError(w, e)
		return
	}
	if e = productionService.Shipment(id, in, productionAccess(r)); e != nil {
		workflowError(w, e)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
// ListProductionSupplies handles GET /production/supplies — lists all production supplies.
func ListProductionSupplies(w http.ResponseWriter, r *http.Request) {
	data, e := productionService.Supplies(productionAccess(r))
	if e != nil {
		workflowError(w, e)
		return
	}
	workflowJSON(w, http.StatusOK, data)
}
// CreateProductionSupply handles POST /production/supplies — creates a new production supply entry.
func CreateProductionSupply(w http.ResponseWriter, r *http.Request) {
	var in models.ProductionSupplyInput
	if e := decodeProduction(r, &in); e != nil {
		workflowError(w, e)
		return
	}
	id, e := productionService.CreateSupply(in, productionAccess(r))
	if e != nil {
		workflowError(w, e)
		return
	}
	workflowJSON(w, http.StatusCreated, map[string]int64{"id": id})
}
// MoveProductionSupply handles POST /production/supplies/movements — records a movement (in/out) of a production supply.
func MoveProductionSupply(w http.ResponseWriter, r *http.Request) {
	var in models.ProductionSupplyMovementInput
	if e := decodeProduction(r, &in); e != nil {
		workflowError(w, e)
		return
	}
	if e := productionService.MoveSupply(in, productionAccess(r)); e != nil {
		workflowError(w, e)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
