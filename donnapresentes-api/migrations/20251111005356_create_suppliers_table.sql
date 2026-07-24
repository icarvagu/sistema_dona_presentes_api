-- +goose Up
CREATE TABLE suppliers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    cnpj VARCHAR(18),
    state_registration VARCHAR(50),
    contact_person VARCHAR(255),
    email VARCHAR(255),
    landline_phone VARCHAR(20),
    mobile_phone VARCHAR(20),
    responsible_email VARCHAR(255),
    commercial_address TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Evita duplicidade de CNPJ para fornecedores PJ
CREATE UNIQUE INDEX idx_suppliers_cnpj_unique
ON suppliers (cnpj)
WHERE cnpj IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS suppliers;
