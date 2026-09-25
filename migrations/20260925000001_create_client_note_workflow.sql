CREATE TABLE IF NOT EXISTS client_note_workflows (
    id BIGSERIAL PRIMARY KEY,
    sale_id INTEGER NOT NULL UNIQUE REFERENCES sales(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'financeiro'
        CHECK (status IN ('financeiro', 'adm', 'enviadas')),
    sent_email TEXT NOT NULL DEFAULT '',
    sent_at TIMESTAMPTZ,
    updated_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS ix_client_note_workflows_status
    ON client_note_workflows(status, updated_at DESC);
