# ADR 002 — WebSocket notifications with REST recovery

## Context
Agent state and approvals change while the mobile client is open.

## Decision
Use WebSockets for live notifications while retaining REST endpoints as the source of truth.

## Alternatives considered
Polling and Server-Sent Events.

## Tradeoffs
WebSockets require connection management, but support future bidirectional agent control. REST recovery prevents missed event notifications from corrupting state.
