# ADR 001 — Standard-library-first Go

## Context
The project needs to demonstrate Go rather than a framework's abstractions.

## Decision
Use `net/http`, explicit services, repository interfaces, and a small custom server-push WebSocket implementation.

## Alternatives considered
Gin, Echo, Chi, Gorilla WebSocket.

## Tradeoffs
More code is owned by the project, but the architecture is visible and the default build has zero third-party Go dependencies.
