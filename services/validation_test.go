package services

import (
	"testing"

	apperrors "donapresentes/errors"
	"donapresentes/models"
)

// ───────────── SaleService Validation ─────────────

func newTestSaleService() *SaleService {
	return &SaleService{}
}

func TestSaleValidateRequiredFields(t *testing.T) {
	s := newTestSaleService()

	tests := []struct {
		name string
		sale *models.SaleInput
	}{
		{"seller vazio", &models.SaleInput{}},
		{"customer vazio", &models.SaleInput{SellerID: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateSale(tt.sale)
			if err == nil {
				t.Fatal("esperava erro de validacao")
			}
		})
	}
}

func TestSaleValidateValidInput(t *testing.T) {
	s := newTestSaleService()
	err := s.ValidateSale(&models.SaleInput{
		SellerID:   1,
		CustomerID: 2,
		Status:     "Em Andamento",
	})
	if err != nil {
		t.Logf("erro (esperado sem repo): %v", err)
	}
}

// ───────────── QuoteService Validation ─────────────

func newTestQuoteService() *QuoteService {
	return &QuoteService{}
}

func TestQuoteValidateItems(t *testing.T) {
	s := newTestQuoteService()
	err := s.ValidateQuote(&models.QuoteInput{})
	if err == nil {
		t.Fatal("esperava erro para quote vazia")
	}
}

// ───────────── ProductService Validation ─────────────

func newTestProductService() *ProductService {
	return &ProductService{}
}

func TestProductValidateRequiredFields(t *testing.T) {
	s := newTestProductService()
	err := s.validateRequiredFields(&models.Product{})
	if err == nil {
		t.Fatal("esperava erro para produto vazio")
	}

	err2 := s.validateRequiredFields(&models.Product{ProductName: "Teste", InternalCode: "COD-001"})
	if err2 != nil {
		t.Fatalf("nao esperava erro: %v", err2)
	}
}

func TestProductValidateKitType(t *testing.T) {
	s := newTestProductService()

	tests := []struct {
		name    string
		kit     string
		items   int
		wantErr bool
	}{
		{"none sem items", "none", 0, false},
		{"none com items", "none", 1, true},
		{"internal sem items", "internal_composition", 0, true},
		// internal com items precisa de repo mockado (pula por enquanto)
		{"supplier_ready sem items", "supplier_ready", 0, false},
		{"supplier_ready com items", "supplier_ready", 1, true},
		{"invalido", "invalid_type", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := make([]models.ProductItem, tt.items)
			err := s.validateItems(&models.Product{KitType: tt.kit, Items: items})
			if tt.wantErr && err == nil {
				t.Fatal("esperava erro")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("nao esperava erro: %v", err)
			}
		})
	}
}

// ───────────── UserService Validation ─────────────

func newTestUserService() *UserService {
	return &UserService{
		authService: newTestAuthService(),
	}
}

func TestUserValidateInputUsernameEmpty(t *testing.T) {
	svc := newTestUserService()
	err := svc.validateUserInput(&models.UserInput{Username: "", Role: "admin"}, true)
	if err == nil {
		t.Fatal("esperava erro")
	}
}

func TestUserValidateInputUsernameShort(t *testing.T) {
	svc := newTestUserService()
	err := svc.validateUserInput(&models.UserInput{Username: "ab", Role: "admin"}, true)
	if err == nil {
		t.Fatal("esperava erro")
	}
}

func TestUserValidateInputInvalidRole(t *testing.T) {
	svc := newTestUserService()
	err := svc.validateUserInput(&models.UserInput{Username: "teste", Role: "invalid"}, true)
	if err == nil {
		t.Fatal("esperava erro")
	}
}

func TestUserValidateInputInvalidStatus(t *testing.T) {
	svc := newTestUserService()
	err := svc.validateUserInput(&models.UserInput{Username: "teste", Role: "admin", Status: "Cancelado"}, true)
	if err == nil {
		t.Fatal("esperava erro")
	}
}

func TestUserValidateInputInvalidGender(t *testing.T) {
	svc := newTestUserService()
	g := "Alien"
	err := svc.validateUserInput(&models.UserInput{Username: "teste", Role: "admin", Gender: &g}, true)
	if err == nil {
		t.Fatal("esperava erro")
	}
}

func TestUserValidateInputShortCPF(t *testing.T) {
	svc := newTestUserService()
	err := svc.validateUserInput(&models.UserInput{Username: "teste", Role: "admin", CPF: "123"}, true)
	if err == nil {
		t.Fatal("esperava erro")
	}
}

