package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func workflowRequest(role string, permissions ...string) *http.Request {
	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), "role", role)
	ctx = context.WithValue(ctx, "permissions", permissions)
	return req.WithContext(ctx)
}

func TestCanManageSalePending(t *testing.T) {
	allowed := []string{"compras", "producao", "financeiro", "diretoria_financeira", "arte_final"}
	for _, permission := range allowed {
		if !canManageSalePending(workflowRequest("standard", permission)) {
			t.Errorf("permission %q should manage sale pending events", permission)
		}
	}
	if !canManageSalePending(workflowRequest("admin")) {
		t.Error("admin should manage sale pending events")
	}
	if canManageSalePending(workflowRequest("standard", "produtos")) {
		t.Error("unrelated permission should not manage sale pending events")
	}
}
