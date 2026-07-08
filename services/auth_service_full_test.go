package services

import (
	"testing"

	apperrors "donapresentes/errors"
)

func TestPasswordComplexityAllErrors(t *testing.T) {
	s := newTestAuthService()
	tests := []struct {
		name     string
		password string
	}{
		{"muito curta (7)", "Ab1!"},
		{"sem maiuscula", "abcdefgh1@"},
		{"sem minuscula", "ABCDEFGH1@"},
		{"sem digito", "Abcdefgh!@"},
		{"sem especial", "Abcdefgh1"},
		{"vazia", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidatePasswordComplexity(tt.password)
			if err == nil {
				t.Fatal("esperava erro de complexidade")
			}
			if !apperrors.IsValidationError(err) {
				t.Fatalf("esperava ValidationError, got %T: %v", err, err)
			}
		})
	}
}

func TestValidateTokenMalformed(t *testing.T) {
	s := newTestAuthService()
	_, err := s.ValidateToken("not-a-valid-jwt-token-at-all")
	if err == nil {
		t.Fatal("esperava erro para token malformatado")
	}
}

func TestValidateTokenEmpty(t *testing.T) {
	s := newTestAuthService()
	_, err := s.ValidateToken("")
	if err == nil {
		t.Fatal("esperava erro para token vazio")
	}
}

func TestHashPasswordWorks(t *testing.T) {
	s := newTestAuthService()
	hash, err := s.HashPassword("SenhaValida1@")
	if err != nil {
		t.Fatalf("erro ao hash: %v", err)
	}
	if hash == "" {
		t.Fatal("hash nao deveria ser vazio")
	}
}

func TestCheckPasswordEmptyHash(t *testing.T) {
	s := newTestAuthService()
	if s.CheckPassword("senha", "") {
		t.Fatal("hash vazio nao deveria validar")
	}
	if s.CheckPassword("", "$2a$10$hash") {
		t.Fatal("senha vazia nao deveria validar")
	}
}

func TestHashAndCheckPasswordEmpty(t *testing.T) {
	s := newTestAuthService()
	hash, err := s.HashPassword("ab")
	if err != nil {
		t.Fatalf("erro ao hash: %v", err)
	}
	if !s.CheckPassword("ab", hash) {
		t.Fatal("senha correta deveria validar")
	}
	if s.CheckPassword("abc", hash) {
		t.Fatal("senha incorreta nao deveria validar")
	}
}
