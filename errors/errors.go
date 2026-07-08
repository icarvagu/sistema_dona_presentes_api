package errors

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Code       int
	Message    string
	Details    string
	LogMessage string
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) StatusCode() int {
	return e.Code
}

var (

	ErrInvalidInput = &AppError{
		Code:    http.StatusBadRequest,
		Message: "Dados de entrada inválidos",
	}

	ErrInvalidID = &AppError{
		Code:    http.StatusBadRequest,
		Message: "ID inválido",
	}

	ErrInvalidEmail = &AppError{
		Code:    http.StatusBadRequest,
		Message: "E-mail inválido",
	}

	ErrInvalidCPF = &AppError{
		Code:    http.StatusBadRequest,
		Message: "CPF inválido (deve ter 11 dígitos)",
	}

	ErrInvalidCNPJ = &AppError{
		Code:    http.StatusBadRequest,
		Message: "CNPJ inválido (deve ter 14 dígitos)",
	}

	ErrMissingField = &AppError{
		Code:    http.StatusBadRequest,
		Message: "Campo obrigatório não fornecido",
	}

	ErrInvalidJSON = &AppError{
		Code:    http.StatusBadRequest,
		Message: "JSON inválido na requisição",
	}

	ErrNotFound = &AppError{
		Code:    http.StatusNotFound,
		Message: "Recurso não encontrado",
	}

	ErrCustomerNotFound = &AppError{
		Code:    http.StatusNotFound,
		Message: "Cliente não encontrado",
	}

	ErrSupplierNotFound = &AppError{
		Code:    http.StatusNotFound,
		Message: "Supplier not found",
	}

	ErrProductNotFound = &AppError{
		Code:    http.StatusNotFound,
		Message: "Produto não encontrado",
	}

	ErrCarrierNotFound = &AppError{
		Code:    http.StatusNotFound,
		Message: "Carrier not found",
	}

	ErrSaleNotFound = &AppError{
		Code:    http.StatusNotFound,
		Message: "Venda não encontrada",
	}

	ErrQuoteNotFound = &AppError{
		Code:    http.StatusNotFound,
		Message: "Orçamento não encontrado",
	}

	ErrDuplicateEntry = &AppError{
		Code:    http.StatusConflict,
		Message: "Registro duplicado",
	}

	ErrInternalServer = &AppError{
		Code:    http.StatusInternalServerError,
		Message: "Erro interno do servidor",
	}

	ErrDatabaseError = &AppError{
		Code:    http.StatusInternalServerError,
		Message: "Erro ao acessar banco de dados",
	}
)

func NewValidationError(field string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("Validação falhou para: %s", field),
	}
}

func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("%s não encontrado", resource),
	}
}

func NewInvalidFieldError(field, reason string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("Campo '%s' inválido: %s", field, reason),
	}
}

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

func NewDatabaseError(err error) *AppError {
	return &AppError{
		Code:       http.StatusInternalServerError,
		Message:    "Erro ao acessar banco de dados",
		LogMessage: err.Error(),
	}
}

func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:    http.StatusUnauthorized,
		Message: message,
	}
}

func NewForbiddenError(message string) *AppError {
	return &AppError{
		Code:    http.StatusForbidden,
		Message: message,
	}
}

func IsNotFound(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == http.StatusNotFound
	}
	return false
}

func IsValidationError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == http.StatusBadRequest
	}
	return false
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	Code    int    `json:"code"`
}
