-- +goose Up
-- +goose StatementBegin
ALTER TABLE products ADD COLUMN IF NOT EXISTS supplier_code TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE products DROP COLUMN IF EXISTS supplier_code;
-- +goose StatementEnd
