-- +goose Up
ALTER TABLE purchase_orders
    ADD COLUMN IF NOT EXISTS buyer_discount NUMERIC(12,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS negotiation_contact TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS negotiation_notes TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE purchase_orders
    DROP COLUMN IF EXISTS negotiation_notes,
    DROP COLUMN IF EXISTS negotiation_contact,
    DROP COLUMN IF EXISTS buyer_discount;
