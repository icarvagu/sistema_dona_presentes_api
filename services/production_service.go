package services

import (
	"errors"
	"fmt"
	"strings"

	"donapresentes/models"
	"donapresentes/repositories"
)

type ProductionService struct {
	repo         *repositories.ProductionRepository
	auditService *AuditService
}

func NewProductionService(repo *repositories.ProductionRepository, auditService *AuditService) *ProductionService {
	return &ProductionService{repo: repo, auditService: auditService}
}

var productionTransitions = map[string]map[string]bool{
	models.ProductionAwaitingReceipt:    {models.ProductionPartialReceipt: true, models.ProductionInspection: true, models.ProductionBlocked: true},
	models.ProductionPartialReceipt:     {models.ProductionPartialReceipt: true, models.ProductionInspection: true, models.ProductionBlocked: true},
	models.ProductionInspection:         {models.ProductionMaterialOK: true, models.ProductionBlocked: true},
	models.ProductionBlocked:            {models.ProductionPartialReceipt: true, models.ProductionInspection: true, models.ProductionMaterialOK: true},
	models.ProductionMaterialOK:         {models.ProductionPreparingEngraving: true, models.ProductionReadyShipment: true, models.ProductionBlocked: true},
	models.ProductionPreparingEngraving: {models.ProductionSentEngraving: true, models.ProductionBlocked: true},
	models.ProductionSentEngraving:      {models.ProductionAwaitingFirstPiece: true, models.ProductionEngraving: true, models.ProductionBlocked: true},
	models.ProductionAwaitingFirstPiece: {models.ProductionFirstPieceApproved: true, models.ProductionAwaitingFirstPiece: true, models.ProductionBlocked: true},
	models.ProductionFirstPieceApproved: {models.ProductionEngraving: true, models.ProductionBlocked: true},
	models.ProductionEngraving:          {models.ProductionEngravingReturn: true, models.ProductionBlocked: true},
	models.ProductionEngravingReturn:    {models.ProductionReturnInspection: true, models.ProductionBlocked: true},
	models.ProductionReturnInspection:   {models.ProductionReadyShipment: true, models.ProductionBlocked: true},
	models.ProductionReadyShipment:      {models.ProductionShipped: true, models.ProductionBlocked: true},
	models.ProductionShipped:            {models.ProductionDelivered: true, models.ProductionBlocked: true},
	models.ProductionDelivered:          {models.ProductionCompleted: true},
}

