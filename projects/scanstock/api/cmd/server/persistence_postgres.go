//go:build postgres

package main

import (
	"context"
	"errors"
	"github.com/AndersonPaulinoDev/scanstock/api/internal/store"
	"os"
	"strings"
)

func openPersistence(ctx context.Context) (store.Store, func(), error) {
	if os.Getenv("PERSISTENCE") == "memory" {
		return store.NewMemoryStore(), func() {}, nil
	}
	u := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if u == "" {
		return nil, func() {}, errors.New("DATABASE_URL is required")
	}
	st, err := store.NewPostgresStore(ctx, u)
	if err != nil {
		return nil, func() {}, err
	}
	return st, func() { _ = st.Close() }, nil
}
