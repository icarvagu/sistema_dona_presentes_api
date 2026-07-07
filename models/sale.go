package models

import (
	"encoding/json"
	"time"
)

type SaleItem struct {
	ID         int             `json:"id"`
	SaleID     int             `json:"sale_id"`
	ProductID  int             `json:"product_id"`
	Product    *Product        `json:"product,omitempty"`
	Quantity   int             `json:"quantity"`
	UnitPrice  float64         `json:"unit_price"`
	TotalPrice float64         `json:"total_price"`
	Discount   float64         `json:"discount"`
	Engravings json.RawMessage `json:"engravings,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

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
	IsEvent               bool            `json:"is_event"`
	DeliveryAddress       string          `json:"delivery_address"`
	DeliveryDate          *time.Time      `json:"delivery_date,omitempty"`
	DepartureDate         *time.Time      `json:"departure_date,omitempty"`
	ArrivalDate           *time.Time      `json:"arrival_date,omitempty"`
	Priority              string          `json:"priority"`
	CareOf                string          `json:"care_of"`
	EmailNF               string          `json:"email_nf"`
	EmailFinanceiro       string          `json:"email_financeiro"`
	OrdemCompra           string          `json:"ordem_compra"`
	ObservacoesExternas   string          `json:"observacoes_externas"`
	ObservacoesInternas   string          `json:"observacoes_internas"`
	LayoutURLs            []string        `json:"layout_urls,omitempty"`
	TotalValue            float64         `json:"total"`
	Items                 []SaleItem      `json:"items,omitempty"`
	Carriers              []Carrier       `json:"carriers,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

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
	EmailNF               string          `json:"email_nf"`
	EmailFinanceiro       string          `json:"email_financeiro"`
	OrdemCompra           string          `json:"ordem_compra"`
	ObservacoesExternas   string          `json:"observacoes_externas"`
	ObservacoesInternas   string          `json:"observacoes_internas"`
	LayoutURLs            []string        `json:"layout_urls,omitempty"`
	Items                 []SaleItemInput `json:"items,omitempty"`
	CarrierIDs            []int           `json:"carrier_ids,omitempty"`
}

type SaleItemInput struct {
	ProductID  int             `json:"product_id"`
	Quantity   int             `json:"quantity"`
	UnitPrice  float64         `json:"unit_price"`
	Engravings json.RawMessage `json:"engravings,omitempty"`
}
