-- +goose Up
-- +goose StatementBegin
CREATE TABLE clientes (
    id SERIAL PRIMARY KEY,
    tipo_cliente VARCHAR(20) NOT NULL CHECK (tipo_cliente IN ('PF', 'PJ')),
    situacao VARCHAR(20) NOT NULL CHECK (situacao IN ('Ativo', 'Inativo')),
    nome_empresa_pessoa TEXT NOT NULL,
    cnpj VARCHAR(18),
    cpf VARCHAR(14),
    email TEXT,
    telefone_comercial VARCHAR(20),
    celular VARCHAR(20),
    site TEXT,
    observacoes TEXT,
    criado_em TIMESTAMP WITH TIME ZONE DEFAULT now(),
    atualizado_em TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE TABLE enderecos_cliente (
    id SERIAL PRIMARY KEY,
    cliente_id INTEGER NOT NULL,
    tipo_endereco VARCHAR(20) NOT NULL CHECK (tipo_endereco IN ('comercial', 'entrega')),
    endereco TEXT NOT NULL,
    criado_em TIMESTAMP WITH TIME ZONE DEFAULT now(),
    atualizado_em TIMESTAMP WITH TIME ZONE DEFAULT now(),
    CONSTRAINT fk_cliente
      FOREIGN KEY(cliente_id)
        REFERENCES clientes(id)
        ON DELETE CASCADE
);

CREATE TABLE contatos_adicionais (
    id SERIAL PRIMARY KEY,
    cliente_id INTEGER NOT NULL,
    nome TEXT NOT NULL,
    email TEXT,
    telefone VARCHAR(20),
    criado_em TIMESTAMP WITH TIME ZONE DEFAULT now(),
    atualizado_em TIMESTAMP WITH TIME ZONE DEFAULT now(),
    CONSTRAINT fk_cliente
      FOREIGN KEY(cliente_id)
        REFERENCES clientes(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_clientes_tipo ON clientes(tipo_cliente);
CREATE INDEX idx_clientes_situacao ON clientes(situacao);
CREATE INDEX idx_enderecos_cliente_id ON enderecos_cliente(cliente_id);
CREATE INDEX idx_contatos_cliente_id ON contatos_adicionais(cliente_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS contatos_adicionais CASCADE;
DROP TABLE IF EXISTS enderecos_cliente CASCADE;
DROP TABLE IF EXISTS clientes CASCADE;
-- +goose StatementEnd
