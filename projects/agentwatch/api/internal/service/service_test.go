package service

import (
	"context"
	"testing"
	"time"

	"github.com/AndersonPaulinoDev/agentwatch/api/internal/domain"
	"github.com/AndersonPaulinoDev/agentwatch/api/internal/realtime"
	"github.com/AndersonPaulinoDev/agentwatch/api/internal/store"
)

func testServices() (context.Context, *AuthService, *AgentWatchService) {
	ctx := context.Background()
	st := store.NewMemoryStore()
	h := realtime.NewHub()
	go h.Run()
	return ctx, NewAuthService(st, "test-secret", time.Minute, time.Hour), NewAgentWatchService(st, h)
}
func TestRegisterLoginAndRefresh(t *testing.T) {
	ctx, a, _ := testServices()
	u, pair, err := a.Register(ctx, "test@example.com", "password123", "developer")
	if err != nil || u.ID == "" || pair.AccessToken == "" {
		t.Fatalf("register failed: %v", err)
	}
	_, pair2, err := a.Login(ctx, "test@example.com", "password123")
	if err != nil || pair2.RefreshToken == "" {
		t.Fatalf("login failed: %v", err)
	}
	if _, err := a.Refresh(ctx, pair2.RefreshToken); err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
}
func TestTaskTransitionRules(t *testing.T) {
	ctx, _, svc := testServices()
	p, err := svc.CreateProject(ctx, "P")
	if err != nil {
		t.Fatal(err)
	}
	task, err := svc.CreateTask(ctx, p.ID, "", "Build", "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpdateTaskStatus(ctx, task.ID, domain.TaskCompleted); err == nil {
		t.Fatal("queued -> completed should be rejected")
	}
	task, err = svc.UpdateTaskStatus(ctx, task.ID, domain.TaskRunning)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpdateTaskStatus(ctx, task.ID, domain.TaskCompleted); err != nil {
		t.Fatalf("running -> completed should pass: %v", err)
	}
}
func TestApprovalCanOnlyBeDecidedOnce(t *testing.T) {
	ctx, _, svc := testServices()
	a, err := svc.RequestApproval(ctx, "p", "a", "", "delete files", "generated output")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DecideApproval(ctx, a.ID, "approved"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DecideApproval(ctx, a.ID, "rejected"); err == nil {
		t.Fatal("expected second decision to fail")
	}
}
