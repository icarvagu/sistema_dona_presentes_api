package models

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
