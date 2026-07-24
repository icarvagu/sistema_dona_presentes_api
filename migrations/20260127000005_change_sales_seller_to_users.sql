-- Remover foreign key antiga
ALTER TABLE sales DROP CONSTRAINT IF EXISTS fk_seller;

-- Adicionar nova foreign key apontando para users
ALTER TABLE sales 
ADD CONSTRAINT fk_seller 
FOREIGN KEY(seller_id) 
REFERENCES users(id) 
ON DELETE RESTRICT;

