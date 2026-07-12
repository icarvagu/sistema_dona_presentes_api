package models

import (
	"encoding/json"
	"time"
)

const (
	// Purchase status constants representing the procurement lifecycle.
	PurchasePending            = "Pendente de Compra"
	PurchaseReview             = "Em Conferência de Compras"
	PurchaseWaitingCorel       = "Aguardando Arquivo Corel"
	PurchaseCorelReady         = "Corel Anexado - Liberado para Compra"
	PurchaseMaterialEmailSent  = "E-mail Material Enviado"
	PurchaseEngravingEmailSent = "E-mail Gravação Enviado"
	PurchaseWaitingAcceptance  = "Awaiting Supplier Acceptance"
	PurchaseWaitingPayment     = "Aguardando Pagamento"
	PurchaseBought             = "Pedido Comprado"
	PurchaseReleased           = "Liberado para Produção"
	PurchaseInEngraving        = "Pedido na Gravação"
	PurchaseWaitingFirstPiece  = "Aguardando Aprovação da Gravação"
	PurchaseMaterialIssue      = "Pendência de Material"
	PurchaseEngravingIssue     = "Pendência de Gravação"
)

// PurchaseOrder represents a procurement order that manages the purchasing
// of materials and engraving services for a sale.
type PurchaseOrder struct {
	ID                     int                  `json:"id"`
	SaleID                 int                  `json:"sale_id"`
	GeneralNumber          string               `json:"general_number"`
	Status                 string               `json:"status"`
	StatusUpdatedAt        time.Time            `json:"status_updated_at"`
	BuyerID                *int                 `json:"buyer_id,omitempty"`
	MaterialSupplierID     *int                 `json:"material_supplier_id,omitempty"`
	EngravingSupplierID    *int                 `json:"engraving_supplier_id,omitempty"`
	IsSample               bool                 `json:"is_sample"`
	SampleHasEngraving     bool                 `json:"sample_has_engraving"`
	HasEngraving           bool                 `json:"has_engraving"`
	CorelRequired          bool                 `json:"corel_required"`
	CorelRequestedAt       *time.Time           `json:"corel_requested_at,omitempty"`
	CorelRequestedBy       *int                 `json:"corel_requested_by,omitempty"`
	CorelAttachedAt        *time.Time           `json:"corel_attached_at,omitempty"`
	CorelAttachedBy        *int                 `json:"corel_attached_by,omitempty"`
	MaterialUnitCost       float64              `json:"material_unit_cost"`
	MaterialTotalCost      float64              `json:"material_total_cost"`
	EngravingCost          float64              `json:"engraving_cost"`
	FreightCost            float64              `json:"freight_cost"`
	OtherCost              float64              `json:"other_cost"`
	BuyerDiscount          float64              `json:"buyer_discount"`
	NegotiationContact     string               `json:"negotiation_contact"`
	NegotiationNotes       string               `json:"negotiation_notes"`
	MaterialDeadline       *time.Time           `json:"material_deadline,omitempty"`
	EngravingDeadline      *time.Time           `json:"engraving_deadline,omitempty"`
	PaymentMethod          string               `json:"payment_method"`
	RequiresAdvancePayment bool                 `json:"requires_advance_payment"`
	MaterialAccepted       bool                 `json:"material_accepted"`
	MaterialAcceptedAt     *time.Time           `json:"material_accepted_at,omitempty"`
	EngravingAccepted      bool                 `json:"engraving_accepted"`
	EngravingAcceptedAt    *time.Time           `json:"engraving_accepted_at,omitempty"`
	FirstPieceRequired     bool                 `json:"first_piece_required"`
	FirstPieceStatus       string               `json:"first_piece_status"`
	FirstPieceURL          string               `json:"first_piece_url"`
	CommercialNotes        string               `json:"commercial_notes"`
	PurchaseNotes          string               `json:"purchase_notes"`
	ProductionReleasedAt   *time.Time           `json:"production_released_at,omitempty"`
	CreatedAt              time.Time            `json:"created_at"`
	UpdatedAt              time.Time            `json:"updated_at"`
	Sale                   *Sale                `json:"sale,omitempty"`
	MaterialSupplier       *Supplier            `json:"material_supplier,omitempty"`
	EngravingSupplier      *Supplier            `json:"engraving_supplier,omitempty"`
	Attachments            []PurchaseAttachment `json:"attachments"`
	Emails                 []PurchaseEmail      `json:"emails"`
	Payments               []PurchasePayment    `json:"payments"`
	Issues                 []PurchaseIssue      `json:"issues"`
	History                []PurchaseHistory    `json:"history"`
}

