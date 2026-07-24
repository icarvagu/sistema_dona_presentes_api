-- +goose Up
ALTER TABLE suppliers ADD COLUMN IF NOT EXISTS website TEXT;

-- +goose Down
ALTER TABLE suppliers DROP COLUMN IF EXISTS website;
