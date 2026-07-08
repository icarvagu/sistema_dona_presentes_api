package models

import "time"

type Supplier struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	CNPJ              *string   `json:"cnpj,omitempty"`
	StateRegistration *string   `json:"state_registration,omitempty"`
	ContactPerson     *string   `json:"contact_person,omitempty"`
	Email             *string   `json:"email,omitempty"`
	LandlinePhone     *string   `json:"landline_phone,omitempty"`
	MobilePhone       *string   `json:"mobile_phone,omitempty"`
	ResponsibleEmail  *string   `json:"responsible_email,omitempty"`
	CommercialAddress *string   `json:"commercial_address,omitempty"`
	PostalCode        *string   `json:"postal_code,omitempty"`
	Website           *string   `json:"website,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
