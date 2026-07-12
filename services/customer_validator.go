package services

import (
	apperrors "donapresentes/errors"
	"regexp"
)

// ValidateEmail checks that the email address matches a standard pattern
// (local@domain.tld). Empty emails are allowed (returns nil).
func ValidateEmail(email string) error {
	if email == "" {
		return nil
	}
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	if !regexp.MustCompile(pattern).MatchString(email) {
		return apperrors.ErrInvalidEmail
	}
	return nil
}

// ValidateCPF checks that the CPF string contains exactly 11 digits after
// stripping non-digit characters. Empty CPFs are allowed (returns nil).
func ValidateCPF(cpf string) error {
	if cpf == "" {
		return nil
	}
	cpf = regexp.MustCompile(`\D`).ReplaceAllString(cpf, "")
	if len(cpf) != 11 {
		return apperrors.ErrInvalidCPF
	}
	return nil
}

// ValidateCNPJ checks that the CNPJ string contains exactly 14 digits after
// stripping non-digit characters. Empty CNPJs are allowed (returns nil).
func ValidateCNPJ(cnpj string) error {
	if cnpj == "" {
		return nil
	}
	cnpj = regexp.MustCompile(`\D`).ReplaceAllString(cnpj, "")
	if len(cnpj) != 14 {
		return apperrors.ErrInvalidCNPJ
	}
	return nil
}
