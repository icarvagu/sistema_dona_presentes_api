package services

import (
	"testing"

	apperrors "donapresentes/errors"
	"donapresentes/models"
)

// ── Mock Supplier ──

type mockSupplierRepo struct {
	suppliers []models.Supplier
}

func (m *mockSupplierRepo) GetAll() ([]models.Supplier, error) {
	return m.suppliers, nil
}
func (m *mockSupplierRepo) GetByID(id int) (*models.Supplier, error) {
	for i := range m.suppliers {
		if m.suppliers[i].ID == id {
			return &m.suppliers[i], nil
		}
	}
	return nil, apperrors.ErrSupplierNotFound
}
func (m *mockSupplierRepo) Create(s *models.Supplier) error {
	s.ID = len(m.suppliers) + 1
	m.suppliers = append(m.suppliers, *s)
	return nil
}
func (m *mockSupplierRepo) Update(id int, s *models.Supplier) error {
	for i := range m.suppliers {
		if m.suppliers[i].ID == id {
			m.suppliers[i] = *s
			m.suppliers[i].ID = id
			return nil
		}
	}
	return apperrors.ErrSupplierNotFound
}
func (m *mockSupplierRepo) Delete(id int) error {
	for i := range m.suppliers {
		if m.suppliers[i].ID == id {
			m.suppliers = append(m.suppliers[:i], m.suppliers[i+1:]...)
			return nil
		}
	}
	return apperrors.ErrSupplierNotFound
}

// ── Mock Carrier ──

type mockCarrierRepo struct {
	carriers []models.Carrier
}

func (m *mockCarrierRepo) GetAll() ([]models.Carrier, error) {
	return m.carriers, nil
}
func (m *mockCarrierRepo) GetByID(id int) (*models.Carrier, error) {
	for i := range m.carriers {
		if m.carriers[i].ID == id {
			return &m.carriers[i], nil
		}
	}
	return nil, apperrors.ErrCarrierNotFound
}
func (m *mockCarrierRepo) Create(c *models.Carrier) error {
	c.ID = len(m.carriers) + 1
	m.carriers = append(m.carriers, *c)
	return nil
}
func (m *mockCarrierRepo) Update(id int, c *models.Carrier) error {
	for i := range m.carriers {
		if m.carriers[i].ID == id {
			m.carriers[i] = *c
			m.carriers[i].ID = id
			return nil
		}
	}
	return apperrors.ErrCarrierNotFound
}
func (m *mockCarrierRepo) Delete(id int) error {
	for i := range m.carriers {
		if m.carriers[i].ID == id {
			m.carriers = append(m.carriers[:i], m.carriers[i+1:]...)
			return nil
		}
	}
	return apperrors.ErrCarrierNotFound
}

// ────── Supplier Tests ──────

func TestSupplierCreateValid(t *testing.T) {
	mock := &mockSupplierRepo{}
	s := NewSupplierService(mock, nil)
	email := "contato@fornecedor.com"

	err := s.Create(&models.Supplier{Name: "Fornecedor Ltda", Email: &email})
	if err != nil {
		t.Fatalf("erro ao criar fornecedor valido: %v", err)
	}
}

func TestSupplierCreateMissingName(t *testing.T) {
	mock := &mockSupplierRepo{}
	s := NewSupplierService(mock, nil)

	err := s.Create(&models.Supplier{Name: ""})
	if err == nil {
		t.Fatal("esperava erro para nome vazio")
	}
	if !apperrors.IsValidationError(err) {
		t.Fatalf("esperava ValidationError, got %T: %v", err, err)
	}
}

func TestSupplierCreateInvalidEmail(t *testing.T) {
	mock := &mockSupplierRepo{}
	s := NewSupplierService(mock, nil)
	email := "invalido"

	err := s.Create(&models.Supplier{Name: "Teste", Email: &email})
	if err == nil {
		t.Fatal("esperava erro para email invalido")
	}
}

