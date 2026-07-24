-- +goose Up
-- +goose StatementBegin
ALTER TABLE products DROP CONSTRAINT IF EXISTS products_internal_code_key;
ALTER TABLE products ADD CONSTRAINT products_internal_code_color_key UNIQUE (internal_code, color);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE products DROP CONSTRAINT IF EXISTS products_internal_code_color_key;
ALTER TABLE products ADD CONSTRAINT products_internal_code_key UNIQUE (internal_code);
-- +goose StatementEnd
