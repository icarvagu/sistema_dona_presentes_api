CREATE TABLE art_final_tasks (
    id BIGSERIAL PRIMARY KEY,
    panel VARCHAR(20) NOT NULL CHECK (panel IN ('pending', 'media')),
    category VARCHAR(30) NOT NULL CHECK (category IN ('layout', 'alteracao', 'corel', 'gravacao', 'video', 'post', 'banner', 'mala')),
    title VARCHAR(180) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    sale_id INTEGER REFERENCES sales(id) ON DELETE CASCADE,
    purchase_id INTEGER REFERENCES purchase_orders(id) ON DELETE CASCADE,
    due_date DATE,
    priority SMALLINT NOT NULL DEFAULT 0 CHECK (priority BETWEEN 0 AND 3),
    tags TEXT[] NOT NULL DEFAULT '{}',
    attachment_url TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'done', 'cancelled')),
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    completed_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE art_final_stories (
    id BIGSERIAL PRIMARY KEY,
    sale_id INTEGER NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
    sale_item_id INTEGER REFERENCES sale_items(id) ON DELETE CASCADE,
    title VARCHAR(180) NOT NULL,
    tags TEXT[] NOT NULL DEFAULT '{}',
    checked BOOLEAN NOT NULL DEFAULT FALSE,
    checked_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    checked_at TIMESTAMPTZ,
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX ux_art_final_stories_sale_item ON art_final_stories(sale_item_id) WHERE sale_item_id IS NOT NULL;
CREATE UNIQUE INDEX ux_art_final_stories_sale ON art_final_stories(sale_id) WHERE sale_item_id IS NULL;
CREATE INDEX ix_art_final_tasks_queue ON art_final_tasks(panel, category, status, due_date, priority DESC);
CREATE INDEX ix_art_final_stories_checked ON art_final_stories(checked, created_at DESC);

CREATE TABLE art_final_audit_log (
    id BIGSERIAL PRIMARY KEY,
    entity_type VARCHAR(30) NOT NULL,
    entity_id BIGINT NOT NULL,
    action VARCHAR(40) NOT NULL,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX ix_art_final_audit_entity ON art_final_audit_log(entity_type, entity_id, created_at DESC);

