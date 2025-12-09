package models

import "time"

type Funcionario struct {
	ID               int        `json:"id"`
	NomeCompleto     string     `json:"nome_completo"`
	CPF              string     `json:"cpf"`
	RG               *string    `json:"rg,omitempty"`                // Nullable
	DataNascimento   *time.Time `json:"data_nascimento,omitempty"`   // Nullable
	Sexo             *string    `json:"sexo,omitempty"`              // Nullable (Masculino, Feminino, Outro)
	Situacao         string     `json:"situacao"`                     // Ativo, Inativo
	EmailContato     *string    `json:"email_contato,omitempty"`     // Nullable
	EnderecoCompleto *string    `json:"endereco_completo,omitempty"` // Nullable
	TelefonesContato *string    `json:"telefones_contato,omitempty"` // Nullable
	Observacoes      *string    `json:"observacoes,omitempty"`       // Nullable
	CriadoEm         time.Time  `json:"criado_em"`
}
