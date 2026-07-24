-- +goose Up
-- +goose StatementBegin
ALTER TABLE quotes
  ADD COLUMN IF NOT EXISTS feedback_datetime TIMESTAMP WITH TIME ZONE NULL,
  ADD COLUMN IF NOT EXISTS feedback_observation TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE quotes
  DROP COLUMN IF EXISTS feedback_datetime,
  DROP COLUMN IF EXISTS feedback_observation;
-- +goose StatementEnd
