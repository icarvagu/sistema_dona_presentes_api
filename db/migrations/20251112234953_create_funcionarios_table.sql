-- +goose Up
CREATE TABLE funcionarios (
    id SERIAL PRIMARY KEY,
    nome_completo VARCHAR(255) NOT NULL,
    cpf VARCHAR(14) UNIQUE NOT NULL,
    rg VARCHAR(20),
    data_nascimento DATE,
    sexo VARCHAR(10) CHECK (sexo IN ('Masculino', 'Feminino', 'Outro')),
    situacao VARCHAR(10) NOT NULL CHECK (situacao IN ('Ativo', 'Inativo')),
    email_contato VARCHAR(255),
    endereco_completo TEXT,
    telefones_contato TEXT,
    observacoes TEXT,
    criado_em TIMESTAMP DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS funcionarios;
