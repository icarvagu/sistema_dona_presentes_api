package models

import "time"

type Product struct {
	ID                int           `json:"id"`
	ProductName       string        `json:"product_name"`
	InternalCode      string        `json:"internal_code"`
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
	KitType           string        `json:"kit_type"` // none | internal_composition | supplier_ready
	IsComposition     bool          `json:"is_composition"`           // true = produto é composição (tem filhos)
	Items             []ProductItem `json:"items,omitempty"`          // produtos filhos (quando is_composition = true)
	Source            string        `json:"source,omitempty"`         // origem do cadastro, ex: "xbz" ou vazio/manual
	ImportedAt        *time.Time    `json:"imported_at,omitempty"`    // data da primeira importação (se houver)
	LastSyncedAt      *time.Time    `json:"last_synced_at,omitempty"` // última sincronização com fonte externa
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}
