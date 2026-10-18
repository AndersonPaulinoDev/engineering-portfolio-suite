CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
  id text PRIMARY KEY,
  email text UNIQUE NOT NULL,
  password_hash text NOT NULL,
  role text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS projects (
  id text PRIMARY KEY,
  name text NOT NULL,
  archived boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS agents (
  id text PRIMARY KEY,
  name text NOT NULL,
  project_id text REFERENCES projects(id) ON DELETE CASCADE,
  type text NOT NULL,
  state text NOT NULL,
  current_task_id text,
  started_at timestamptz NOT NULL,
  last_heartbeat timestamptz NOT NULL,
  latest_message text NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS tasks (
  id text PRIMARY KEY,
  project_id text REFERENCES projects(id) ON DELETE CASCADE,
  agent_id text REFERENCES agents(id) ON DELETE SET NULL,
  title text NOT NULL,
  description text NOT NULL DEFAULT '',
  status text NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);
CREATE TABLE IF NOT EXISTS events (
  id text PRIMARY KEY,
  type text NOT NULL,
  project_id text REFERENCES projects(id) ON DELETE CASCADE,
  agent_id text REFERENCES agents(id) ON DELETE SET NULL,
  task_id text REFERENCES tasks(id) ON DELETE SET NULL,
  payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_events_created_at ON events(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_events_project_id ON events(project_id);
CREATE TABLE IF NOT EXISTS approvals (
  id text PRIMARY KEY,
  project_id text REFERENCES projects(id) ON DELETE CASCADE,
  agent_id text REFERENCES agents(id) ON DELETE SET NULL,
  task_id text REFERENCES tasks(id) ON DELETE SET NULL,
  action text NOT NULL,
  reason text NOT NULL DEFAULT '',
  status text NOT NULL,
  created_at timestamptz NOT NULL,
  decided_at timestamptz
);
