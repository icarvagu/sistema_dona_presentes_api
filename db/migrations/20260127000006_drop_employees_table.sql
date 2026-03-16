-- +goose Up
-- +goose StatementBegin
-- Remover tabela employees (após migração de dados)
DROP TABLE IF EXISTS employees CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Recriar tabela employees (se necessário para rollback)
CREATE TABLE employees (
    id SERIAL PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL,
    cpf VARCHAR(14) UNIQUE NOT NULL,
    rg VARCHAR(20),
    birth_date DATE,
    gender VARCHAR(10) CHECK (gender IN ('Masculino', 'Feminino', 'Outro')),
    status VARCHAR(10) NOT NULL CHECK (status IN ('Ativo', 'Inativo')),
    contact_email VARCHAR(255),
    full_address TEXT,
    contact_phone TEXT,
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
-- +goose StatementEnd
