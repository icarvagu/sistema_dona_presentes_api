package models

import "time"

type Endereco struct {
	ID           int       `json:"id"`
	ClienteID    int       `json:"cliente_id"`
	TipoEndereco string    `json:"tipo_endereco"` // "comercial" or "entrega"
	Endereco     string    `json:"endereco"`
	CriadoEm     time.Time `json:"criado_em"`
	AtualizadoEm time.Time `json:"atualizado_em"`
}

type ContatoAdicional struct {
	ID           int       `json:"id"`
	ClienteID    int       `json:"cliente_id"`
	Nome         string    `json:"nome"`
	Email        string    `json:"email,omitempty"`
	Telefone     string    `json:"telefone,omitempty"`
	CriadoEm     time.Time `json:"criado_em"`
	AtualizadoEm time.Time `json:"atualizado_em"`
}

type Cliente struct {
	ID                 int                `json:"id"`
	TipoCliente        string             `json:"tipo_cliente"` // "PF" or "PJ"
	Situacao           string             `json:"situacao"`     // "Ativo" or "Inativo"
	NomeEmpresaPessoa  string             `json:"nome_empresa_pessoa"`
	CNPJ               *string            `json:"cnpj,omitempty"`
	CPF                *string            `json:"cpf,omitempty"`
	Email              string             `json:"email"`
	TelefoneComercial  string             `json:"telefone_comercial,omitempty"`
	Celular            string             `json:"celular,omitempty"`
	Site               *string            `json:"site,omitempty"`
	Observacoes        *string            `json:"observacoes,omitempty"`
	Enderecos          []Endereco         `json:"enderecos,omitempty"`
	ContatosAdicionais []ContatoAdicional `json:"contatos_adicionais,omitempty"`
	CriadoEm           time.Time          `json:"criado_em"`
	AtualizadoEm       time.Time          `json:"atualizado_em"`
}
