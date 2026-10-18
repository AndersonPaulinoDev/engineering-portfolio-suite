# OpsTrack

Offline-first field operations software for technicians, inspectors, supervisors, and maintenance teams.

## What it demonstrates

OpsTrack is intentionally built around operational edge cases rather than generic CRUD. The backend owns work-order state transitions and authorization. The mobile client can queue mutations in SQLite when connectivity fails and replay those operations later without replacing entire server records.

## Implemented MVP

- PostgreSQL-backed users, work orders, audit events, and sync results
- worker, supervisor, and admin roles
- explicit work-order state machine
- assignment checks for workers
- append-only audit events
- technician notes and evidence references
- idempotent offline sync operations
- conflict responses for stale offline transitions
- Expo SQLite pending-operation queue
- work-order list and detail routes
- QR/barcode equipment scanner
- audit-trail screen
- HTTP workflow tests
- OpenAPI schema and ADRs

## Demo mode

The following accounts are created only when `DEMO_MODE=true`:

```text
worker@opstrack.dev / demo1234
supervisor@opstrack.dev / demo1234
admin@opstrack.dev / demo1234
```

The Expo app shows that they are demo credentials and requires the reviewer to press **Sign in**.

## Run with PostgreSQL

```bash
cd projects/opstrack
cp .env.example .env
# Set AUTH_SECRET in .env to the output of: openssl rand -hex 32
docker compose up --build
```

The API runs with the `postgres` build tag and uses `DATABASE_URL`. PostgreSQL applies `api/migrations/001_init.sql` on first startup.

## Backend tests

```bash
cd api
PERSISTENCE=memory go test ./...
go vet ./...
```

CI also runs the tagged PostgreSQL integration test against a real PostgreSQL service.

## Mobile flow

```text
Sign in
  -> assigned work
      -> work-order detail
          -> status transition
          -> technician note
          -> offline queue on failure
      -> equipment scanner
      -> audit trail
```

For web:

```bash
cd mobile
npm install
npx expo install --check
EXPO_PUBLIC_API_URL=http://localhost:8081 npm run web
```

## Offline conflict model

The client queues operations, not whole records. Replaying an operation checks current server state again. If the server has moved to an incompatible state, the API returns `conflict`; it does not silently overwrite newer information.

That makes offline behavior visible and reviewable rather than hiding synchronization errors.
