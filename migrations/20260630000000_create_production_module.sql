CREATE TABLE production_orders (
    id BIGSERIAL PRIMARY KEY,
    purchase_id INTEGER NOT NULL UNIQUE REFERENCES purchase_orders(id) ON DELETE RESTRICT,
    sale_id INTEGER NOT NULL UNIQUE REFERENCES sales(id) ON DELETE RESTRICT,
    status VARCHAR(40) NOT NULL DEFAULT 'AGUARDANDO_RECEBIMENTO',
    has_engraving BOOLEAN NOT NULL DEFAULT FALSE,
    first_piece_required BOOLEAN NOT NULL DEFAULT FALSE,
    priority SMALLINT NOT NULL DEFAULT 1 CHECK (priority BETWEEN 0 AND 3),
    owner_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,
    released_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (status IN ('AGUARDANDO_RECEBIMENTO','RECEBIMENTO_PARCIAL','EM_CONFERENCIA','MATERIAL_OK','BLOQUEADO','SEPARANDO_GRAVACAO','ENVIADO_GRAVACAO','AGUARDANDO_PRIMEIRA_PECA','PRIMEIRA_PECA_APROVADA','EM_GRAVACAO','RETORNO_GRAVACAO','CONFERENCIA_RETORNO','PRONTO_EXPEDICAO','EXPEDIDO','ENTREGUE','CONCLUIDO'))
);

CREATE TABLE production_order_items (
    id BIGSERIAL PRIMARY KEY,
    production_order_id BIGINT NOT NULL REFERENCES production_orders(id) ON DELETE CASCADE,
    sale_item_id INTEGER NOT NULL UNIQUE REFERENCES sale_items(id) ON DELETE RESTRICT,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    expected_quantity INTEGER NOT NULL CHECK (expected_quantity > 0),
    received_quantity INTEGER NOT NULL DEFAULT 0 CHECK (received_quantity >= 0),
    approved_quantity INTEGER NOT NULL DEFAULT 0 CHECK (approved_quantity >= 0),
    rejected_quantity INTEGER NOT NULL DEFAULT 0 CHECK (rejected_quantity >= 0),
    has_engraving BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (received_quantity <= expected_quantity),
    CHECK (approved_quantity + rejected_quantity <= received_quantity)
);

CREATE TABLE production_receipts (
    id BIGSERIAL PRIMARY KEY,
    production_order_id BIGINT NOT NULL REFERENCES production_orders(id) ON DELETE CASCADE,
    idempotency_key VARCHAR(100) NOT NULL,
    invoice_number VARCHAR(80) NOT NULL DEFAULT '',
    invoice_key VARCHAR(60) NOT NULL DEFAULT '',
    invoice_url TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    received_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(production_order_id,idempotency_key)
);
CREATE TABLE production_receipt_items (
    id BIGSERIAL PRIMARY KEY,
    receipt_id BIGINT NOT NULL REFERENCES production_receipts(id) ON DELETE CASCADE,
    production_item_id BIGINT NOT NULL REFERENCES production_order_items(id) ON DELETE RESTRICT,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    UNIQUE(receipt_id,production_item_id)
);

CREATE TABLE production_occurrences (
    id BIGSERIAL PRIMARY KEY,
    production_order_id BIGINT NOT NULL REFERENCES production_orders(id) ON DELETE CASCADE,
    production_item_id BIGINT REFERENCES production_order_items(id) ON DELETE SET NULL,
    kind VARCHAR(40) NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('PARCIAL','BLOQUEANTE')),
    quantity INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    description TEXT NOT NULL,
    attachment_url TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'ABERTA' CHECK (status IN ('ABERTA','RESOLVIDA','CANCELADA')),
    opened_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    resolved_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    resolution TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE TABLE production_engraving_events (
    id BIGSERIAL PRIMARY KEY,
    production_order_id BIGINT NOT NULL REFERENCES production_orders(id) ON DELETE CASCADE,
    event_type VARCHAR(40) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT '',
    quantity INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    carrier_id INTEGER REFERENCES carriers(id) ON DELETE SET NULL,
    tracking_code VARCHAR(120) NOT NULL DEFAULT '',
    file_url TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    idempotency_key VARCHAR(100) NOT NULL,
    created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(production_order_id,idempotency_key)
);

