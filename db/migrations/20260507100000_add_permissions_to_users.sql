-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS permissions TEXT[] NOT NULL DEFAULT '{}';

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS permissions;
