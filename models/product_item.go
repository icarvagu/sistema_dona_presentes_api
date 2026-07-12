package models

import "time"

// ProductItem links a child product to a parent product composition (kit).
type ProductItem struct {
	ID              int       `json:"id"`
	ProductParentID int       `json:"product_parent_id"`
	ProductID       int       `json:"product_id"`
	Quantity        int       `json:"quantity"`
	Product         *Product  `json:"product,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ProductItemInput is the DTO for adding a child item to a product composition.
type ProductItemInput struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}
