ALTER TABLE sales
  ADD COLUMN customer_id INTEGER,
  ADD CONSTRAINT fk_sales_customer
    FOREIGN KEY (customer_id)
    REFERENCES customers(id)
    ON DELETE RESTRICT;

CREATE INDEX idx_sales_customer_id ON sales(customer_id);

