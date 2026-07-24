-- +goose Up
ALTER TABLE quotes
  ADD COLUMN IF NOT EXISTS installments INTEGER NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS installment_dates JSONB NOT NULL DEFAULT '[]'::jsonb;

-- +goose Down
ALTER TABLE quotes
  DROP COLUMN IF EXISTS installment_dates,
  DROP COLUMN IF EXISTS installments;
