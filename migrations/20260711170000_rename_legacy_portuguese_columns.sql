-- +goose Up
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

-- +goose Down
ALTER TABLE quotes RENAME COLUMN freight_tax_id_sender TO freight_cnpj_solicitante;
ALTER TABLE quotes RENAME COLUMN freight_tax_id_origin TO freight_cnpj_cpf_origem;
ALTER TABLE quotes RENAME COLUMN freight_tax_id_dest TO freight_cnpj_cpf_destino;
ALTER TABLE quotes RENAME COLUMN freight_tax_id_payer TO freight_cnpj_devedor;

ALTER TABLE sales RENAME COLUMN invoice_email TO email_nf;
ALTER TABLE sales RENAME COLUMN financial_email TO email_financeiro;
ALTER TABLE sales RENAME COLUMN purchase_order TO ordem_compra;

ALTER TABLE customers RENAME COLUMN trade_name TO fantasy_name;
ALTER TABLE customers RENAME COLUMN company_name TO razao_social;
ALTER TABLE customers RENAME COLUMN state_registration TO inscricao_estadual;
ALTER TABLE customers RENAME COLUMN city_registration TO inscricao_municipal;
ALTER TABLE customers RENAME COLUMN responsible TO responsavel;