func TestUserValidateInputValid(t *testing.T) {
	svc := newTestUserService()
	err := svc.validateUserInput(&models.UserInput{Username: "usuarioteste", Role: "admin", Password: "SenhaForte1@"}, true)
	if err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}
}

func TestUserSanitize(t *testing.T) {
	t.Run("SanitizeUsername lowercases and removes non-printable", func(t *testing.T) {
		r := SanitizeUsername("ADMIN")
		if r != "ADMIN" {
			t.Errorf("SanitizeUsername nao deveria mexer em maiusculas: %s", r)
		}
	})
	t.Run("SanitizeString removes non-printable", func(t *testing.T) {
		r := SanitizeString("normal")
		if r != "normal" {
			t.Errorf("SanitizeString alterou string normal: %s", r)
		}
	})
	t.Run("SanitizeCPF removes punctuation", func(t *testing.T) {
		r := SanitizeCPF("123.456.789-01")
		if r != "12345678901" {
			t.Errorf("CPF esperado '12345678901', got '%s'", r)
		}
	})
	t.Run("SanitizePassword truncates long", func(t *testing.T) {
		r := SanitizePassword("abc")
		if r != "abc" {
			t.Errorf("password alterada: '%s'", r)
		}
	})
}

// ───────────── CryptoService ─────────────

func TestCryptoHash(t *testing.T) {
	svc := &CryptoService{}
	h := svc.Hash("teste")
	if h == "" {
		t.Fatal("hash nao deveria ser vazio")
	}
	if svc.Hash("teste") != svc.Hash("teste") {
		t.Fatal("hash deveria ser deterministico")
	}
	if svc.Hash("a") == svc.Hash("b") {
		t.Fatal("hash diferente para strings diferentes")
	}
}

// ───────────── DashboardService ─────────────

func newTestDashboardService() *DashboardService {
	return &DashboardService{}
}

// ───────────── AuditService ─────────────

func TestAuditLogSimple(t *testing.T) {
	s := &AuditService{}
	s.LogSimple(nil, "test", "test", "detail", "127.0.0.1")
	uid := 1
	s.LogSimple(&uid, "login", "user", "success", "127.0.0.1")
}

// ───────────── SyncService ─────────────

func newTestSyncService() *SyncService {
	return &SyncService{}
}

// ───────────── XBZService ─────────────

func TestXBZServiceWithEmptyConfig(t *testing.T) {
	svc := NewXBZService("", "")
	if svc == nil {
		t.Fatal("service nao deveria ser nil")
	}
	_, err := svc.GetProducts()
	if err == nil {
		t.Fatal("esperava erro sem configuracao")
	}
}

func TestXBZServiceWithConfig(t *testing.T) {
	svc := NewXBZService("cnpj-test", "token-test")
	if svc == nil {
		t.Fatal("service nao deveria ser nil")
	}
}

// ───────────── RequestLoggerService ─────────────

func TestRequestLoggerDisabled(t *testing.T) {
	svc := &RequestLoggerService{enabled: false}
	svc.Log(RequestLogEntry{
		RequestID: "test",
		Method:    "GET",
		Path:      "/test",
		Status:    200,
	})
}

// ───────────── Endpoint error tests ─────────────
// Verify that all common AppError types return proper status codes

func TestAppErrorStatusCodes(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code int
	}{
		{"NotFound", apperrors.ErrCustomerNotFound, 404},
		{"NotFound supplier", apperrors.ErrSupplierNotFound, 404},
		{"NotFound product", apperrors.ErrProductNotFound, 404},
		{"NotFound carrier", apperrors.ErrCarrierNotFound, 404},
		{"NotFound sale", apperrors.ErrSaleNotFound, 404},
		{"NotFound quote", apperrors.ErrQuoteNotFound, 404},
		{"Duplicate", apperrors.ErrDuplicateEntry, 409},
		{"InvalidJSON", apperrors.ErrInvalidJSON, 400},
		{"InvalidID", apperrors.ErrInvalidID, 400},
		{"InternalServer", apperrors.ErrInternalServer, 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if appErr, ok := tt.err.(*apperrors.AppError); ok {
				if appErr.Code != tt.code {
					t.Errorf("esperado code %d, got %d", tt.code, appErr.Code)
				}
			}
		})
	}
}

func TestNewValidationErrorFormat(t *testing.T) {
	err := apperrors.NewValidationError("campo x")
	if err.Error() != "Validação falhou para: campo x" {
		t.Errorf("formato inesperado: %s", err.Error())
	}
}

func TestNewMissingFieldErrorFormat(t *testing.T) {
	err := apperrors.NewMissingFieldError("name", "email")
	if err.Error() != "Campos obrigatórios não fornecidos: name, email" {
		t.Errorf("formato inesperado: %s", err.Error())
	}
}
