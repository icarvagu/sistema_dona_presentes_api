package models

import "time"

type Funcionario struct {
	ID               int        `json:"id"`
	NomeCompleto     string     `json:"nome_completo"`
	CPF              string     `json:"cpf"`
	RG               *string    `json:"rg"`                // Nullable
	DataNascimento   *time.Time `json:"data_nascimento"`   // Nullable
	Sexo             *string    `json:"sexo"`              // Nullable (Masculino, Feminino, Outro)
	Situacao         string     `json:"situacao"`          // Ativo, Inativo
	EmailContato     *string    `json:"email_contato"`     // Nullable
	EnderecoCompleto *string    `json:"endereco_completo"` // Nullable
	TelefonesContato *string    `json:"telefones_contato"` // Nullable
	Observacoes      *string    `json:"observacoes"`       // Nullable
	CriadoEm         time.Time  `json:"criado_em"`
}
