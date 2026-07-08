-- +goose Up
ALTER TABLE users
ADD COLUMN IF NOT EXISTS failed_login_attempts INT NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS locked_until TIMESTAMPTZ;

-- +goose Down
ALTER TABLE users
DROP COLUMN IF EXISTS locked_until,
DROP COLUMN IF EXISTS failed_login_attempts;
