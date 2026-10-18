# ScanStock

Mobile-first inventory software built around an append-only movement ledger.

## What it demonstrates

ScanStock focuses on correctness under retry, concurrent clients, and intermittent connectivity. Stock on hand is derived from inventory transactions instead of maintained as a directly mutable quantity field.

## Implemented MVP

- PostgreSQL-backed users, products, and inventory transactions
- append-only inventory movement ledger
- stock-on-hand projection
- safe retry through unique idempotency keys
- serializable PostgreSQL write transaction
- negative-stock guard
- normalized inbound/outbound quantity semantics
- Expo SQLite offline queue
- barcode scanner
- product detail screen
- receive, ship, and damage actions
- transaction-ledger screen
- HTTP idempotency test
- OpenAPI schema and ADR

## Demo mode

Created only when `DEMO_MODE=true`:

```text
stock@scanstock.dev / demo1234
manager@scanstock.dev / demo1234
```

The app does not auto-login.

## Run with PostgreSQL

```bash
cd projects/scanstock
cp .env.example .env
# Set AUTH_SECRET in .env to the output of: openssl rand -hex 32
docker compose up --build
```

The API is started with:

```bash
go run -tags postgres ./cmd/server
```

## Backend tests

```bash
cd api
PERSISTENCE=memory go test ./...
go vet ./...
```

CI additionally exercises the PostgreSQL store with a real database.

## Ledger behavior

Receiving and transfer-in movements normalize to positive quantities. Shipping, damage, and transfer-out normalize to negative quantities. Corrections preserve the signed quantity supplied by the caller.

A retry with the same idempotency key returns the original transaction instead of applying inventory twice.

## Mobile flow

```text
Sign in
  -> stock dashboard
      -> product detail
          -> receive / ship / damage
          -> recent movements
      -> barcode scanner
      -> full transaction ledger
```
