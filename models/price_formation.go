package models

import "time"

type ProductPriceFormation struct {
	ProductID            int       `json:"product_id"`
	TaxesPercent         float64   `json:"taxes_percent"`
	OverheadPercent      float64   `json:"overhead_percent"`
	CommissionPercent    float64   `json:"commission_percent"`
	DesiredMarginPercent float64   `json:"desired_margin_percent"`
	SuggestedSelling     float64   `json:"suggested_selling_price"`
	FinalSelling         float64   `json:"final_selling_price"`
	Notes                string    `json:"notes,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type FinancialReportItem struct {
	ProductID        int     `json:"product_id"`
	ProductName      string  `json:"product_name"`
	InternalCode     string  `json:"internal_code"`
	KitType          string  `json:"kit_type"`
	CostPrice        float64 `json:"cost_price"`
	SellingPrice     float64 `json:"selling_price"`
	MarginValue      float64 `json:"margin_value"`
	MarginPercent    float64 `json:"margin_percent"`
	MovesStock       bool    `json:"moves_stock"`
	EnabledForInvoice bool   `json:"enabled_for_invoice"`
}
