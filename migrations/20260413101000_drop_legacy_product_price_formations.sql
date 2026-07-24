-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS product_price_formations;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
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