func TestSupplierGetAll(t *testing.T) {
	mock := &mockSupplierRepo{
		suppliers: []models.Supplier{
			{ID: 1, Name: "Fornecedor A"},
			{ID: 2, Name: "Fornecedor B"},
		},
	}
	s := NewSupplierService(mock, nil)
	list, err := s.GetAll()
	if err != nil {
		t.Fatalf("erro ao listar: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("esperado 2, got %d", len(list))
	}
}

func TestSupplierGetByID(t *testing.T) {
	mock := &mockSupplierRepo{
		suppliers: []models.Supplier{
			{ID: 10, Name: "Fornecedor X"},
		},
	}
	s := NewSupplierService(mock, nil)
	supplier, err := s.GetByID(10)
	if err != nil {
		t.Fatalf("erro ao buscar: %v", err)
	}
	if supplier.Name != "Fornecedor X" {
		t.Errorf("esperado 'Fornecedor X', got '%s'", supplier.Name)
	}
}

func TestSupplierGetByIDNotFound(t *testing.T) {
	mock := &mockSupplierRepo{}
	s := NewSupplierService(mock, nil)
	_, err := s.GetByID(999)
	if err == nil {
		t.Fatal("esperava erro para ID inexistente")
	}
}

func TestSupplierUpdate(t *testing.T) {
	mock := &mockSupplierRepo{
		suppliers: []models.Supplier{
			{ID: 1, Name: "Original"},
		},
	}
	s := NewSupplierService(mock, nil)
	email := "novo@email.com"

	err := s.Update(1, &models.Supplier{Name: "Atualizado", Email: &email})
	if err != nil {
		t.Fatalf("erro ao atualizar: %v", err)
	}
}

func TestSupplierDelete(t *testing.T) {
	mock := &mockSupplierRepo{
		suppliers: []models.Supplier{
			{ID: 1, Name: "Fornecedor"},
		},
	}
	s := NewSupplierService(mock, nil)

	err := s.Delete(1)
	if err != nil {
		t.Fatalf("erro ao deletar: %v", err)
	}
	if len(mock.suppliers) != 0 {
		t.Fatalf("esperado 0 apos delete, got %d", len(mock.suppliers))
	}
}

func TestSupplierDeleteNotFound(t *testing.T) {
	mock := &mockSupplierRepo{}
	s := NewSupplierService(mock, nil)

	err := s.Delete(999)
	if err == nil {
		t.Fatal("esperava erro para ID inexistente")
	}
}

func TestSupplierUpdateNotFound(t *testing.T) {
	mock := &mockSupplierRepo{}
	s := NewSupplierService(mock, nil)
	err := s.Update(999, &models.Supplier{Name: "Teste"})
	if err == nil {
		t.Fatal("esperava erro para ID inexistente no update")
	}
}

// ────── Carrier Tests ──────

func TestCarrierCreateValid(t *testing.T) {
	mock := &mockCarrierRepo{}
	s := NewCarrierService(mock, nil)

	email := "logistica@transp.com"
	err := s.Create(&models.Carrier{
		Name:         "Transportadora Teste",
		CarrierType:  "Pessoa Jurídica",
		Email:        &email,
	})
	if err != nil {
		t.Fatalf("erro ao criar transportadora valida: %v", err)
	}
}

func TestCarrierCreateMissingName(t *testing.T) {
	mock := &mockCarrierRepo{}
	s := NewCarrierService(mock, nil)

	err := s.Create(&models.Carrier{CarrierType: "Pessoa Jurídica"})
	if err == nil {
		t.Fatal("esperava erro para nome vazio")
	}
}

func TestCarrierCreateMissingType(t *testing.T) {
	mock := &mockCarrierRepo{}
	s := NewCarrierService(mock, nil)

	err := s.Create(&models.Carrier{Name: "Teste"})
	if err == nil {
		t.Fatal("esperava erro para tipo vazio")
	}
}

func TestCarrierCreateInvalidType(t *testing.T) {
	mock := &mockCarrierRepo{}
	s := NewCarrierService(mock, nil)

	err := s.Create(&models.Carrier{Name: "Teste", CarrierType: "Tipo Invalido"})
	if err == nil {
		t.Fatal("esperava erro para tipo invalido")
	}
}

func TestCarrierGetAll(t *testing.T) {
	mock := &mockCarrierRepo{
		carriers: []models.Carrier{
			{ID: 1, Name: "Transp A"},
			{ID: 2, Name: "Transp B"},
		},
	}
	s := NewCarrierService(mock, nil)
	list, err := s.GetAll()
	if err != nil {
		t.Fatalf("erro ao listar: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("esperado 2, got %d", len(list))
	}
}

func TestCarrierGetByID(t *testing.T) {
	mock := &mockCarrierRepo{
		carriers: []models.Carrier{
			{ID: 5, Name: "Transp X"},
		},
	}
	s := NewCarrierService(mock, nil)
	c, err := s.GetByID(5)
	if err != nil {
		t.Fatalf("erro ao buscar: %v", err)
	}
	if c.Name != "Transp X" {
		t.Errorf("esperado 'Transp X', got '%s'", c.Name)
	}
}

func TestCarrierGetByIDNotFound(t *testing.T) {
	mock := &mockCarrierRepo{}
	s := NewCarrierService(mock, nil)
	_, err := s.GetByID(999)
	if err == nil {
		t.Fatal("esperava erro para ID inexistente")
	}
}

func TestCarrierUpdate(t *testing.T) {
	mock := &mockCarrierRepo{
		carriers: []models.Carrier{
			{ID: 1, Name: "Original", CarrierType: "Pessoa Jurídica"},
		},
	}
	s := NewCarrierService(mock, nil)
	err := s.Update(1, &models.Carrier{Name: "Atualizada", CarrierType: "Pessoa Física"})
	if err != nil {
		t.Fatalf("erro ao atualizar: %v", err)
	}
}

func TestCarrierDelete(t *testing.T) {
	mock := &mockCarrierRepo{
		carriers: []models.Carrier{
			{ID: 1, Name: "Transp", CarrierType: "Pessoa Jurídica"},
		},
	}
	s := NewCarrierService(mock, nil)
	err := s.Delete(1)
	if err != nil {
		t.Fatalf("erro ao deletar: %v", err)
	}
	if len(mock.carriers) != 0 {
		t.Fatalf("esperado 0 apos delete, got %d", len(mock.carriers))
	}
}
