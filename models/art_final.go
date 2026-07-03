package models

import "time"

type ArtFinalAccess struct {
	UserID            int
	Admin             bool
	ArtFinal          bool
	Sales             bool
	Purchases         bool
	ProductionProfile bool
	Manage            bool
	Layout            bool
	Corel             bool
	Engraving         bool
	Media             bool
	Production        bool
	Stories           bool
	Marketing         bool
}

type LayoutRequestItemInput struct {
	EntityType    string `json:"entity_type"`
	ItemID        int    `json:"item_id"`
	GroupKey      string `json:"group_key"`
	SourceFileURL string `json:"source_file_url"`
}

type LayoutRequestInput struct {
	SourceType    string                   `json:"source_type"`
	SourceID      int                      `json:"source_id"`
	Title         string                   `json:"title"`
	Instructions  string                   `json:"instructions"`
	FileMode      string                   `json:"file_mode"`
	CommonFileURL string                   `json:"common_file_url"`
	Priority      int                      `json:"priority"`
	DueAt         *time.Time               `json:"due_at,omitempty"`
	AssignedTo    *int                     `json:"assigned_to,omitempty"`
	Items         []LayoutRequestItemInput `json:"items"`
}

type LayoutTransitionInput struct {
	Status string `json:"status"`
	Note   string `json:"note"`
}

type LayoutMessageInput struct {
	Message string `json:"message"`
	FileURL string `json:"file_url"`
}

type LayoutVersionInput struct {
	FileURL string `json:"file_url"`
	Label   string `json:"label"`
}

type LayoutDecisionInput struct {
	Status string `json:"status"`
	Note   string `json:"note"`
}

type LayoutJobInput struct {
	Status          string     `json:"status"`
	ResponsibleID   *int       `json:"responsible_id,omitempty"`
	ExternalContact string     `json:"external_contact"`
	DueAt           *time.Time `json:"due_at,omitempty"`
	FileURL         string     `json:"file_url"`
	Tags            []string   `json:"tags"`
	Reason          string     `json:"reason"`
}

type StoryLifecycleInput struct {
	Status      string     `json:"status"`
	Observation string     `json:"observation"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type ArtFinalFilters struct {
	Category string
	Status   string
	Search   string
	DateFrom *time.Time
	DateTo   *time.Time
}

type ArtFinalTaskInput struct {
	Panel         string     `json:"panel"`
	Category      string     `json:"category"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	SaleID        *int       `json:"sale_id,omitempty"`
	PurchaseID    *int       `json:"purchase_id,omitempty"`
	DueDate       *time.Time `json:"due_date,omitempty"`
	DueAt         *time.Time `json:"due_at,omitempty"`
	AssignedTo    *int       `json:"assigned_to,omitempty"`
	Channel       string     `json:"channel"`
	Format        string     `json:"format"`
	Priority      int        `json:"priority"`
	Tags          []string   `json:"tags"`
	AttachmentURL string     `json:"attachment_url"`
	Status        string     `json:"status"`
}

type ArtFinalStoryInput struct {
	SaleID     int      `json:"sale_id"`
	SaleItemID *int     `json:"sale_item_id,omitempty"`
	Title      string   `json:"title"`
	Tags       []string `json:"tags"`
}
