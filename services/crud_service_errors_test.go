package services

import (
	"testing"

	apperrors "donapresentes/errors"
)

func TestValidateEmailErrors(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{"sem arroba", "invalido"},
		{"com injection", "a@b.com' OR 1=1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if err == nil {
				t.Fatal("esperava erro de email invalido")
			}
		})
	}
}

func TestValidateEmailEmpty(t *testing.T) {
	err := ValidateEmail("")
	if err != nil {
		t.Fatalf("email vazio (opcional) nao deveria dar erro: %v", err)
	}
}

func TestValidateEmailValid(t *testing.T) {
	emails := []string{"cliente@exemplo.com", "teste@dominio.com.br", "a.b@c.co"}
	for _, e := range emails {
		t.Run(e, func(t *testing.T) {
			if err := ValidateEmail(e); err != nil {
				t.Fatalf("nao esperava erro para %s: %v", e, err)
			}
		})
	}
}

func TestValidateCPF(t *testing.T) {
	tests := []struct {
		cpf     string
		wantErr bool
	}{
		{"", false},  // vazio é válido (opcional)
		{"123", true},
		{"12345678901", false}, // 11 digitos passa (só verifica length)
		{"000.000.000-00", false}, // sanitiza e 11 digitos
	}
	for _, tt := range tests {
		t.Run(tt.cpf, func(t *testing.T) {
			err := ValidateCPF(tt.cpf)
			if tt.wantErr && err == nil {
				t.Fatal("esperava erro")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("nao esperava erro: %v", err)
			}
		})
	}
}

func TestValidateCNPJ(t *testing.T) {
	tests := []struct {
		cnpj    string
		wantErr bool
	}{
		{"", false},  // vazio é válido (opcional)
		{"123", true},
		{"00.000.000/0000-00", false}, // 14 digitos sanitizados
	}
	for _, tt := range tests {
		t.Run(tt.cnpj, func(t *testing.T) {
			err := ValidateCNPJ(tt.cnpj)
			if tt.wantErr && err == nil {
				t.Fatal("esperava erro")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("nao esperava erro: %v", err)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{" normal ", " normal "},
		{"<script>alert(1)</script>", "<script>alert(1)</script>"},
		{"", ""},
		{"  ", "  "},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("esperado '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestSanitizeUsername(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{" admin ", " admin "},
		{"ADMIN", "ADMIN"},
		{"Usuario Teste", "Usuario Teste"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SanitizeUsername(tt.input)
			if result != tt.expected {
				t.Errorf("esperado '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestSanitizeCPF(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"123.456.789-01", "12345678901"},
		{" 111.222.333-44 ", "11122233344"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SanitizeCPF(tt.input)
			if result != tt.expected {
				t.Errorf("esperado '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestSanitizePassword(t *testing.T) {
	s := SanitizePassword("MinhaSenha@123")
	if s != "MinhaSenha@123" {
		t.Errorf("esperado senha original, got '%s'", s)
	}
	if SanitizePassword("") != "" {
		t.Error("senha vazia deveria retornar vazio")
	}
	long := string(make([]byte, 200))
	if len(SanitizePassword(long)) != 128 {
		t.Error("senha longa deveria ser truncada para 128")
	}
}

func TestAppErrorIsNotFoundForGenericError(t *testing.T) {
	if apperrors.IsNotFound(nil) {
		t.Fatal("nil nao deveria ser not found")
	}
}

func TestAppErrorIsValidationForGenericError(t *testing.T) {
	if apperrors.IsValidationError(nil) {
		t.Fatal("nil nao deveria ser validation error")
	}
}
