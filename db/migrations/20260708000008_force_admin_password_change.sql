-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE users SET must_change_password = TRUE WHERE username = 'donnapresentesadm';

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS must_change_password;
