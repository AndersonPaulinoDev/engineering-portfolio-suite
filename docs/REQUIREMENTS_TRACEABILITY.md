# Requirements Traceability

| Requirement | AgentWatch | OpsTrack | ScanStock |
|---|---|---|---|
| Go backend | `api/` | `api/` | `api/` |
| Expo + TypeScript | `mobile/` | `mobile/` | `mobile/` |
| PostgreSQL runtime | `internal/store/postgres.go` | `internal/store/postgres.go` | `internal/store/postgres.go` |
| PostgreSQL migrations | `migrations/001_init.sql` | `migrations/001_init.sql` | `migrations/001_init.sql` |
| Persisted users | AgentWatch store | OpsTrack store | ScanStock store |
| Explicit demo mode | `DEMO_MODE=true` only | `DEMO_MODE=true` only | `DEMO_MODE=true` only |
| Required auth secret | startup fails without it | startup fails without it | startup fails without it |
| HTTP route tests | projects/auth routes | work-order flow/audit | products/transactions |
| Realtime test | WebSocket handshake + event | n/a | n/a |
| Offline-first | optional | SQLite pending mutations | SQLite pending mutations |
| QR/barcode | n/a | equipment scanner | product scanner |
| Auditability | events + approvals | audit events | inventory ledger |
| Idempotency | state guards | stored sync results | transaction idempotency keys |
| Multi-screen mobile | tabbed operations UI | dashboard/detail/scanner/audit | dashboard/detail/scanner/ledger |
| CI | root workflow | root workflow | root workflow |
| Correct GitHub namespace | `AndersonPaulinoDev` | `AndersonPaulinoDev` | `AndersonPaulinoDev` |

## Persistence rule

The APIs have two persistence implementations:

- **PostgreSQL**: intended runtime; compiled with `-tags postgres`; used by Docker Compose and CI.
- **Memory**: explicit test/demo implementation; server startup requires `PERSISTENCE=memory` when this build is used.

This keeps fast unit/HTTP tests independent while preventing accidental memory persistence in the portfolio runtime.
