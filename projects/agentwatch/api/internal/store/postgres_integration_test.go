//go:build postgres

package store

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/AndersonPaulinoDev/agentwatch/api/internal/domain"
)

func TestPostgresStoreRoundTrip(t *testing.T) {
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

	user := domain.User{ID: "pg_user_" + time.Now().Format("150405.000000000"), Email: "pg-test-" + time.Now().Format("150405.000000000") + "@example.com", PasswordHash: "hash", Role: "developer", CreatedAt: time.Now().UTC()}
	if err := st.CreateUser(ctx, user); err != nil {
		t.Fatal(err)
	}
	got, err := st.UserByEmail(ctx, user.Email)
	if err != nil || got.ID != user.ID {
		t.Fatalf("user round trip failed: %#v %v", got, err)
	}

	project := domain.Project{ID: "pg_project_" + time.Now().Format("150405.000000000"), Name: "Postgres Test", CreatedAt: time.Now().UTC()}
	if err := st.CreateProject(ctx, project); err != nil {
		t.Fatal(err)
	}
	projects, err := st.ListProjects(ctx)
	if err != nil || len(projects) == 0 {
		t.Fatalf("project round trip failed: %v", err)
	}
}
