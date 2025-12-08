package models

import "time"

type Produto struct {
	ID               int         `json:"id"`
	NomeProduto      string      `json:"nome_produto"`
	CodigoInterno    string      `json:"codigo_interno"`
	CodigoFornecedor int         `json:"codigo_fornecedor"`
	Fornecedor       *Fornecedor `json:"fornecedor,omitempty"`
	GrupoProduto     string      `json:"grupo_produto,omitempty"`
	Descricao        string      `json:"descricao,omitempty"`
	Fotos            []string    `json:"fotos,omitempty"`
	NCM              string      `json:"ncm,omitempty"`
	OrigemMaterial   string      `json:"origem_material,omitempty"`
	Estoque          int         `json:"estoque"`
	CriadoEm         time.Time   `json:"criado_em"`
	AtualizadoEm     time.Time   `json:"atualizado_em"`
}
