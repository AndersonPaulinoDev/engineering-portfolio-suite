//go:build !postgres

package main

import (
	"context"
	"errors"
	"github.com/AndersonPaulinoDev/scanstock/api/internal/store"
	"os"
)

func openPersistence(_ context.Context) (store.Store, func(), error) {
	if os.Getenv("PERSISTENCE") != "memory" {
		return nil, func() {}, errors.New("use `go run -tags postgres ./cmd/server` for PostgreSQL or set PERSISTENCE=memory explicitly")
	}
	return store.NewMemoryStore(), func() {}, nil
}
