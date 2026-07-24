ALTER TABLE products DROP CONSTRAINT IF EXISTS products_internal_code_key;
ALTER TABLE products ADD CONSTRAINT products_internal_code_color_key UNIQUE (internal_code, color);

