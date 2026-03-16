-- +goose Up
-- +goose StatementBegin
ALTER TABLE sales ADD COLUMN total_value NUMERIC(10, 2) DEFAULT 0.00;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sales DROP COLUMN IF EXISTS total_value;
-- +goose StatementEnd
