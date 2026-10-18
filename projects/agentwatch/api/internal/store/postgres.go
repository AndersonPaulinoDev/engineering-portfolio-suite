//go:build postgres

package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/AndersonPaulinoDev/agentwatch/api/internal/domain"
	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &PostgresStore{db: db}, nil
}

func (p *PostgresStore) Close() error { return p.db.Close() }

func (p *PostgresStore) CreateUser(ctx context.Context, v domain.User) error {
	_, err := p.db.ExecContext(ctx, `INSERT INTO users(id,email,password_hash,role,created_at) VALUES($1,$2,$3,$4,$5)`, v.ID, v.Email, v.PasswordHash, v.Role, v.CreatedAt)
	if err != nil {
		return ErrConflict
	}
	return nil
}
func (p *PostgresStore) UserByEmail(ctx context.Context, email string) (domain.User, error) {
	var v domain.User
	err := p.db.QueryRowContext(ctx, `SELECT id,email,password_hash,role,created_at FROM users WHERE lower(email)=lower($1)`, email).Scan(&v.ID, &v.Email, &v.PasswordHash, &v.Role, &v.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return v, ErrNotFound
	}
	return v, err
}
func (p *PostgresStore) UserByID(ctx context.Context, id string) (domain.User, error) {
	var v domain.User
	err := p.db.QueryRowContext(ctx, `SELECT id,email,password_hash,role,created_at FROM users WHERE id=$1`, id).Scan(&v.ID, &v.Email, &v.PasswordHash, &v.Role, &v.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return v, ErrNotFound
	}
	return v, err
}
func (p *PostgresStore) CreateProject(ctx context.Context, v domain.Project) error {
	_, err := p.db.ExecContext(ctx, `INSERT INTO projects(id,name,archived,created_at) VALUES($1,$2,$3,$4)`, v.ID, v.Name, v.Archived, v.CreatedAt)
	return err
}
func (p *PostgresStore) ListProjects(ctx context.Context) ([]domain.Project, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT id,name,archived,created_at FROM projects ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Project{}
	for rows.Next() {
		var v domain.Project
		if err := rows.Scan(&v.ID, &v.Name, &v.Archived, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (p *PostgresStore) UpsertAgent(ctx context.Context, v domain.Agent) error {
	_, err := p.db.ExecContext(ctx, `INSERT INTO agents(id,name,project_id,type,state,current_task_id,started_at,last_heartbeat,latest_message)
VALUES($1,$2,NULLIF($3,''),$4,$5,NULLIF($6,''),$7,$8,$9)
ON CONFLICT(id) DO UPDATE SET name=excluded.name,project_id=excluded.project_id,type=excluded.type,state=excluded.state,current_task_id=excluded.current_task_id,last_heartbeat=excluded.last_heartbeat,latest_message=excluded.latest_message`, v.ID, v.Name, v.ProjectID, v.Type, v.State, v.CurrentTaskID, v.StartedAt, v.LastHeartbeat, v.LatestMessage)
	return err
}
func (p *PostgresStore) ListAgents(ctx context.Context) ([]domain.Agent, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT id,name,COALESCE(project_id,''),type,state,COALESCE(current_task_id,''),started_at,last_heartbeat,latest_message FROM agents ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Agent{}
	for rows.Next() {
		var v domain.Agent
		if err := rows.Scan(&v.ID, &v.Name, &v.ProjectID, &v.Type, &v.State, &v.CurrentTaskID, &v.StartedAt, &v.LastHeartbeat, &v.LatestMessage); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (p *PostgresStore) AgentByID(ctx context.Context, id string) (domain.Agent, error) {
	var v domain.Agent
	err := p.db.QueryRowContext(ctx, `SELECT id,name,COALESCE(project_id,''),type,state,COALESCE(current_task_id,''),started_at,last_heartbeat,latest_message FROM agents WHERE id=$1`, id).Scan(&v.ID, &v.Name, &v.ProjectID, &v.Type, &v.State, &v.CurrentTaskID, &v.StartedAt, &v.LastHeartbeat, &v.LatestMessage)
	if errors.Is(err, sql.ErrNoRows) {
		return v, ErrNotFound
	}
	return v, err
}
func (p *PostgresStore) CreateTask(ctx context.Context, v domain.Task) error {
	_, err := p.db.ExecContext(ctx, `INSERT INTO tasks(id,project_id,agent_id,title,description,status,created_at,updated_at) VALUES($1,NULLIF($2,''),NULLIF($3,''),$4,$5,$6,$7,$8)`, v.ID, v.ProjectID, v.AgentID, v.Title, v.Description, v.Status, v.CreatedAt, v.UpdatedAt)
	return err
}
func (p *PostgresStore) UpdateTask(ctx context.Context, v domain.Task) error {
	r, err := p.db.ExecContext(ctx, `UPDATE tasks SET project_id=NULLIF($2,''),agent_id=NULLIF($3,''),title=$4,description=$5,status=$6,updated_at=$7 WHERE id=$1`, v.ID, v.ProjectID, v.AgentID, v.Title, v.Description, v.Status, v.UpdatedAt)
	if err != nil {
		return err
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
func (p *PostgresStore) TaskByID(ctx context.Context, id string) (domain.Task, error) {
	var v domain.Task
	err := p.db.QueryRowContext(ctx, `SELECT id,COALESCE(project_id,''),COALESCE(agent_id,''),title,description,status,created_at,updated_at FROM tasks WHERE id=$1`, id).Scan(&v.ID, &v.ProjectID, &v.AgentID, &v.Title, &v.Description, &v.Status, &v.CreatedAt, &v.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return v, ErrNotFound
	}
	return v, err
}
func (p *PostgresStore) ListTasks(ctx context.Context) ([]domain.Task, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT id,COALESCE(project_id,''),COALESCE(agent_id,''),title,description,status,created_at,updated_at FROM tasks ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Task{}
	for rows.Next() {
		var v domain.Task
		if err := rows.Scan(&v.ID, &v.ProjectID, &v.AgentID, &v.Title, &v.Description, &v.Status, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (p *PostgresStore) AppendEvent(ctx context.Context, v domain.Event) error {
	payload, _ := json.Marshal(v.Payload)
	_, err := p.db.ExecContext(ctx, `INSERT INTO events(id,type,project_id,agent_id,task_id,payload,created_at) VALUES($1,$2,NULLIF($3,''),NULLIF($4,''),NULLIF($5,''),$6,$7)`, v.ID, v.Type, v.ProjectID, v.AgentID, v.TaskID, payload, v.CreatedAt)
	return err
}
func (p *PostgresStore) ListEvents(ctx context.Context, limit int) ([]domain.Event, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := p.db.QueryContext(ctx, `SELECT id,type,COALESCE(project_id,''),COALESCE(agent_id,''),COALESCE(task_id,''),payload,created_at FROM events ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Event{}
	for rows.Next() {
		var v domain.Event
		var raw []byte
		if err := rows.Scan(&v.ID, &v.Type, &v.ProjectID, &v.AgentID, &v.TaskID, &raw, &v.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(raw, &v.Payload)
		out = append(out, v)
	}
	return out, rows.Err()
}
func (p *PostgresStore) CreateApproval(ctx context.Context, v domain.Approval) error {
	_, err := p.db.ExecContext(ctx, `INSERT INTO approvals(id,project_id,agent_id,task_id,action,reason,status,created_at,decided_at) VALUES($1,NULLIF($2,''),NULLIF($3,''),NULLIF($4,''),$5,$6,$7,$8,$9)`, v.ID, v.ProjectID, v.AgentID, v.TaskID, v.Action, v.Reason, v.Status, v.CreatedAt, v.DecidedAt)
	return err
}
func (p *PostgresStore) ApprovalByID(ctx context.Context, id string) (domain.Approval, error) {
	var v domain.Approval
	err := p.db.QueryRowContext(ctx, `SELECT id,COALESCE(project_id,''),COALESCE(agent_id,''),COALESCE(task_id,''),action,reason,status,created_at,decided_at FROM approvals WHERE id=$1`, id).Scan(&v.ID, &v.ProjectID, &v.AgentID, &v.TaskID, &v.Action, &v.Reason, &v.Status, &v.CreatedAt, &v.DecidedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return v, ErrNotFound
	}
	return v, err
}
func (p *PostgresStore) UpdateApproval(ctx context.Context, v domain.Approval) error {
	r, err := p.db.ExecContext(ctx, `UPDATE approvals SET status=$2,decided_at=$3 WHERE id=$1`, v.ID, v.Status, v.DecidedAt)
	if err != nil {
		return err
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
func (p *PostgresStore) ListApprovals(ctx context.Context) ([]domain.Approval, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT id,COALESCE(project_id,''),COALESCE(agent_id,''),COALESCE(task_id,''),action,reason,status,created_at,decided_at FROM approvals ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Approval{}
	for rows.Next() {
		var v domain.Approval
		if err := rows.Scan(&v.ID, &v.ProjectID, &v.AgentID, &v.TaskID, &v.Action, &v.Reason, &v.Status, &v.CreatedAt, &v.DecidedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
