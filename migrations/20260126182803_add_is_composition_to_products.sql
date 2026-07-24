-- +goose Up
-- +goose StatementBegin
ALTER TABLE products ADD COLUMN is_composition BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE products DROP COLUMN IF EXISTS is_composition;
-- +goose StatementEnd
