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

