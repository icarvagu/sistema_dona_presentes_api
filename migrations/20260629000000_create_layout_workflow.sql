-- +goose Up
-- +goose StatementBegin
CREATE TABLE layout_requests (
    id BIGSERIAL PRIMARY KEY,
    source_type VARCHAR(20) NOT NULL CHECK (source_type IN ('quote','sale')),
    source_id INTEGER NOT NULL,
    title VARCHAR(180) NOT NULL,
    instructions TEXT NOT NULL DEFAULT '',
    file_mode VARCHAR(10) NOT NULL DEFAULT 'separate' CHECK (file_mode IN ('common','separate')),
    common_file_url TEXT NOT NULL DEFAULT '',
    status VARCHAR(30) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','requested','in_progress','awaiting_approval','changes_requested','approved','cancelled')),
    priority SMALLINT NOT NULL DEFAULT 1 CHECK (priority BETWEEN 0 AND 3),
    due_at TIMESTAMPTZ,
    assigned_to INTEGER REFERENCES users(id) ON DELETE SET NULL,
    requested_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    source_snapshot JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX ix_layout_requests_queue ON layout_requests(status,due_at,priority DESC);
CREATE INDEX ix_layout_requests_source ON layout_requests(source_type,source_id);

ALTER TABLE art_final_tasks
    ADD COLUMN due_at TIMESTAMPTZ,
    ADD COLUMN assigned_to INTEGER REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN channel VARCHAR(60) NOT NULL DEFAULT '',
    ADD COLUMN format VARCHAR(60) NOT NULL DEFAULT '';
UPDATE art_final_tasks SET due_at=due_date::timestamp WHERE due_date IS NOT NULL AND due_at IS NULL;

CREATE TABLE layout_request_items (
    id BIGSERIAL PRIMARY KEY,
    request_id BIGINT NOT NULL REFERENCES layout_requests(id) ON DELETE CASCADE,
    entity_type VARCHAR(20) NOT NULL CHECK (entity_type IN ('quote_item','sale_item')),
    item_id INTEGER NOT NULL,
    group_key VARCHAR(80) NOT NULL DEFAULT '',
    source_file_url TEXT NOT NULL DEFAULT '',
    item_snapshot JSONB NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'requested' CHECK (status IN ('requested','in_progress','awaiting_approval','changes_requested','approved','cancelled')),
    product_received_at TIMESTAMPTZ,
    product_received_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(request_id,entity_type,item_id)
);
CREATE INDEX ix_layout_request_items_request ON layout_request_items(request_id,id);
CREATE INDEX ix_layout_request_items_source ON layout_request_items(entity_type,item_id);

CREATE TABLE layout_request_groups (
    id BIGSERIAL PRIMARY KEY,
    request_id BIGINT NOT NULL REFERENCES layout_requests(id) ON DELETE CASCADE,
    group_key VARCHAR(80) NOT NULL,
    file_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(request_id,group_key)
);

ALTER TABLE item_layout_versions
    ADD COLUMN request_item_id BIGINT REFERENCES layout_request_items(id) ON DELETE SET NULL;
CREATE INDEX ix_item_layout_versions_request_item ON item_layout_versions(request_item_id,version DESC);

CREATE TABLE layout_corel_jobs (
    id BIGSERIAL PRIMARY KEY,
    request_item_id BIGINT NOT NULL UNIQUE REFERENCES layout_request_items(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','sent','received','cancelled')),
    responsible_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    external_contact VARCHAR(180) NOT NULL DEFAULT '',
    due_at TIMESTAMPTZ,
    file_url TEXT NOT NULL DEFAULT '',
    tags TEXT[] NOT NULL DEFAULT '{}',
    reason TEXT NOT NULL DEFAULT '',
    sent_at TIMESTAMPTZ,
    received_at TIMESTAMPTZ,
    updated_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE layout_engraving_jobs (
    id BIGSERIAL PRIMARY KEY,
    request_item_id BIGINT NOT NULL UNIQUE REFERENCES layout_request_items(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','in_progress','ready','approved','cancelled')),
    responsible_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    due_at TIMESTAMPTZ,
    file_url TEXT NOT NULL DEFAULT '',
    tags TEXT[] NOT NULL DEFAULT '{}',
    reason TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    updated_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE layout_job_versions (
    id BIGSERIAL PRIMARY KEY,
    kind VARCHAR(20) NOT NULL CHECK (kind IN ('corel','engraving')),
    request_item_id BIGINT NOT NULL REFERENCES layout_request_items(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL,
    responsible_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    external_contact VARCHAR(180) NOT NULL DEFAULT '',
    due_at TIMESTAMPTZ,
    file_url TEXT NOT NULL DEFAULT '',
    tags TEXT[] NOT NULL DEFAULT '{}',
    reason TEXT NOT NULL DEFAULT '',
    created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(kind,request_item_id,version)
);
CREATE INDEX ix_layout_job_versions_item ON layout_job_versions(kind,request_item_id,version DESC);

ALTER TABLE sale_items
    ADD COLUMN product_received_at TIMESTAMPTZ,
    ADD COLUMN product_received_by INTEGER REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE art_final_stories
    ADD COLUMN lifecycle_status VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (lifecycle_status IN ('draft','returned','corrected','published','expired')),
    ADD COLUMN observation TEXT NOT NULL DEFAULT '',
    ADD COLUMN published_at TIMESTAMPTZ,
    ADD COLUMN expires_at TIMESTAMPTZ,
    ADD COLUMN updated_by INTEGER REFERENCES users(id) ON DELETE SET NULL;
UPDATE art_final_stories SET lifecycle_status=CASE WHEN checked THEN 'published' ELSE 'draft' END,
    published_at=CASE WHEN checked THEN checked_at ELSE NULL END;

CREATE OR REPLACE VIEW art_final_layout_agenda AS
SELECT r.id request_id,i.id request_item_id,r.title,r.status request_status,i.status item_status,
       r.priority,r.due_at,r.assigned_to,r.source_type,r.source_id,i.entity_type,i.item_id,
       i.group_key,i.product_received_at,
       c.status corel_status,c.due_at corel_due_at,
       e.status engraving_status,e.due_at engraving_due_at,
       (SELECT v.file_url FROM item_layout_versions v WHERE v.request_item_id=i.id ORDER BY v.version DESC LIMIT 1) latest_file_url
FROM layout_requests r JOIN layout_request_items i ON i.request_id=r.id
LEFT JOIN layout_corel_jobs c ON c.request_item_id=i.id
LEFT JOIN layout_engraving_jobs e ON e.request_item_id=i.id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS art_final_layout_agenda;
ALTER TABLE sale_items DROP COLUMN IF EXISTS product_received_by,DROP COLUMN IF EXISTS product_received_at;
DROP TABLE IF EXISTS layout_job_versions;
ALTER TABLE art_final_tasks DROP COLUMN IF EXISTS format,DROP COLUMN IF EXISTS channel,DROP COLUMN IF EXISTS assigned_to,DROP COLUMN IF EXISTS due_at;
ALTER TABLE art_final_stories DROP COLUMN IF EXISTS updated_by,DROP COLUMN IF EXISTS expires_at,DROP COLUMN IF EXISTS published_at,DROP COLUMN IF EXISTS observation,DROP COLUMN IF EXISTS lifecycle_status;
DROP TABLE IF EXISTS layout_engraving_jobs;
DROP TABLE IF EXISTS layout_corel_jobs;
ALTER TABLE item_layout_versions DROP COLUMN IF EXISTS request_item_id;
DROP TABLE IF EXISTS layout_request_groups;
DROP TABLE IF EXISTS layout_request_items;
DROP TABLE IF EXISTS layout_requests;
-- +goose StatementEnd
