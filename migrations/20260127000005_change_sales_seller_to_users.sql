-- +goose Up
-- +goose StatementBegin
-- Remover foreign key antiga
ALTER TABLE sales DROP CONSTRAINT IF EXISTS fk_seller;

-- Adicionar nova foreign key apontando para users
ALTER TABLE sales 
ADD CONSTRAINT fk_seller 
FOREIGN KEY(seller_id) 
REFERENCES users(id) 
ON DELETE RESTRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sales DROP CONSTRAINT IF EXISTS fk_seller;
ALTER TABLE sales 
ADD CONSTRAINT fk_seller 
FOREIGN KEY(seller_id) 
REFERENCES employees(id) 
ON DELETE RESTRICT;
-- +goose StatementEnd
