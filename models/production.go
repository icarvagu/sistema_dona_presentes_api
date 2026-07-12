package models

import (
	"encoding/json"
	"time"
)

const (
	// Production status constants representing the lifecycle of a production order.
	ProductionAwaitingReceipt    = "AGUARDANDO_RECEBIMENTO"    // waiting for material receipt
	ProductionPartialReceipt     = "RECEBIMENTO_PARCIAL"       // partial material received
	ProductionInspection         = "EM_CONFERENCIA"             // under inspection
	ProductionMaterialOK         = "MATERIAL_OK"               // material approved
	ProductionBlocked            = "BLOQUEADO"                  // blocked
	ProductionPreparingEngraving = "SEPARANDO_GRAVACAO"        // preparing for engraving
	ProductionSentEngraving      = "ENVIADO_GRAVACAO"          // sent to engraving
	ProductionAwaitingFirstPiece = "AGUARDANDO_PRIMEIRA_PECA"  // waiting for first piece
	ProductionFirstPieceApproved = "PRIMEIRA_PECA_APROVADA"    // first piece approved
	ProductionEngraving          = "EM_GRAVACAO"               // in engraving
	ProductionEngravingReturn    = "RETORNO_GRAVACAO"          // returned from engraving
	ProductionReturnInspection   = "CONFERENCIA_RETORNO"       // return inspection
	ProductionReadyShipment      = "PRONTO_EXPEDICAO"          // ready for shipment
	ProductionShipped            = "EXPEDIDO"                  // shipped
	ProductionDelivered          = "ENTREGUE"                  // delivered
	ProductionCompleted          = "CONCLUIDO"                  // completed
)

// ProductionOrder represents a production work order that tracks the full
// lifecycle from material receipt through engraving and final shipment.
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

// ProductionFilters defines the criteria for querying production orders.
type ProductionFilters struct {
	Status, Search    string
	OwnerID, Priority *int
}
// ProductionAccess encodes the permissions a user has over production resources.
type ProductionAccess struct {
	UserID                                                                             int
	Admin, Production, Inspection, Purchases, Sales, Finance, Board, Logistics, Driver bool
}

// CanView reports whether the user can view production data.
func (a ProductionAccess) CanView() bool {
	return a.Admin || a.Production || a.Inspection || a.Purchases || a.Sales || a.Finance || a.Board || a.Logistics || a.Driver
}
// CanOperate reports whether the user can modify production data.
func (a ProductionAccess) CanOperate() bool { return a.Admin || a.Production || a.Inspection }

// ProductionReceiptItemInput describes a single item within a material receipt.
type ProductionReceiptItemInput struct {
	ItemID   int64 `json:"item_id"`
	Quantity int   `json:"quantity"`
}
// ProductionReceiptInput is the DTO for recording a material receipt.
type ProductionReceiptInput struct {
	IdempotencyKey string                       `json:"idempotency_key"`
	InvoiceNumber  string                       `json:"invoice_number"`
	InvoiceKey     string                       `json:"invoice_key"`
	InvoiceURL     string                       `json:"invoice_url"`
	Notes          string                       `json:"notes"`
	Items          []ProductionReceiptItemInput `json:"items"`
}
// ProductionOccurrenceInput records a problem or event within a production order.
type ProductionOccurrenceInput struct {
	ItemID        *int64 `json:"item_id,omitempty"`
	Kind          string `json:"kind"`
	Severity      string `json:"severity"`
	Quantity      int    `json:"quantity"`
	Description   string `json:"description"`
	AttachmentURL string `json:"attachment_url"`
}
// ProductionOccurrenceResolutionInput carries the resolution description for an occurrence.
type ProductionOccurrenceResolutionInput struct {
	Resolution string `json:"resolution"`
}
// ProductionTransitionInput is used to advance a production order to a new status.
type ProductionTransitionInput struct {
	ToStatus        string `json:"to_status"`
	Note            string `json:"note"`
	ExpectedVersion int    `json:"expected_version"`
}
// ProductionEventInput records an engraving-related event for a production order.
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
// ProductionVolumeInput describes a physical volume/package within a shipment.
type ProductionVolumeInput struct {
	Label    string  `json:"label"`
	WeightKG float64 `json:"weight_kg"`
	LengthCM float64 `json:"length_cm"`
	WidthCM  float64 `json:"width_cm"`
	HeightCM float64 `json:"height_cm"`
}
// ProductionFiscalInput records a fiscal document (e.g. invoice) linked to a production order.
type ProductionFiscalInput struct {
	DocumentType   string     `json:"document_type"`
	DocumentNumber string     `json:"document_number"`
	AccessKey      string     `json:"access_key"`
	FileURL        string     `json:"file_url"`
	IssuedAt       *time.Time `json:"issued_at,omitempty"`
}
// ProductionShipmentInput records shipment details for a production order.
type ProductionShipmentInput struct {
	Method        string `json:"method"`
	CarrierID     *int   `json:"carrier_id,omitempty"`
	DriverID      *int   `json:"driver_id,omitempty"`
	TrackingCode  string `json:"tracking_code"`
	TrackingURL   string `json:"tracking_url"`
	PostalService string `json:"postal_service"`
	ProofURL      string `json:"proof_url"`
}
// ProductionSupplyInput defines a supply item used in production.
type ProductionSupplyInput struct {
	Name            string  `json:"name"`
	Unit            string  `json:"unit"`
	CurrentQuantity float64 `json:"current_quantity"`
	MinimumQuantity float64 `json:"minimum_quantity"`
}
// ProductionSupplyMovementInput records a stock movement (in/out) for a production supply.
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
// ProductionAssignmentInput is used to assign an owner and priority to a production order.
type ProductionAssignmentInput struct {
	OwnerID  *int `json:"owner_id,omitempty"`
	Priority int  `json:"priority"`
}
