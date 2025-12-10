package models

import "time"

type Product struct {
    ID             int       `json:"id"`
    ProductName    string    `json:"product_name"`
    InternalCode   string    `json:"internal_code"`
    SupplierID     int       `json:"supplier_id"`
    Supplier       *Supplier `json:"supplier,omitempty"`
    ProductGroup   string    `json:"product_group,omitempty"`
    Description    string    `json:"description,omitempty"`
    Photos         []string  `json:"photos,omitempty"`
    NCM            string    `json:"ncm,omitempty"`
    MaterialOrigin string    `json:"material_origin,omitempty"`
    Stock          int       `json:"stock"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`
}
