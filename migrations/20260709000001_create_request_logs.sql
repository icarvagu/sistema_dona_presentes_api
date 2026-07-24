CREATE TABLE IF NOT EXISTS request_logs (
    id BIGSERIAL PRIMARY KEY,
    request_id VARCHAR(32) NOT NULL,
    method VARCHAR(10) NOT NULL,
    path VARCHAR(512) NOT NULL,
    status INTEGER NOT NULL,
    duration_ms INTEGER NOT NULL,
    remote_addr VARCHAR(45),
    user_agent VARCHAR(512),
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    slow BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_request_logs_created_at ON request_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_request_logs_request_id ON request_logs(request_id);
CREATE INDEX IF NOT EXISTS idx_request_logs_status ON request_logs(status);
CREATE INDEX IF NOT EXISTS idx_request_logs_slow ON request_logs(slow) WHERE slow = TRUE;

