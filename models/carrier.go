package models

import "time"

type Carrier struct {
    ID              int        `json:"id"`
    Name            string     `json:"name"`
    CNPJ            *string    `json:"cnpj,omitempty"`
    CarrierType     string     `json:"carrier_type"`
    Email           *string    `json:"email,omitempty"`
    LandlinePhone   *string    `json:"landline_phone,omitempty"`
    MobilePhone     *string    `json:"mobile_phone,omitempty"`
    FullAddress     *string    `json:"full_address,omitempty"`
    ContactName     *string    `json:"contact_name,omitempty"`
    ContactPhone    *string    `json:"contact_phone,omitempty"`
    Website         *string    `json:"website,omitempty"`
    CreatedAt       time.Time  `json:"created_at"`
    UpdatedAt       time.Time  `json:"updated_at"`
}
