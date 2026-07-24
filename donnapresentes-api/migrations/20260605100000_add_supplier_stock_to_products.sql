-- +goose Up
-- +goose StatementBegin
ALTER TABLE products
ADD COLUMN supplier_stock INTEGER DEFAULT 0 NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE products
DROP COLUMN IF EXISTS supplier_stock;
-- +goose StatementEnd