CREATE TABLE production_volumes (
    id BIGSERIAL PRIMARY KEY,
    production_order_id BIGINT NOT NULL REFERENCES production_orders(id) ON DELETE CASCADE,
    label VARCHAR(80) NOT NULL,
    weight_kg NUMERIC(12,3) NOT NULL DEFAULT 0 CHECK (weight_kg >= 0),
    length_cm NUMERIC(12,2) NOT NULL DEFAULT 0,
    width_cm NUMERIC(12,2) NOT NULL DEFAULT 0,
    height_cm NUMERIC(12,2) NOT NULL DEFAULT 0,
    created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE production_fiscal_documents (
    id BIGSERIAL PRIMARY KEY,
    production_order_id BIGINT NOT NULL REFERENCES production_orders(id) ON DELETE CASCADE,
    document_type VARCHAR(30) NOT NULL,
    document_number VARCHAR(80) NOT NULL,
    access_key VARCHAR(60) NOT NULL DEFAULT '',
    file_url TEXT NOT NULL DEFAULT '',
    issued_at TIMESTAMPTZ,
    created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(production_order_id,document_type,document_number)
);
CREATE TABLE production_shipments (
    id BIGSERIAL PRIMARY KEY,
    production_order_id BIGINT NOT NULL UNIQUE REFERENCES production_orders(id) ON DELETE CASCADE,
    method VARCHAR(30) NOT NULL,
    carrier_id INTEGER REFERENCES carriers(id) ON DELETE SET NULL,
    driver_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    tracking_code VARCHAR(120) NOT NULL DEFAULT '',
    tracking_url TEXT NOT NULL DEFAULT '',
    postal_service VARCHAR(60) NOT NULL DEFAULT '',
    proof_url TEXT NOT NULL DEFAULT '',
    shipped_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    updated_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE production_supplies (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(160) NOT NULL UNIQUE,
    unit VARCHAR(20) NOT NULL,
    current_quantity NUMERIC(14,3) NOT NULL DEFAULT 0,
    minimum_quantity NUMERIC(14,3) NOT NULL DEFAULT 0,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE production_supply_movements (
    id BIGSERIAL PRIMARY KEY,
    supply_id BIGINT NOT NULL REFERENCES production_supplies(id) ON DELETE RESTRICT,
    production_order_id BIGINT REFERENCES production_orders(id) ON DELETE SET NULL,
    movement_type VARCHAR(10) NOT NULL CHECK (movement_type IN ('ENTRADA','SAIDA','AJUSTE')),
    quantity NUMERIC(14,3) NOT NULL CHECK (quantity > 0),
    reason TEXT NOT NULL,
    created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE production_history (
    id BIGSERIAL PRIMARY KEY,
    production_order_id BIGINT NOT NULL REFERENCES production_orders(id) ON DELETE CASCADE,
    action VARCHAR(80) NOT NULL,
    from_status VARCHAR(40) NOT NULL DEFAULT '',
    to_status VARCHAR(40) NOT NULL DEFAULT '',
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
ALTER TABLE notifications ADD COLUMN production_order_id BIGINT REFERENCES production_orders(id) ON DELETE CASCADE;

CREATE INDEX ix_production_orders_queue ON production_orders(status,priority DESC,updated_at);
CREATE INDEX ix_production_items_order ON production_order_items(production_order_id);
CREATE INDEX ix_production_occurrences_open ON production_occurrences(production_order_id,severity) WHERE status='ABERTA';
CREATE INDEX ix_production_history_order ON production_history(production_order_id,created_at DESC);
CREATE INDEX ix_production_shipments_tracking ON production_shipments(tracking_code) WHERE tracking_code<>'';

INSERT INTO production_orders(purchase_id,sale_id,status,has_engraving,first_piece_required,released_at)
SELECT p.id,p.sale_id,'AGUARDANDO_RECEBIMENTO',p.has_engraving,p.first_piece_required,COALESCE(p.production_released_at,NOW())
FROM purchase_orders p WHERE p.production_released_at IS NOT NULL OR p.status='Liberado para Produção'
ON CONFLICT (purchase_id) DO NOTHING;
INSERT INTO production_order_items(production_order_id,sale_item_id,product_id,expected_quantity,has_engraving)
SELECT po.id,si.id,si.product_id,si.quantity,COALESCE(si.engravings,'[]'::jsonb)<>'[]'::jsonb
FROM production_orders po JOIN sale_items si ON si.sale_id=po.sale_id
ON CONFLICT (sale_item_id) DO NOTHING;

