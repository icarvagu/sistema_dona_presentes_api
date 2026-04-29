-- +goose Up
-- +goose StatementBegin
ALTER TABLE products ADD COLUMN selling_price DECIMAL(10,2) DEFAULT 0.00;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE products DROP COLUMN IF EXISTS selling_price;
-- +goose StatementEnd
