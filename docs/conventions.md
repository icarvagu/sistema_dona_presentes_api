# Dona Presentes - Coding Conventions (Go Backend)

Data: 2026-04-29

## 1. Nomenclatura
- Adotar inglês para nomes de funções/métodos/campos novos no backend Go.
- Em refactors de API pública interna, manter aliases temporários para compatibilidade até remoção planejada.
- Evitar mistura de idiomas no mesmo fluxo (ex.: preferir `GetByInternalCode` em vez de novos `GetByCodigoInterno`).

## 2. Tratamento de Erros
- Sempre retornar erro com contexto quando fizer wrap: `fmt.Errorf("contexto: %w", err)`.
- Em camada HTTP, padronizar resposta JSON usando `middleware.ErrorHandler`.
- Não usar `panic` para falhas operacionais de bootstrap ou integração; preferir fail-fast explícito com log e saída controlada.
- Para erros de banco conhecidos (FK, unique), mapear para erros de domínio reutilizáveis (`errors` package).

## 3. Logging
- Preferir logs estruturados em chave=valor para eventos de request/integração.
- Não logar payload bruto de APIs externas ou dados sensíveis.
- Diferenciar logs operacionais:
  - `INFO`: eventos de ciclo normal (início/fim de sync, subida de servidor)
  - `WARN`: degradação sem queda total (feature desabilitada por env ausente)
  - `ERROR`: falhas com impacto funcional
- Evitar logs redundantes em caminhos de alto volume (ex.: preflight CORS).

## 4. Configuração
- Segredos e credenciais somente por variável de ambiente.
- Quando configuração obrigatória estiver ausente, retornar erro explícito com mensagem objetiva.

## 5. Refactors Incrementais
- Mudanças de padronização devem ser feitas em pequenos lotes com commit claro.
- Para renomeações em uso, aplicar fase de transição:
  1. Introduzir novo nome.
  2. Manter alias de compatibilidade.
  3. Migrar chamadas.
  4. Remover alias em ciclo posterior.

## 6. Checklist Rápido de PR
- Código compila e testes mínimos do escopo passam.
- Sem credencial hardcoded.
- Sem logs sensíveis.
- Erros com contexto e mapeamento consistente.
- Nomes seguem padrão definido.
