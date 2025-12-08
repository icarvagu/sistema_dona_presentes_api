-- +goose Up
-- +goose StatementBegin
CREATE TABLE vendas (
    id SERIAL PRIMARY KEY,
    vendedor_id INTEGER NOT NULL,
    forma_pagamento TEXT NOT NULL,
    parcelas INTEGER NOT NULL CHECK (parcelas > 0),
    prazo_dias INTEGER DEFAULT 0,
    inicio_primeira_parcela TIMESTAMP WITH TIME ZONE,
    criado_em TIMESTAMP WITH TIME ZONE DEFAULT now(),
    atualizado_em TIMESTAMP WITH TIME ZONE DEFAULT now(),
    CONSTRAINT fk_vendedor
      FOREIGN KEY(vendedor_id)
        REFERENCES funcionarios(id)
        ON DELETE RESTRICT
);

CREATE TABLE venda_itens (
    id SERIAL PRIMARY KEY,
    venda_id INTEGER NOT NULL,
    produto_id INTEGER NOT NULL,
    quantidade INTEGER NOT NULL CHECK (quantidade > 0),
    valor_unitario NUMERIC(10, 2) NOT NULL,
    valor_total NUMERIC(10, 2) NOT NULL,
    criado_em TIMESTAMP WITH TIME ZONE DEFAULT now(),
    atualizado_em TIMESTAMP WITH TIME ZONE DEFAULT now(),
    CONSTRAINT fk_venda
      FOREIGN KEY(venda_id)
        REFERENCES vendas(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_produto
      FOREIGN KEY(produto_id)
        REFERENCES produtos(id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_vendas_vendedor_id ON vendas(vendedor_id);
CREATE INDEX idx_venda_itens_venda_id ON venda_itens(venda_id);
CREATE INDEX idx_venda_itens_produto_id ON venda_itens(produto_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS venda_itens CASCADE;
DROP TABLE IF EXISTS vendas CASCADE;
-- +goose StatementEnd
