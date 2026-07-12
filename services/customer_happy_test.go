package services

import (
	"testing"

	apperrors "donapresentes/errors"
	"donapresentes/models"
)

type mockCustomerRepo struct {
	customers []models.Customer
}

func (m *mockCustomerRepo) GetAll() ([]models.Customer, error) {
	return m.customers, nil
}
func (m *mockCustomerRepo) GetByID(id int) (*models.Customer, error) {
	for i := range m.customers {
		if m.customers[i].ID == id {
			return &m.customers[i], nil
		}
	}
	return nil, apperrors.ErrCustomerNotFound
}
func (m *mockCustomerRepo) Create(c *models.Customer) error {
	c.ID = len(m.customers) + 1
	m.customers = append(m.customers, *c)
	return nil
}
func (m *mockCustomerRepo) Update(id int, c *models.Customer) error {
	for i := range m.customers {
		if m.customers[i].ID == id {
			m.customers[i] = *c
			m.customers[i].ID = id
			return nil
		}
	}
	return apperrors.ErrCustomerNotFound
}
func (m *mockCustomerRepo) Delete(id int) error {
	for i := range m.customers {
		if m.customers[i].ID == id {
			m.customers = append(m.customers[:i], m.customers[i+1:]...)
			return nil
		}
	}
	return apperrors.ErrCustomerNotFound
}

func TestCustomerValidatesRequiredFields(t *testing.T) {
	s := NewCustomerService(&mockCustomerRepo{}, nil)

	tests := []struct {
		name     string
		customer *models.Customer
	}{
		{"nome vazio", &models.Customer{CustomerType: "PF", Status: "Ativo"}},
		{"tipo vazio", &models.Customer{Name: "Teste", Status: "Ativo"}},
		{"status vazio", &models.Customer{Name: "Teste", CustomerType: "PF"}},
		{"tipo invalido", &models.Customer{Name: "Teste", CustomerType: "XYZ", Status: "Ativo"}},
		{"status invalido", &models.Customer{Name: "Teste", CustomerType: "PF", Status: "Cancelado"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.Create(tt.customer)
			if err == nil {
				t.Fatal("esperava erro de validacao")
			}
			if !apperrors.IsValidationError(err) {
				t.Fatalf("esperava ValidationError, got %T: %v", err, err)
			}
		})
	}
}

func TestCustomerCreateValidPF(t *testing.T) {
	mock := &mockCustomerRepo{}
	s := NewCustomerService(mock, nil)

	cpf := "12345678901"
	customer := &models.Customer{
		Name:         "João Silva",
		CustomerType: "PF",
		Status:       "Ativo",
		Email:        "joao@email.com",
		CPF:          &cpf,
	}

	err := s.Create(customer)
	if err != nil {
		t.Fatalf("erro ao criar: %v", err)
	}
	if customer.ID == 0 {
		t.Fatal("ID nao deveria ser 0 apos criar")
	}
}

func TestCustomerCreateValidPJ(t *testing.T) {
	mock := &mockCustomerRepo{}
	s := NewCustomerService(mock, nil)
	cnpj := "11222333000181"

	err := s.Create(&models.Customer{
		Name:         "Empresa Ltda",
		CustomerType: "PJ",
		Status:       "Ativo",
		Email:        "contato@empresa.com",
		CNPJ:         &cnpj,
	})
	if err != nil {
		t.Fatalf("erro ao criar: %v", err)
	}
}

func TestCustomerGetAll(t *testing.T) {
	mock := &mockCustomerRepo{
		customers: []models.Customer{
			{ID: 1, Name: "Cliente A", CustomerType: "PF", Status: "Ativo"},
			{ID: 2, Name: "Cliente B", CustomerType: "PJ", Status: "Ativo"},
		},
	}
	s := NewCustomerService(mock, nil)
	customers, err := s.GetAll()
	if err != nil {
		t.Fatalf("erro ao listar: %v", err)
	}
	if len(customers) != 2 {
		t.Fatalf("esperado 2 clientes, got %d", len(customers))
	}
}

func TestCustomerGetByID(t *testing.T) {
	mock := &mockCustomerRepo{
		customers: []models.Customer{
			{ID: 5, Name: "Cliente A", CustomerType: "PF", Status: "Ativo"},
		},
	}
	s := NewCustomerService(mock, nil)
	c, err := s.GetByID(5)
	if err != nil {
		t.Fatalf("erro ao buscar: %v", err)
	}
	if c.Name != "Cliente A" {
		t.Errorf("esperado 'Cliente A', got '%s'", c.Name)
	}
}

func TestCustomerGetByIDNotFound(t *testing.T) {
	mock := &mockCustomerRepo{}
	s := NewCustomerService(mock, nil)
	_, err := s.GetByID(999)
	if err == nil {
		t.Fatal("esperava erro para ID inexistente")
	}
}

func TestCustomerUpdate(t *testing.T) {
	mock := &mockCustomerRepo{
		customers: []models.Customer{
			{ID: 1, Name: "Original", CustomerType: "PF", Status: "Ativo", Email: "a@b.com"},
		},
	}
	s := NewCustomerService(mock, nil)
	err := s.Update(1, &models.Customer{
		Name:         "Atualizado",
		CustomerType: "PJ",
		Status:       "Inativo",
		Email:        "novo@email.com",
	})
	if err != nil {
		t.Fatalf("erro ao atualizar: %v", err)
	}
}

func TestCustomerDelete(t *testing.T) {
	mock := &mockCustomerRepo{
		customers: []models.Customer{
			{ID: 1, Name: "Cliente", CustomerType: "PF", Status: "Ativo"},
		},
	}
	s := NewCustomerService(mock, nil)
	err := s.Delete(1)
	if err != nil {
		t.Fatalf("erro ao deletar: %v", err)
	}
	customers, _ := s.GetAll()
	if len(customers) != 0 {
		t.Fatalf("esperado 0 clientes apos delete, got %d", len(customers))
	}
}

func TestCustomerCreateWithCPFAndCNPJValidations(t *testing.T) {
	mock := &mockCustomerRepo{}
	s := NewCustomerService(mock, nil)

	t.Run("PF com CPF valido", func(t *testing.T) {
		cpf := "52998224725"
		err := s.Create(&models.Customer{
			Name: "Teste", CustomerType: "PF", Status: "Ativo", CPF: &cpf,
		})
		if err != nil {
			t.Fatalf("CPF valido reprovado: %v", err)
		}
	})

	t.Run("PJ com CNPJ valido", func(t *testing.T) {
		cnpj := "11222333000181"
		err := s.Create(&models.Customer{
			Name: "Teste", CustomerType: "PJ", Status: "Ativo", CNPJ: &cnpj,
		})
		if err != nil {
			t.Fatalf("CNPJ valido reprovado: %v", err)
		}
	})
}

func TestCustomerEmailValidation(t *testing.T) {
	mock := &mockCustomerRepo{}
	s := NewCustomerService(mock, nil)

	err := s.Create(&models.Customer{
		Name: "Teste", CustomerType: "PF", Status: "Ativo", Email: "email-invalido",
	})
	if err == nil {
		t.Fatal("esperava erro para email invalido")
	}
}
