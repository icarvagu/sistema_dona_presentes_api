-- Todos os produtos atuais devem passar pela análise de impostos e autorização.
UPDATE products
SET pending_approval = TRUE,
    origin = '',
    updated_at = NOW();

-- Novos registros também aguardam autorização por padrão.
ALTER TABLE products ALTER COLUMN pending_approval SET DEFAULT TRUE;

