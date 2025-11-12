-- +goose Up
CREATE TABLE transportadoras (
    id SERIAL PRIMARY KEY,
    nome_transportadora VARCHAR(255) NOT NULL,
    tipo_transportadora VARCHAR(20) NOT NULL CHECK (tipo_transportadora IN ('Pessoa Jurídica', 'Pessoa Física')),
    email VARCHAR(255),
    telefone_fixo VARCHAR(20),
    celular VARCHAR(20),
    endereco_completo TEXT,
    contato_principal_nome VARCHAR(255),
    contato_principal_telefone VARCHAR(20),
    site VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS transportadoras;
