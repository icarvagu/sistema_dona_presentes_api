-- +goose Up
ALTER TABLE products
  ADD COLUMN IF NOT EXISTS last_cost_val1 NUMERIC(10,2) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS last_cost_val2 NUMERIC(10,2) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS last_cost_val3 NUMERIC(10,2) NOT NULL DEFAULT 0;

UPDATE products
SET last_cost_val1 = last_cost
WHERE last_cost_val1 = 0
  AND last_cost > 0;

-- +goose Down
ALTER TABLE products
  DROP COLUMN IF EXISTS last_cost_val1,
  DROP COLUMN IF EXISTS last_cost_val2,
  DROP COLUMN IF EXISTS last_cost_val3;
