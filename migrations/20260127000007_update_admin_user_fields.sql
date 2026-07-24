-- +goose Up
-- +goose StatementBegin
-- Atualizar usuário administrador com campos obrigatórios
-- Usar CPF específico para admin para evitar conflitos
UPDATE users 
SET 
    full_name = COALESCE(full_name, 'Administrador Dona Presentes'),
    cpf = COALESCE(cpf, '99999999999'),
    status = COALESCE(status, 'Ativo')
WHERE username = 'donnapresentesadm';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Não há necessidade de rollback para esta migration
-- +goose StatementEnd