func ValidProductionTransition(from, to string, hasEngraving, firstPiece bool) error {
	if !productionTransitions[from][to] {
		return errors.New("transição de produção inválida")
	}
	if from == models.ProductionMaterialOK && hasEngraving && to == models.ProductionReadyShipment {
		return errors.New("pedido com gravação deve concluir o fluxo de gravação")
	}
	if from == models.ProductionMaterialOK && !hasEngraving && to == models.ProductionPreparingEngraving {
		return errors.New("pedido sem gravação deve seguir para expedição")
	}
	if from == models.ProductionSentEngraving && firstPiece && to == models.ProductionEngraving {
		return errors.New("a primeira peça precisa ser aprovada")
	}
	if from == models.ProductionSentEngraving && !firstPiece && to == models.ProductionAwaitingFirstPiece {
		return errors.New("o pedido não exige primeira peça")
	}
	return nil
}
func (s *ProductionService) List(f models.ProductionFilters, a models.ProductionAccess) ([]models.ProductionOrder, error) {
	if !a.CanView() {
		return nil, errors.New("acesso negado")
	}
	return s.repo.List(f, a)
}
func (s *ProductionService) Get(id int64, a models.ProductionAccess) (*models.ProductionOrder, error) {
	if !a.CanView() {
		return nil, errors.New("acesso negado")
	}
	o, e := s.repo.Get(id)
	if e != nil {
		return nil, e
	}
	if !a.Admin && a.Sales && !(a.Production || a.Inspection || a.Purchases || a.Finance || a.Board || a.Logistics) && o.SellerID != a.UserID {
		return nil, errors.New("acesso negado")
	}
	if !a.Admin && a.Driver && !(a.Production || a.Inspection || a.Purchases || a.Sales || a.Finance || a.Board || a.Logistics) && !s.repo.DriverAssigned(id, a.UserID) {
		return nil, errors.New("acesso negado")
	}
	return o, nil
}
func (s *ProductionService) Dashboard(f models.ProductionFilters, a models.ProductionAccess) (map[string]interface{}, error) {
	orders, e := s.List(f, a)
	if e != nil {
		return nil, e
	}
	counts := map[string]int{}
	for _, o := range orders {
		counts[o.Status]++
	}
	return map[string]interface{}{"counts": counts, "orders": orders, "total": len(orders)}, nil
}
func (s *ProductionService) Assign(id int64, in models.ProductionAssignmentInput, a models.ProductionAccess) error {
	if !(a.Admin || a.Production) {
		return errors.New("atribuição exige Produção")
	}
	if in.Priority < 0 || in.Priority > 3 {
		return errors.New("prioridade deve estar entre 0 e 3")
	}
	err := s.repo.Assign(id, in, a.UserID)
	if err == nil {
		s.auditService.LogSimple(&a.UserID, "production_assigned", "production", fmt.Sprintf("order_id=%d", id), "")
	}
	return err
}
func (s *ProductionService) Receipt(id int64, in models.ProductionReceiptInput, a models.ProductionAccess) (*models.ProductionOrder, error) {
	if !a.CanOperate() && !a.Purchases {
		return nil, errors.New("acesso negado")
	}
	if strings.TrimSpace(in.IdempotencyKey) == "" || len(in.Items) == 0 {
		return nil, errors.New("chave de idempotência e itens são obrigatórios")
	}
	for _, i := range in.Items {
		if i.ItemID <= 0 || i.Quantity <= 0 {
			return nil, errors.New("item e quantidade devem ser válidos")
		}
	}
	p, err := s.repo.AddReceipt(id, in, a.UserID)
	if err == nil {
		s.auditService.LogSimple(&a.UserID, "production_receipt", "production", fmt.Sprintf("order_id=%d items=%d", id, len(in.Items)), "")
	}
	return p, err
}
func (s *ProductionService) Occurrence(id int64, in models.ProductionOccurrenceInput, a models.ProductionAccess) (*models.ProductionOrder, error) {
	if !a.CanOperate() && !a.Purchases {
		return nil, errors.New("acesso negado")
	}
	in.Severity = strings.ToUpper(in.Severity)
	if in.Severity != "PARCIAL" && in.Severity != "BLOQUEANTE" {
		return nil, errors.New("severidade inválida")
	}
	if strings.TrimSpace(in.Description) == "" {
		return nil, errors.New("descrição obrigatória")
	}
	return s.repo.AddOccurrence(id, in, a.UserID)
}
func (s *ProductionService) ResolveOccurrence(id, occ int64, in models.ProductionOccurrenceResolutionInput, a models.ProductionAccess) error {
	if !a.CanOperate() && !a.Purchases {
		return errors.New("acesso negado")
	}
	if strings.TrimSpace(in.Resolution) == "" {
		return errors.New("resolução obrigatória")
	}
	return s.repo.ResolveOccurrence(id, occ, in.Resolution, a.UserID)
}
func (s *ProductionService) Transition(id int64, in models.ProductionTransitionInput, a models.ProductionAccess) error {
	o, e := s.Get(id, a)
	if e != nil {
		return e
	}
	to := strings.ToUpper(in.ToStatus)
	if e = ValidProductionTransition(o.Status, to, o.HasEngraving, o.FirstPieceRequired); e != nil {
		return e
	}
	switch {
	case to == models.ProductionFirstPieceApproved || o.Status == models.ProductionAwaitingFirstPiece:
		if !(a.Admin || a.Sales || a.Production) {
			return errors.New("primeira peça exige Vendas ou Produção")
		}
	case to == models.ProductionMaterialOK || to == models.ProductionReturnInspection:
		if !(a.Admin || a.Production || a.Inspection) {
			return errors.New("etapa exige Produção/Conferência")
		}
	case to == models.ProductionShipped || to == models.ProductionDelivered || to == models.ProductionCompleted:
		if !(a.Admin || a.Logistics || a.Driver) {
			return errors.New("etapa exige Logística ou Motorista")
		}
	default:
		if !(a.Admin || a.Production) {
			return errors.New("transição exige Produção")
		}
	}
	if o.Status == models.ProductionAwaitingFirstPiece && to == models.ProductionAwaitingFirstPiece && strings.TrimSpace(in.Note) == "" {
		return errors.New("informe o motivo da reprovação")
	}
	err := s.repo.Transition(id, o.Status, to, in.Note, in.ExpectedVersion, a.UserID)
	if err == nil {
		s.auditService.LogSimple(&a.UserID, "production_transition", "production", fmt.Sprintf("order_id=%d from=%s to=%s", id, o.Status, to), "")
	}
	return err
}
func (s *ProductionService) Event(id int64, in models.ProductionEventInput, a models.ProductionAccess) error {
	if !(a.Admin || a.Production || a.Logistics || a.Driver) {
		return errors.New("acesso negado")
	}
	if in.IdempotencyKey == "" || in.EventType == "" {
		return errors.New("evento e chave de idempotência são obrigatórios")
	}
	if a.Driver && !(a.Admin || a.Production || a.Logistics) && !s.repo.DriverAssigned(id, a.UserID) {
		return errors.New("motorista não atribuído a esta ordem")
	}
	return s.repo.AddEvent(id, in, a.UserID)
}
func (s *ProductionService) Volume(id int64, in models.ProductionVolumeInput, a models.ProductionAccess) error {
	if !(a.Admin || a.Production || a.Logistics) {
		return errors.New("acesso negado")
	}
	if in.Label == "" {
		return errors.New("identificação do volume obrigatória")
	}
	return s.repo.AddVolume(id, in, a.UserID)
}
func (s *ProductionService) Fiscal(id int64, in models.ProductionFiscalInput, a models.ProductionAccess) error {
	if !(a.Admin || a.Finance || a.Board || a.Logistics) {
		return errors.New("acesso negado")
	}
	if in.DocumentType == "" || in.DocumentNumber == "" {
		return errors.New("tipo e número do documento são obrigatórios")
	}
	return s.repo.AddFiscal(id, in, a.UserID)
}
func (s *ProductionService) Shipment(id int64, in models.ProductionShipmentInput, a models.ProductionAccess) error {
	if !(a.Admin || a.Logistics || a.Driver) {
		return errors.New("acesso negado")
	}
	if in.Method == "" {
		return errors.New("modalidade de expedição obrigatória")
	}
	if a.Driver && !(a.Admin || a.Logistics) && !s.repo.DriverAssigned(id, a.UserID) {
		return errors.New("motorista não atribuído a esta ordem")
	}
	return s.repo.UpsertShipment(id, in, a.UserID)
}
func (s *ProductionService) Supplies(a models.ProductionAccess) ([]map[string]interface{}, error) {
	if !(a.Admin || a.Production || a.Purchases) {
		return nil, errors.New("acesso negado")
	}
	return s.repo.ListSupplies()
}
func (s *ProductionService) CreateSupply(in models.ProductionSupplyInput, a models.ProductionAccess) (int64, error) {
	if !(a.Admin || a.Production || a.Purchases) {
		return 0, errors.New("acesso negado")
	}
	if in.Name == "" || in.Unit == "" {
		return 0, errors.New("nome e unidade obrigatórios")
	}
	if in.CurrentQuantity != 0 {
		return 0, errors.New("register the supply with zero balance then record the entry with supplier and invoice")
	}
	return s.repo.CreateSupply(in)
}
func (s *ProductionService) MoveSupply(in models.ProductionSupplyMovementInput, a models.ProductionAccess) error {
	if !(a.Admin || a.Production || a.Purchases) {
		return errors.New("acesso negado")
	}
	if in.Quantity <= 0 {
		return errors.New("quantidade deve ser positiva")
	}
	if in.MovementType == "ENTRADA" && (in.SupplierID == nil || strings.TrimSpace(in.InvoiceNumber) == "" || strings.TrimSpace(in.InvoiceURL) == "") {
		return errors.New("supply entry requires supplier, invoice and file")
	}
	return s.repo.MoveSupply(in, a.UserID)
}
