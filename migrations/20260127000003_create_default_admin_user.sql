-- Criar usuário administrador padrão
-- Username: donnapresentesadm
-- Password: admin123
INSERT INTO users (username, password_hash, role) 
VALUES (
    'donnapresentesadm',
    '$2b$12$5hS1N5HdrPI0JE5zjc4s9OLSecG0g28i9cbvFhNdsgzrf2341RzVS',
    'admin'
) ON CONFLICT (username) DO NOTHING;
