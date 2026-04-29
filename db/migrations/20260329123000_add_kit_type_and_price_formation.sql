-- +goose Up
-- +goose StatementBegin
ALTER TABLE products
  ADD COLUMN IF NOT EXISTS kit_type TEXT NOT NULL DEFAULT 'none';

UPDATE products
SET kit_type = CASE
  WHEN is_composition = TRUE THEN 'internal_composition'
  ELSE 'none'
END
WHERE kit_type IS NULL OR kit_type = 'none';

ALTER TABLE products
  ADD CONSTRAINT products_kit_type_check
  CHECK (kit_type IN ('none', 'internal_composition', 'supplier_ready'));

CREATE TABLE IF NOT EXISTS product_price_formations (
  product_id INTEGER PRIMARY KEY REFERENCES products(id) ON DELETE CASCADE,
  taxes_percent NUMERIC(8,2) NOT NULL DEFAULT 0,
  overhead_percent NUMERIC(8,2) NOT NULL DEFAULT 0,
  commission_percent NUMERIC(8,2) NOT NULL DEFAULT 0,
  desired_margin_percent NUMERIC(8,2) NOT NULL DEFAULT 0,
  suggested_selling_price NUMERIC(12,2) NOT NULL DEFAULT 0,
  final_selling_price NUMERIC(12,2) NOT NULL DEFAULT 0,
  notes TEXT,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS product_price_formations;

ALTER TABLE products DROP CONSTRAINT IF EXISTS products_kit_type_check;
ALTER TABLE products DROP COLUMN IF EXISTS kit_type;
-- +goose StatementEnd
