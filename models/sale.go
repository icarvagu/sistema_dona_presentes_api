package models

import (
	"encoding/json"
	"time"
)

// SaleItem represents a single line item within a sale, including the product,
// quantity, pricing, and any engraving/personalization details.
type SaleItem struct {
	ID             int             `json:"id"`
	SaleID         int             `json:"sale_id"`
	ProductID      int             `json:"product_id"`
	Product        *Product        `json:"product,omitempty"`
	Quantity       int             `json:"quantity"`
	UnitPrice      float64         `json:"unit_price"`
	TotalPrice     float64         `json:"total_price"`
	Discount       float64         `json:"discount"`
	PriceFormation json.RawMessage `json:"price_formation,omitempty"`
	Engravings     json.RawMessage `json:"engravings,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// Sale represents a confirmed sale order with customer, items, payment,
// delivery, and carrier information.
type Sale struct {
	ID                    int             `json:"id"`
	SellerID              int             `json:"seller_id"`
	Seller                *User           `json:"seller,omitempty"`
	CustomerID            int             `json:"customer_id"`
	Customer              *Customer       `json:"customer,omitempty"`
	PaymentMethod         string          `json:"payment_method"`
	Installments          int             `json:"installments"`
	PaymentTermDays       int             `json:"payment_term_days,omitempty"`
	FirstInstallmentStart *time.Time      `json:"first_installment_start,omitempty"`
	InstallmentDates      json.RawMessage `json:"installment_dates,omitempty"`
	Status                string          `json:"status"`
	IsEvent               bool            `json:"is_event"` // true when the sale is tied to a specific event/occasion
	DeliveryAddress       string          `json:"delivery_address"`
	DeliveryDate          *time.Time      `json:"delivery_date,omitempty"`
	DepartureDate         *time.Time      `json:"departure_date,omitempty"`
	ArrivalDate           *time.Time      `json:"arrival_date,omitempty"`
	Priority              string          `json:"priority"`
	CareOf                string          `json:"care_of"`
	InvoiceEmail               string          `json:"invoice_email"`
	FinancialEmail       string          `json:"financial_email"`
	PurchaseOrder           string          `json:"purchase_order"`
	ExternalNotes   string          `json:"external_notes"`
	InternalNotes   string          `json:"internal_notes"`
	LayoutURLs            []string        `json:"layout_urls,omitempty"`
	TotalValue            float64         `json:"total"`
	Items                 []SaleItem      `json:"items,omitempty"`
	Carriers              []Carrier       `json:"carriers,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

// SaleInput is the DTO for creating or updating a sale.
type SaleInput struct {
	SellerID              int             `json:"seller_id"`
	CustomerID            int             `json:"customer_id"`
	PaymentMethod         string          `json:"payment_method"`
	Installments          int             `json:"installments"`
	PaymentTermDays       int             `json:"payment_term_days,omitempty"`
	FirstInstallmentStart *time.Time      `json:"first_installment_start,omitempty"`
	InstallmentDates      json.RawMessage `json:"installment_dates,omitempty"`
	Status                string          `json:"status"`
	IsEvent               bool            `json:"is_event"`
	DeliveryAddress       string          `json:"delivery_address"`
	DeliveryDate          *time.Time      `json:"delivery_date,omitempty"`
	DepartureDate         *time.Time      `json:"departure_date,omitempty"`
	ArrivalDate           *time.Time      `json:"arrival_date,omitempty"`
	Priority              string          `json:"priority"`
	CareOf                string          `json:"care_of"`
	InvoiceEmail               string          `json:"invoice_email"`
	FinancialEmail       string          `json:"financial_email"`
	PurchaseOrder           string          `json:"purchase_order"`
	ExternalNotes   string          `json:"external_notes"`
	InternalNotes   string          `json:"internal_notes"`
	LayoutURLs            []string        `json:"layout_urls,omitempty"`
	Items                 []SaleItemInput `json:"items,omitempty"`
	CarrierIDs            []int           `json:"carrier_ids,omitempty"`
}

// SaleItemInput is the DTO for a single item line within a sale creation request.
type SaleItemInput struct {
	ProductID  int             `json:"product_id"`
	Quantity   int             `json:"quantity"`
	UnitPrice  float64         `json:"unit_price"`
	Engravings json.RawMessage `json:"engravings,omitempty"`
}
