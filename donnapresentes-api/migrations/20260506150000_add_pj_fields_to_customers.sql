-- +goose Up
-- +goose StatementBegin
ALTER TABLE customers
  ADD COLUMN IF NOT EXISTS fantasy_name          TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS razao_social           TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS inscricao_estadual     TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS inscricao_municipal    TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS responsavel            TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS contact_financial_name  TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS contact_financial_email TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS contact_financial_phone TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS contact_nf_name         TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS contact_nf_email        TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS contact_nf_phone        TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS contact_commercial_name  TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS contact_commercial_email TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS contact_commercial_phone TEXT NOT NULL DEFAULT '';

ALTER TABLE customer_addresses
  DROP CONSTRAINT IF EXISTS customer_addresses_address_type_check;

ALTER TABLE customer_addresses
  ADD CONSTRAINT customer_addresses_address_type_check
    CHECK (address_type IN ('comercial', 'entrega', 'cobrança'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE customers
  DROP COLUMN IF EXISTS fantasy_name,
  DROP COLUMN IF EXISTS razao_social,
  DROP COLUMN IF EXISTS inscricao_estadual,
  DROP COLUMN IF EXISTS inscricao_municipal,
  DROP COLUMN IF EXISTS responsavel,
  DROP COLUMN IF EXISTS contact_financial_name,
  DROP COLUMN IF EXISTS contact_financial_email,
  DROP COLUMN IF EXISTS contact_financial_phone,
  DROP COLUMN IF EXISTS contact_nf_name,
  DROP COLUMN IF EXISTS contact_nf_email,
  DROP COLUMN IF EXISTS contact_nf_phone,
  DROP COLUMN IF EXISTS contact_commercial_name,
  DROP COLUMN IF EXISTS contact_commercial_email,
  DROP COLUMN IF EXISTS contact_commercial_phone;
-- +goose StatementEnd
