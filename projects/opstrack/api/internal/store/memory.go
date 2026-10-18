package store

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"

	"github.com/AndersonPaulinoDev/opstrack/api/internal/domain"
)

type MemoryStore struct {
	mu        sync.RWMutex
	users     map[string]domain.User
	work      map[string]domain.WorkOrder
	audit     []domain.AuditEntry
	processed map[string]domain.SyncResult
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users:     map[string]domain.User{},
		work:      map[string]domain.WorkOrder{},
		audit:     []domain.AuditEntry{},
		processed: map[string]domain.SyncResult{},
	}
}

func (m *MemoryStore) CreateUser(_ context.Context, u domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, existing := range m.users {
		if strings.EqualFold(existing.Email, u.Email) {
			return errors.New("user already exists")
		}
	}
	m.users[u.ID] = u
	return nil
}
func (m *MemoryStore) UserByEmail(_ context.Context, email string) (domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, u := range m.users {
		if strings.EqualFold(u.Email, email) {
			return u, nil
		}
	}
	return domain.User{}, ErrNotFound
}

func (m *MemoryStore) ListWork(_ context.Context, user string, role domain.Role) ([]domain.WorkOrder, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]domain.WorkOrder, 0, len(m.work))
	for _, w := range m.work {
		if role != domain.RoleWorker || w.AssignedEmployee == user {
			out = append(out, w)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DueDate.Before(out[j].DueDate) })
	return out, nil
}
func (m *MemoryStore) GetWork(_ context.Context, id string) (domain.WorkOrder, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	w, ok := m.work[id]
	if !ok {
		return domain.WorkOrder{}, ErrNotFound
	}
	return w, nil
}
func (m *MemoryStore) SaveWork(_ context.Context, w domain.WorkOrder) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.work[w.ID] = w
	return nil
}
func (m *MemoryStore) ByEquipment(_ context.Context, tag string) ([]domain.WorkOrder, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := []domain.WorkOrder{}
	for _, w := range m.work {
		if w.EquipmentTag == tag {
			out = append(out, w)
		}
	}
	return out, nil
}
func (m *MemoryStore) AppendAudit(_ context.Context, e domain.AuditEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.audit = append(m.audit, e)
	return nil
}
func (m *MemoryStore) ListAudit(_ context.Context) ([]domain.AuditEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := append([]domain.AuditEntry(nil), m.audit...)
	sort.Slice(out, func(i, j int) bool { return out[i].Timestamp.After(out[j].Timestamp) })
	return out, nil
}
func (m *MemoryStore) SyncResult(_ context.Context, id string) (domain.SyncResult, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.processed[id]
	return v, ok, nil
}
func (m *MemoryStore) SaveSyncResult(_ context.Context, v domain.SyncResult, _ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processed[v.ID] = v
	return nil
}
