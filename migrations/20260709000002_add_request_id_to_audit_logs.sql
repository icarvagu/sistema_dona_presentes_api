-- +goose Up
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS request_id VARCHAR(32);

CREATE INDEX IF NOT EXISTS idx_audit_logs_request_id ON audit_logs(request_id);

-- +goose Down
ALTER TABLE audit_logs DROP COLUMN IF EXISTS request_id;
