-- +goose Up
-- +goose StatementBegin
ALTER TABLE sales ADD COLUMN IF NOT EXISTS is_event BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE sales ADD COLUMN IF NOT EXISTS delivery_address TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sales DROP COLUMN IF EXISTS is_event;
ALTER TABLE sales DROP COLUMN IF EXISTS delivery_address;
-- +goose StatementEnd
