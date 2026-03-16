-- +goose Up
-- +goose StatementBegin
ALTER TABLE sales
  ADD COLUMN customer_id INTEGER,
  ADD CONSTRAINT fk_sales_customer
    FOREIGN KEY (customer_id)
    REFERENCES customers(id)
    ON DELETE RESTRICT;

CREATE INDEX idx_sales_customer_id ON sales(customer_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sales
  DROP CONSTRAINT IF EXISTS fk_sales_customer,
  DROP COLUMN IF EXISTS customer_id;
DROP INDEX IF EXISTS idx_sales_customer_id;
-- +goose StatementEnd

