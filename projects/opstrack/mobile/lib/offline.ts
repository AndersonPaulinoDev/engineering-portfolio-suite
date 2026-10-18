import * as SQLite from 'expo-sqlite';
import type { SyncOperation } from './api';

const db = SQLite.openDatabaseSync('opstrack.db');

db.execSync(`
  CREATE TABLE IF NOT EXISTS pending_ops (
    id TEXT PRIMARY KEY,
    payload TEXT NOT NULL,
    created_at TEXT NOT NULL
  );
`);

export function queueOperation(operation: SyncOperation) {
  db.runSync(
    'INSERT OR REPLACE INTO pending_ops(id, payload, created_at) VALUES(?, ?, ?)',
    operation.id,
    JSON.stringify(operation),
    new Date().toISOString(),
  );
}

export function pendingOperations(): SyncOperation[] {
  return db
    .getAllSync<{ id: string; payload: string }>(
      'SELECT id, payload FROM pending_ops ORDER BY created_at',
    )
    .map((row) => JSON.parse(row.payload) as SyncOperation);
}

export function removeOperation(id: string) {
  db.runSync('DELETE FROM pending_ops WHERE id = ?', id);
}
