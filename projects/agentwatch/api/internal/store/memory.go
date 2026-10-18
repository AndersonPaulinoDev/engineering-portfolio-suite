package store

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/AndersonPaulinoDev/agentwatch/api/internal/domain"
)

type MemoryStore struct {
	mu        sync.RWMutex
	users     map[string]domain.User
	projects  map[string]domain.Project
	agents    map[string]domain.Agent
	tasks     map[string]domain.Task
	events    []domain.Event
	approvals map[string]domain.Approval
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users: map[string]domain.User{}, projects: map[string]domain.Project{},
		agents: map[string]domain.Agent{}, tasks: map[string]domain.Task{},
		approvals: map[string]domain.Approval{}, events: []domain.Event{},
	}
}

func (m *MemoryStore) CreateUser(_ context.Context, v domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.users {
		if strings.EqualFold(u.Email, v.Email) {
			return ErrConflict
		}
	}
	m.users[v.ID] = v
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
func (m *MemoryStore) UserByID(_ context.Context, id string) (domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.users[id]
	if !ok {
		return domain.User{}, ErrNotFound
	}
	return v, nil
}
func (m *MemoryStore) CreateProject(_ context.Context, v domain.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.projects[v.ID] = v
	return nil
}
func (m *MemoryStore) ListProjects(_ context.Context) ([]domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]domain.Project, 0, len(m.projects))
	for _, v := range m.projects {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}
func (m *MemoryStore) UpsertAgent(_ context.Context, v domain.Agent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.agents[v.ID] = v
	return nil
}
func (m *MemoryStore) ListAgents(_ context.Context) ([]domain.Agent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]domain.Agent, 0, len(m.agents))
	for _, v := range m.agents {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
func (m *MemoryStore) AgentByID(_ context.Context, id string) (domain.Agent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.agents[id]
	if !ok {
		return domain.Agent{}, ErrNotFound
	}
	return v, nil
}
func (m *MemoryStore) CreateTask(_ context.Context, v domain.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks[v.ID] = v
	return nil
}
func (m *MemoryStore) UpdateTask(_ context.Context, v domain.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.tasks[v.ID]; !ok {
		return ErrNotFound
	}
	m.tasks[v.ID] = v
	return nil
}
func (m *MemoryStore) TaskByID(_ context.Context, id string) (domain.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.tasks[id]
	if !ok {
		return domain.Task{}, ErrNotFound
	}
	return v, nil
}
func (m *MemoryStore) ListTasks(_ context.Context) ([]domain.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]domain.Task, 0, len(m.tasks))
	for _, v := range m.tasks {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}
func (m *MemoryStore) AppendEvent(_ context.Context, v domain.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, v)
	return nil
}
func (m *MemoryStore) ListEvents(_ context.Context, limit int) ([]domain.Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if limit <= 0 || limit > len(m.events) {
		limit = len(m.events)
	}
	start := len(m.events) - limit
	out := append([]domain.Event(nil), m.events[start:]...)
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}
func (m *MemoryStore) CreateApproval(_ context.Context, v domain.Approval) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.approvals[v.ID] = v
	return nil
}
func (m *MemoryStore) ApprovalByID(_ context.Context, id string) (domain.Approval, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.approvals[id]
	if !ok {
		return domain.Approval{}, ErrNotFound
	}
	return v, nil
}
func (m *MemoryStore) UpdateApproval(_ context.Context, v domain.Approval) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.approvals[v.ID]; !ok {
		return ErrNotFound
	}
	m.approvals[v.ID] = v
	return nil
}
func (m *MemoryStore) ListApprovals(_ context.Context) ([]domain.Approval, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]domain.Approval, 0, len(m.approvals))
	for _, v := range m.approvals {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}
