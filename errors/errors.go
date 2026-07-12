// Package errors provides a unified application error type and a set of
// pre-defined error instances for the Dona Presentes API.
package errors

import (
	"fmt"
	"net/http"
)

// AppError represents an application-level error with an HTTP status code,
// a user-facing message, optional internal details, and an optional log message
// for server-side diagnostics.
type AppError struct {
	Code       int
	Message    string
	Details    string
	LogMessage string
}

// Error implements the error interface by returning the user-facing Message.
func (e *AppError) Error() string {
	return e.Message
}

// StatusCode returns the HTTP status code associated with the error.
func (e *AppError) StatusCode() int {
	return e.Code
}

var (

	// ErrInvalidInput is returned when the request body contains invalid data.
	ErrInvalidInput = &AppError{
		Code:    http.StatusBadRequest,
		Message: "Dados de entrada inválidos",
	}

	// ErrInvalidID is returned when an ID parameter is not a valid number or format.
	ErrInvalidID = &AppError{
		Code:    http.StatusBadRequest,
		Message: "ID inválido",
	}

	// ErrInvalidEmail is returned when the provided email address fails validation.
	ErrInvalidEmail = &AppError{
		Code:    http.StatusBadRequest,
		Message: "E-mail inválido",
	}

	// ErrInvalidCPF is returned when the provided CPF does not match the expected format.
	ErrInvalidCPF = &AppError{
		Code:    http.StatusBadRequest,
		Message: "CPF inválido (deve ter 11 dígitos)",
	}

	// ErrInvalidCNPJ is returned when the provided CNPJ does not match the expected format.
	ErrInvalidCNPJ = &AppError{
		Code:    http.StatusBadRequest,
		Message: "CNPJ inválido (deve ter 14 dígitos)",
	}

	// ErrMissingField is returned when one or more required fields are absent.
	ErrMissingField = &AppError{
		Code:    http.StatusBadRequest,
		Message: "Campo obrigatório não fornecido",
	}

	// ErrInvalidJSON is returned when the request body cannot be parsed as valid JSON.
	ErrInvalidJSON = &AppError{
		Code:    http.StatusBadRequest,
		Message: "JSON inválido na requisição",
	}

	// ErrNotFound is a generic not-found error for any resource.
	ErrNotFound = &AppError{
		Code:    http.StatusNotFound,
		Message: "Recurso não encontrado",
	}

	// ErrCustomerNotFound is returned when a customer record cannot be found.
	ErrCustomerNotFound = &AppError{
		Code:    http.StatusNotFound,
		Message: "Cliente não encontrado",
	}

	// ErrSupplierNotFound is returned when a supplier record cannot be found.
	ErrSupplierNotFound = &AppError{
		Code:    http.StatusNotFound,
		Message: "Supplier not found",
	}

	// ErrProductNotFound is returned when a product record cannot be found.
	ErrProductNotFound = &AppError{
		Code:    http.StatusNotFound,
		Message: "Produto não encontrado",
	}

	// ErrCarrierNotFound is returned when a carrier record cannot be found.
	ErrCarrierNotFound = &AppError{
		Code:    http.StatusNotFound,
		Message: "Carrier not found",
	}

	// ErrSaleNotFound is returned when a sale record cannot be found.
	ErrSaleNotFound = &AppError{
		Code:    http.StatusNotFound,
		Message: "Venda não encontrada",
	}

	// ErrQuoteNotFound is returned when a quote record cannot be found.
	ErrQuoteNotFound = &AppError{
		Code:    http.StatusNotFound,
		Message: "Orçamento não encontrado",
	}

	// ErrDuplicateEntry is returned when attempting to insert a record that
	// violates a uniqueness constraint.
	ErrDuplicateEntry = &AppError{
		Code:    http.StatusConflict,
		Message: "Registro duplicado",
	}

	// ErrInternalServer is a generic internal server error.
	ErrInternalServer = &AppError{
		Code:    http.StatusInternalServerError,
		Message: "Erro interno do servidor",
	}

	// ErrDatabaseError is returned when a database operation fails unexpectedly.
	ErrDatabaseError = &AppError{
		Code:    http.StatusInternalServerError,
		Message: "Erro ao acessar banco de dados",
	}
)

// NewValidationError creates a validation error for the given field name.
func NewValidationError(field string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("Validação falhou para: %s", field),
	}
}

// NewNotFoundError creates a not-found error for the named resource.
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("%s não encontrado", resource),
	}
}

// NewInvalidFieldError creates a validation error with a specific field name and reason.
func NewInvalidFieldError(field, reason string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("Campo '%s' inválido: %s", field, reason),
	}
}

// NewMissingFieldError creates an error listing the required fields that were not provided.
func NewMissingFieldError(fields ...string) *AppError {
	fieldStr := ""
	for i, f := range fields {
		if i > 0 {
			fieldStr += ", "
		}
		fieldStr += f
	}
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("Campos obrigatórios não fornecidos: %s", fieldStr),
	}
}

// NewDatabaseError wraps a database error as an AppError. If the provided error
// is already an AppError with a log message, it is returned as-is; otherwise
// a new internal server error is created with the original error logged.
func NewDatabaseError(err error) *AppError {
	if inner, ok := err.(*AppError); ok {
		if inner.LogMessage != "" {
			return inner
		}
		return &AppError{
			Code:       http.StatusInternalServerError,
			Message:    "Erro ao acessar banco de dados",
			LogMessage: inner.Message,
		}
	}
	return &AppError{
		Code:       http.StatusInternalServerError,
		Message:    "Erro ao acessar banco de dados",
		LogMessage: err.Error(),
	}
}

// NewUnauthorizedError creates an authentication error with the given message.
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:    http.StatusUnauthorized,
		Message: message,
	}
}

// NewForbiddenError creates an authorization error with the given message.
func NewForbiddenError(message string) *AppError {
	return &AppError{
		Code:    http.StatusForbidden,
		Message: message,
	}
}

// IsNotFound reports whether err is an AppError with HTTP status 404.
func IsNotFound(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == http.StatusNotFound
	}
	return false
}

// IsValidationError reports whether err is an AppError with HTTP status 400.
func IsValidationError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == http.StatusBadRequest
	}
	return false
}

// ErrorResponse is the JSON structure returned to API clients when an error occurs.
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	Code    int    `json:"code"`
}
