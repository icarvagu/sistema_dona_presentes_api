# PRO-18 Refactor Backlog

Data: 2026-04-29

## Objetivo
Levantamento incremental de padronização, debloat e otimização com foco em baixo risco e ganho rápido.

## Priorização

### P0 (fazer primeiro)
- Remover hardcode de credenciais XBZ em controller.
  - Status: concluído neste ciclo.
  - Arquivo: `controllers/product_xbz_controller.go`
- Garantir comportamento explícito quando integração XBZ não estiver configurada.
  - Status: concluído neste ciclo (retorno HTTP 503 com mensagem clara).
  - Arquivo: `controllers/product_xbz_controller.go`

### P1 (próximos ciclos)
- Padronizar inicialização da aplicação para evitar `panic` em `init()` e centralizar fail-fast com logs estruturados no `main()`.
  - Status: concluído neste ciclo.
  - Arquivo: `main.go`
- Reduzir ruído de logs de infraestrutura (`middleware/error_handler.go` e `services/xbz_service.go`), definindo nível e formato padrão.
  - Status: concluído neste ciclo.
  - Arquivos: `middleware/error_handler.go`, `services/xbz_service.go`
- Padronizar nomenclatura de métodos para português ou inglês (evitar mistura: `GetByCodigoInterno`, `SyncProductsFromXBZ`, etc.).
  - Status: em andamento (escopo XBZ e SyncService concluídos com aliases de compatibilidade).
  - Escopo inicial: `controllers/`, `services/`, `repositories/`

### P2 (estrutural)
- Mapear e remover coleções/artefatos legados que não refletem mais o modelo atual (`insomnia-employees-collection.json` e referências históricas de migration).
  - Status: em andamento (`insomnia-employees-collection.json` removido neste ciclo).
  - Observação: migrations históricas devem ser preservadas; limpeza focada em artefatos operacionais/documentais.
- Criar guideline curto de convenções (nomes, erros, logs) para evitar regressão de padrão.

## Próxima ação proposta
1. Concluir padronização de nomenclatura no restante do domínio (não-XBZ) em `controllers/`, `services/`, `repositories/`.
2. Mapear e remover artefatos legados operacionais (coleções insomnia fora de uso).
