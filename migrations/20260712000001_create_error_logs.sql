CREATE TABLE IF NOT EXISTS error_logs (
    id BIGSERIAL PRIMARY KEY,
    request_id VARCHAR(32),
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    error_code INTEGER NOT NULL DEFAULT 500,
    error_message TEXT,
    stack_trace TEXT,
    path TEXT,
    method VARCHAR(10),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_error_logs_created_at ON error_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_error_logs_error_code ON error_logs(error_code);
CREATE INDEX IF NOT EXISTS idx_error_logs_request_id ON error_logs(request_id);

