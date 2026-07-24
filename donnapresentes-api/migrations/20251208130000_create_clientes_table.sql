-- +goose Up
-- +goose StatementBegin
CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    customer_type VARCHAR(20) NOT NULL CHECK (customer_type IN ('PF', 'PJ')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('Ativo', 'Inativo')),
    name TEXT NOT NULL,
    cnpj VARCHAR(18),
    cpf VARCHAR(14),
    email TEXT,
    business_phone VARCHAR(20),
    mobile_phone VARCHAR(20),
    website TEXT,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE TABLE customer_addresses (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL,
    address_type VARCHAR(20) NOT NULL CHECK (address_type IN ('comercial', 'entrega')),
    address TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    CONSTRAINT fk_customer
      FOREIGN KEY(customer_id)
        REFERENCES customers(id)
        ON DELETE CASCADE
);

CREATE TABLE additional_contacts (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    email TEXT,
    phone VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    CONSTRAINT fk_customer
      FOREIGN KEY(customer_id)
        REFERENCES customers(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_customers_type ON customers(customer_type);
CREATE INDEX idx_customers_status ON customers(status);
CREATE INDEX idx_customer_addresses_customer_id ON customer_addresses(customer_id);
CREATE INDEX idx_additional_contacts_customer_id ON additional_contacts(customer_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS additional_contacts CASCADE;
DROP TABLE IF EXISTS customer_addresses CASCADE;
DROP TABLE IF EXISTS customers CASCADE;
-- +goose StatementEnd