// PurchaseUpdateInput is the DTO for updating procurement fields on a purchase order.
type PurchaseUpdateInput struct {
	BuyerID                *int       `json:"buyer_id,omitempty"`
	MaterialSupplierID     *int       `json:"material_supplier_id,omitempty"`
	EngravingSupplierID    *int       `json:"engraving_supplier_id,omitempty"`
	IsSample               bool       `json:"is_sample"`
	SampleHasEngraving     bool       `json:"sample_has_engraving"`
	HasEngraving           bool       `json:"has_engraving"`
	MaterialUnitCost       float64    `json:"material_unit_cost"`
	MaterialTotalCost      float64    `json:"material_total_cost"`
	EngravingCost          float64    `json:"engraving_cost"`
	FreightCost            float64    `json:"freight_cost"`
	OtherCost              float64    `json:"other_cost"`
	BuyerDiscount          float64    `json:"buyer_discount"`
	NegotiationContact     string     `json:"negotiation_contact"`
	NegotiationNotes       string     `json:"negotiation_notes"`
	MaterialDeadline       *time.Time `json:"material_deadline,omitempty"`
	EngravingDeadline      *time.Time `json:"engraving_deadline,omitempty"`
	PaymentMethod          string     `json:"payment_method"`
	RequiresAdvancePayment bool       `json:"requires_advance_payment"`
	FirstPieceRequired     bool       `json:"first_piece_required"`
	CommercialNotes        string     `json:"commercial_notes"`
	PurchaseNotes          string     `json:"purchase_notes"`
}

// PurchaseActionInput carries an action (e.g. send email, attach file) for a purchase order.
type PurchaseActionInput struct {
	Action      string   `json:"action"`
	Observation string   `json:"observation"`
	URL         string   `json:"url"`
	FileName    string   `json:"file_name"`
	Recipient   string   `json:"recipient"`
	Attachments []string `json:"attachments"`
}

