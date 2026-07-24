-- Seed para testar o PDF: fornecedor, produto, cliente e venda

-- Fornecedor
INSERT INTO suppliers (name, cnpj, contact_person, email, landline_phone, mobile_phone, commercial_address)
VALUES ('Fornecedor Teste Ltda', '12.345.678/0001-99', 'João Silva', 'contato@fornecedor.com', '(11) 3000-0000', '(11) 99999-0000', 'Av. Paulista, 1000, São Paulo, SP');

-- Produto
INSERT INTO products (product_name, internal_code, supplier_id, product_group, description, photos, ncm, stock, selling_price, is_composition)
VALUES ('Chaveiro PVC Alto Relevo', 'DN1961', 1, 'Brindes', 'Chaveiro personalizado', '{}', '39269090', 100, 23.99, false);

-- Cliente PF
INSERT INTO customers (customer_type, status, name, cpf, email, business_phone, mobile_phone)
VALUES ('PF', 'Ativo', 'Maria Santos', '12345678900', 'maria@email.com', '(11) 3000-1111', '(11) 98888-0000');

-- Endereço do cliente PF
INSERT INTO customer_addresses (customer_id, address_type, address)
VALUES (1, 'entrega', 'Rua das Flores, 123 - Centro, São Paulo/SP - CEP 03011-010');

-- Cliente PJ
INSERT INTO customers (customer_type, status, name, cnpj, email, business_phone, mobile_phone)
VALUES ('PJ', 'Ativo', 'Empresa Teste Ltda', '98765432000111', 'contato@empresa.com', '(11) 3000-2222', '(11) 97777-0000');

INSERT INTO customer_addresses (customer_id, address_type, address)
VALUES (2, 'comercial', 'Av. Brasil, 500 - São Paulo/SP - CEP 03011-010');

-- Venda (seller_id=1 admin, customer_id=1 Maria Santos)
INSERT INTO sales (seller_id, customer_id, payment_method, installments, payment_term_days, first_installment_start, total_value)
VALUES (1, 1, 'PIX', 1, 0, '2026-03-20'::timestamp, 47.98);

-- Itens da venda
INSERT INTO sale_items (sale_id, product_id, quantity, unit_price, total_price)
VALUES (1, 1, 2, 23.99, 47.98);

