package models

import "time"

type Fornecedor struct {
	ID                 int       `json:"id"`
	FantasyName        string    `json:"nome_fantasia_ou_razao_social"`
	CNPJ               *string   `json:"cnpj"`                    // Nullable
	StateRegistration  *string   `json:"inscricao_estadual"`      // Nullable
	ContactResponsible *string   `json:"responsavel_atendimento"` // Nullable
	GeneralEmail       *string   `json:"email_geral"`             // Nullable
	LandlinePhone      *string   `json:"telefone_fixo"`           // Nullable
	MobilePhone        *string   `json:"celular"`                 // Nullable
	ResponsibleEmail   *string   `json:"email_responsavel"`       // Nullable
	CommercialAddress  *string   `json:"endereco_comercial"`      // Nullable
	CreatedAt          time.Time `json:"criado_em"`
	UpdatedAt          time.Time `json:"atualizado_em"`
}
