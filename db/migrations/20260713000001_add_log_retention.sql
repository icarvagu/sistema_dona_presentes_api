-- +goose Up
-- Add scheduled cleanup for log tables: keep only last 2 months

CREATE OR REPLACE FUNCTION cleanup_old_logs() RETURNS void AS $func$
BEGIN
    DELETE FROM error_logs WHERE created_at < NOW() - INTERVAL '2 months';
    DELETE FROM audit_logs WHERE created_at < NOW() - INTERVAL '2 months';
    DELETE FROM request_logs WHERE created_at < NOW() - INTERVAL '2 months';
END;
$func$ LANGUAGE plpgsql;

-- +goose Down
DROP FUNCTION IF EXISTS cleanup_old_logs();
