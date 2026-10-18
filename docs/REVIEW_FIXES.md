# Portfolio Review Fixes — v0.2

This document records the concrete changes made after the first scaffold review.

## 1. Mobile source readability

**Before:** major Expo screens were compressed into a few extremely long lines.

**Now:** TypeScript/TSX is formatted as normal reviewable source. AgentWatch is split into API/auth/components, while OpsTrack and ScanStock have multiple routed screens and supporting modules. The final audit rejects TS/TSX source lines over 200 characters.

## 2. Real PostgreSQL persistence

**Before:** migrations and Compose files existed, but runtime data used only memory repositories.

**Now:** all three APIs implement repository interfaces with PostgreSQL adapters behind the `postgres` build tag. Docker Compose starts PostgreSQL, applies migrations, and launches the API with `go run -tags postgres ./cmd/server`. In-memory persistence must be selected explicitly and is used for deterministic tests.

Persisted domains:

- AgentWatch: users, projects, agents, tasks, events, approvals
- OpsTrack: users, work orders, audit events, sync-operation results
- ScanStock: users, products, inventory transactions

## 3. Test coverage of real boundaries

The suite now contains 19 named Go tests across service, HTTP, realtime, and database-integration layers.

Notable boundary tests:

- AgentWatch raw WebSocket upgrade + event-delivery test
- authenticated and unauthorized HTTP route tests
- OpsTrack login → work list → transition → note → audit workflow
- ScanStock HTTP idempotency replay test
- PostgreSQL round-trip integration tests for every project under the `postgres` build tag

The WebSocket test caught and drove a real fix: the request logging wrapper now preserves `http.Hijacker`, which is required for WebSocket upgrades.

## 4. OpsTrack and ScanStock depth

OpsTrack now includes:

- work-order dashboard
- routed work-order detail
- backend-owned state transitions
- notes/evidence
- SQLite pending-operation queue
- equipment scanner
- audit trail
- conflict-aware sync model

ScanStock now includes:

- stock dashboard
- routed product detail
- receive/ship/damage operations
- append-only ledger
- SQLite pending-operation queue
- barcode scanner
- transaction ledger
- serializable PostgreSQL transaction writes
- idempotent replay protection

Each project has expanded architecture documentation covering its domain model and failure behavior.

## 5. Security/demo separation

- API startup requires `AUTH_SECRET` and rejects values shorter than 32 characters.
- Docker Compose requires the secret explicitly; no built-in runtime fallback exists.
- `.env.example` leaves the secret blank and documents `openssl rand -hex 32`.
- Demo accounts are seeded only when `DEMO_MODE=true`.
- Expo clients show demo credentials as demo-only and do not auto-login.

## 6. Repository identity

All Go modules and imports now use the correct account namespace:

```text
github.com/AndersonPaulinoDev/...
```

## Local validation performed

- `gofmt` clean for all Go packages
- `go vet ./...` passes for all three APIs
- default service/HTTP/realtime tests pass for all three APIs
- short secrets are rejected at actual server startup
- JSON and YAML files parse successfully
- 22 TypeScript/TSX files syntax-transpile successfully
- no TS/TSX source line exceeds 200 characters

## Validation intentionally left to CI / a Docker-enabled machine

The artifact build environment does not provide Docker/PostgreSQL and could not install the full Expo dependency tree from npm. Therefore:

- live PostgreSQL integration tests are configured in GitHub Actions but were not executed against a local database here
- full Expo `tsc --noEmit` is configured in CI but was not executed locally with installed SDK packages

The repository does not claim those checks passed locally.
