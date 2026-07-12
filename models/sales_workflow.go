package models

import (
	"encoding/json"
	"time"
)

// QuoteFeedbackEvent records a scheduled feedback interaction for a quotation.
type QuoteFeedbackEvent struct {
	ID          int64      `json:"id"`
	QuoteID     int        `json:"quote_id"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	Observation string     `json:"observation"`
	CreatedBy   int        `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ItemLayoutVersion tracks a versioned layout file attached to a sale or quote item.
type ItemLayoutVersion struct {
	ID             int64           `json:"id"`
	EntityType     string          `json:"entity_type"`
	ItemID         int             `json:"item_id"`
	Version        int             `json:"version"`
	Label          string          `json:"label"`
	FileURL        string          `json:"file_url"`
	ItemSnapshot   json.RawMessage `json:"item_snapshot"`
	ApprovalStatus string          `json:"approval_status"`
	ApprovalNote   string          `json:"approval_note"`
	ApprovedBy     *int            `json:"approved_by,omitempty"`
	ApprovedAt     *time.Time      `json:"approved_at,omitempty"`
	CreatedBy      int             `json:"created_by"`
	CreatedAt      time.Time       `json:"created_at"`
}

// FinancialAnalysisInput carries the result of a financial review for a sale or quote.
type FinancialAnalysisInput struct {
	Status        string   `json:"status"`
	Tags          []string `json:"tags"`
	Observation   string   `json:"observation"`
	AttachmentURL string   `json:"attachment_url"`
}

// SalePendingInput records a pending task or blocker on a sale.
type SalePendingInput struct {
	Sector      string `json:"sector"`
	Description string `json:"description"`
	Blocking    bool   `json:"blocking"`
}

// EngravingApprovalInput captures the approval/rejection response for an engraving proof.
type EngravingApprovalInput struct {
	Response    string `json:"response"`
	Observation string `json:"observation"`
}

// QuoteConversionInput maps withdrawal dates when converting a quote into a sale.
type QuoteConversionInput struct {
	WithdrawalDates map[string]*time.Time `json:"withdrawal_dates"`
}
