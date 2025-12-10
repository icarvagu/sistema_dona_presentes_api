package models

import "time"

type Carrier struct {
    ID              int       `json:"id"`
    Name            string    `json:"name"`
    CarrierType     string    `json:"carrier_type"`
    Email           string    `json:"email"`
    LandlinePhone   string    `json:"landline_phone"`
    MobilePhone     string    `json:"mobile_phone"`
    FullAddress     string    `json:"full_address"`
    ContactName     string    `json:"contact_name"`
    ContactPhone    string    `json:"contact_phone"`
    Website         string    `json:"website"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}
