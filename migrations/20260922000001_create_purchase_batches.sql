CREATE TABLE IF NOT EXISTS purchase_batches (
    id SERIAL PRIMARY KEY,
    supplier_id INTEGER REFERENCES suppliers(id) ON DELETE SET NULL,
    status VARCHAR(60) NOT NULL DEFAULT 'Aberta',
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS purchase_batch_orders (
    batch_id INTEGER NOT NULL REFERENCES purchase_batches(id) ON DELETE CASCADE,
    purchase_id INTEGER NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    PRIMARY KEY (batch_id, purchase_id)
);

CREATE TABLE IF NOT EXISTS purchase_batch_items (
    id SERIAL PRIMARY KEY,
    batch_id INTEGER NOT NULL REFERENCES purchase_batches(id) ON DELETE CASCADE,
    product_id INTEGER REFERENCES products(id) ON DELETE SET NULL,
    product_name TEXT NOT NULL,
    total_quantity NUMERIC(12, 3) NOT NULL CHECK (total_quantity > 0),
    allocations JSONB NOT NULL DEFAULT '[]'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_purchase_batch_orders_purchase ON purchase_batch_orders(purchase_id);
CREATE INDEX IF NOT EXISTS idx_purchase_batch_items_batch ON purchase_batch_items(batch_id);
