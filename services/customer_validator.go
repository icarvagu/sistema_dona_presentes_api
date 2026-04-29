package services

import (
	apperrors "donapresentes/errors"
	"regexp"
)

// ValidateEmail valida formato de email
func ValidateEmail(email string) error {
	if email == "" {
		return nil // email é opcional
	}
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	if !regexp.MustCompile(pattern).MatchString(email) {
		return apperrors.ErrInvalidEmail
	}
	return nil
}

// ValidateCPF valida formato de CPF
func ValidateCPF(cpf string) error {
	if cpf == "" {
		return nil // CPF é opcional
	}
	cpf = regexp.MustCompile(`\D`).ReplaceAllString(cpf, "")
	if len(cpf) != 11 {
		return apperrors.ErrInvalidCPF
	}
	return nil
}

// ValidateCNPJ valida formato de CNPJ
func ValidateCNPJ(cnpj string) error {
	if cnpj == "" {
		return nil // CNPJ é opcional
	}
	cnpj = regexp.MustCompile(`\D`).ReplaceAllString(cnpj, "")
	if len(cnpj) != 14 {
		return apperrors.ErrInvalidCNPJ
	}
	return nil
}
