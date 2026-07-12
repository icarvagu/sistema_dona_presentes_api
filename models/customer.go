// Package models defines the data structures used throughout the Dona Presentes API.
//
// It includes entity definitions, input/output DTOs, and shared types for the business management system.
package models

import "time"

// Address represents a physical address associated with a customer.
type Address struct {
    ID          int       `json:"id"`
    CustomerID  int       `json:"customer_id"`
    AddressType string    `json:"address_type"`
    AddressLine string    `json:"address"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// AdditionalContact represents an extra point of contact linked to a customer,
// such as an assistant, department head, or secondary buyer.
type AdditionalContact struct {
    ID         int       `json:"id"`
    CustomerID int       `json:"customer_id"`
    Name       string    `json:"name"`
    Email      string    `json:"email,omitempty"`
    Phone      string    `json:"phone,omitempty"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}

// Customer represents a client registered in the system, either individual (PF) or company (PJ).
type Customer struct {
	ID                     int                 `json:"id"`
	CustomerType           string              `json:"customer_type"` // "PF" (individual) or "PJ" (company)
    Status                 string              `json:"status"`
    Name                   string              `json:"name"`
    TradeName            string              `json:"trade_name,omitempty"`
    CompanyName            string              `json:"company_name,omitempty"`
    StateRegistration      string              `json:"state_registration,omitempty"`
    CityRegistration     string              `json:"city_registration,omitempty"`
    Responsible            string              `json:"responsible,omitempty"`
    ContactFinancialName   string              `json:"contact_financial_name,omitempty"`
    ContactFinancialEmail  string              `json:"contact_financial_email,omitempty"`
    ContactFinancialPhone  string              `json:"contact_financial_phone,omitempty"`
    ContactNFName          string              `json:"contact_nf_name,omitempty"`
    ContactNFEmail         string              `json:"contact_nf_email,omitempty"`
    ContactNFPhone         string              `json:"contact_nf_phone,omitempty"`
    ContactCommercialName  string              `json:"contact_commercial_name,omitempty"`
    ContactCommercialEmail string              `json:"contact_commercial_email,omitempty"`
    ContactCommercialPhone string              `json:"contact_commercial_phone,omitempty"`
    CNPJ                   *string             `json:"cnpj,omitempty"`
    CPF                    *string             `json:"cpf,omitempty"`
    Email                  string              `json:"email"`
    BusinessPhone          string              `json:"business_phone,omitempty"`
    MobilePhone            string              `json:"mobile_phone,omitempty"`
    Website                *string             `json:"website,omitempty"`
    Notes                  *string             `json:"notes,omitempty"`
    Addresses              []Address           `json:"addresses,omitempty"`
    AdditionalContact      []AdditionalContact `json:"additional_contacts,omitempty"`
    CreatedAt              time.Time           `json:"created_at"`
    UpdatedAt              time.Time           `json:"updated_at"`
}
