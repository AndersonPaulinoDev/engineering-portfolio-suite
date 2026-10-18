# AgentWatch

Realtime control center for coding agents, background jobs, approvals, and project activity.

## Why this exists

AgentWatch is the flagship portfolio project. It demonstrates realtime backend engineering, explicit workflow state, authentication, PostgreSQL persistence, HTTP/WebSocket testing, and a shared Expo iOS/Android/web client.

## Product flow

A developer signs in, sees active projects and agents, inspects tasks, and handles approval requests without sitting in front of every terminal process. Domain events are written to the store and then broadcast to connected clients through the realtime hub.

## Implemented MVP

- register, login, refresh, logout
- PBKDF2 password hashing
- HMAC-signed access and refresh tokens
- PostgreSQL users, projects, agents, tasks, events, approvals
- explicit memory repository for isolated tests
- task transition validation
- approval requests and one-time decisions
- WebSocket event delivery
- request logging, health and readiness endpoints
- readable Expo dashboard for iOS, Android, and web
- HTTP tests plus a raw WebSocket handshake/event-delivery test
- OpenAPI schema and architecture ADRs

## Demo mode

The repository contains demo credentials for portfolio review:

```text
demo@agentwatch.dev
demo1234
```

They are created only when `DEMO_MODE=true`. The mobile client does not log in automatically.

## Run with PostgreSQL

The intended local runtime is Docker Compose:

```bash
cd projects/agentwatch
cp .env.example .env
# Set AUTH_SECRET in .env to the output of: openssl rand -hex 32
docker compose up --build
```

Compose starts PostgreSQL, applies the migration, waits for database readiness, and runs:

```bash
go run -tags postgres ./cmd/server
```

Health endpoint:

```text
http://localhost:8080/health
```

## Run backend tests

```bash
cd api
PERSISTENCE=memory go test ./...
go vet ./...
```

The memory store is for isolated tests. Server startup will not silently fall back to it. `AUTH_SECRET` must be at least 32 characters.

## Mobile/web

```bash
cd mobile
npm install
npx expo install --check
EXPO_PUBLIC_API_URL=http://localhost:8080 npm run web
```

For a physical device, use the development machine's LAN or Tailscale address.

## Architecture

```text
Expo iOS / Android / Web
        |
   HTTPS + WebSocket
        |
   HTTP middleware
        |
 Auth + domain services
        |
   Store interface
    /         \
Memory     PostgreSQL
 tests      runtime
        |
  durable events
        |
 realtime hub
        |
 connected clients
```

See `docs/architecture.md` for failure handling, concurrency, persistence, and testing boundaries.
