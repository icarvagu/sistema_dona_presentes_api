-- +goose Up
ALTER TABLE sales RENAME COLUMN observacoes_externas TO external_notes;
ALTER TABLE sales RENAME COLUMN observacoes_internas TO internal_notes;

-- +goose Down
ALTER TABLE sales RENAME COLUMN external_notes TO observacoes_externas;
ALTER TABLE sales RENAME COLUMN internal_notes TO observacoes_internas;