// PurchaseAttachment represents a file attached to a purchase order.
type PurchaseAttachment struct {
	ID         int       `json:"id"`
	PurchaseID int       `json:"purchase_id"`
	Category   string    `json:"category"`
	FileName   string    `json:"file_name"`
	URL        string    `json:"url"`
	UploadedBy *int      `json:"uploaded_by,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// PurchaseEmail records an email sent as part of the procurement process.
type PurchaseEmail struct {
	ID          int             `json:"id"`
	PurchaseID  int             `json:"purchase_id"`
	Kind        string          `json:"kind"`
	Recipient   string          `json:"recipient"`
	Subject     string          `json:"subject"`
	Body        string          `json:"body"`
	Observation string          `json:"observation"`
	Attachments json.RawMessage `json:"attachments"`
	SentBy      *int            `json:"sent_by,omitempty"`
	SentAt      time.Time       `json:"sent_at"`
}

// PurchasePayment represents a payment request linked to a purchase order.
type PurchasePayment struct {
	ID            int        `json:"id"`
	PurchaseID    int        `json:"purchase_id"`
	CostType      string     `json:"cost_type"`
	SupplierID    *int       `json:"supplier_id,omitempty"`
	Amount        float64    `json:"amount"`
	Method        string     `json:"method"`
	Status        string     `json:"status"`
	Justification string     `json:"justification"`
	ReceiptURL    string     `json:"receipt_url"`
	RequestedBy   *int       `json:"requested_by,omitempty"`
	ApprovedBy    *int       `json:"approved_by,omitempty"`
	RequestedAt   time.Time  `json:"requested_at"`
	ApprovedAt    *time.Time `json:"approved_at,omitempty"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// PurchasePaymentInput is the DTO for requesting a payment on a purchase order.
type PurchasePaymentInput struct {
	CostType      string  `json:"cost_type"`
	SupplierID    *int    `json:"supplier_id,omitempty"`
	Amount        float64 `json:"amount"`
	Method        string  `json:"method"`
	Justification string  `json:"justification"`
	ReceiptURL    string  `json:"receipt_url"`
}

// PurchaseIssue represents a problem (material or engraving) reported on a purchase order.
type PurchaseIssue struct {
	ID                 int             `json:"id"`
	PurchaseID         int             `json:"purchase_id"`
	IssueType          string          `json:"issue_type"`
	Description        string          `json:"description"`
	Attachments        json.RawMessage `json:"attachments"`
	SupplierID         *int            `json:"supplier_id,omitempty"`
	Solution           string          `json:"solution"`
	OccurrenceDate     time.Time       `json:"occurrence_date"`
	ResolutionDeadline time.Time       `json:"resolution_deadline"`
	Priority           int             `json:"priority"`
	Status             string          `json:"status"`
	OpenedBy           *int            `json:"opened_by,omitempty"`
	ResolvedBy         *int            `json:"resolved_by,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	ResolvedAt         *time.Time      `json:"resolved_at,omitempty"`
}

// PurchaseIssueInput is the DTO for reporting a new issue on a purchase order.
type PurchaseIssueInput struct {
	IssueType          string    `json:"issue_type"`
	Description        string    `json:"description"`
	Attachments        []string  `json:"attachments"`
	SupplierID         *int      `json:"supplier_id,omitempty"`
	Solution           string    `json:"solution"`
	OccurrenceDate     time.Time `json:"occurrence_date"`
	ResolutionDeadline time.Time `json:"resolution_deadline"`
	Priority           int       `json:"priority"`
}

// PurchaseIssueUpdateInput is the DTO for updating the resolution or status of an issue.
type PurchaseIssueUpdateInput struct {
	Solution           string     `json:"solution"`
	ResolutionDeadline *time.Time `json:"resolution_deadline,omitempty"`
	Status             string     `json:"status"`
}

// PurchaseHistory records a status change or action taken on a purchase order.
type PurchaseHistory struct {
	ID         int       `json:"id"`
	PurchaseID int       `json:"purchase_id"`
	Action     string    `json:"action"`
	FromStatus string    `json:"from_status"`
	ToStatus   string    `json:"to_status"`
	Details    string    `json:"details"`
	UserID     *int      `json:"user_id,omitempty"`
	UserName   string    `json:"user_name"`
	CreatedAt  time.Time `json:"created_at"`
}

// Notification represents a system notification sent to users about purchase events.
type Notification struct {
	ID                  int        `json:"id"`
	PurchaseID          *int       `json:"purchase_id,omitempty"`
	RecipientUserID     *int       `json:"recipient_user_id,omitempty"`
	RecipientPermission string     `json:"recipient_permission"`
	NotificationType    string     `json:"notification_type"`
	Message             string     `json:"message"`
	ReadAt              *time.Time `json:"read_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
}

// PurchaseFinancialSummary aggregates financial data for a purchase order for reporting.
type PurchaseFinancialSummary struct {
	PurchaseID    int     `json:"purchase_id"`
	GeneralNumber string  `json:"general_number"`
	CustomerName  string  `json:"customer_name"`
	CostType      string  `json:"cost_type"`
	SupplierName  string  `json:"supplier_name"`
	Amount        float64 `json:"amount"`
	Method        string  `json:"method"`
	Status        string  `json:"status"`
	ReceiptURL    string  `json:"receipt_url"`
}
