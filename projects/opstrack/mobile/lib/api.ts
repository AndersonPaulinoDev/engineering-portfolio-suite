export type Status =
  | 'new'
  | 'assigned'
  | 'accepted'
  | 'in_progress'
  | 'blocked'
  | 'completed'
  | 'verified'
  | 'cancelled';

export type WorkOrder = {
  id: string;
  title: string;
  description: string;
  location: string;
  priority: string;
  assigned_employee: string;
  due_date: string;
  status: Status;
  instructions: string;
  equipment_tag: string;
  notes: string[];
  completion_evidence: string[];
  updated_at: string;
};

export type AuditEntry = {
  id: string;
  actor: string;
  action: string;
  work_order: string;
  metadata?: Record<string, string>;
  timestamp: string;
};

export type SyncOperation = {
  id: string;
  type: 'transition' | 'note' | 'evidence';
  work_order_id: string;
  status?: Status;
  note?: string;
  evidence?: string;
};

export type SyncResult = {
  id: string;
  state: 'synced' | 'conflict' | 'failed';
  error?: string;
};

const base = process.env.EXPO_PUBLIC_API_URL ?? 'http://localhost:8081';

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

export function login(email: string, password: string) {
  return request<{ access_token: string; user: { id: string; email: string; role: string } }>(
    '/api/v1/auth/login',
    undefined,
    { method: 'POST', body: JSON.stringify({ email, password }) },
  );
}

export const listWork = (token: string) => request<WorkOrder[]>('/api/v1/work-orders', token);
export const getWorkOrder = (token: string, id: string) =>
  request<WorkOrder>(`/api/v1/work-orders/${id}`, token);
export const listAudit = (token: string) => request<AuditEntry[]>('/api/v1/audit', token);
export const byEquipment = (token: string, tag: string) =>
  request<WorkOrder[]>(`/api/v1/equipment/${encodeURIComponent(tag)}/work-orders`, token);

export function transition(token: string, id: string, status: Status) {
  return request<WorkOrder>(`/api/v1/work-orders/${id}/transition`, token, {
    method: 'POST',
    body: JSON.stringify({ status }),
  });
}

export function addNote(token: string, id: string, note: string) {
  return request<WorkOrder>(`/api/v1/work-orders/${id}/notes`, token, {
    method: 'POST',
    body: JSON.stringify({ note }),
  });
}

export function addEvidence(token: string, id: string, uri: string) {
  return request<WorkOrder>(`/api/v1/work-orders/${id}/evidence`, token, {
    method: 'POST',
    body: JSON.stringify({ uri }),
  });
}

export function syncOps(token: string, operations: SyncOperation[]) {
  return request<SyncResult[]>('/api/v1/sync', token, {
    method: 'POST',
    body: JSON.stringify(operations),
  });
}
