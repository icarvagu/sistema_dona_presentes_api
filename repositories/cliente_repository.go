package repositories

import (
	"database/sql"
	"donapresentes/models"
	apperrors "donapresentes/errors"
	"regexp"
)

type ClienteRepository struct {
	db *sql.DB
}

func NewClienteRepository(db *sql.DB) *ClienteRepository {
	return &ClienteRepository{db: db}
}

// Validation helpers
func ValidateEmail(email string) error {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	if !regexp.MustCompile(pattern).MatchString(email) {
		return apperrors.ErrInvalidEmail
	}
	return nil
}

func ValidateCPF(cpf string) error {
	cpf = regexp.MustCompile(`\D`).ReplaceAllString(cpf, "")
	if len(cpf) != 11 {
		return apperrors.ErrInvalidCPF
	}
	return nil
}

func ValidateCNPJ(cnpj string) error {
	cnpj = regexp.MustCompile(`\D`).ReplaceAllString(cnpj, "")
	if len(cnpj) != 14 {
		return apperrors.ErrInvalidCNPJ
	}
	return nil
}

// GetAll returns all clientes with enderecos and contatos
func (r *ClienteRepository) GetAll() ([]models.Cliente, error) {
	rows, err := r.db.Query(`SELECT id, tipo_cliente, situacao, nome_empresa_pessoa, cnpj, cpf, email, telefone_comercial, celular, site, observacoes, criado_em, atualizado_em FROM clientes`)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	var clientes []models.Cliente
	for rows.Next() {
		var c models.Cliente
		err := rows.Scan(&c.ID, &c.TipoCliente, &c.Situacao, &c.NomeEmpresaPessoa, &c.CNPJ, &c.CPF, &c.Email, &c.TelefoneComercial, &c.Celular, &c.Site, &c.Observacoes, &c.CriadoEm, &c.AtualizadoEm)
		if err != nil {
			return nil, apperrors.NewDatabaseError(err)
		}

		// Fetch enderecos
		enderecos, _ := r.GetEnderecos(c.ID)
		c.Enderecos = enderecos

		// Fetch contatos
		contatos, _ := r.GetContatosAdicionais(c.ID)
		c.ContatosAdicionais = contatos

		clientes = append(clientes, c)
	}
	return clientes, nil
}

// GetByID returns a single cliente by ID with related data
func (r *ClienteRepository) GetByID(id int) (*models.Cliente, error) {
	var c models.Cliente
	err := r.db.QueryRow(`SELECT id, tipo_cliente, situacao, nome_empresa_pessoa, cnpj, cpf, email, telefone_comercial, celular, site, observacoes, criado_em, atualizado_em FROM clientes WHERE id=$1`, id).
		Scan(&c.ID, &c.TipoCliente, &c.Situacao, &c.NomeEmpresaPessoa, &c.CNPJ, &c.CPF, &c.Email, &c.TelefoneComercial, &c.Celular, &c.Site, &c.Observacoes, &c.CriadoEm, &c.AtualizadoEm)
	if err != nil {
		return nil, err
	}

	enderecos, _ := r.GetEnderecos(c.ID)
	c.Enderecos = enderecos

	contatos, _ := r.GetContatosAdicionais(c.ID)
	c.ContatosAdicionais = contatos

	return &c, nil
}

// Create inserts a new cliente with validation
func (r *ClienteRepository) Create(c *models.Cliente) error {
	// Validate required fields
	if c.NomeEmpresaPessoa == "" || c.Email == "" || c.TipoCliente == "" || c.Situacao == "" {
		return apperrors.NewMissingFieldError("nome_empresa_pessoa", "email", "tipo_cliente", "situacao")
	}

	// Validate tipo_cliente
	if c.TipoCliente != "PF" && c.TipoCliente != "PJ" {
		return apperrors.NewInvalidFieldError("tipo_cliente", "deve ser 'PF' ou 'PJ'")
	}

	// Validate situacao
	if c.Situacao != "Ativo" && c.Situacao != "Inativo" {
		return apperrors.NewInvalidFieldError("situacao", "deve ser 'Ativo' ou 'Inativo'")
	}

	// Validate email
	if err := ValidateEmail(c.Email); err != nil {
		return err
	}

	// Validate CPF/CNPJ
	if c.TipoCliente == "PF" {
		if c.CPF == nil || *c.CPF == "" {
			return apperrors.NewMissingFieldError("CPF")
		}
		if err := ValidateCPF(*c.CPF); err != nil {
			return err
		}
	} else if c.TipoCliente == "PJ" {
		if c.CNPJ == nil || *c.CNPJ == "" {
			return apperrors.NewMissingFieldError("CNPJ")
		}
		if err := ValidateCNPJ(*c.CNPJ); err != nil {
			return err
		}
	}

	// Insert cliente
	err := r.db.QueryRow(
		`INSERT INTO clientes (tipo_cliente, situacao, nome_empresa_pessoa, cnpj, cpf, email, telefone_comercial, celular, site, observacoes) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id, criado_em, atualizado_em`,
		c.TipoCliente, c.Situacao, c.NomeEmpresaPessoa, c.CNPJ, c.CPF, c.Email, c.TelefoneComercial, c.Celular, c.Site, c.Observacoes).
		Scan(&c.ID, &c.CriadoEm, &c.AtualizadoEm)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}

	// Insert enderecos
	for i := range c.Enderecos {
		c.Enderecos[i].ClienteID = c.ID
		_ = r.CreateEndereco(&c.Enderecos[i])
	}

	// Insert contatos adicionais
	for i := range c.ContatosAdicionais {
		c.ContatosAdicionais[i].ClienteID = c.ID
		_ = r.CreateContatoAdicional(&c.ContatosAdicionais[i])
	}

	return nil
}

