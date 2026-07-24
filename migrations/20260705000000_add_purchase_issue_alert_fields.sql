ALTER TABLE purchase_issues
    ADD COLUMN IF NOT EXISTS occurrence_date DATE,
    ADD COLUMN IF NOT EXISTS priority SMALLINT NOT NULL DEFAULT 1 CHECK (priority BETWEEN 1 AND 3);

UPDATE purchase_issues
SET occurrence_date = created_at::date
WHERE occurrence_date IS NULL;

ALTER TABLE purchase_issues
    ALTER COLUMN occurrence_date SET NOT NULL,
    ALTER COLUMN occurrence_date SET DEFAULT CURRENT_DATE;

CREATE INDEX IF NOT EXISTS idx_purchase_issues_alert_queue
    ON purchase_issues(status, priority DESC, resolution_deadline ASC);

