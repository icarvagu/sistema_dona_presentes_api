package models

import (
	"encoding/json"
	"time"
)

type QuoteFeedbackEvent struct {
	ID          int64      `json:"id"`
	QuoteID     int        `json:"quote_id"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	Observation string     `json:"observation"`
	CreatedBy   int        `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
}

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

type FinancialAnalysisInput struct {
	Status        string   `json:"status"`
	Tags          []string `json:"tags"`
	Observation   string   `json:"observation"`
	AttachmentURL string   `json:"attachment_url"`
}

type SalePendingInput struct {
	Sector      string `json:"sector"`
	Description string `json:"description"`
	Blocking    bool   `json:"blocking"`
}

type EngravingApprovalInput struct {
	Response    string `json:"response"`
	Observation string `json:"observation"`
}

type QuoteConversionInput struct {
	WithdrawalDates map[string]*time.Time `json:"withdrawal_dates"`
}
