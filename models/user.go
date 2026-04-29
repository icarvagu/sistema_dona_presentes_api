package models

import "time"

// User representa um usuário do sistema (também é um vendedor)
type User struct {
	// Campos de autenticação
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // Não serializar senha no JSON
	Role         string    `json:"role"` // "admin" ou "standard"
	
	// Campos de Employee (dados pessoais do vendedor)
	FullName     string     `json:"full_name"`
	CPF          string     `json:"cpf"`
	RG           *string    `json:"rg,omitempty"`
	BirthDate    *time.Time `json:"birth_date,omitempty"`
	Gender       *string    `json:"gender,omitempty"`
	Status       string     `json:"status"` // "Ativo" ou "Inativo"
	ContactEmail *string    `json:"contact_email,omitempty"`
	FullAddress  *string    `json:"full_address,omitempty"`
	ContactPhone *string    `json:"contact_phone,omitempty"`
	Notes        *string    `json:"notes,omitempty"`
	
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// UserInput representa os dados de entrada para criar/atualizar um usuário
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

// LoginInput representa os dados de entrada para login
type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse representa a resposta do login
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
