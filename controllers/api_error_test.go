package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "donapresentes/errors"
	"donapresentes/middleware"
)

func TestMiddlewareErrorHandlerReturnsJSON(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		middleware.ErrorHandler(w, apperrors.NewUnauthorizedError("test error"), http.StatusUnauthorized)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	var resp apperrors.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("expected JSON error response, got: %v", err)
	}
	if resp.Code != 401 {
		t.Errorf("expected code 401, got %d", resp.Code)
	}
	if resp.Error != "test error" {
		t.Errorf("expected 'test error', got '%s'", resp.Error)
	}
}

func TestErrorHandlerHidesInternalDetails(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		statusCode int
		wantMsg    string
	}{
		{
			name:       "AppError usa Message",
			err:        apperrors.NewNotFoundError("Cliente"),
			statusCode: 404,
			wantMsg:    "Cliente não encontrado",
		},
		{
			name:       "erro generico nao vaza detalhes internos",
			err:        errors.New("sql: column x does not exist"),
			statusCode: 500,
			wantMsg:    "Erro interno do servidor",
		},
		{
			name:       "AppError de validacao mantem mensagem",
			err:        apperrors.NewValidationError("Header inválido"),
			statusCode: 400,
			wantMsg:    "Validação falhou para: Header inválido",
		},
		{
			name:       "erro de autenticacao mantem mensagem",
			err:        apperrors.NewUnauthorizedError("Token inválido"),
			statusCode: 401,
			wantMsg:    "Token inválido",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			middleware.ErrorHandler(rec, tt.err, tt.statusCode)

			var resp apperrors.ErrorResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("expected JSON, got: %v", err)
			}
			if resp.Error != tt.wantMsg {
				t.Errorf("expected '%s', got '%s'", tt.wantMsg, resp.Error)
			}
		})
	}
}

func TestErrorResponseAlwaysHasCode(t *testing.T) {
	errs := []error{
		apperrors.NewNotFoundError("Teste"),
		apperrors.NewValidationError("teste"),
		apperrors.NewUnauthorizedError("teste"),
		apperrors.NewDatabaseError(errors.New("internal db error")),
		errors.New("generic raw error"),
	}

	for _, err := range errs {
		rec := httptest.NewRecorder()
		middleware.ErrorHandler(rec, err, http.StatusBadRequest)

		var resp apperrors.ErrorResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("expected JSON, got: %v for error %v", err, err)
		}
		if resp.Code == 0 {
			t.Errorf("code is 0 for error: %v", err)
		}
		if resp.Error == "" {
			t.Errorf("error message is empty for: %v", err)
		}
	}
}
