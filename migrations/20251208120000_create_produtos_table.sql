CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    product_name TEXT NOT NULL,
    internal_code TEXT NOT NULL UNIQUE,
    supplier_id INTEGER NOT NULL,
    product_group TEXT,
    description TEXT,
    photos TEXT[],
    ncm TEXT,
    material_origin TEXT,
    stock INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    CONSTRAINT fk_supplier
      FOREIGN KEY(supplier_id)
        REFERENCES suppliers(id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_products_supplier_id ON products(supplier_id);

