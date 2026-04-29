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
  - Status: concluído na fase de migração (restam apenas aliases de compatibilidade planejados para remoção pós-validação do PR).
  - Escopo inicial: `controllers/`, `services/`, `repositories/`

### P2 (estrutural)
- Mapear e remover coleções/artefatos legados que não refletem mais o modelo atual (`insomnia-employees-collection.json` e referências históricas de migration).
  - Status: concluído (`insomnia-employees-collection.json`, `insomnia-customers-collection.zip` e duplicatas frontend de kits/products/sales removidas; inventário final sem duplicidade direta remanescente).
  - Observação: migrations históricas devem ser preservadas; limpeza focada em artefatos operacionais/documentais.
- Criar guideline curto de convenções (nomes, erros, logs) para evitar regressão de padrão.
  - Status: concluído (documentado em `docs/conventions.md`).

## Próxima ação proposta
1. Encerrar PRO-18 após revisão e merge do PR de remoção de aliases.
2. Consolidar aprendizados de padronização em issues futuras de manutenção.

## Blocker Atual
- Status: aguardando revisão/aprovação do PR `#6` (`https://github.com/IgorAsVI/dona_presentes/pull/6`).
- Unblock owner: `IgorAsVI` (reviewer solicitado).
- Unblock action: aprovar/solicitar ajustes no PR para seguir com remoção de aliases e fechamento da issue.
