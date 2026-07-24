-- +goose Up
CREATE TABLE carriers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    carrier_type VARCHAR(20) NOT NULL CHECK (carrier_type IN ('Pessoa Jurídica', 'Pessoa Física')),
    email VARCHAR(255),
    landline_phone VARCHAR(20),
    mobile_phone VARCHAR(20),
    full_address TEXT,
    contact_name VARCHAR(255),
    contact_phone VARCHAR(20),
    website VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS carriers;
