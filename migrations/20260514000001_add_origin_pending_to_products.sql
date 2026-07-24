-- +goose Up
-- +goose StatementBegin
ALTER TABLE products ADD COLUMN IF NOT EXISTS origin TEXT NOT NULL DEFAULT '';
ALTER TABLE products ADD COLUMN IF NOT EXISTS pending_approval BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE products DROP COLUMN IF EXISTS origin;
ALTER TABLE products DROP COLUMN IF EXISTS pending_approval;
-- +goose StatementEnd
