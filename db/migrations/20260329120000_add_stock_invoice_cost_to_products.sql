-- +goose Up
-- +goose StatementBegin
ALTER TABLE products
  ADD COLUMN IF NOT EXISTS moves_stock BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS enabled_for_invoice BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS cost_price NUMERIC(10,2) NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE products
  DROP COLUMN IF EXISTS moves_stock,
  DROP COLUMN IF EXISTS enabled_for_invoice,
  DROP COLUMN IF EXISTS cost_price;
-- +goose StatementEnd
