CREATE TABLE IF NOT EXISTS purchase_requests (
    id SERIAL PRIMARY KEY,
    request_type VARCHAR(20) NOT NULL CHECK (request_type IN ('supply', 'internal')),
    product TEXT NOT NULL,
    product_link TEXT NOT NULL DEFAULT '',
    supplier TEXT NOT NULL DEFAULT '',
    quantity DOUBLE PRECISION,
    description TEXT NOT NULL DEFAULT '',
    attachments JSONB NOT NULL DEFAULT '[]'::jsonb,
    status VARCHAR(40) NOT NULL DEFAULT 'Aguardando compra',
    requested_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_purchase_requests_status ON purchase_requests(status);
CREATE INDEX IF NOT EXISTS idx_purchase_requests_type ON purchase_requests(request_type);
