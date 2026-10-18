package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AndersonPaulinoDev/opstrack/api/internal/domain"
	"github.com/AndersonPaulinoDev/opstrack/api/internal/store"
)

type Service struct{ store store.Store }

func New(st store.Store) *Service { return &Service{store: st} }

func (s *Service) SeedDemo(ctx context.Context) error {
	now := time.Now().UTC()
	rows := []domain.WorkOrder{
		{ID: "wo_1001", Title: "Inspect rooftop HVAC", Description: "Verify filter, belt and condensate line.", Location: "Building A - Roof", Priority: "high", AssignedEmployee: "usr_worker", DueDate: now.Add(6 * time.Hour), Status: domain.StatusAssigned, Instructions: "Photograph filter before and after.", EquipmentTag: "HVAC-A-04", UpdatedAt: now},
		{ID: "wo_1002", Title: "Replace loading dock light", Description: "LED fixture is intermittent.", Location: "Dock 3", Priority: "medium", AssignedEmployee: "usr_worker", DueDate: now.Add(24 * time.Hour), Status: domain.StatusAccepted, Instructions: "Lock out circuit before replacement.", EquipmentTag: "ELEC-D3-12", UpdatedAt: now},
		{ID: "wo_1003", Title: "Inspect fire door", Description: "Monthly compliance check.", Location: "North stairwell", Priority: "low", AssignedEmployee: "usr_worker", DueDate: now.Add(48 * time.Hour), Status: domain.StatusInProgress, Instructions: "Record latch and closer condition.", EquipmentTag: "FIRE-N-02", UpdatedAt: now},
	}
	for _, row := range rows {
		if err := s.store.SaveWork(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) List(ctx context.Context, user string, role domain.Role) ([]domain.WorkOrder, error) {
	return s.store.ListWork(ctx, user, role)
}
func (s *Service) Get(ctx context.Context, id string) (domain.WorkOrder, error) {
	w, err := s.store.GetWork(ctx, id)
	if errors.Is(err, store.ErrNotFound) {
		return domain.WorkOrder{}, errors.New("work order not found")
	}
	return w, err
}
func (s *Service) Transition(ctx context.Context, actor string, role domain.Role, id string, to domain.Status) (domain.WorkOrder, error) {
	w, err := s.Get(ctx, id)
	if err != nil {
		return w, err
	}
	if role == domain.RoleWorker && w.AssignedEmployee != actor {
		return w, errors.New("not assigned to this worker")
	}
	if !allowed(w.Status, to, role) {
		return w, fmt.Errorf("invalid transition %s -> %s", w.Status, to)
	}
	from := w.Status
	w.Status = to
	w.UpdatedAt = time.Now().UTC()
	if err := s.store.SaveWork(ctx, w); err != nil {
		return w, err
	}
	_ = s.store.AppendAudit(ctx, domain.AuditEntry{ID: newID("aud"), Actor: actor, Action: "work_order.status_changed", WorkOrderID: id, Metadata: map[string]string{"from": string(from), "to": string(to)}, Timestamp: w.UpdatedAt})
	return w, nil
}
func allowed(from, to domain.Status, role domain.Role) bool {
	transitions := map[domain.Status][]domain.Status{
		domain.StatusNew:        {domain.StatusAssigned, domain.StatusCancelled},
		domain.StatusAssigned:   {domain.StatusAccepted, domain.StatusCancelled},
		domain.StatusAccepted:   {domain.StatusInProgress, domain.StatusBlocked, domain.StatusCancelled},
		domain.StatusInProgress: {domain.StatusBlocked, domain.StatusCompleted},
		domain.StatusBlocked:    {domain.StatusInProgress, domain.StatusCancelled},
		domain.StatusCompleted:  {domain.StatusVerified, domain.StatusInProgress},
		domain.StatusVerified:   {},
		domain.StatusCancelled:  {},
	}
	for _, candidate := range transitions[from] {
		if candidate == to {
			return !(to == domain.StatusVerified && role == domain.RoleWorker)
		}
	}
	return false
}
func (s *Service) AddNote(ctx context.Context, actor, id, note string) (domain.WorkOrder, error) {
	note = strings.TrimSpace(note)
	if note == "" {
		return domain.WorkOrder{}, errors.New("note required")
	}
	w, err := s.Get(ctx, id)
	if err != nil {
		return w, err
	}
	w.Notes = append(w.Notes, note)
	w.UpdatedAt = time.Now().UTC()
	if err := s.store.SaveWork(ctx, w); err != nil {
		return w, err
	}
	_ = s.store.AppendAudit(ctx, domain.AuditEntry{ID: newID("aud"), Actor: actor, Action: "work_order.note_added", WorkOrderID: id, Timestamp: w.UpdatedAt})
	return w, nil
}
func (s *Service) AddEvidence(ctx context.Context, actor, id, uri string) (domain.WorkOrder, error) {
	uri = strings.TrimSpace(uri)
	if uri == "" {
		return domain.WorkOrder{}, errors.New("evidence uri required")
	}
	w, err := s.Get(ctx, id)
	if err != nil {
		return w, err
	}
	w.CompletionEvidence = append(w.CompletionEvidence, uri)
	w.UpdatedAt = time.Now().UTC()
	if err := s.store.SaveWork(ctx, w); err != nil {
		return w, err
	}
	_ = s.store.AppendAudit(ctx, domain.AuditEntry{ID: newID("aud"), Actor: actor, Action: "work_order.evidence_added", WorkOrderID: id, Timestamp: w.UpdatedAt})
	return w, nil
}
func (s *Service) ByEquipment(ctx context.Context, tag string) ([]domain.WorkOrder, error) {
	return s.store.ByEquipment(ctx, tag)
}
func (s *Service) Audit(ctx context.Context) ([]domain.AuditEntry, error) {
	return s.store.ListAudit(ctx)
}
func (s *Service) Sync(ctx context.Context, actor string, role domain.Role, ops []domain.SyncOperation) []domain.SyncResult {
	results := make([]domain.SyncResult, 0, len(ops))
	for _, op := range ops {
		if old, ok, err := s.store.SyncResult(ctx, op.ID); err == nil && ok {
			results = append(results, old)
			continue
		}
		result := domain.SyncResult{ID: op.ID, State: "synced"}
		switch op.Type {
		case "transition":
			if _, err := s.Transition(ctx, actor, role, op.WorkOrderID, op.Status); err != nil {
				result.State = "conflict"
				result.Error = err.Error()
			}
		case "note":
			if _, err := s.AddNote(ctx, actor, op.WorkOrderID, op.Note); err != nil {
				result.State = "failed"
				result.Error = err.Error()
			}
		case "evidence":
			if _, err := s.AddEvidence(ctx, actor, op.WorkOrderID, op.Evidence); err != nil {
				result.State = "failed"
				result.Error = err.Error()
			}
		default:
			result.State = "failed"
			result.Error = "unsupported operation"
		}
		_ = s.store.SaveSyncResult(ctx, result, actor)
		results = append(results, result)
	}
	return results
}
func newID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return prefix + "_" + hex.EncodeToString(b)
}
