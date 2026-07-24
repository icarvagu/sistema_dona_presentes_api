ALTER TABLE production_orders
    ADD COLUMN conference_started_at TIMESTAMPTZ,
    ADD COLUMN conference_completed_at TIMESTAMPTZ,
    ADD COLUMN conference_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE production_supply_movements
    ADD COLUMN supplier_id INTEGER REFERENCES suppliers(id) ON DELETE SET NULL,
    ADD COLUMN invoice_number VARCHAR(80) NOT NULL DEFAULT '',
    ADD COLUMN invoice_url TEXT NOT NULL DEFAULT '';

