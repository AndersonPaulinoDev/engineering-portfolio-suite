import * as SQLite from 'expo-sqlite';
import type { SyncOperation } from './api';

const db = SQLite.openDatabaseSync('scanstock.db');

db.execSync(`
  CREATE TABLE IF NOT EXISTS pending_ops (
    id TEXT PRIMARY KEY,
    payload TEXT NOT NULL,
    created_at TEXT NOT NULL
  );
`);

export function queue(operation: SyncOperation) {
  db.runSync(
    'INSERT OR REPLACE INTO pending_ops(id, payload, created_at) VALUES(?, ?, ?)',
    operation.id,
    JSON.stringify(operation),
    new Date().toISOString(),
  );
}

export function all(): SyncOperation[] {
  return db
    .getAllSync<{ payload: string }>('SELECT payload FROM pending_ops ORDER BY created_at')
    .map((row) => JSON.parse(row.payload) as SyncOperation);
}

export function remove(id: string) {
  db.runSync('DELETE FROM pending_ops WHERE id = ?', id);
}