// Update updates an existing cliente
func (r *ClienteRepository) Update(id int, c *models.Cliente) error {
	// Validate required fields
	if c.NomeEmpresaPessoa == "" || c.Email == "" || c.TipoCliente == "" || c.Situacao == "" {
		return apperrors.NewMissingFieldError("nome_empresa_pessoa", "email", "tipo_cliente", "situacao")
	}

	if c.TipoCliente != "PF" && c.TipoCliente != "PJ" {
		return apperrors.NewInvalidFieldError("tipo_cliente", "deve ser 'PF' ou 'PJ'")
	}

	if c.Situacao != "Ativo" && c.Situacao != "Inativo" {
		return apperrors.NewInvalidFieldError("situacao", "deve ser 'Ativo' ou 'Inativo'")
	}

	if err := ValidateEmail(c.Email); err != nil {
		return err
	}

	if c.TipoCliente == "PF" {
		if c.CPF == nil || *c.CPF == "" {
			return apperrors.NewMissingFieldError("CPF")
		}
		if err := ValidateCPF(*c.CPF); err != nil {
			return err
		}
	} else if c.TipoCliente == "PJ" {
		if c.CNPJ == nil || *c.CNPJ == "" {
			return apperrors.NewMissingFieldError("CNPJ")
		}
		if err := ValidateCNPJ(*c.CNPJ); err != nil {
			return err
		}
	}

	_, err := r.db.Exec(
		`UPDATE clientes SET tipo_cliente=$1, situacao=$2, nome_empresa_pessoa=$3, cnpj=$4, cpf=$5, email=$6, telefone_comercial=$7, celular=$8, site=$9, observacoes=$10, atualizado_em=NOW() WHERE id=$11`,
		c.TipoCliente, c.Situacao, c.NomeEmpresaPessoa, c.CNPJ, c.CPF, c.Email, c.TelefoneComercial, c.Celular, c.Site, c.Observacoes, id)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

// Delete removes a cliente and related enderecos/contatos (cascade)
func (r *ClienteRepository) Delete(id int) error {
	res, err := r.db.Exec("DELETE FROM clientes WHERE id=$1", id)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Endereco methods
func (r *ClienteRepository) GetEnderecos(clienteID int) ([]models.Endereco, error) {
	rows, err := r.db.Query(`SELECT id, cliente_id, tipo_endereco, endereco, criado_em, atualizado_em FROM enderecos_cliente WHERE cliente_id=$1`, clienteID)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	var enderecos []models.Endereco
	for rows.Next() {
		var e models.Endereco
		err := rows.Scan(&e.ID, &e.ClienteID, &e.TipoEndereco, &e.Endereco, &e.CriadoEm, &e.AtualizadoEm)
		if err != nil {
			return nil, apperrors.NewDatabaseError(err)
		}
		enderecos = append(enderecos, e)
	}
	return enderecos, nil
}

func (r *ClienteRepository) CreateEndereco(e *models.Endereco) error {
	if e.Endereco == "" {
		return apperrors.NewMissingFieldError("endereco")
	}
	if e.TipoEndereco != "comercial" && e.TipoEndereco != "entrega" {
		return apperrors.NewInvalidFieldError("tipo_endereco", "deve ser 'comercial' ou 'entrega'")
	}
	err := r.db.QueryRow(
		`INSERT INTO enderecos_cliente (cliente_id, tipo_endereco, endereco) VALUES ($1,$2,$3) RETURNING id, criado_em, atualizado_em`,
		e.ClienteID, e.TipoEndereco, e.Endereco).
		Scan(&e.ID, &e.CriadoEm, &e.AtualizadoEm)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

// Contato adicional methods
func (r *ClienteRepository) GetContatosAdicionais(clienteID int) ([]models.ContatoAdicional, error) {
	rows, err := r.db.Query(`SELECT id, cliente_id, nome, email, telefone, criado_em, atualizado_em FROM contatos_adicionais WHERE cliente_id=$1`, clienteID)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}
	defer rows.Close()

	var contatos []models.ContatoAdicional
	for rows.Next() {
		var c models.ContatoAdicional
		err := rows.Scan(&c.ID, &c.ClienteID, &c.Nome, &c.Email, &c.Telefone, &c.CriadoEm, &c.AtualizadoEm)
		if err != nil {
			return nil, apperrors.NewDatabaseError(err)
		}
		contatos = append(contatos, c)
	}
	return contatos, nil
}

func (r *ClienteRepository) CreateContatoAdicional(c *models.ContatoAdicional) error {
	if c.Nome == "" {
		return apperrors.NewMissingFieldError("nome")
	}
	err := r.db.QueryRow(
		`INSERT INTO contatos_adicionais (cliente_id, nome, email, telefone) VALUES ($1,$2,$3,$4) RETURNING id, criado_em, atualizado_em`,
		c.ClienteID, c.Nome, c.Email, c.Telefone).
		Scan(&c.ID, &c.CriadoEm, &c.AtualizadoEm)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}
