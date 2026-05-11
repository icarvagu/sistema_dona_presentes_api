-- +goose Up
-- +goose StatementBegin
ALTER TABLE sales
  ADD COLUMN IF NOT EXISTS status VARCHAR(50) NOT NULL DEFAULT 'Pendente';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sales DROP COLUMN IF EXISTS status;
-- +goose StatementEnd
