CREATE TABLE purchase_orders (
    id SERIAL PRIMARY KEY,
    sale_id INTEGER NOT NULL UNIQUE REFERENCES sales(id) ON DELETE CASCADE,
    general_number VARCHAR(80) NOT NULL UNIQUE,
    status VARCHAR(80) NOT NULL DEFAULT 'Pendente de Compra',
    buyer_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    material_supplier_id INTEGER REFERENCES suppliers(id) ON DELETE SET NULL,
    engraving_supplier_id INTEGER REFERENCES suppliers(id) ON DELETE SET NULL,
    is_sample BOOLEAN NOT NULL DEFAULT FALSE,
    sample_has_engraving BOOLEAN NOT NULL DEFAULT FALSE,
    has_engraving BOOLEAN NOT NULL DEFAULT FALSE,
    corel_required BOOLEAN NOT NULL DEFAULT FALSE,
    corel_requested_at TIMESTAMPTZ,
    corel_requested_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    corel_attached_at TIMESTAMPTZ,
    corel_attached_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    material_unit_cost NUMERIC(12,2) NOT NULL DEFAULT 0,
    material_total_cost NUMERIC(12,2) NOT NULL DEFAULT 0,
    engraving_cost NUMERIC(12,2) NOT NULL DEFAULT 0,
    freight_cost NUMERIC(12,2) NOT NULL DEFAULT 0,
    other_cost NUMERIC(12,2) NOT NULL DEFAULT 0,
    material_deadline DATE,
    engraving_deadline DATE,
    payment_method VARCHAR(60) NOT NULL DEFAULT '',
    requires_advance_payment BOOLEAN NOT NULL DEFAULT FALSE,
    material_accepted BOOLEAN NOT NULL DEFAULT FALSE,
    material_accepted_at TIMESTAMPTZ,
    engraving_accepted BOOLEAN NOT NULL DEFAULT FALSE,
    engraving_accepted_at TIMESTAMPTZ,
    first_piece_required BOOLEAN NOT NULL DEFAULT FALSE,
    first_piece_status VARCHAR(80) NOT NULL DEFAULT '',
    first_piece_url TEXT NOT NULL DEFAULT '',
    commercial_notes TEXT NOT NULL DEFAULT '',
    purchase_notes TEXT NOT NULL DEFAULT '',
    production_released_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE purchase_attachments (
    id SERIAL PRIMARY KEY,
    purchase_id INTEGER NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    category VARCHAR(40) NOT NULL,
    file_name TEXT NOT NULL,
    url TEXT NOT NULL,
    uploaded_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE purchase_emails (
    id SERIAL PRIMARY KEY,
    purchase_id INTEGER NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    kind VARCHAR(30) NOT NULL CHECK (kind IN ('material', 'gravacao')),
    recipient TEXT NOT NULL,
    subject TEXT NOT NULL,
    body TEXT NOT NULL,
    observation TEXT NOT NULL DEFAULT '',
    attachments JSONB NOT NULL DEFAULT '[]'::jsonb,
    sent_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE purchase_payments (
    id SERIAL PRIMARY KEY,
    purchase_id INTEGER NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    cost_type VARCHAR(30) NOT NULL,
    supplier_id INTEGER REFERENCES suppliers(id) ON DELETE SET NULL,
    amount NUMERIC(12,2) NOT NULL CHECK (amount >= 0),
    method VARCHAR(30) NOT NULL,
    status VARCHAR(60) NOT NULL DEFAULT 'Pendente de Pagamento',
    justification TEXT NOT NULL DEFAULT '',
    receipt_url TEXT NOT NULL DEFAULT '',
    requested_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    approved_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE purchase_issues (
    id SERIAL PRIMARY KEY,
    purchase_id INTEGER NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    issue_type VARCHAR(20) NOT NULL CHECK (issue_type IN ('material', 'gravacao')),
    description TEXT NOT NULL,
    attachments JSONB NOT NULL DEFAULT '[]'::jsonb,
    supplier_id INTEGER REFERENCES suppliers(id) ON DELETE SET NULL,
    solution TEXT NOT NULL DEFAULT '',
    resolution_deadline DATE NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'Aberta',
    opened_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    resolved_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE TABLE purchase_history (
    id SERIAL PRIMARY KEY,
    purchase_id INTEGER NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    action VARCHAR(80) NOT NULL,
    from_status VARCHAR(80) NOT NULL DEFAULT '',
    to_status VARCHAR(80) NOT NULL DEFAULT '',
    details TEXT NOT NULL DEFAULT '',
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    purchase_id INTEGER REFERENCES purchase_orders(id) ON DELETE CASCADE,
    recipient_user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    recipient_permission VARCHAR(60) NOT NULL DEFAULT '',
    notification_type VARCHAR(60) NOT NULL,
    message TEXT NOT NULL,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_purchase_orders_status ON purchase_orders(status);
CREATE INDEX idx_purchase_orders_sample ON purchase_orders(is_sample);
CREATE INDEX idx_purchase_payments_status ON purchase_payments(status);
CREATE INDEX idx_purchase_issues_status ON purchase_issues(status);
CREATE INDEX idx_notifications_recipient ON notifications(recipient_user_id, recipient_permission, read_at);

INSERT INTO purchase_orders (
    sale_id, general_number, status, material_supplier_id, has_engraving,
    corel_required, material_unit_cost, material_total_cost, payment_method,
    commercial_notes
)
SELECT s.id, s.id::text, 'Pendente de Compra',
       MIN(NULLIF(p.supplier_id, 0)),
       COALESCE(BOOL_OR(COALESCE(si.engravings, '[]'::jsonb) <> '[]'::jsonb), FALSE),
       COALESCE(BOOL_OR(COALESCE(si.engravings, '[]'::jsonb) <> '[]'::jsonb), FALSE),
       COALESCE(MIN(p.cost_price), 0),
       COALESCE(SUM(si.quantity * p.cost_price), 0),
       COALESCE(s.payment_method, ''),
       COALESCE(s.observacoes_internas, '')
FROM sales s
JOIN sale_items si ON si.sale_id = s.id
JOIN products p ON p.id = si.product_id
WHERE s.status IN ('Pendente de Compra', 'Aprovado pelo Cliente')
GROUP BY s.id
ON CONFLICT (sale_id) DO NOTHING;

