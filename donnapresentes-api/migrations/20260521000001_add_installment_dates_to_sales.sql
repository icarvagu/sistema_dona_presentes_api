-- +goose Up
ALTER TABLE sales ADD COLUMN installment_dates JSONB;

-- +goose Down
ALTER TABLE sales DROP COLUMN installment_dates;
