# ADR 001 — Operation queue for offline sync
## Context
Workers can lose connectivity after beginning a job.
## Decision
Persist mutations as uniquely identified operations in SQLite and replay them against the API.
## Alternatives considered
Last-write-wins document replacement and disabling edits while offline.
## Tradeoffs
Operation replay requires conflict logic, but avoids silently overwriting server changes.
