# Donna Presentes - Padrões da Codebase (Backend)

## Regras de Ouro

1. **Toda interação com DB é via repository** — nunca `db.Query()` em controllers/services
2. **Toda operação de escrita tem audit log** — `auditService.LogSimple()` em Create/Update/Delete
3. **Todo erro 5xx é logado em `error_logs`** — via `ErrorHandlerWithRequest` ou `RecoveryMiddleware`
4. **Toda migration nova usa `-- +goose StatementBegin/End`** para blocos multi-statement
5. **API segue camadas**: `routes → controllers → services → repositories → models`
6. **Nomes de campos em inglês** — models, JSON tags, colunas SQL consistentemente
7. **Todo campo opcional usa `omitempty`** — sem zero-values desnecessários no JSON
8. **Tokens aleatórios via `crypto/rand`** — nunca `time.Now().UnixNano()`


## Camadas da Arquitetura

```
routes/        → registra endpoints HTTP, sem lógica
controllers/   → parse request, chama service, write response, sem lógica de negócio
services/      → lógica de negócio, validação, orquestração
repositories/  → acesso a dados, queries SQL parametrizadas
models/        → structs Go, JSON tags, DTOs de input/output
middleware/    → auth, CORS, rate limiting, logging, recovery
errors/        → tipos de erro padronizados (AppError)
```


## Convenções de Código

### Models
- Usar `*string` com `json:"...,omitempty"` para campos nullable (padrão `Supplier`)
- Structs de input separados dos structs de response (ex: `QuoteInput`, `Quote`)
- Campo `PasswordHash` sempre `json:"-"` (nunca expor)

### Services
- Construtor `NewXxxService` recebe repositories + `*AuditService`
- Validação retorna `apperrors.NewValidationError` / `NewMissingFieldError`
- Erro de DB retorna `apperrors.NewDatabaseError(err)`
- Erro "not found" retorna `apperrors.NewNotFoundError`
- `sql.ErrNoRows` → erro amigável (nunca raw)

### Repositories
- SQL 100% parametrizado (`$1`, `$2`), nunca concatenação de string
- `SELECT` com JOIN explícito, sem N+1 queries
- Exported type + constructor + métodos CRUD padronizados

### Controllers
- `InitXxxService()` no topo, usa `config.DB` e `auditService` package-level
- Handler: `json.NewDecoder(r.Body).Decode()` → `service.Xxx()` → `json.NewEncoder(w).Encode()`
- Erro via `middleware.ErrorHandler(w, err, statusCode)`
- Package-level `var xxxService *services.XxxService`

### Middleware
- `ErrorHandlerWithRequest(w, r, err, statusCode)` para logar erro 5xx no DB
- `AuthMiddleware` usa `apperrors.NewUnauthorizedError` (não ValidationError)
- Ordem da chain no main.go: Tracing → CORS → Security → Validation → RateLimit → Timeout → NormalizePath → StructuredLogging → Recovery


## Logs e Auditoria

### Tabelas de log
- `audit_logs` — ações de negócio (create/update/delete, login, transições)
- `error_logs` — erros 5xx e panics
- `request_logs` — métricas HTTP (feature-flag `LOG_DB=true`)

### Cleanup automático
- Migration `20260713000001_add_log_retention.sql` — função `cleanup_old_logs()`
- Chamada no startup: `config.DB.Exec("SELECT cleanup_old_logs()")`
- Remove registros > 2 meses

### Exemplo de audit log
```go
s.auditService.LogSimple(&userID, "supplier_created", "supplier", 
    fmt.Sprintf("name=%s", supplier.Name), "")
```


## Testes

### Sempre rodar antes de commitar
```bash
cd sistema_dona_presentes_api && go test ./... 2>&1
```

### Sempre atualizar OpenAPI
```bash
# api.yml deve refletir TODOS os campos dos models Go e TODAS as rotas registradas
cd sistema_dona_presentes_api
python3 -c "import yaml; yaml.safe_load(open('api.yml'))" 2>&1
```

### Checklist pré-commit
- [ ] `go test ./...` passa
- [ ] `api.yml` atualizado se houve mudança em models/routes
- [ ] `auditService` injetado em novos services
- [ ] Migration nova com `-- +goose StatementBegin/End` se tiver `BEGIN...END`


## Segurança

- **JWT**: HS256, 15min access, 7d refresh, secret do env `JWT_SECRET`
- **CPF**: criptografado AES-256-GCM com `ENCRYPTION_KEY` do env
- **CPF hash**: HMAC-SHA-256 com salt para deduplicação (sem rainbow table)
- **Tokens**: `GenerateRefreshToken` e `GeneratePasswordResetToken` usam `crypto/rand`
- **Lockout**: 5 tentativas → bloqueio 15 min
- **Senha**: bcrypt, mínimo 8 chars + upper + lower + digit + special


## Regras para Agentes / AI

### Sempre antes de commit:
1. **Rodar testes**: `go test ./...` → 100% pass
2. **Build limpo**: `go build ./...` → sem erros
3. **Atualizar `api.yml`**: todo campo novo/alterado em models ou rota nova deve estar no spec
4. **Validar YAML**: `python3 -c "import yaml; yaml.safe_load(open('api.yml'))"` 
5. **Criar migration** para toda alteração de schema do banco

### Checklist:
- [ ] `go test ./...` → 100% pass
- [ ] `go build ./...` → sem erro
- [ ] `api.yml` atualizado (models + routes)
- [ ] Migration nova se alterou tabelas
- [ ] `auditService` no construtor de service novo
- [ ] `-- +goose StatementBegin/End` em migration com `BEGIN...END`
