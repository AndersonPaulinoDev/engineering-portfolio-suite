package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/AndersonPaulinoDev/agentwatch/api/internal/domain"
	"github.com/AndersonPaulinoDev/agentwatch/api/internal/realtime"
	"github.com/AndersonPaulinoDev/agentwatch/api/internal/store"
)

type AgentWatchService struct {
	st  store.Store
	hub *realtime.Hub
}

func NewAgentWatchService(st store.Store, hub *realtime.Hub) *AgentWatchService {
	return &AgentWatchService{st: st, hub: hub}
}

func (s *AgentWatchService) CreateProject(ctx context.Context, name string) (domain.Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.Project{}, errors.New("name required")
	}
	v := domain.Project{ID: newID("prj"), Name: name, CreatedAt: time.Now().UTC()}
	err := s.st.CreateProject(ctx, v)
	if err == nil {
		s.emit(ctx, domain.Event{ID: newID("evt"), Type: "project.created", ProjectID: v.ID, CreatedAt: time.Now().UTC(), Payload: map[string]any{"name": v.Name}})
	}
	return v, err
}
func (s *AgentWatchService) Projects(ctx context.Context) ([]domain.Project, error) {
	return s.st.ListProjects(ctx)
}
func (s *AgentWatchService) Agents(ctx context.Context) ([]domain.Agent, error) {
	return s.st.ListAgents(ctx)
}
func (s *AgentWatchService) Heartbeat(ctx context.Context, id, name, projectID, kind string, state domain.AgentState, message string) (domain.Agent, error) {
	if id == "" {
		id = newID("agt")
	}
	old, err := s.st.AgentByID(ctx, id)
	now := time.Now().UTC()
	a := domain.Agent{ID: id, Name: name, ProjectID: projectID, Type: kind, State: state, StartedAt: now, LastHeartbeat: now, LatestMessage: message}
	if err == nil {
		a.StartedAt = old.StartedAt
		if a.Name == "" {
			a.Name = old.Name
		}
		if a.ProjectID == "" {
			a.ProjectID = old.ProjectID
		}
		if a.Type == "" {
			a.Type = old.Type
		}
	}
	if a.Name == "" {
		return domain.Agent{}, errors.New("agent name required")
	}
	if a.State == "" {
		a.State = domain.AgentIdle
	}
	if err := s.st.UpsertAgent(ctx, a); err != nil {
		return domain.Agent{}, err
	}
	s.emit(ctx, domain.Event{ID: newID("evt"), Type: "agent.heartbeat", ProjectID: a.ProjectID, AgentID: a.ID, CreatedAt: now, Payload: map[string]any{"state": a.State, "message": message}})
	return a, nil
}
func (s *AgentWatchService) CreateTask(ctx context.Context, projectID, agentID, title, description string) (domain.Task, error) {
	if strings.TrimSpace(title) == "" {
		return domain.Task{}, errors.New("title required")
	}
	now := time.Now().UTC()
	t := domain.Task{ID: newID("tsk"), ProjectID: projectID, AgentID: agentID, Title: strings.TrimSpace(title), Description: description, Status: domain.TaskQueued, CreatedAt: now, UpdatedAt: now}
	if err := s.st.CreateTask(ctx, t); err != nil {
		return domain.Task{}, err
	}
	s.emit(ctx, domain.Event{ID: newID("evt"), Type: "task.created", ProjectID: projectID, AgentID: agentID, TaskID: t.ID, CreatedAt: now, Payload: map[string]any{"title": t.Title}})
	return t, nil
}
func (s *AgentWatchService) Tasks(ctx context.Context) ([]domain.Task, error) {
	return s.st.ListTasks(ctx)
}
func (s *AgentWatchService) UpdateTaskStatus(ctx context.Context, id string, status domain.TaskStatus) (domain.Task, error) {
	t, err := s.st.TaskByID(ctx, id)
	if err != nil {
		return domain.Task{}, err
	}
	if !validTaskTransition(t.Status, status) {
		return domain.Task{}, errors.New("invalid task transition")
	}
	t.Status = status
	t.UpdatedAt = time.Now().UTC()
	if err := s.st.UpdateTask(ctx, t); err != nil {
		return domain.Task{}, err
	}
	s.emit(ctx, domain.Event{ID: newID("evt"), Type: "task." + string(status), ProjectID: t.ProjectID, AgentID: t.AgentID, TaskID: t.ID, CreatedAt: t.UpdatedAt})
	return t, nil
}
func validTaskTransition(from, to domain.TaskStatus) bool {
	allowed := map[domain.TaskStatus]map[domain.TaskStatus]bool{domain.TaskQueued: {domain.TaskRunning: true, domain.TaskCancelled: true}, domain.TaskRunning: {domain.TaskWaiting: true, domain.TaskFailed: true, domain.TaskCompleted: true, domain.TaskCancelled: true}, domain.TaskWaiting: {domain.TaskRunning: true, domain.TaskFailed: true, domain.TaskCompleted: true, domain.TaskCancelled: true}, domain.TaskFailed: {domain.TaskQueued: true, domain.TaskCancelled: true}}
	return allowed[from][to]
}
func (s *AgentWatchService) RequestApproval(ctx context.Context, projectID, agentID, taskID, action, reason string) (domain.Approval, error) {
	if action == "" {
		return domain.Approval{}, errors.New("action required")
	}
	a := domain.Approval{ID: newID("apr"), ProjectID: projectID, AgentID: agentID, TaskID: taskID, Action: action, Reason: reason, Status: "pending", CreatedAt: time.Now().UTC()}
	if err := s.st.CreateApproval(ctx, a); err != nil {
		return domain.Approval{}, err
	}
	s.emit(ctx, domain.Event{ID: newID("evt"), Type: "approval.requested", ProjectID: projectID, AgentID: agentID, TaskID: taskID, CreatedAt: a.CreatedAt, Payload: map[string]any{"approval_id": a.ID, "action": action, "reason": reason}})
	return a, nil
}
func (s *AgentWatchService) DecideApproval(ctx context.Context, id, decision string) (domain.Approval, error) {
	if decision != "approved" && decision != "rejected" {
		return domain.Approval{}, errors.New("decision must be approved or rejected")
	}
	a, err := s.st.ApprovalByID(ctx, id)
	if err != nil {
		return domain.Approval{}, err
	}
	if a.Status != "pending" {
		return domain.Approval{}, errors.New("approval already decided")
	}
	now := time.Now().UTC()
	a.Status = decision
	a.DecidedAt = &now
	if err := s.st.UpdateApproval(ctx, a); err != nil {
		return domain.Approval{}, err
	}
	s.emit(ctx, domain.Event{ID: newID("evt"), Type: "approval." + decision, ProjectID: a.ProjectID, AgentID: a.AgentID, TaskID: a.TaskID, CreatedAt: now, Payload: map[string]any{"approval_id": a.ID}})
	return a, nil
}
func (s *AgentWatchService) Approvals(ctx context.Context) ([]domain.Approval, error) {
	return s.st.ListApprovals(ctx)
}
func (s *AgentWatchService) Events(ctx context.Context, limit int) ([]domain.Event, error) {
	return s.st.ListEvents(ctx, limit)
}
func (s *AgentWatchService) emit(ctx context.Context, e domain.Event) {
	_ = s.st.AppendEvent(ctx, e)
	s.hub.Broadcast(e)
}

