CREATE TABLE IF NOT EXISTS financial_icms_parameters (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(160) NOT NULL,
    aliquota NUMERIC(8,4) NOT NULL CHECK (aliquota >= 0),
    credit_percent NUMERIC(8,4) NOT NULL DEFAULT 100 CHECK (credit_percent BETWEEN 0 AND 100),
    description TEXT NOT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_financial_icms_parameters_rate
    ON financial_icms_parameters(aliquota) WHERE active = TRUE;

CREATE TABLE IF NOT EXISTS financial_icms_entries (
    id BIGSERIAL PRIMARY KEY,
    purchase_id INTEGER REFERENCES purchase_orders(id) ON DELETE SET NULL,
    production_order_id BIGINT REFERENCES production_orders(id) ON DELETE SET NULL,
    supplier_id INTEGER REFERENCES suppliers(id) ON DELETE SET NULL,
    invoice_number VARCHAR(80) NOT NULL,
    invoice_key VARCHAR(60) NOT NULL DEFAULT '',
    issued_at TIMESTAMPTZ NOT NULL,
    tax_base NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (tax_base >= 0),
    aliquota NUMERIC(8,4) NOT NULL DEFAULT 0 CHECK (aliquota >= 0),
    icms_value NUMERIC(14,2) NOT NULL DEFAULT 0 CHECK (icms_value >= 0),
    xml_url TEXT NOT NULL DEFAULT '',
    source VARCHAR(20) NOT NULL DEFAULT 'xml',
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(invoice_number, invoice_key, supplier_id)
);

CREATE INDEX IF NOT EXISTS ix_financial_icms_entries_month
    ON financial_icms_entries(issued_at, aliquota);

INSERT INTO financial_icms_parameters (name, aliquota, credit_percent, description)
VALUES
    ('ICMS 18% · dentro do estado', 18, 100, 'Operação interna (SP).'),
    ('ICMS 12% · interestadual Sul/Sudeste', 12, 100, 'Origem Sul/Sudeste para SP.'),
    ('ICMS 7% · interestadual Norte/Nordeste/CO', 7, 100, 'Origem Norte, Nordeste ou Centro-Oeste.'),
    ('ICMS 4% · importado', 4, 100, 'Mercadoria com conteúdo de importação.'),
    ('Sem ICMS · Simples Nacional', 0, 0, 'Fornecedor do Simples, sem destaque.')
ON CONFLICT DO NOTHING;
