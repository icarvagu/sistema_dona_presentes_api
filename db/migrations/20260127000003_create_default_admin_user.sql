-- +goose Up
-- +goose StatementBegin
-- Criar usuário administrador padrão
-- Username: donnapresentesadm
-- Password: senhadonna
-- Nota: Campos full_name, cpf e status serão preenchidos pela migration 20260127000007
INSERT INTO users (username, password_hash, role) 
VALUES (
    'donnapresentesadm',
    '$2a$10$.Ckklx0U.j/pOxw2SQ09Je6cBKzfzx/Tm0uQxxf46/MWPIh4B7202',
    'admin'
) ON CONFLICT (username) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM users WHERE username = 'donnapresentesadm';
-- +goose StatementEnd
