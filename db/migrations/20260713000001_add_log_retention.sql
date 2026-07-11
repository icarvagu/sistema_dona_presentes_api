-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION cleanup_old_logs() RETURNS void AS $$
BEGIN
    DELETE FROM error_logs WHERE created_at < NOW() - INTERVAL '2 months';
    DELETE FROM audit_logs WHERE created_at < NOW() - INTERVAL '2 months';
    DELETE FROM request_logs WHERE created_at < NOW() - INTERVAL '2 months';
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION IF EXISTS cleanup_old_logs();
