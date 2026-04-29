# Alias Deprecation Plan (Post-PR Validation)

Data: 2026-04-29
Issue: PRO-18

## Objetivo
Remover aliases temporários introduzidos para compatibilidade durante a padronização de nomenclatura, sem regressão funcional.

## Aliases atualmente ativos
- `repositories.ProductRepository.GetByCodigoInterno` -> `GetByInternalCode`
- `services.SyncService.Sincronizar` -> `Synchronize`
- `services.SyncService.garantirFornecedor` -> `ensureSupplier`
- `controllers.SyncProductsFromXBZ` -> `SyncProductsFromXBZHandler`

## Critérios para remoção
1. PR de refatoração aprovado e mergeado.
2. Nenhuma chamada interna restante aos aliases (`grep` no repositório).
3. Testes do backend passando (`go test ./...`).
4. Endpoint de sync XBZ validado em ambiente de desenvolvimento.

## Passo a passo
1. Migrar qualquer referência residual para os nomes padrão.
2. Remover aliases em lote único e pequeno.
3. Rodar `go test ./...`.
4. Abrir PR específico: "remove temporary compatibility aliases".

## Risco e mitigação
- Risco: quebrar chamadas legadas internas.
- Mitigação: busca textual + testes + rollout em PR dedicado e pequeno.
