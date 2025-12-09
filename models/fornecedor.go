package models

import "time"

type Fornecedor struct {
	ID                 int       `json:"id"`
	FantasyName        string    `json:"nome_fantasia_ou_razao_social"`
	CNPJ               *string   `json:"cnpj,omitempty"`                    // Nullable
	StateRegistration  *string   `json:"inscricao_estadual,omitempty"`      // Nullable
	ContactResponsible *string   `json:"responsavel_atendimento,omitempty"` // Nullable
	GeneralEmail       *string   `json:"email_geral,omitempty"`             // Nullable
	LandlinePhone      *string   `json:"telefone_fixo,omitempty"`           // Nullable
	MobilePhone        *string   `json:"celular,omitempty"`                 // Nullable
	ResponsibleEmail   *string   `json:"email_responsavel,omitempty"`       // Nullable
	CommercialAddress  *string   `json:"endereco_comercial,omitempty"`      // Nullable
	CreatedAt          time.Time `json:"criado_em"`
	UpdatedAt          time.Time `json:"atualizado_em"`
}
