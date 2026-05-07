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
    ID                     int                 `json:"id"`
    CustomerType           string              `json:"customer_type"`
    Status                 string              `json:"status"`
    Name                   string              `json:"name"`
    FantasyName            string              `json:"fantasy_name,omitempty"`
    RazaoSocial            string              `json:"razao_social,omitempty"`
    InscricaoEstadual      string              `json:"inscricao_estadual,omitempty"`
    InscricaoMunicipal     string              `json:"inscricao_municipal,omitempty"`
    Responsavel            string              `json:"responsavel,omitempty"`
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
