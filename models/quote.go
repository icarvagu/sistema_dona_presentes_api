package models

import (
	"encoding/json"
	"time"
)

// QuoteItem represents a single line item in a quotation, including cost
// breakdown, price formation data, and profit/margin calculations.
type QuoteItem struct {
	ID                  int      `json:"id"`
	QuoteID             int      `json:"quote_id"`
	ProductID           int      `json:"product_id"`
	Product             *Product `json:"product,omitempty"`
	Quantity            int      `json:"quantity"`
	UnitPrice           float64  `json:"unit_price"`
	TotalPrice          float64  `json:"total_price"`
	PersonalizationType string   `json:"personalization_type,omitempty"`

	DNCode                  string          `json:"dn_code,omitempty"`
	DescriptionSummary      string          `json:"description_summary,omitempty"`
	IsKit                   bool            `json:"is_kit"`
	BaseCostUnit            float64         `json:"base_cost_unit"`
	LaborCost               float64         `json:"labor_cost"`
	ExtraUnitCost1          float64         `json:"extra_unit_cost1"`
	ExtraUnitCost2          float64         `json:"extra_unit_cost2"`
	EngravingCost           float64         `json:"engraving_cost"`
	UrgencyFee              float64         `json:"urgency_fee"`
	LogisticsCost           float64         `json:"logistics_cost"`
	FreightCost             float64         `json:"freight_cost"`
	TaxPercent              float64         `json:"tax_percent"`
	StPercent               float64         `json:"st_percent"`
	LossIndexPercent        float64         `json:"loss_index_percent"`
	ImportedLaborPercent    float64         `json:"imported_labor_percent"`
	MgmtCommissionPercent   float64         `json:"mgmt_commission_percent"`
	SellerCommissionPercent float64         `json:"seller_commission_percent"`
	AgencyCommissionPercent float64         `json:"agency_commission_percent"`
	PublicityPercent        float64         `json:"publicity_percent"`
	ScrapIndex              float64         `json:"scrap_index"`
	OverPercent             float64         `json:"over_percent"`
	FinancialFactor         float64         `json:"financial_factor"`
	FinancialPercent        float64         `json:"financial_percent"`
	SaleUnitValue           float64         `json:"sale_unit_value"`
	TransportApart          float64         `json:"transport_apart"`
	ProductionCostCalc      float64         `json:"production_cost_calc"`
	TransportCostCalc       float64         `json:"transport_cost_calc"`
	AdditionalCostsCalc     float64         `json:"additional_costs_calc"`
	ProfitCalc              float64         `json:"profit_calc"`
	MarginPercentCalc       float64         `json:"margin_percent_calc"`
	CostUnitCalc            float64         `json:"cost_unit_calc"`
	Engravings              json.RawMessage `json:"engravings,omitempty"`
	HasPriceFormation       bool            `json:"has_price_formation"`
	Discount                float64         `json:"discount"`
	CreatedAt               time.Time       `json:"created_at"`
	UpdatedAt               time.Time       `json:"updated_at"`
}

