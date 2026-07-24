CREATE TABLE sales (
    id SERIAL PRIMARY KEY,
    seller_id INTEGER NOT NULL,
    payment_method TEXT NOT NULL,
    installments INTEGER NOT NULL CHECK (installments > 0),
    payment_term_days INTEGER DEFAULT 0,
    first_installment_start TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    CONSTRAINT fk_seller
      FOREIGN KEY(seller_id)
        REFERENCES employees(id)
        ON DELETE RESTRICT
);

CREATE TABLE sale_items (
    id SERIAL PRIMARY KEY,
    sale_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC(10, 2) NOT NULL,
    total_price NUMERIC(10, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    CONSTRAINT fk_sale
      FOREIGN KEY(sale_id)
        REFERENCES sales(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_product
      FOREIGN KEY(product_id)
        REFERENCES products(id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_sales_seller_id ON sales(seller_id);
CREATE INDEX idx_sale_items_sale_id ON sale_items(sale_id);
CREATE INDEX idx_sale_items_product_id ON sale_items(product_id);

