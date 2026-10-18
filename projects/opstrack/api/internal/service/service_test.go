package service

import (
	"context"
	"testing"

	"github.com/AndersonPaulinoDev/opstrack/api/internal/domain"
	"github.com/AndersonPaulinoDev/opstrack/api/internal/store"
)

func testService(t *testing.T) *Service {
	t.Helper()
	svc := New(store.NewMemoryStore())
	if err := svc.SeedDemo(context.Background()); err != nil {
		t.Fatal(err)
	}
	return svc
}

func TestRejectInvalidTransition(t *testing.T) {
	svc := testService(t)
	if _, err := svc.Transition(context.Background(), "usr_worker", domain.RoleWorker, "wo_1001", domain.StatusCompleted); err == nil {
		t.Fatal("assigned -> completed must be rejected")
	}
}

func TestWorkerFlowCreatesAudit(t *testing.T) {
	ctx := context.Background()
	svc := testService(t)
	if _, err := svc.Transition(ctx, "usr_worker", domain.RoleWorker, "wo_1001", domain.StatusAccepted); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Transition(ctx, "usr_worker", domain.RoleWorker, "wo_1001", domain.StatusInProgress); err != nil {
		t.Fatal(err)
	}
	audit, err := svc.Audit(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(audit) != 2 {
		t.Fatalf("want 2 audit events, got %d", len(audit))
	}
}

func TestOfflineSyncIdempotent(t *testing.T) {
	ctx := context.Background()
	svc := testService(t)
	op := domain.SyncOperation{ID: "op1", Type: "transition", WorkOrderID: "wo_1001", Status: domain.StatusAccepted}
	first := svc.Sync(ctx, "usr_worker", domain.RoleWorker, []domain.SyncOperation{op})
	second := svc.Sync(ctx, "usr_worker", domain.RoleWorker, []domain.SyncOperation{op})
	if first[0] != second[0] {
		t.Fatal("duplicate sync operation should return same result")
	}
	audit, _ := svc.Audit(ctx)
	if len(audit) != 1 {
		t.Fatalf("expected one audit event, got %d", len(audit))
	}
}

func TestNoteAndEvidenceArePersisted(t *testing.T) {
	ctx := context.Background()
	svc := testService(t)
	if _, err := svc.AddNote(ctx, "usr_worker", "wo_1001", "Filter is dirty"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddEvidence(ctx, "usr_worker", "wo_1001", "file://photo.jpg"); err != nil {
		t.Fatal(err)
	}
	work, err := svc.Get(ctx, "wo_1001")
	if err != nil {
		t.Fatal(err)
	}
	if len(work.Notes) != 1 || len(work.CompletionEvidence) != 1 {
		t.Fatalf("expected note and evidence to persist: %#v", work)
	}
}
