package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apperrors "donapresentes/errors"
)

func assertErrorResponse(t *testing.T, body []byte, expectedCode int, expectedMsg string) {
	t.Helper()
	var resp apperrors.ErrorResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("resposta nao e JSON valido: %v (body: %s)", err, string(body))
	}
	if resp.Code != expectedCode {
		t.Errorf("code: esperado %d, got %d", expectedCode, resp.Code)
	}
	if resp.Error == "" {
		t.Error("error message vazia")
	}
	if expectedMsg != "" && !strings.Contains(resp.Error, expectedMsg) {
		t.Errorf("message: esperado contendo '%s', got '%s'", expectedMsg, resp.Error)
	}
}

func TestAllErrorPathsAreJSON(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		statusCode  int
		wantCode    int
		wantMsgHint string
	}{
		{"UnauthorizedError", apperrors.NewUnauthorizedError("Token invalido"), http.StatusUnauthorized, 401, "Token invalido"},
		{"ValidationError", apperrors.NewValidationError("campo x"), http.StatusBadRequest, 400, "campo x"},
		{"NotFoundError", apperrors.NewNotFoundError("Cliente"), http.StatusNotFound, 404, "Cliente"},
		{"ForbiddenError", apperrors.NewForbiddenError("Acesso negado"), http.StatusForbidden, 403, "Acesso negado"},
		{"DatabaseError", apperrors.NewDatabaseError(errors.New("connection refused")), http.StatusInternalServerError, 500, "Erro ao acessar banco de dados"},
		{"DuplicateEntry", apperrors.ErrDuplicateEntry, http.StatusConflict, 409, "Registro duplicado"},
		{"InvalidJSON", apperrors.ErrInvalidJSON, http.StatusBadRequest, 400, "JSON inválido"},
		{"InvalidID", apperrors.ErrInvalidID, http.StatusBadRequest, 400, "ID inválido"},
		{"MissingField", apperrors.NewMissingFieldError("email"), http.StatusBadRequest, 400, "email"},
		{"InternalServer", apperrors.ErrInternalServer, http.StatusInternalServerError, 500, "Erro interno"},
		{"GenericError", errors.New("internal: column x not found"), 500, 500, "Erro interno do servidor"},
		{"GenericErrorBadRequest", errors.New("some random error"), 400, 400, "Erro interno do servidor"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
				rec := httptest.NewRecorder()
			ErrorHandler(rec, tt.err, tt.statusCode)

			assertErrorResponse(t, rec.Body.Bytes(), tt.wantCode, tt.wantMsgHint)
			if rec.Code != tt.wantCode {
				t.Errorf("HTTP status: esperado %d, got %d", tt.wantCode, rec.Code)
			}
		})
	}
}

func TestGenericErrorDoesNotLeakInternalDetails(t *testing.T) {
	internalMsg := "column 'x' does not exist | stack: /usr/local/go/src/database/sql/... | password=secret123"
	err := errors.New(internalMsg)

	rec := httptest.NewRecorder()
	ErrorHandler(rec, err, 500)

	var resp apperrors.ErrorResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Error == internalMsg {
		t.Errorf("vazou detalhe interno na resposta: %s", resp.Error)
	}
	if resp.Error != "Erro interno do servidor" {
		t.Errorf("esperado 'Erro interno do servidor', got '%s'", resp.Error)
	}
}

func TestAllErrorsReturnValidJSON(t *testing.T) {
	errs := []struct {
		err  error
		name string
	}{
		{apperrors.NewUnauthorizedError("x"), "unauthorized"},
		{apperrors.NewValidationError("x"), "validation"},
		{apperrors.NewNotFoundError("x"), "notfound"},
		{apperrors.NewForbiddenError("x"), "forbidden"},
		{apperrors.NewDatabaseError(errors.New("x")), "database"},
		{apperrors.NewMissingFieldError("x"), "missingfield"},
		{apperrors.ErrDuplicateEntry, "duplicate"},
		{apperrors.ErrInvalidID, "invalidid"},
		{apperrors.ErrInvalidJSON, "invalidjson"},
		{errors.New("raw error"), "raw"},
	}
	for _, e := range errs {
		t.Run(e.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			ErrorHandler(rec, e.err, 400)
			var resp apperrors.ErrorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("JSON invalido para erro %s: %v", e.name, err)
			}
			if resp.Error == "" {
				t.Errorf("error message vazia para %s", e.name)
			}
			if resp.Code == 0 {
				t.Errorf("code zero para %s", e.name)
			}
		})
	}
}
