package models

import "time"

type Transportadora struct {
	ID                       int       `json:"id"`
	NomeTransportadora       string    `json:"nome_transportadora"`
	TipoTransportadora       string    `json:"tipo_transportadora"` // Pessoa Jurídica ou Física
	Email                    string    `json:"email"`
	TelefoneFixo             string    `json:"telefone_fixo"`
	Celular                  string    `json:"celular"`
	EnderecoCompleto         string    `json:"endereco_completo"`
	ContatoPrincipalNome     string    `json:"contato_principal_nome"`
	ContatoPrincipalTelefone string    `json:"contato_principal_telefone"`
	Site                     string    `json:"site"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}
