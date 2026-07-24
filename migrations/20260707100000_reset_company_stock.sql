-- +goose Up
-- O estoque da empresa começa zerado e será alimentado manualmente.
-- O campo stock continua representando a disponibilidade informada pelo fornecedor.
UPDATE products SET supplier_stock = 0 WHERE supplier_stock <> 0;

-- +goose Down
-- Não é possível restaurar com segurança os valores incorretos vindos do fornecedor.
