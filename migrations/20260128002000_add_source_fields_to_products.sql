-- +goose Up
-- +goose StatementBegin
ALTER TABLE products
  ADD COLUMN source TEXT,
  ADD COLUMN imported_at TIMESTAMP WITH TIME ZONE,
  ADD COLUMN last_synced_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX idx_products_source_imported_at
  ON products(source, imported_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_products_source_imported_at;

ALTER TABLE products
  DROP COLUMN IF EXISTS last_synced_at,
  DROP COLUMN IF EXISTS imported_at,
  DROP COLUMN IF EXISTS source;
-- +goose StatementEnd

