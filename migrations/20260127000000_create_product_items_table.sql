CREATE TABLE product_items (
    id SERIAL PRIMARY KEY,
    product_parent_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(product_parent_id, product_id)
);

CREATE INDEX idx_product_items_product_parent_id ON product_items(product_parent_id);
CREATE INDEX idx_product_items_product_id ON product_items(product_id);

