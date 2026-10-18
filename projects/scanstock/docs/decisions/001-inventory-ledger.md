# ADR 001 — Append-only inventory ledger
## Context
A mutable quantity field hides why stock changed and makes retry correctness difficult.
## Decision
Represent every movement as an immutable transaction and derive stock from the sum.
## Alternatives considered
Directly update a product quantity column.
## Tradeoffs
Reads need aggregation or projections, but auditability, idempotency, and debugging improve substantially.
