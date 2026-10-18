export type TransactionType =
  | 'received'
  | 'shipped'
  | 'damaged'
  | 'correction'
  | 'count_adjustment'
  | 'transfer_in'
  | 'transfer_out';

export type Product = {
  id: string;
  sku: string;
  upc: string;
  name: string;
  description: string;
  location: string;
  reorder_level: number;
  created_at: string;
  updated_at: string;
};

export type Stock = {
  product: Product;
  quantity: number;
  needs_reorder: boolean;
};

export type InventoryTransaction = {
  transaction_id: string;
  product_id: string;
  quantity_change: number;
  transaction_type: TransactionType;
  user_id: string;
  device_id: string;
  reference: string;
  timestamp: string;
  idempotency_key: string;
};

export type SyncOperation = {
  id: string;
  product_id: string;
  quantity_change: number;
  transaction_type: TransactionType;
  device_id: string;
  reference: string;
};

export type SyncResult = {
  id: string;
  state: 'synced' | 'conflict' | 'failed';
  transaction?: InventoryTransaction;
  error?: string;
};

const base = process.env.EXPO_PUBLIC_API_URL ?? 'http://localhost:8082';

async function request<T>(path: string, token?: string, init: RequestInit = {}): Promise<T> {
  const response = await fetch(`${base}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(init.headers ?? {}),
    },
  });
  const body = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(body?.error?.message ?? `HTTP ${response.status}`);
  }
  return body.data as T;
}

export const login = (email: string, password: string) =>
  request<{ access_token: string; user: { id: string; email: string; role: string } }>(
    '/api/v1/auth/login',
    undefined,
    { method: 'POST', body: JSON.stringify({ email, password }) },
  );

export const listStock = (token: string) => request<Stock[]>('/api/v1/products', token);
export const byCode = (token: string, code: string) =>
  request<Stock>(`/api/v1/products/code/${encodeURIComponent(code)}`, token);
export const listTransactions = (token: string, productId = '') =>
  request<InventoryTransaction[]>(
    `/api/v1/transactions${productId ? `?product_id=${encodeURIComponent(productId)}` : ''}`,
    token,
  );

export function transact(
  token: string,
  input: {
    product_id: string;
    quantity_change: number;
    transaction_type: TransactionType;
    device_id: string;
    reference: string;
    idempotency_key: string;
  },
) {
  return request<InventoryTransaction>('/api/v1/transactions', token, {
    method: 'POST',
    body: JSON.stringify(input),
  });
}

export function syncOps(token: string, operations: SyncOperation[]) {
  return request<SyncResult[]>('/api/v1/sync', token, {
    method: 'POST',
    body: JSON.stringify(operations),
  });
}
