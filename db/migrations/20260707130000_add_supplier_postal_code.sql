-- +goose Up
ALTER TABLE suppliers ADD COLUMN IF NOT EXISTS postal_code VARCHAR(9) NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE suppliers DROP COLUMN IF EXISTS postal_code;
