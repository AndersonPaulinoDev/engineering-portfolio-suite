//go:build !postgres

package main

import (
	"context"
	"errors"
	"os"

	"github.com/AndersonPaulinoDev/opstrack/api/internal/store"
)

func openPersistence(_ context.Context) (store.Store, func(), error) {
	if os.Getenv("PERSISTENCE") != "memory" {
		return nil, func() {}, errors.New("use `go run -tags postgres ./cmd/server` for PostgreSQL or set PERSISTENCE=memory explicitly")
	}
	return store.NewMemoryStore(), func() {}, nil
}
