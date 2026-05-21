-- +goose Up
-- +goose StatementBegin
ALTER TABLE products ADD COLUMN IF NOT EXISTS color TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE products DROP COLUMN IF EXISTS color;
-- +goose StatementEnd
