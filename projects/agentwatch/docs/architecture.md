# AgentWatch Architecture

## Goals

AgentWatch must make asynchronous developer work observable without making the realtime transport the source of truth. The REST API and PostgreSQL store remain authoritative; WebSocket messages are low-latency notifications.

## Components

```text
Expo client
  |-- REST queries and mutations
  `-- WebSocket notifications
          |
      HTTP server
          |
  auth middleware + handlers
          |
   AgentWatch service
      |          |
      |          `--> realtime Hub
      |
      `--> Store interface
             |-- MemoryStore (tests)
             `-- PostgresStore (runtime)
```

## Persistence

`store.Store` owns persistence contracts for users, projects, agents, tasks, events, and approvals. Domain services do not know which implementation is active.

The server does not silently choose memory persistence. A normal build requires explicit `PERSISTENCE=memory`; the Docker/CI runtime uses `-tags postgres` and `DATABASE_URL`.

## Realtime model

A domain change is persisted first, then an event is broadcast. If a realtime client disconnects or drops a message, it can recover by re-querying REST state.

The hub uses channels and one subscription channel per client. Slow subscribers cannot block domain writes indefinitely.

## WebSocket middleware requirement

The HTTP logging wrapper preserves `http.Hijacker`. A regression test opens a raw WebSocket upgrade, verifies `101 Switching Protocols`, triggers a domain event, and reads the delivered text frame. This test exists because an earlier logging wrapper broke WebSocket upgrades while ordinary HTTP tests still passed.

## Authentication

Passwords are PBKDF2-hashed. Access and refresh tokens are HMAC-signed. `AUTH_SECRET` is mandatory at startup. Demo credentials are seeded only under `DEMO_MODE=true`.

## Failure model

- Database errors return server errors rather than mutating memory as fallback.
- Realtime failure does not invalidate durable state.
- Invalid task transitions are rejected in the service layer.
- Approval decisions are one-way after leaving `pending`.
- Missing/invalid access tokens are rejected by middleware.

## Testing boundaries

- service tests cover auth, task transitions, and approval invariants
- HTTP tests cover authorization and routes consumed by the client
- realtime test covers actual upgrade and event delivery
- PostgreSQL integration test runs with `-tags postgres`