func SeedDemo(auth *AuthService, svc *AgentWatchService) error {
	ctx := context.Background()
	_, _, err := auth.Register(ctx, "demo@agentwatch.dev", "demo1234", "developer")
	if err != nil && err != store.ErrConflict {
		return err
	}
	p1, _ := svc.CreateProject(ctx, "Ders Compiler")
	p2, _ := svc.CreateProject(ctx, "Anderson AI OS")
	p3, _ := svc.CreateProject(ctx, "NeusicWave")
	agents := []struct {
		id, name, project, kind string
		state                   domain.AgentState
	}{
		{"agt_codex", "Codex Builder", p1.ID, "coding", domain.AgentRunning}, {"agt_review", "Review Agent", p1.ID, "review", domain.AgentIdle}, {"agt_orchestrator", "Orchestrator", p2.ID, "orchestrator", domain.AgentWaiting}, {"agt_tests", "Test Runner", p2.ID, "testing", domain.AgentRunning}, {"agt_audio", "Audio Worker", p3.ID, "worker", domain.AgentCompleted},
	}
	for _, x := range agents {
		_, _ = svc.Heartbeat(ctx, x.id, x.name, x.project, x.kind, x.state, "demo heartbeat")
	}
	for i, title := range []string{"Implement parser diagnostics", "Refactor workspace rail", "Run Go tests", "Review permissions", "Package Linux build", "Audit event stream", "Fix sync conflict", "Write ADR", "Profile worker pool", "Create mobile inbox", "Validate CI", "Document API"} {
		project := p1.ID
		if i%3 == 1 {
			project = p2.ID
		}
		if i%3 == 2 {
			project = p3.ID
		}
		_, _ = svc.CreateTask(ctx, project, "", title, "Seeded portfolio demo task")
	}
	_, _ = svc.RequestApproval(ctx, p1.ID, "agt_codex", "", "Delete generated build output", "14 generated files will be replaced")
	return nil
}
