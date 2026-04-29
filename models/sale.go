package models

import "time"

type SaleItem struct {
	ID         int       `json:"id"`
	SaleID     int       `json:"sale_id"`
	ProductID  int       `json:"product_id"`
	Product    *Product  `json:"product,omitempty"`
	Quantity   int       `json:"quantity"`
	UnitPrice  float64   `json:"unit_price"`
	TotalPrice float64   `json:"total_price"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Sale struct {
	ID                    int        `json:"id"`
	SellerID              int        `json:"seller_id"`
	Seller                *User      `json:"seller,omitempty"`
	CustomerID            int        `json:"customer_id"`
	Customer              *Customer  `json:"customer,omitempty"`
	PaymentMethod         string     `json:"payment_method"`
	Installments          int        `json:"installments"`
	PaymentTermDays       int        `json:"payment_term_days,omitempty"`
	FirstInstallmentStart *time.Time `json:"first_installment_start,omitempty"`
	TotalValue            float64    `json:"total"`
	Items                 []SaleItem `json:"items,omitempty"`
	Carriers              []Carrier  `json:"carriers,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type SaleInput struct {
	SellerID              int             `json:"seller_id"`
	CustomerID            int             `json:"customer_id"`
	PaymentMethod         string          `json:"payment_method"`
	Installments          int             `json:"installments"`
	PaymentTermDays       int             `json:"payment_term_days,omitempty"`
	FirstInstallmentStart *time.Time      `json:"first_installment_start,omitempty"`
	Items                 []SaleItemInput `json:"items,omitempty"`
	CarrierIDs            []int           `json:"carrier_ids,omitempty"`
}

type SaleItemInput struct {
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

