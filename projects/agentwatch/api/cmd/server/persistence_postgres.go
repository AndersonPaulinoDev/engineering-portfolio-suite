//go:build postgres

package main

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/AndersonPaulinoDev/agentwatch/api/internal/store"
)

func openPersistence(ctx context.Context) (store.Store, func(), error) {
	if os.Getenv("PERSISTENCE") == "memory" {
		return store.NewMemoryStore(), func() {}, nil
	}
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return nil, func() {}, errors.New("DATABASE_URL is required for PostgreSQL persistence")
	}
	st, err := store.NewPostgresStore(ctx, databaseURL)
	if err != nil {
		return nil, func() {}, err
	}
	return st, func() { _ = st.Close() }, nil
}
