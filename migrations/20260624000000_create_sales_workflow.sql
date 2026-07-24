ALTER TABLE quotes
  ADD COLUMN IF NOT EXISTS important BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS converted_sale_id INTEGER NULL;

ALTER TABLE sales
  ADD COLUMN IF NOT EXISTS quote_id INTEGER NULL,
  ADD COLUMN IF NOT EXISTS seller_approved_at TIMESTAMPTZ NULL,
  ADD COLUMN IF NOT EXISTS seller_approved_by INTEGER NULL,
  ADD COLUMN IF NOT EXISTS released_to_purchases_at TIMESTAMPTZ NULL;

CREATE UNIQUE INDEX IF NOT EXISTS ux_sales_quote_id ON sales(quote_id) WHERE quote_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS ix_quotes_converted_sale_id ON quotes(converted_sale_id) WHERE converted_sale_id IS NOT NULL;

ALTER TABLE sale_items
  ADD COLUMN IF NOT EXISTS quote_item_id INTEGER NULL,
  ADD COLUMN IF NOT EXISTS personalization_type TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS engraving_withdrawal_date DATE NULL;
CREATE INDEX IF NOT EXISTS ix_sale_items_quote_item_id ON sale_items(quote_item_id) WHERE quote_item_id IS NOT NULL;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_quotes_converted_sale_id') THEN
    ALTER TABLE quotes ADD CONSTRAINT fk_quotes_converted_sale_id FOREIGN KEY (converted_sale_id) REFERENCES sales(id) ON DELETE SET NULL NOT VALID;
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_sales_quote_id') THEN
    ALTER TABLE sales ADD CONSTRAINT fk_sales_quote_id FOREIGN KEY (quote_id) REFERENCES quotes(id) ON DELETE SET NULL NOT VALID;
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_sales_seller_approved_by') THEN
    ALTER TABLE sales ADD CONSTRAINT fk_sales_seller_approved_by FOREIGN KEY (seller_approved_by) REFERENCES users(id) ON DELETE SET NULL NOT VALID;
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_sale_items_quote_item_id') THEN
    ALTER TABLE sale_items ADD CONSTRAINT fk_sale_items_quote_item_id FOREIGN KEY (quote_item_id) REFERENCES quote_items(id) ON DELETE SET NULL NOT VALID;
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS quote_feedback_events (
  id BIGSERIAL PRIMARY KEY,
  quote_id INTEGER NOT NULL REFERENCES quotes(id) ON DELETE CASCADE,
  scheduled_at TIMESTAMPTZ NULL,
  observation TEXT NOT NULL,
  created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS ix_quote_feedback_events_quote ON quote_feedback_events(quote_id, created_at DESC);

CREATE TABLE IF NOT EXISTS item_layout_versions (
  id BIGSERIAL PRIMARY KEY,
  entity_type TEXT NOT NULL CHECK (entity_type IN ('quote_item','sale_item')),
  item_id INTEGER NOT NULL,
  version INTEGER NOT NULL,
  label TEXT NOT NULL,
  file_url TEXT NOT NULL,
  item_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
  approval_status TEXT NOT NULL DEFAULT 'pending' CHECK (approval_status IN ('pending','approved','changes_requested','rejected')),
  approval_note TEXT NOT NULL DEFAULT '',
  approved_by INTEGER NULL REFERENCES users(id) ON DELETE RESTRICT,
  approved_at TIMESTAMPTZ NULL,
  created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(entity_type, item_id, version)
);
CREATE INDEX IF NOT EXISTS ix_item_layout_versions_item ON item_layout_versions(entity_type, item_id, version DESC);

CREATE TABLE IF NOT EXISTS financial_analyses (
  id BIGSERIAL PRIMARY KEY,
  sale_id INTEGER NOT NULL UNIQUE REFERENCES sales(id) ON DELETE CASCADE,
  status TEXT NOT NULL CHECK (status IN ('pending','approved','rejected')),
  tags TEXT[] NOT NULL DEFAULT '{}',
  observation TEXT NOT NULL DEFAULT '',
  attachment_url TEXT NOT NULL DEFAULT '',
  analyzed_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  analyzed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS financial_analysis_events (
  id BIGSERIAL PRIMARY KEY,
  sale_id INTEGER NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
  status TEXT NOT NULL,
  tags TEXT[] NOT NULL DEFAULT '{}',
  observation TEXT NOT NULL DEFAULT '',
  attachment_url TEXT NOT NULL DEFAULT '',
  created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS layout_approval_events (
  id BIGSERIAL PRIMARY KEY,
  layout_version_id BIGINT NOT NULL REFERENCES item_layout_versions(id) ON DELETE CASCADE,
  status TEXT NOT NULL,
  note TEXT NOT NULL DEFAULT '',
  created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sale_payment_receipts (
  id BIGSERIAL PRIMARY KEY,
  sale_id INTEGER NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
  file_url TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','validated','rejected')),
  observation TEXT NOT NULL DEFAULT '',
  uploaded_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  validated_by INTEGER NULL REFERENCES users(id) ON DELETE RESTRICT,
  validated_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sale_pending_events (
  id BIGSERIAL PRIMARY KEY,
  sale_id INTEGER NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
  sector TEXT NOT NULL,
  description TEXT NOT NULL,
  blocking BOOLEAN NOT NULL DEFAULT TRUE,
  resolved_at TIMESTAMPTZ NULL,
  resolved_by INTEGER NULL REFERENCES users(id) ON DELETE RESTRICT,
  resolution_note TEXT NOT NULL DEFAULT '',
  created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS ix_sale_pending_events_open ON sale_pending_events(sale_id, sector) WHERE resolved_at IS NULL;

CREATE TABLE IF NOT EXISTS engraving_approvals (
  id BIGSERIAL PRIMARY KEY,
  sale_item_id INTEGER NOT NULL REFERENCES sale_items(id) ON DELETE CASCADE,
  response TEXT NOT NULL CHECK (response IN ('APROVADO','ERRADO','COR_DIFERENTE','NOVA_FOTO')),
  observation TEXT NOT NULL DEFAULT '',
  responded_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  responded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  channel TEXT NULL,
  recorded_by INTEGER NULL REFERENCES users(id) ON DELETE RESTRICT,
  recorded_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS seller_monthly_targets (
  id BIGSERIAL PRIMARY KEY,
  seller_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  month DATE NOT NULL,
  target_value NUMERIC(14,2) NOT NULL CHECK (target_value >= 0),
  created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(seller_id, month)
);

INSERT INTO quote_feedback_events (quote_id, scheduled_at, observation, created_by, created_at)
SELECT q.id, q.feedback_datetime, q.feedback_observation, q.seller_id, COALESCE(q.updated_at, now())
FROM quotes q
WHERE (q.feedback_datetime IS NOT NULL OR q.feedback_observation <> '')
  AND NOT EXISTS (SELECT 1 FROM quote_feedback_events e WHERE e.quote_id = q.id);

