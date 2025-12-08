-- +goose Up
-- +goose StatementBegin
CREATE TABLE produtos (
    id SERIAL PRIMARY KEY,
    nome_produto TEXT NOT NULL,
    codigo_interno TEXT NOT NULL UNIQUE,
    codigo_fornecedor INTEGER NOT NULL,
    grupo_produto TEXT,
    descricao TEXT,
    fotos TEXT[],
    ncm TEXT,
    origem_material TEXT,
    estoque INTEGER DEFAULT 0,
    criado_em TIMESTAMP WITH TIME ZONE DEFAULT now(),
    atualizado_em TIMESTAMP WITH TIME ZONE DEFAULT now(),
    CONSTRAINT fk_fornecedor
      FOREIGN KEY(codigo_fornecedor)
        REFERENCES fornecedores(id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_produtos_codigo_fornecedor ON produtos(codigo_fornecedor);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS produtos CASCADE;
-- +goose StatementEnd
