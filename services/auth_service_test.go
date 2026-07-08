package services

import (
	"testing"
	"time"

	apperrors "donapresentes/errors"
	"donapresentes/models"
)

func newTestAuthService() *AuthService {
	return &AuthService{
		jwtSecret:     []byte("test-secret-para-testes-32bytes!!"),
		accessExpiry:  15 * time.Minute,
		refreshExpiry: 7 * 24 * time.Hour,
		auditService:  &AuditService{},
	}
}

func TestValidatePasswordComplexity(t *testing.T) {
	auth := newTestAuthService()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"muito curta", "Ab1!", true},
		{"sem maiúscula", "abcdefgh1!", true},
		{"sem minúscula", "ABCDEFGH1!", true},
		{"sem dígito", "Abcdefgh!", true},
		{"sem especial", "Abcdefgh1", true},
		{"válida", "Abcdefgh1!", false},
		{"válida com especiais", "MinhaSenha@2024!", false},
		{"8 chars sem especial", "Abcd1234", true},
		{"8 chars ok", "Abcdef1@", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auth.ValidatePasswordComplexity(tt.password)
			if tt.wantErr && err == nil {
				t.Fatalf("esperava erro, mas não houve")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("não esperava erro, mas houve: %v", err)
			}
		})
	}
}

func TestHashAndCheckPassword(t *testing.T) {
	auth := newTestAuthService()

	password := "SenhaForte@123"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("erro ao hash: %v", err)
	}

	if !auth.CheckPassword(password, hash) {
		t.Fatalf("senha correta deveria validar")
	}

	if auth.CheckPassword("SenhaErrada@456", hash) {
		t.Fatalf("senha incorreta não deveria validar")
	}
}

func TestValidateToken(t *testing.T) {
	auth := newTestAuthService()

	user := &models.User{ID: 1, Username: "testuser", Role: "admin"}

	token, err := auth.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("erro ao gerar token: %v", err)
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		t.Fatalf("erro ao validar token: %v", err)
	}

	if claims["username"] != "testuser" {
		t.Fatalf("username esperado 'testuser', got %v", claims["username"])
	}
	if claims["role"] != "admin" {
		t.Fatalf("role esperado 'admin', got %v", claims["role"])
	}
	if int(claims["user_id"].(float64)) != 1 {
		t.Fatalf("user_id esperado 1, got %v", claims["user_id"])
	}

	_, err = auth.ValidateToken("token-invalido")
	if err == nil {
		t.Fatalf("token inválido deveria retornar erro")
	}
}

func TestTokenExpiry(t *testing.T) {
	auth := newTestAuthService()
	auth.accessExpiry = -1 * time.Hour

	user := &models.User{ID: 1, Username: "testuser", Role: "standard"}

	token, err := auth.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("erro ao gerar token expirado: %v", err)
	}

	_, err = auth.ValidateToken(token)
	if err == nil {
		t.Fatalf("token expirado deveria retornar erro")
	}
}

func TestGetJWTSecret(t *testing.T) {
	auth := newTestAuthService()

	secret := auth.GetJWTSecret()
	if string(secret) != "test-secret-para-testes-32bytes!!" {
		t.Fatalf("JWT_SECRET não corresponde")
	}
}

func TestGenerateAccessTokenHasTypeClaim(t *testing.T) {
	auth := newTestAuthService()

	user := &models.User{ID: 5, Username: "joao", Role: "gerente"}
	token, err := auth.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("erro ao gerar token: %v", err)
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		t.Fatalf("erro ao validar token: %v", err)
	}

	if claims["type"] != "access" {
		t.Fatalf("type claim esperado 'access', got %v", claims["type"])
	}
	if claims["user_id"].(float64) != 5 {
		t.Fatalf("user_id esperado 5, got %v", claims["user_id"])
	}
}

func TestAppErrorIsValidationError(t *testing.T) {
	err := apperrors.NewValidationError("teste")
	if !apperrors.IsValidationError(err) {
		t.Fatalf("NewValidationError deveria ser validation error")
	}
}

func TestAppErrorIsNotFound(t *testing.T) {
	err := apperrors.NewNotFoundError("recurso")
	if !apperrors.IsNotFound(err) {
		t.Fatalf("NewNotFoundError deveria ser not found")
	}
}
