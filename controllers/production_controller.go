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

var productionService *services.ProductionService

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
func ListProductionSupplies(w http.ResponseWriter, r *http.Request) {
	data, e := productionService.Supplies(productionAccess(r))
	if e != nil {
		workflowError(w, e)
		return
	}
	workflowJSON(w, http.StatusOK, data)
}
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
