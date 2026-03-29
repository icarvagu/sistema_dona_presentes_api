package models

import "time"

type QuoteItem struct {
	ID                  int       `json:"id"`
	QuoteID             int       `json:"quote_id"`
	ProductID           int       `json:"product_id"`
	Product             *Product  `json:"product,omitempty"`
	Quantity            int       `json:"quantity"`
	UnitPrice           float64   `json:"unit_price"`
	TotalPrice          float64   `json:"total_price"`
	PersonalizationType string    `json:"personalization_type,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type Quote struct {
	ID                 int         `json:"id"`
	QuoteNumber        string      `json:"quote_number,omitempty"`
	SellerID           int         `json:"seller_id"`
	Seller             *User       `json:"seller,omitempty"`
	CustomerID         int         `json:"customer_id"`
	Customer           *Customer   `json:"customer,omitempty"`
	ResponsibleName    string      `json:"responsible_name"`
	QuoteValidUntil    *time.Time  `json:"quote_valid_until,omitempty"`
	ProductionLeadTime string      `json:"production_lead_time"`
	TotalValue         float64     `json:"total"`
	Items              []QuoteItem `json:"items,omitempty"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
}

type QuoteInput struct {
	QuoteNumber        string           `json:"quote_number,omitempty"`
	SellerID           int              `json:"seller_id"`
	CustomerID         int              `json:"customer_id"`
	ResponsibleName    string           `json:"responsible_name"`
	QuoteValidUntil    *time.Time       `json:"quote_valid_until,omitempty"`
	ProductionLeadTime string           `json:"production_lead_time"`
	Items              []QuoteItemInput `json:"items,omitempty"`
}

type QuoteItemInput struct {
	ProductID           int     `json:"product_id"`
	Quantity            int     `json:"quantity"`
	UnitPrice           float64 `json:"unit_price"`
	PersonalizationType string  `json:"personalization_type,omitempty"`
}