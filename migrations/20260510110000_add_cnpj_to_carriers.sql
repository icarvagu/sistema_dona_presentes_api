-- +goose Up
ALTER TABLE carriers ADD COLUMN IF NOT EXISTS cnpj VARCHAR(20);

-- +goose Down
ALTER TABLE carriers DROP COLUMN IF EXISTS cnpj;
