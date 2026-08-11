CREATE TABLE IF NOT EXISTS workflow_alerts (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    sale_id INTEGER REFERENCES sales(id) ON DELETE SET NULL,
    quote_id INTEGER REFERENCES quotes(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    scheduled_at TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL DEFAULT 'scheduled',
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT workflow_alerts_status_check CHECK (status IN ('scheduled', 'done')),
    CONSTRAINT workflow_alerts_single_context_check CHECK (sale_id IS NULL OR quote_id IS NULL)
);

CREATE INDEX IF NOT EXISTS idx_workflow_alerts_user_scheduled
    ON workflow_alerts (user_id, scheduled_at);

CREATE INDEX IF NOT EXISTS idx_workflow_alerts_sale
    ON workflow_alerts (sale_id)
    WHERE sale_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_workflow_alerts_quote
    ON workflow_alerts (quote_id)
    WHERE quote_id IS NOT NULL;
