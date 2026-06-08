-- +goose Up
ALTER TABLE sales ADD COLUMN observacoes_externas TEXT NOT NULL DEFAULT '';
ALTER TABLE sales ADD COLUMN observacoes_internas TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE sales DROP COLUMN observacoes_externas;
ALTER TABLE sales DROP COLUMN observacoes_internas;
