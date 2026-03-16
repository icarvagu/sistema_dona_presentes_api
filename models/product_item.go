package models

import "time"

// ProductItem representa um item dentro de uma composição de produto
// Relaciona um produto pai (com is_composition = true) com produtos filhos
type ProductItem struct {
	ID              int       `json:"id"`
	ProductParentID int       `json:"product_parent_id"` // produto pai (composição)
	ProductID       int       `json:"product_id"`         // produto filho
	Quantity        int       `json:"quantity"`
	Product         *Product  `json:"product,omitempty"`  // produto filho completo
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ProductItemInput representa os dados de entrada para um item de composição
type ProductItemInput struct {
	ProductID int `json:"product_id"` // produto filho
	Quantity  int `json:"quantity"`
}
