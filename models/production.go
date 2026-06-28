package models

import (
	"encoding/json"
	"time"
)

const (
	ProductionAwaitingReceipt    = "AGUARDANDO_RECEBIMENTO"
	ProductionPartialReceipt     = "RECEBIMENTO_PARCIAL"
	ProductionInspection         = "EM_CONFERENCIA"
	ProductionMaterialOK         = "MATERIAL_OK"
	ProductionBlocked            = "BLOQUEADO"
	ProductionPreparingEngraving = "SEPARANDO_GRAVACAO"
	ProductionSentEngraving      = "ENVIADO_GRAVACAO"
	ProductionAwaitingFirstPiece = "AGUARDANDO_PRIMEIRA_PECA"
	ProductionFirstPieceApproved = "PRIMEIRA_PECA_APROVADA"
	ProductionEngraving          = "EM_GRAVACAO"
	ProductionEngravingReturn    = "RETORNO_GRAVACAO"
	ProductionReturnInspection   = "CONFERENCIA_RETORNO"
	ProductionReadyShipment      = "PRONTO_EXPEDICAO"
	ProductionShipped            = "EXPEDIDO"
	ProductionDelivered          = "ENTREGUE"
	ProductionCompleted          = "CONCLUIDO"
)

type ProductionOrder struct {
	ID                    int64           `json:"id"`
	PurchaseID            int             `json:"purchase_id"`
	SaleID                int             `json:"sale_id"`
	Status                string          `json:"status"`
	HasEngraving          bool            `json:"has_engraving"`
	FirstPieceRequired    bool            `json:"first_piece_required"`
	Priority              int             `json:"priority"`
	OwnerID               *int            `json:"owner_id,omitempty"`
	Version               int             `json:"version"`
	ReleasedAt            time.Time       `json:"released_at"`
	CompletedAt           *time.Time      `json:"completed_at,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
	ConferenceStartedAt   *time.Time      `json:"conference_started_at,omitempty"`
	ConferenceCompletedAt *time.Time      `json:"conference_completed_at,omitempty"`
	ConferenceUserID      *int            `json:"conference_user_id,omitempty"`
	GeneralNumber         string          `json:"general_number"`
	CustomerName          string          `json:"customer_name"`
	SellerID              int             `json:"seller_id"`
	DepartureDate         *time.Time      `json:"departure_date,omitempty"`
	MaterialDeadline      *time.Time      `json:"material_deadline,omitempty"`
	EngravingDeadline     *time.Time      `json:"engraving_deadline,omitempty"`
	Items                 json.RawMessage `json:"items,omitempty"`
	Receipts              json.RawMessage `json:"receipts,omitempty"`
	Occurrences           json.RawMessage `json:"occurrences,omitempty"`
	EngravingEvents       json.RawMessage `json:"engraving_events,omitempty"`
	Volumes               json.RawMessage `json:"volumes,omitempty"`
	FiscalDocuments       json.RawMessage `json:"fiscal_documents,omitempty"`
	Shipment              json.RawMessage `json:"shipment,omitempty"`
	History               json.RawMessage `json:"history,omitempty"`
}

type ProductionFilters struct {
	Status, Search    string
	OwnerID, Priority *int
}
type ProductionAccess struct {
	UserID                                                                             int
	Admin, Production, Inspection, Purchases, Sales, Finance, Board, Logistics, Driver bool
}

func (a ProductionAccess) CanView() bool {
	return a.Admin || a.Production || a.Inspection || a.Purchases || a.Sales || a.Finance || a.Board || a.Logistics || a.Driver
}
func (a ProductionAccess) CanOperate() bool { return a.Admin || a.Production || a.Inspection }

type ProductionReceiptItemInput struct {
	ItemID   int64 `json:"item_id"`
	Quantity int   `json:"quantity"`
}
type ProductionReceiptInput struct {
	IdempotencyKey string                       `json:"idempotency_key"`
	InvoiceNumber  string                       `json:"invoice_number"`
	InvoiceKey     string                       `json:"invoice_key"`
	InvoiceURL     string                       `json:"invoice_url"`
	Notes          string                       `json:"notes"`
	Items          []ProductionReceiptItemInput `json:"items"`
}
type ProductionOccurrenceInput struct {
	ItemID        *int64 `json:"item_id,omitempty"`
	Kind          string `json:"kind"`
	Severity      string `json:"severity"`
	Quantity      int    `json:"quantity"`
	Description   string `json:"description"`
	AttachmentURL string `json:"attachment_url"`
}
type ProductionOccurrenceResolutionInput struct {
	Resolution string `json:"resolution"`
}
type ProductionTransitionInput struct {
	ToStatus        string `json:"to_status"`
	Note            string `json:"note"`
	ExpectedVersion int    `json:"expected_version"`
}
type ProductionEventInput struct {
	EventType      string `json:"event_type"`
	Status         string `json:"status"`
	TrackingCode   string `json:"tracking_code"`
	FileURL        string `json:"file_url"`
	Notes          string `json:"notes"`
	IdempotencyKey string `json:"idempotency_key"`
	Quantity       int    `json:"quantity"`
	CarrierID      *int   `json:"carrier_id,omitempty"`
}
type ProductionVolumeInput struct {
	Label    string  `json:"label"`
	WeightKG float64 `json:"weight_kg"`
	LengthCM float64 `json:"length_cm"`
	WidthCM  float64 `json:"width_cm"`
	HeightCM float64 `json:"height_cm"`
}
type ProductionFiscalInput struct {
	DocumentType   string     `json:"document_type"`
	DocumentNumber string     `json:"document_number"`
	AccessKey      string     `json:"access_key"`
	FileURL        string     `json:"file_url"`
	IssuedAt       *time.Time `json:"issued_at,omitempty"`
}
type ProductionShipmentInput struct {
	Method        string `json:"method"`
	CarrierID     *int   `json:"carrier_id,omitempty"`
	DriverID      *int   `json:"driver_id,omitempty"`
	TrackingCode  string `json:"tracking_code"`
	TrackingURL   string `json:"tracking_url"`
	PostalService string `json:"postal_service"`
	ProofURL      string `json:"proof_url"`
}
type ProductionSupplyInput struct {
	Name            string  `json:"name"`
	Unit            string  `json:"unit"`
	CurrentQuantity float64 `json:"current_quantity"`
	MinimumQuantity float64 `json:"minimum_quantity"`
}
type ProductionSupplyMovementInput struct {
	SupplyID      int64   `json:"supply_id"`
	OrderID       *int64  `json:"order_id,omitempty"`
	MovementType  string  `json:"movement_type"`
	Quantity      float64 `json:"quantity"`
	Reason        string  `json:"reason"`
	SupplierID    *int    `json:"supplier_id,omitempty"`
	InvoiceNumber string  `json:"invoice_number"`
	InvoiceURL    string  `json:"invoice_url"`
}
type ProductionAssignmentInput struct {
	OwnerID  *int `json:"owner_id,omitempty"`
	Priority int  `json:"priority"`
}
