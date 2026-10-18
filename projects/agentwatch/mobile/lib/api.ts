export type Project = {
  id: string;
  name: string;
  archived: boolean;
  created_at: string;
};

export type Agent = {
  id: string;
  name: string;
  project_id: string;
  type: string;
  state: string;
  last_heartbeat: string;
  latest_message: string;
};

export type Task = {
  id: string;
  project_id: string;
  agent_id?: string;
  title: string;
  description: string;
  status: string;
  created_at: string;
};

export type Approval = {
  id: string;
  project_id: string;
  agent_id: string;
  task_id?: string;
  action: string;
  reason: string;
  status: string;
  created_at: string;
};

const base = process.env.EXPO_PUBLIC_API_URL ?? 'http://localhost:8080';

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
  return request<{
    user: { id: string; email: string; role: string };
    tokens: { access_token: string; refresh_token: string };
  }>('/api/v1/auth/login', undefined, {
    method: 'POST',
    body: JSON.stringify({ Email: email, Password: password }),
  });
}

export const listProjects = (token: string) => request<Project[]>('/api/v1/projects', token);
export const listAgents = (token: string) => request<Agent[]>('/api/v1/agents', token);
export const listTasks = (token: string) => request<Task[]>('/api/v1/tasks', token);
export const listApprovals = (token: string) => request<Approval[]>('/api/v1/approvals', token);

export function decideApproval(
  token: string,
  id: string,
  decision: 'approved' | 'rejected',
) {
  return request<Approval>(`/api/v1/approvals/${id}/decision`, token, {
    method: 'POST',
    body: JSON.stringify({ Decision: decision }),
  });
}

export function websocketUrl(token: string) {
  return `${base.replace(/^http/, 'ws')}/ws?token=${encodeURIComponent(token)}`;
}
