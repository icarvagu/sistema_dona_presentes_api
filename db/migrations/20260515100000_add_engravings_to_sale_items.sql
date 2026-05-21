-- +goose Up
ALTER TABLE sale_items
  ADD COLUMN IF NOT EXISTS engravings JSONB DEFAULT '[]'::jsonb;

-- +goose Down
ALTER TABLE sale_items DROP COLUMN IF EXISTS engravings;
