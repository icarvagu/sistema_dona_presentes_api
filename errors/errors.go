package errors

import (
	"fmt"
	"net/http"
)

// AppError representa um erro da aplicação com contexto HTTP
type AppError struct {
	Code       int
	Message    string
	Details    string
	LogMessage string
}

// Error implementa a interface error
func (e *AppError) Error() string {
	return e.Message
}

// StatusCode retorna o código HTTP apropriado
func (e *AppError) StatusCode() int {
	return e.Code
}

// Mensagens de erro predefinidas
var (
	// Erros de validação
	ErrInvalidInput = &AppError{
		Code:    http.StatusBadRequest,
		Message: "Invalid input data",
	}

	ErrInvalidID = &AppError{
		Code:    http.StatusBadRequest,
		Message: "Invalid ID",
	}

	ErrInvalidEmail = &AppError{
		Code:    http.StatusBadRequest,
		Message: "Invalid email",
	}

	ErrInvalidCPF = &AppError{
		Code:    http.StatusBadRequest,
		Message: "Invalid CPF (must have 11 digits)",
	}

	ErrInvalidCNPJ = &AppError{
		Code:    http.StatusBadRequest,
		Message: "Invalid CNPJ (must have 14 digits)",
	}

	ErrMissingField = &AppError{
		Code:    http.StatusBadRequest,
		Message: "Required field missing",
	}

	ErrInvalidJSON = &AppError{
		Code:    http.StatusBadRequest,
		Message: "Invalid JSON in request",
	}

	// Erros de recurso não encontrado
	ErrNotFound = &AppError{
		Code:    http.StatusNotFound,
		Message: "Resource not found",
	}

	ErrCustomerNotFound = &AppError{
		Code:    http.StatusNotFound,
		Message: "Customer not found",
	}

	ErrSupplierNotFound = &AppError{
		Code:    http.StatusNotFound,
		Message: "Supplier not found",
	}

	ErrEmployeeNotFound = &AppError{
		Code:    http.StatusNotFound,
		Message: "Employee not found",
	}

	ErrProductNotFound = &AppError{
		Code:    http.StatusNotFound,
		Message: "Product not found",
	}

	ErrCarrierNotFound = &AppError{
		Code:    http.StatusNotFound,
		Message: "Carrier not found",
	}

	ErrSaleNotFound = &AppError{
		Code:    http.StatusNotFound,
		Message: "Sale not found",
	}

	// Erros de conflito
	ErrDuplicateEntry = &AppError{
		Code:    http.StatusConflict,
		Message: "Registro duplicado",
	}

	// Erros de servidor
	ErrInternalServer = &AppError{
		Code:    http.StatusInternalServerError,
		Message: "Erro interno do servidor",
	}

	ErrDatabaseError = &AppError{
		Code:    http.StatusInternalServerError,
		Message: "Erro ao acessar banco de dados",
	}
)

// NewValidationError cria um erro de validação customizado
func NewValidationError(field string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("Validação falhou para: %s", field),
	}
}

// NewNotFoundError cria um erro de recurso não encontrado
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("%s não encontrado", resource),
	}
}

// NewInvalidFieldError cria um erro para campo inválido
func NewInvalidFieldError(field, reason string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: fmt.Sprintf("Campo '%s' inválido: %s", field, reason),
	}
}

// NewMissingFieldError cria um erro para campo obrigatório
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

// NewDatabaseError cria um erro de banco de dados
func NewDatabaseError(err error) *AppError {
	return &AppError{
		Code:       http.StatusInternalServerError,
		Message:    "Erro ao acessar banco de dados",
		LogMessage: err.Error(),
	}
}

// IsNotFound verifica se é erro de não encontrado
func IsNotFound(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == http.StatusNotFound
	}
	return false
}

// IsValidationError verifica se é erro de validação
func IsValidationError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == http.StatusBadRequest
	}
	return false
}

// ErrorResponse representa a resposta JSON de erro
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	Code    int    `json:"code"`
}
