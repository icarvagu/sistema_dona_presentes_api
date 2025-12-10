package models

import "time"

type Employee struct {
    ID           int        `json:"id"`
    FullName     string     `json:"full_name"`
    CPF          string     `json:"cpf"`
    RG           *string    `json:"rg,omitempty"`
    BirthDate    *time.Time `json:"birth_date,omitempty"`
    Gender       *string    `json:"gender,omitempty"`
    Status       string     `json:"status"`
    ContactEmail *string    `json:"contact_email,omitempty"`
    FullAddress  *string    `json:"full_address,omitempty"`
    ContactPhone *string    `json:"contact_phone,omitempty"`
    Notes        *string    `json:"notes,omitempty"`
    CreatedAt    time.Time  `json:"created_at"`
}
