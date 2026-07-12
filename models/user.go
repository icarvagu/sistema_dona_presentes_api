package models

import "time"

// User represents a system user with authentication and personal data.
type User struct {

	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	Permissions  []string  `json:"permissions"`

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

	FailedLoginAttempts int        `json:"-"`
	LockedUntil         *time.Time `json:"locked_until,omitempty"`
	MustChangePassword  bool       `json:"must_change_password,omitempty"`

	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// UserInput is the DTO for creating or updating a user.
type UserInput struct {
	Username            string     `json:"username"`
	Password            string     `json:"password"`
	Role                string     `json:"role"`
	Permissions         []string   `json:"permissions"`
	FullName            string     `json:"full_name"`
	CPF                 string     `json:"cpf"`
	RG                  *string    `json:"rg,omitempty"`
	BirthDate           *time.Time `json:"birth_date,omitempty"`
	Gender              *string    `json:"gender,omitempty"`
	Status              string     `json:"status"`
	ContactEmail        *string    `json:"contact_email,omitempty"`
	FullAddress         *string    `json:"full_address,omitempty"`
	ContactPhone        *string    `json:"contact_phone,omitempty"`
	Notes               *string    `json:"notes,omitempty"`
	CPFHash             string     `json:"-"`
	MustChangePassword  bool       `json:"must_change_password,omitempty"`
}

// LoginInput contains the credentials for user authentication.
type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse returns the JWT access/refresh tokens and the authenticated user profile.
type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

// RefreshRequest is used to obtain a new access token from a refresh token.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// ForgotPasswordRequest triggers a password reset email for the given username.
type ForgotPasswordRequest struct {
	Username string `json:"username"`
}

// ResetPasswordRequest carries the password reset token and the new password.
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// LogoutRequest invalidates a refresh token, effectively logging the user out.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}
