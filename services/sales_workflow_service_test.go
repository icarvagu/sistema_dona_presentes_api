package services

import "testing"

func TestNormalizeFinancialTag(t *testing.T) {
	tests := map[string]string{
		"50%+boleto":  "50% + BOLETO",
		"so a vista":  "SOMENTE À VISTA",
		"negado":      "NEGADO",
		" serasa ok ": "SERASA OK",
	}
	for input, want := range tests {
		if got := normalizeFinancialTag(input); got != want {
			t.Fatalf("normalizeFinancialTag(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestNormalizeEngravingResponse(t *testing.T) {
	tests := map[string]string{
		"approved":          "APROVADO",
		"changes_requested": "NOVA_FOTO",
		"cor_diferente":     "COR_DIFERENTE",
		" errado ":          "ERRADO",
	}
	for input, want := range tests {
		if got := normalizeEngravingResponse(input); got != want {
			t.Fatalf("normalizeEngravingResponse(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestValidateFileURL(t *testing.T) {
	valid := []string{"/uploads/layout.pdf", "/files/comprovante.png", "https://example.com/layout.pdf"}
	for _, input := range valid {
		if err := validateFileURL(input); err != nil {
			t.Errorf("validateFileURL(%q) returned %v", input, err)
		}
	}
	invalid := []string{"", "layout.pdf", "javascript:alert(1)", "ftp://example.com/file"}
	for _, input := range invalid {
		if err := validateFileURL(input); err == nil {
			t.Errorf("validateFileURL(%q) unexpectedly succeeded", input)
		}
	}
}
