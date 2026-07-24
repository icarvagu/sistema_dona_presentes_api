ALTER TABLE purchase_orders
    ADD COLUMN IF NOT EXISTS status_updated_at TIMESTAMPTZ;

UPDATE purchase_orders po
SET status_updated_at = COALESCE(
    (SELECT MAX(h.created_at) FROM purchase_history h WHERE h.purchase_id = po.id AND h.to_status = po.status),
    po.updated_at,
    po.created_at
)
WHERE status_updated_at IS NULL;

ALTER TABLE purchase_orders
    ALTER COLUMN status_updated_at SET NOT NULL,
    ALTER COLUMN status_updated_at SET DEFAULT NOW();

