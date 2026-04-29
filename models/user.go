package models

import "time"

type User struct {

	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	

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
	UpdatedAt    time.Time  `json:"updated_at"`
}

type UserInput struct {
	Username     string     `json:"username"`
	Password     string     `json:"password"`
	Role         string     `json:"role"`
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
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
