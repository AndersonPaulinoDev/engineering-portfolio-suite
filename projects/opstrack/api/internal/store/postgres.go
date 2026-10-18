//go:build postgres

package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/AndersonPaulinoDev/opstrack/api/internal/domain"
	_ "github.com/lib/pq"
)

type PostgresStore struct{ db *sql.DB }

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

func (p *PostgresStore) CreateUser(ctx context.Context, u domain.User) error {
	_, err := p.db.ExecContext(ctx, `INSERT INTO users(id,email,password_hash,role) VALUES($1,$2,$3,$4) ON CONFLICT(email) DO NOTHING`, u.ID, u.Email, u.PasswordHash, u.Role)
	return err
}
func (p *PostgresStore) UserByEmail(ctx context.Context, email string) (domain.User, error) {
	var u domain.User
	err := p.db.QueryRowContext(ctx, `SELECT id,email,password_hash,role FROM users WHERE lower(email)=lower($1)`, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

func scanWork(scanner interface{ Scan(...any) error }) (domain.WorkOrder, error) {
	var w domain.WorkOrder
	var attachments, evidence, notes []byte
	err := scanner.Scan(&w.ID, &w.Title, &w.Description, &w.Location, &w.Priority, &w.AssignedEmployee, &w.DueDate, &w.Status, &w.Instructions, &w.EquipmentTag, &attachments, &evidence, &notes, &w.UpdatedAt)
	if err != nil {
		return w, err
	}
	_ = json.Unmarshal(attachments, &w.Attachments)
	_ = json.Unmarshal(evidence, &w.CompletionEvidence)
	_ = json.Unmarshal(notes, &w.Notes)
	return w, nil
}

func (p *PostgresStore) ListWork(ctx context.Context, user string, role domain.Role) ([]domain.WorkOrder, error) {
	q := `SELECT id,title,description,location,priority,COALESCE(assigned_employee,''),due_date,status,instructions,COALESCE(equipment_tag,''),attachments,completion_evidence,notes,updated_at FROM work_orders`
	args := []any{}
	if role == domain.RoleWorker {
		q += ` WHERE assigned_employee=$1`
		args = append(args, user)
	}
	q += ` ORDER BY due_date`
	rows, err := p.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.WorkOrder{}
	for rows.Next() {
		w, err := scanWork(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}
func (p *PostgresStore) GetWork(ctx context.Context, id string) (domain.WorkOrder, error) {
	w, err := scanWork(p.db.QueryRowContext(ctx, `SELECT id,title,description,location,priority,COALESCE(assigned_employee,''),due_date,status,instructions,COALESCE(equipment_tag,''),attachments,completion_evidence,notes,updated_at FROM work_orders WHERE id=$1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return w, ErrNotFound
	}
	return w, err
}
func (p *PostgresStore) SaveWork(ctx context.Context, w domain.WorkOrder) error {
	a, _ := json.Marshal(w.Attachments)
	e, _ := json.Marshal(w.CompletionEvidence)
	n, _ := json.Marshal(w.Notes)
	_, err := p.db.ExecContext(ctx, `INSERT INTO work_orders(id,title,description,location,priority,assigned_employee,due_date,status,instructions,equipment_tag,attachments,completion_evidence,notes,updated_at) VALUES($1,$2,$3,$4,$5,NULLIF($6,''),$7,$8,$9,NULLIF($10,''),$11,$12,$13,$14) ON CONFLICT(id) DO UPDATE SET title=excluded.title,description=excluded.description,location=excluded.location,priority=excluded.priority,assigned_employee=excluded.assigned_employee,due_date=excluded.due_date,status=excluded.status,instructions=excluded.instructions,equipment_tag=excluded.equipment_tag,attachments=excluded.attachments,completion_evidence=excluded.completion_evidence,notes=excluded.notes,updated_at=excluded.updated_at`, w.ID, w.Title, w.Description, w.Location, w.Priority, w.AssignedEmployee, w.DueDate, w.Status, w.Instructions, w.EquipmentTag, a, e, n, w.UpdatedAt)
	return err
}
func (p *PostgresStore) ByEquipment(ctx context.Context, tag string) ([]domain.WorkOrder, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT id,title,description,location,priority,COALESCE(assigned_employee,''),due_date,status,instructions,COALESCE(equipment_tag,''),attachments,completion_evidence,notes,updated_at FROM work_orders WHERE equipment_tag=$1 ORDER BY due_date`, tag)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.WorkOrder{}
	for rows.Next() {
		w, err := scanWork(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}
func (p *PostgresStore) AppendAudit(ctx context.Context, e domain.AuditEntry) error {
	m, _ := json.Marshal(e.Metadata)
	_, err := p.db.ExecContext(ctx, `INSERT INTO audit_events(id,actor,action,work_order_id,metadata,timestamp) VALUES($1,$2,$3,NULLIF($4,''),$5,$6)`, e.ID, e.Actor, e.Action, e.WorkOrderID, m, e.Timestamp)
	return err
}
func (p *PostgresStore) ListAudit(ctx context.Context) ([]domain.AuditEntry, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT id,actor,action,COALESCE(work_order_id,''),metadata,timestamp FROM audit_events ORDER BY timestamp DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.AuditEntry{}
	for rows.Next() {
		var e domain.AuditEntry
		var raw []byte
		if err := rows.Scan(&e.ID, &e.Actor, &e.Action, &e.WorkOrderID, &raw, &e.Timestamp); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(raw, &e.Metadata)
		out = append(out, e)
	}
	return out, rows.Err()
}
func (p *PostgresStore) SyncResult(ctx context.Context, id string) (domain.SyncResult, bool, error) {
	var v domain.SyncResult
	err := p.db.QueryRowContext(ctx, `SELECT id,state,COALESCE(error,'') FROM sync_operations WHERE id=$1`, id).Scan(&v.ID, &v.State, &v.Error)
	if errors.Is(err, sql.ErrNoRows) {
		return v, false, nil
	}
	return v, err == nil, err
}
func (p *PostgresStore) SaveSyncResult(ctx context.Context, v domain.SyncResult, user string) error {
	_, err := p.db.ExecContext(ctx, `INSERT INTO sync_operations(id,user_id,state,error) VALUES($1,$2,$3,NULLIF($4,'')) ON CONFLICT(id) DO NOTHING`, v.ID, user, v.State, v.Error)
	return err
}
