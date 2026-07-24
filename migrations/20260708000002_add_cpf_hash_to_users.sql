-- +goose Up
ALTER TABLE users
ADD COLUMN IF NOT EXISTS cpf_hash VARCHAR(64);

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS cpf_hash;
