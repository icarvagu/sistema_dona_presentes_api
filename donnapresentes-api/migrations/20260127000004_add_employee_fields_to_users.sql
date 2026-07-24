-- +goose Up
-- +goose StatementBegin
-- Adicionar campos de Employee na tabela users
ALTER TABLE users 
ADD COLUMN full_name VARCHAR(255),
ADD COLUMN cpf VARCHAR(14) UNIQUE,
ADD COLUMN rg VARCHAR(20),
ADD COLUMN birth_date DATE,
ADD COLUMN gender VARCHAR(10) CHECK (gender IN ('Masculino', 'Feminino', 'Outro')),
ADD COLUMN status VARCHAR(10) DEFAULT 'Ativo' CHECK (status IN ('Ativo', 'Inativo')),
ADD COLUMN contact_email VARCHAR(255),
ADD COLUMN full_address TEXT,
ADD COLUMN contact_phone TEXT,
ADD COLUMN notes TEXT;

-- Criar índices
CREATE INDEX idx_users_cpf ON users(cpf);
CREATE INDEX idx_users_status ON users(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_users_cpf;
ALTER TABLE users 
DROP COLUMN IF EXISTS notes,
DROP COLUMN IF EXISTS contact_phone,
DROP COLUMN IF EXISTS full_address,
DROP COLUMN IF EXISTS contact_email,
DROP COLUMN IF EXISTS status,
DROP COLUMN IF EXISTS gender,
DROP COLUMN IF EXISTS birth_date,
DROP COLUMN IF EXISTS rg,
DROP COLUMN IF EXISTS cpf,
DROP COLUMN IF EXISTS full_name;
-- +goose StatementEnd
