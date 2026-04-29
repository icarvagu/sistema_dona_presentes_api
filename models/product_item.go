package models

import "time"

type ProductItem struct {
	ID              int       `json:"id"`
	ProductParentID int       `json:"product_parent_id"`
	ProductID       int       `json:"product_id"`
	Quantity        int       `json:"quantity"`
	Product         *Product  `json:"product,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ProductItemInput struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}
