package models

import "time"

type VendaItem struct {
	ID            int       `json:"id"`
	VendaID       int       `json:"venda_id"`
	ProdutoID     int       `json:"produto_id"`
	Produto       *Produto  `json:"produto,omitempty"` // Produto completo
	Quantidade    int       `json:"quantidade"`
	ValorUnitario float64   `json:"valor_unitario"`
	ValorTotal    float64   `json:"valor_total"`
	CriadoEm      time.Time `json:"criado_em"`
	AtualizadoEm  time.Time `json:"atualizado_em"`
}

type Venda struct {
	ID                    int          `json:"id"`
	VendedorID            int          `json:"vendedor_id"`
	Vendedor              *Funcionario `json:"vendedor,omitempty"` // Vendedor completo
	FormaPagamento        string       `json:"forma_pagamento"`
	Parcelas              int          `json:"parcelas"`
	PrazoDias             int          `json:"prazo_dias,omitempty"`
	InicioPrimeiraParcela *time.Time   `json:"inicio_primeira_parcela,omitempty"`
	Itens                 []VendaItem  `json:"itens,omitempty"` // Itens da venda
	CriadoEm              time.Time    `json:"criado_em"`
	AtualizadoEm          time.Time    `json:"atualizado_em"`
}

// DTO para entrada de dados (POST/PUT)
type VendaInput struct {
	VendedorID            int              `json:"vendedor_id"`
	FormaPagamento        string           `json:"forma_pagamento"`
	Parcelas              int              `json:"parcelas"`
	PrazoDias             int              `json:"prazo_dias,omitempty"`
	InicioPrimeiraParcela *time.Time       `json:"inicio_primeira_parcela,omitempty"`
	Itens                 []VendaItemInput `json:"itens"`
}

type VendaItemInput struct {
	ProdutoID     int     `json:"produto_id"`
	Quantidade    int     `json:"quantidade"`
	ValorUnitario float64 `json:"valor_unitario"`
}
