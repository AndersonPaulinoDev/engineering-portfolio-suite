//go:build postgres

package store

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/AndersonPaulinoDev/opstrack/api/internal/domain"
)

func TestPostgresWorkOrderAndAuditRoundTrip(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	st, err := NewPostgresStore(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	id := time.Now().Format("150405.000000000")
	user := domain.User{ID: "pg_worker_" + id, Email: "pg-worker-" + id + "@example.com", PasswordHash: "hash", Role: domain.RoleWorker}
	if err := st.CreateUser(ctx, user); err != nil {
		t.Fatal(err)
	}
	work := domain.WorkOrder{ID: "pg_wo_" + id, Title: "Database test", Priority: "high", AssignedEmployee: user.ID, DueDate: time.Now().Add(time.Hour), Status: domain.StatusAssigned, UpdatedAt: time.Now().UTC()}
	if err := st.SaveWork(ctx, work); err != nil {
		t.Fatal(err)
	}
	got, err := st.GetWork(ctx, work.ID)
	if err != nil || got.Title != work.Title {
		t.Fatalf("work round trip failed: %#v %v", got, err)
	}
	audit := domain.AuditEntry{ID: "pg_aud_" + id, Actor: user.ID, Action: "test", WorkOrderID: work.ID, Timestamp: time.Now().UTC()}
	if err := st.AppendAudit(ctx, audit); err != nil {
		t.Fatal(err)
	}
	entries, err := st.ListAudit(ctx)
	if err != nil || len(entries) == 0 {
		t.Fatalf("audit round trip failed: %v", err)
	}
}
