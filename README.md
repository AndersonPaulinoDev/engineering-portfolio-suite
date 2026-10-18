# Engineering Portfolio Suite

Three portfolio applications built from `build-spec.json` to demonstrate product engineering across Go, Expo/React Native, PostgreSQL, realtime systems, offline workflows, testing, and technical documentation.

## Projects

| Project | Product problem | Strongest engineering signal | API |
|---|---|---|---:|
| **AgentWatch** | Monitor coding agents and approval-gated background work | WebSockets, event fan-out, workflow state, auth, realtime tests | 8080 |
| **OpsTrack** | Keep field technicians productive through poor connectivity | state machines, audit history, offline mutation queue, QR scanning | 8081 |
| **ScanStock** | Keep inventory correct despite retries and disconnected devices | append-only ledger, idempotency, serializable PostgreSQL writes | 8082 |

## What changed from the first scaffold

The repository now addresses the issues that would matter in a portfolio review:

- readable TypeScript/TSX instead of compressed single-line UI files
- multi-screen OpsTrack and ScanStock mobile flows
- real PostgreSQL repository implementations for all three APIs
- persisted auth users for OpsTrack and ScanStock, not just domain records
- Docker Compose starts PostgreSQL, applies migrations, and runs the PostgreSQL build
- memory persistence is explicit (`PERSISTENCE=memory`) and intended only for tests/demos
- `AUTH_SECRET` is mandatory; servers no longer fall back to built-in secrets
- demo credentials exist only when `DEMO_MODE=true`
- HTTP endpoint tests cover routes used by the Expo clients
- AgentWatch has a real WebSocket handshake + event-delivery test
- Go modules use `github.com/AndersonPaulinoDev/...`

## Runtime modes

### Portfolio/demo with PostgreSQL

Use Docker Compose inside any project:

```bash
cd projects/agentwatch
AUTH_SECRET="$(openssl rand -hex 32)" DEMO_MODE=true docker compose up --build
```

The Compose configuration:

1. starts PostgreSQL,
2. applies `api/migrations/001_init.sql`,
3. waits for database readiness,
4. runs `go run -tags postgres ./cmd/server`,
5. enables demo seed data only when `DEMO_MODE=true`.

### Fast backend tests without external services

The normal test suite uses the in-memory repository implementation so domain and HTTP tests run quickly and deterministically:

```bash
make test
```

Memory mode is not selected silently by the servers. A non-PostgreSQL binary requires `PERSISTENCE=memory` explicitly.

## Security defaults

- `AUTH_SECRET` is required at startup and must be at least 32 characters.
- No production fallback secret exists.
- Demo accounts are seeded only with `DEMO_MODE=true`.
- `.env.example` intentionally leaves `AUTH_SECRET` blank; generate one before startup.
- Authorization is enforced by API middleware, not just mobile UI state.
- PostgreSQL writes use parameterized SQL.

## Mobile apps

The Expo apps target iOS, Android, and web from shared TypeScript code.

### AgentWatch

- explicit login
- overview metrics
- agent states
- task queue
- approval decisions
- live WebSocket refreshes

### OpsTrack

- explicit demo login
- work-order dashboard
- work-order detail route
- notes and status transitions
- SQLite offline mutation queue
- equipment scanner
- audit-trail screen

### ScanStock

- explicit demo login
- stock dashboard
- product detail route
- inventory movement controls
- barcode scanner
- transaction-ledger screen
- SQLite offline queue

## Toolchain

- Go 1.23+ source compatibility
- Expo SDK 58 / React Native 0.88
- TypeScript
- PostgreSQL 17 in local Compose
- GitHub Actions

## Validation

Backend validation performed in this workspace:

```bash
cd projects/agentwatch/api && PERSISTENCE=memory go test ./...
cd projects/opstrack/api && PERSISTENCE=memory go test ./...
cd projects/scanstock/api && PERSISTENCE=memory go test ./...
```

GitHub Actions is configured to compile and integration-test the PostgreSQL code with the `postgres` build tag against a PostgreSQL service. The local artifact-building environment used here does not provide Docker/PostgreSQL, so those database integration jobs have not been claimed as locally executed.

## Repository layout

```text
engineering-portfolio-suite/
├── build-spec.json
├── projects/
│   ├── agentwatch/
│   │   ├── api/
│   │   ├── mobile/
│   │   └── docs/
│   ├── opstrack/
│   └── scanstock/
├── docs/
└── .github/workflows/
```

See `docs/REQUIREMENTS_TRACEABILITY.md` for requirement-to-code evidence.


See `docs/REVIEW_FIXES.md` for the review issue → implementation evidence map.
