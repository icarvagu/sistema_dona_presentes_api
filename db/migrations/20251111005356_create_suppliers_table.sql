-- +goose Up
CREATE TABLE fornecedores (
    id SERIAL PRIMARY KEY,
    nome_fantasia_ou_razao_social VARCHAR(255) NOT NULL,
    cnpj VARCHAR(18),
    inscricao_estadual VARCHAR(50),
    responsavel_atendimento VARCHAR(255),
    email_geral VARCHAR(255),
    telefone_fixo VARCHAR(20),
    celular VARCHAR(20),
    email_responsavel VARCHAR(255),
    endereco_comercial TEXT,
    criado_em TIMESTAMP DEFAULT NOW(),
    atualizado_em TIMESTAMP DEFAULT NOW()
);

-- Evita duplicidade de CNPJ para fornecedores PJ
CREATE UNIQUE INDEX idx_fornecedores_cnpj_unico
ON fornecedores (cnpj)
WHERE cnpj IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS fornecedores;
