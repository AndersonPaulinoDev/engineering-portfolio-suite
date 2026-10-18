# OpsTrack Architecture

## Goal

Field work must remain usable when a technician loses connectivity. The design therefore treats the mobile device as an occasionally connected client, not as a permanently online frontend.

## Components

```text
Expo app
  |-- SQLite pending_ops queue
  |-- work-order screens
  |-- equipment scanner
  |
  v
Go HTTP API
  |-- auth / RBAC
  |-- transition validation
  |-- sync replay
  |
  v
Store interface
  |-- MemoryStore for isolated tests
  `-- PostgresStore for runtime
          |-- users
          |-- work_orders
          |-- audit_events
          `-- sync_operations
```

## Work-order state machine

The server validates all transitions. A worker cannot jump directly from `assigned` to `completed`, and a worker cannot verify their own completed job. These rules are enforced server-side even if a client sends a handcrafted request.

## Offline synchronization

The mobile app queues operations such as:

```text
transition(work_order, status)
note(work_order, text)
evidence(work_order, uri)
```

It does not queue full-record replacements. On reconnect, `/api/v1/sync` replays each operation against current server state.

Each operation ID is persisted in `sync_operations`. Replaying an already-processed ID returns the stored result, which makes retries safe.

## Conflict behavior

If a queued transition is no longer valid because the server changed while the device was offline, the operation returns `conflict`. The newer server state is not overwritten.

## Auditability

Status changes, notes, and evidence create append-only audit records. Normal API users cannot edit audit entries.

## Authentication and persistence

Auth users live in the same repository abstraction as work data, so PostgreSQL foreign keys remain valid. `AUTH_SECRET` is required. Demo users exist only under `DEMO_MODE=true`.

## Testing

- service tests cover transition rules, audit creation, offline idempotency, and note/evidence persistence
- HTTP tests perform login, list work, transition a job, add a note, and read audit history
- PostgreSQL integration tests verify user/work-order/audit round trips
