package services

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestBuildPurchaseEmailIncludesBodyAndAttachment(t *testing.T) {
	attachment := "data:text/plain;base64," + base64.StdEncoding.EncodeToString([]byte("corel-test"))
	message, err := buildPurchaseEmail("compras@example.com", "fornecedor@example.com", "Pedido 42", "Corpo automático", []string{attachment})
	if err != nil {
		t.Fatalf("buildPurchaseEmail returned error: %v", err)
	}
	content := string(message)
	for _, expected := range []string{"Subject: Pedido 42", "Corpo automático", `filename="anexo-1.plain"`} {
		if !strings.Contains(content, expected) {
			t.Fatalf("expected MIME message to contain %q", expected)
		}
	}
}
