# Build Status

## AgentWatch — portfolio MVP

Implemented:

- persisted users/projects/agents/tasks/events/approvals via PostgreSQL adapter
- memory repository for isolated tests only
- registration, login, refresh and logout
- approval workflow
- task state-transition guards
- realtime WebSocket event fan-out
- HTTP route tests
- WebSocket handshake and event-delivery test
- readable Expo dashboard with explicit demo login
- Docker PostgreSQL runtime

## OpsTrack — portfolio MVP

Implemented:

- PostgreSQL-backed users, work orders, audit events and sync results
- role-aware worker/supervisor/admin authentication
- explicit work-order transition state machine
- notes and evidence
- operation-based offline synchronization
- idempotent sync replay
- Expo SQLite mutation queue
- work-order detail screen
- equipment scanner
- audit-trail screen
- HTTP workflow tests

## ScanStock — portfolio MVP

Implemented:

- PostgreSQL-backed users, products and inventory movements
- append-only stock ledger
- server-calculated stock-on-hand
- serializable inventory write transaction
- idempotency-key replay handling
- negative-stock protection
- Expo SQLite mutation queue
- product detail screen
- barcode scanner
- ledger screen
- HTTP idempotency test

## Remaining polish before publishing

- capture screenshots and short demo recordings on real devices/web
- run PostgreSQL integration CI once pushed to GitHub
- add deployment infrastructure only after choosing a hosting target
- optionally add mobile component tests after UI stabilizes