// Quote represents a commercial quotation sent to a customer before a sale is confirmed.
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

	// New fields
	QuoteDate           *time.Time      `json:"quote_date,omitempty"`
	CareOf              string          `json:"care_of"`
	SalesChannel        string          `json:"sales_channel"`
	Observations        string          `json:"observations"`
	FeedbackDateTime    *time.Time      `json:"feedback_datetime,omitempty"`
	FeedbackObservation string          `json:"feedback_observation,omitempty"`
	PaymentMethod       string          `json:"payment_method"`
	Installments        int             `json:"installments"`
	InstallmentDates    json.RawMessage `json:"installment_dates,omitempty"`
	CarrierID           *int            `json:"carrier_id,omitempty"`
	Carrier             *Carrier        `json:"carrier,omitempty"`
	FreightValue        float64         `json:"freight_value"`

	// Freight info
	FreightTaxIDSender string          `json:"freight_tax_id_sender"`
	FreightTaxIDOrigin   string          `json:"freight_tax_id_origin"`
	FreightTaxIDDest  string          `json:"freight_tax_id_dest"`
	FreightTaxIDPayer     string          `json:"freight_tax_id_payer"`
	FreightTipoTransporte  string          `json:"freight_tipo_transporte"`
	FreightContato         string          `json:"freight_contato"`
	FreightCidadeOrigem    string          `json:"freight_cidade_origem"`
	FreightCidadeDestino   string          `json:"freight_cidade_destino"`
	FreightMaterial        string          `json:"freight_material"`
	FreightTipoFrete       string          `json:"freight_tipo_frete"`
	FreightProduto         string          `json:"freight_produto"`
	FreightTipoEmbalagem   string          `json:"freight_tipo_embalagem"`
	FreightQuantidade      int             `json:"freight_quantidade"`
	FreightVolumes         json.RawMessage `json:"freight_volumes"`
	FreightValorNota       float64         `json:"freight_valor_nota"`
	FreightPesoReal        float64         `json:"freight_peso_real"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// QuoteInput is the DTO for creating or updating a quotation.
type QuoteInput struct {
	QuoteNumber        string           `json:"quote_number,omitempty"`
	SellerID           int              `json:"seller_id"`
	CustomerID         int              `json:"customer_id"`
	ResponsibleName    string           `json:"responsible_name"`
	QuoteValidUntil    *time.Time       `json:"quote_valid_until,omitempty"`
	ProductionLeadTime string           `json:"production_lead_time"`
	Items              []QuoteItemInput `json:"items,omitempty"`

	// New fields
	QuoteDate           *time.Time      `json:"quote_date,omitempty"`
	CareOf              string          `json:"care_of"`
	SalesChannel        string          `json:"sales_channel"`
	Observations        string          `json:"observations"`
	FeedbackDateTime    *time.Time      `json:"feedback_datetime,omitempty"`
	FeedbackObservation string          `json:"feedback_observation,omitempty"`
	PaymentMethod       string          `json:"payment_method"`
	Installments        int             `json:"installments"`
	InstallmentDates    json.RawMessage `json:"installment_dates,omitempty"`
	CarrierID           *int            `json:"carrier_id,omitempty"`
	FreightValue        float64         `json:"freight_value"`

	// Freight info
	FreightTaxIDSender string          `json:"freight_tax_id_sender"`
	FreightTaxIDOrigin   string          `json:"freight_tax_id_origin"`
	FreightTaxIDDest  string          `json:"freight_tax_id_dest"`
	FreightTaxIDPayer     string          `json:"freight_tax_id_payer"`
	FreightTipoTransporte  string          `json:"freight_tipo_transporte"`
	FreightContato         string          `json:"freight_contato"`
	FreightCidadeOrigem    string          `json:"freight_cidade_origem"`
	FreightCidadeDestino   string          `json:"freight_cidade_destino"`
	FreightMaterial        string          `json:"freight_material"`
	FreightTipoFrete       string          `json:"freight_tipo_frete"`
	FreightProduto         string          `json:"freight_produto"`
	FreightTipoEmbalagem   string          `json:"freight_tipo_embalagem"`
	FreightQuantidade      int             `json:"freight_quantidade"`
	FreightVolumes         json.RawMessage `json:"freight_volumes"`
	FreightValorNota       float64         `json:"freight_valor_nota"`
	FreightPesoReal        float64         `json:"freight_peso_real"`
}

// QuoteItemInput is the DTO for a single item line within a quotation request.
type QuoteItemInput struct {
	ProductID           int     `json:"product_id"`
	Quantity            int     `json:"quantity"`
	UnitPrice           float64 `json:"unit_price"`
	PersonalizationType string  `json:"personalization_type,omitempty"`

	DNCode                  string          `json:"dn_code,omitempty"`
	DescriptionSummary      string          `json:"description_summary,omitempty"`
	IsKit                   bool            `json:"is_kit"`
	BaseCostUnit            float64         `json:"base_cost_unit"`
	LaborCost               float64         `json:"labor_cost"`
	ExtraUnitCost1          float64         `json:"extra_unit_cost1"`
	ExtraUnitCost2          float64         `json:"extra_unit_cost2"`
	EngravingCost           float64         `json:"engraving_cost"`
	UrgencyFee              float64         `json:"urgency_fee"`
	LogisticsCost           float64         `json:"logistics_cost"`
	FreightCost             float64         `json:"freight_cost"`
	TaxPercent              float64         `json:"tax_percent"`
	StPercent               float64         `json:"st_percent"`
	LossIndexPercent        float64         `json:"loss_index_percent"`
	ImportedLaborPercent    float64         `json:"imported_labor_percent"`
	MgmtCommissionPercent   float64         `json:"mgmt_commission_percent"`
	SellerCommissionPercent float64         `json:"seller_commission_percent"`
	AgencyCommissionPercent float64         `json:"agency_commission_percent"`
	PublicityPercent        float64         `json:"publicity_percent"`
	ScrapIndex              float64         `json:"scrap_index"`
	OverPercent             float64         `json:"over_percent"`
	FinancialFactor         float64         `json:"financial_factor"`
	FinancialPercent        float64         `json:"financial_percent"`
	SaleUnitValue           float64         `json:"sale_unit_value"`
	TransportApart          float64         `json:"transport_apart"`
	ProductionCostCalc      float64         `json:"production_cost_calc"`
	TransportCostCalc       float64         `json:"transport_cost_calc"`
	AdditionalCostsCalc     float64         `json:"additional_costs_calc"`
	ProfitCalc              float64         `json:"profit_calc"`
	MarginPercentCalc       float64         `json:"margin_percent_calc"`
	CostUnitCalc            float64         `json:"cost_unit_calc"`
	Engravings              json.RawMessage `json:"engravings,omitempty"`
	HasPriceFormation       bool            `json:"has_price_formation"`
	Discount                float64         `json:"discount"`
}
