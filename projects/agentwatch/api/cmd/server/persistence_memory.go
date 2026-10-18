//go:build !postgres

package main

import (
	"context"
	"errors"
	"os"

	"github.com/AndersonPaulinoDev/agentwatch/api/internal/store"
)

func openPersistence(_ context.Context) (store.Store, func(), error) {
	if os.Getenv("PERSISTENCE") != "memory" {
		return nil, func() {}, errors.New("this build has no PostgreSQL driver; use `go run -tags postgres ./cmd/server` or set PERSISTENCE=memory explicitly for tests/demo")
	}
	return store.NewMemoryStore(), func() {}, nil
}
