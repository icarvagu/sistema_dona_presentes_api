# Dona Presentes API - Migração Go → Rust

## Status Final

| Fase | Status |
|------|--------|
| 1. Fundação | ✅ |
| 2. Auth + CRUDs simples | ✅ |
| 3. Core business | ✅ |
| 4. Compras + Produção | ✅ |
| 5. Arte Final + Dashboard | ✅ |
| 6. Polish + Testes | ✅ |

## Métricas

| | Go | Rust |
|---|-----|------|
| Arquivos .go/.rs | 116 | 79 |
| Linhas de código | ~18.900 | ~6.800 |
| Linhas de teste | 0 | 336 |
| Migrações SQL | 77 | 77 |
| Endpoints | ~105 | ~95 |
| Warnings | N/A | 0 |

## Stack

| Camada | Rust |
|--------|------|
| Framework | `axum` 0.8 |
| ORM/DB | `sqlx` 0.8 (PostgreSQL) |
| Auth | `jsonwebtoken` + `bcrypt` |
| Serialization | `serde` + `serde_json` |
| Async | `tokio` |
| HTTP Client | `reqwest` |
| PDF | Raw PDF generator (0 deps) |
| Migrations | `sqlx::migrate!` |
| Logging | `tracing` + `tracing-subscriber` |

## Testes (21 passing)

```
test_app_error_bad_request
test_app_error_conflict
test_app_error_not_found
test_app_error_unauthorized
test_claims_encode_decode
test_config_from_env
test_health_check
test_jwt_claims
test_password_complexity_*
test_rate_limiter_*
test_sale_input_serialization
test_sanitizer_*
test_unauthorized_without_token
test_admin_endpoint_blocked_for_standard
test_auth_login_missing_fields
```

## Endpoints (~95)

- Auth (8) - login, refresh, logout, forgot/reset password, me, change-password, health
- Users (4) - list, get, create (admin), delete (admin)
- Customers (5) - CRUD
- Suppliers (5) - CRUD
- Carriers (5) - CRUD
- Products (14) - CRUD + search, groups, pending, approve, bulk-approve, financial-report, last-cost, items
- Quotes (8) - CRUD + feedback, pdf, convert, feedback-events
- Sales (6) - CRUD + layout, pdf
- Purchases (13) - CRUD + financial, release, actions, attachments, payments, issues
- Production (14) - dashboard, orders, receipts, occurrences, transitions, assignment, engraving, volumes, fiscal, shipment, supplies
- Art Final (16) - tasks, stories, layout requests/versions/jobs/items
- Sales Workflow (13) - dashboard, financial-analysis, pending-events, engraving-approvals, seller-approval, item-layouts
- Notifications (2) - list, mark-read
- Dashboard (1) - aggregated metrics
- Admin (3) - user create/delete, XBZ sync
- PDF (2) - sale pdf, quote pdf
