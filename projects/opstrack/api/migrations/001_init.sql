CREATE TABLE users(id text PRIMARY KEY,email text UNIQUE NOT NULL,password_hash text NOT NULL,role text NOT NULL,created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE work_orders(id text PRIMARY KEY,title text NOT NULL,description text NOT NULL DEFAULT '',location text NOT NULL DEFAULT '',priority text NOT NULL,assigned_employee text REFERENCES users(id),due_date timestamptz,status text NOT NULL,instructions text NOT NULL DEFAULT '',equipment_tag text,attachments jsonb NOT NULL DEFAULT '[]',completion_evidence jsonb NOT NULL DEFAULT '[]',notes jsonb NOT NULL DEFAULT '[]',updated_at timestamptz NOT NULL DEFAULT now());
CREATE INDEX idx_work_orders_assignee_status ON work_orders(assigned_employee,status);
CREATE INDEX idx_work_orders_equipment_tag ON work_orders(equipment_tag);
CREATE TABLE audit_events(id text PRIMARY KEY,actor text NOT NULL,action text NOT NULL,work_order_id text REFERENCES work_orders(id),metadata jsonb NOT NULL DEFAULT '{}',timestamp timestamptz NOT NULL DEFAULT now());
CREATE INDEX idx_audit_work_order_time ON audit_events(work_order_id,timestamp DESC);
CREATE TABLE sync_operations(id text PRIMARY KEY,user_id text NOT NULL,state text NOT NULL,error text,created_at timestamptz NOT NULL DEFAULT now());
