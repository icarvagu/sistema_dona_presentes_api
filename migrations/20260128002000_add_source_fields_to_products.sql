ALTER TABLE products
  ADD COLUMN source TEXT,
  ADD COLUMN imported_at TIMESTAMP WITH TIME ZONE,
  ADD COLUMN last_synced_at TIMESTAMP WITH TIME ZONE;

CREATE INDEX idx_products_source_imported_at
  ON products(source, imported_at);

