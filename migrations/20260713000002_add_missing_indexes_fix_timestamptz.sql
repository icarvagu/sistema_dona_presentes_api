
-- ============================================================
-- Missing indexes (critical for query performance)
-- ============================================================

-- carriers: zero indexes — every lookup by name/cnpj/type requires full scan
CREATE INDEX IF NOT EXISTS idx_carriers_name ON carriers(name);
CREATE INDEX IF NOT EXISTS idx_carriers_cnpj ON carriers(cnpj) WHERE cnpj IS NOT NULL;

-- purchase_attachments: FK lookup by purchase_id was unindexed
CREATE INDEX IF NOT EXISTS idx_purchase_attachments_purchase ON purchase_attachments(purchase_id);

-- purchase_emails: FK lookup by purchase_id was unindexed
CREATE INDEX IF NOT EXISTS idx_purchase_emails_purchase ON purchase_emails(purchase_id);

-- sale_payment_receipts: FK lookup by sale_id was unindexed
CREATE INDEX IF NOT EXISTS idx_sale_payment_receipts_sale ON sale_payment_receipts(sale_id);

-- financial_analysis_events: FK lookup by sale_id was unindexed
CREATE INDEX IF NOT EXISTS idx_financial_analysis_events_sale ON financial_analysis_events(sale_id);

-- layout_approval_events: FK lookup by layout_version_id was unindexed
CREATE INDEX IF NOT EXISTS idx_layout_approval_events_version ON layout_approval_events(layout_version_id);

-- quotes: FK lookup by carrier_id was unindexed
CREATE INDEX IF NOT EXISTS idx_quotes_carrier ON quotes(carrier_id) WHERE carrier_id IS NOT NULL;

-- ============================================================
-- Standardize TIMESTAMP → TIMESTAMPTZ across tables
-- ============================================================

ALTER TABLE suppliers
    ALTER COLUMN created_at TYPE TIMESTAMPTZ,
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ;

ALTER TABLE users
    ALTER COLUMN created_at TYPE TIMESTAMPTZ,
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ;

ALTER TABLE product_items
    ALTER COLUMN created_at TYPE TIMESTAMPTZ,
    ALTER COLUMN updated_at TYPE TIMESTAMPTZ;


