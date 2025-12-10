package models

import "time"

type Address struct {
    ID          int       `json:"id"`
    CustomerID  int       `json:"customer_id"`
    AddressType string    `json:"address_type"`
    AddressLine string    `json:"address"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type AdditionalContact struct {
    ID         int       `json:"id"`
    CustomerID int       `json:"customer_id"`
    Name       string    `json:"name"`
    Email      string    `json:"email,omitempty"`
    Phone      string    `json:"phone,omitempty"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}

type Customer struct {
    ID                int                 `json:"id"`
    CustomerType      string              `json:"customer_type"`
    Status            string              `json:"status"`
    Name              string              `json:"name"`
    CNPJ              *string             `json:"cnpj,omitempty"`
    CPF               *string             `json:"cpf,omitempty"`
    Email             string              `json:"email"`
    BusinessPhone     string              `json:"business_phone,omitempty"`
    MobilePhone       string              `json:"mobile_phone,omitempty"`
    Website           *string             `json:"website,omitempty"`
    Notes             *string             `json:"notes,omitempty"`
    Addresses         []Address           `json:"addresses,omitempty"`
    AdditionalContact []AdditionalContact `json:"additional_contacts,omitempty"`
    CreatedAt         time.Time           `json:"created_at"`
    UpdatedAt         time.Time           `json:"updated_at"`
}
