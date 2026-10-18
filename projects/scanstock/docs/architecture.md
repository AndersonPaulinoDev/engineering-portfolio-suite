# ScanStock Architecture

## Goal

Inventory must remain correct when mobile clients retry requests, work offline, or submit movements concurrently. The design therefore makes transactions—not a mutable quantity column—the source of truth.

## Components

```text
Expo app
  |-- SQLite pending_ops queue
  |-- barcode scanner
  |-- product detail
  |-- ledger view
  |
  v
Go HTTP API
  |-- authentication
  |-- movement normalization
  |-- idempotency handling
  |
  v
Store interface
  |-- MemoryStore for tests
  `-- PostgresStore for runtime
          |-- users
          |-- products
          `-- inventory_transactions
```

## Inventory ledger

Every stock change creates an `inventory_transactions` row. Stock on hand is calculated as the sum of movement quantities for a product.

Examples:

```text
+40 received
-5 shipped
-2 damaged
+1 correction
```

## Idempotency

Every mobile write supplies an idempotency key. PostgreSQL stores that key under a unique constraint. A retry looks up and returns the existing transaction instead of creating another movement.

## Transaction isolation

The PostgreSQL adapter records a movement inside a serializable transaction. It checks current stock and refuses an outbound write that would make stock negative.

## Offline model

The mobile SQLite queue stores operation payloads until the server is reachable. Replayed operations use their queue ID as the idempotency key, so network retries and manual resyncs do not duplicate stock.

## Authentication

Users are persisted through the repository. `AUTH_SECRET` is mandatory. Demo credentials are seeded only under `DEMO_MODE=true`.

## Testing

- service tests cover ledger quantities, idempotency, negative stock, sync replay, and barcode lookup
- HTTP tests exercise products and duplicate transaction submission
- PostgreSQL integration tests verify real ledger writes and replay semantics
