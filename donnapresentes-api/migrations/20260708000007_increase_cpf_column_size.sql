-- +goose Up
-- CPF will now store encrypted value (base64) so VARCHAR(14) is too small
ALTER TABLE users ALTER COLUMN cpf TYPE TEXT;
DROP INDEX IF EXISTS idx_users_cpf;

-- +goose Down
ALTER TABLE users ALTER COLUMN cpf TYPE VARCHAR(14);
CREATE INDEX idx_users_cpf ON users(cpf);
