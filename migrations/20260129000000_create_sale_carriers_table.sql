-- +goose Up
-- Tabela de junção para relacionamento many-to-many entre sales e carriers
CREATE TABLE IF NOT EXISTS sale_carriers (
    id SERIAL PRIMARY KEY,
    sale_id INTEGER NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
    carrier_id INTEGER NOT NULL REFERENCES carriers(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(sale_id, carrier_id)
);

CREATE INDEX idx_sale_carriers_sale_id ON sale_carriers(sale_id);
CREATE INDEX idx_sale_carriers_carrier_id ON sale_carriers(carrier_id);

-- +goose Down
DROP INDEX IF EXISTS idx_sale_carriers_carrier_id;
DROP INDEX IF EXISTS idx_sale_carriers_sale_id;
DROP TABLE IF EXISTS sale_carriers;
