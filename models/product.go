package models

import "time"

type Product struct {
	ID                int           `json:"id"`
	ProductName       string        `json:"product_name"`
	InternalCode      string        `json:"internal_code"`
	SupplierCode      string        `json:"supplier_code"`
	SupplierID        int           `json:"supplier_id"`
	Supplier          *Supplier     `json:"supplier,omitempty"`
	ProductGroup      string        `json:"product_group,omitempty"`
	Description       string        `json:"description,omitempty"`
	Photos            []string      `json:"photos"`
	NCM               string        `json:"ncm,omitempty"`
	MaterialOrigin    string        `json:"material_origin,omitempty"`
	Stock             int           `json:"stock"`
	MovesStock        bool          `json:"moves_stock"`
	EnabledForInvoice bool          `json:"enabled_for_invoice"`
	CostPrice         float64       `json:"cost_price"`
	SellingPrice      float64       `json:"selling_price"`
	KitType           string        `json:"kit_type"`
	IsComposition     bool          `json:"is_composition"`
	Items             []ProductItem `json:"items,omitempty"`
	Source            string        `json:"source,omitempty"`
	ImportedAt        *time.Time    `json:"imported_at,omitempty"`
	LastSyncedAt      *time.Time    `json:"last_synced_at,omitempty"`
	Color             string        `json:"color"`
	Origin            string        `json:"origin"`
	PendingApproval   bool          `json:"pending_approval"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`

	// Campos para produtos sem API
	LastCost          float64       `json:"last_cost"`
	LastCostDate      *time.Time    `json:"last_cost_date"`
	LastCostQty1      int           `json:"last_cost_qty1"`
	LastCostQty2      int           `json:"last_cost_qty2"`
	LastCostQty3      int           `json:"last_cost_qty3"`
	LastCostUser      string        `json:"last_cost_user"`
}
