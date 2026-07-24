ALTER TABLE quotes RENAME COLUMN freight_cnpj_solicitante TO freight_tax_id_sender;
ALTER TABLE quotes RENAME COLUMN freight_cnpj_cpf_origem TO freight_tax_id_origin;
ALTER TABLE quotes RENAME COLUMN freight_cnpj_cpf_destino TO freight_tax_id_dest;
ALTER TABLE quotes RENAME COLUMN freight_cnpj_devedor TO freight_tax_id_payer;

ALTER TABLE sales RENAME COLUMN email_nf TO invoice_email;
ALTER TABLE sales RENAME COLUMN email_financeiro TO financial_email;
ALTER TABLE sales RENAME COLUMN ordem_compra TO purchase_order;

ALTER TABLE customers RENAME COLUMN fantasy_name TO trade_name;
ALTER TABLE customers RENAME COLUMN razao_social TO company_name;
ALTER TABLE customers RENAME COLUMN inscricao_estadual TO state_registration;
ALTER TABLE customers RENAME COLUMN inscricao_municipal TO city_registration;
ALTER TABLE customers RENAME COLUMN responsavel TO responsible;

