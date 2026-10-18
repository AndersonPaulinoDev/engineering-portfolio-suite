# Portfolio Demo Scripts

## AgentWatch — 3 minute demo
1. Sign in with the seeded developer account.
2. Show three projects and five seeded agents.
3. Open Tasks and explain why status transitions are validated by the Go service.
4. Open Approvals and approve or reject the seeded destructive-action request.
5. Trigger an agent heartbeat with the API and show the dashboard update over WebSocket.
6. Open `docs/architecture.md` and explain REST-as-source-of-truth plus realtime notification delivery.

**What this proves:** Go services, auth, event systems, concurrency, realtime mobile/web UX, state machines, technical decisions.

## OpsTrack — 3 minute demo
1. Sign in as the seeded worker.
2. Show assigned work orders and their operational state.
3. Move `wo_1001` from assigned -> accepted -> in progress.
4. Scan an equipment tag from the mobile client.
5. Explain that failed network mutations are written to the SQLite queue.
6. Reconnect and sync; show how a stale state returns `conflict` instead of silently overwriting server data.
7. Show the append-only audit endpoint.

**What this proves:** mobile product engineering, offline-first behavior, workflow correctness, device APIs, conflict handling, auditability.

## ScanStock — 3 minute demo
1. Show stock-on-hand values derived from ledger entries.
2. Scan a product barcode.
3. Receive one unit and show the quantity projection change.
4. Retry the same API request with the same idempotency key and show that inventory changes only once.
5. Attempt to ship more units than available and show the rejection.
6. Open the PostgreSQL migration and show the ledger table plus `stock_on_hand` view.

**What this proves:** transactional domain modeling, idempotency, mobile retry safety, inventory correctness, SQL design.
