//go:build postgres

package store

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/AndersonPaulinoDev/scanstock/api/internal/domain"
)

func TestPostgresLedgerAndIdempotency(t *testing.T) {
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
	user := domain.User{ID: "pg_stock_" + id, Email: "pg-stock-" + id + "@example.com", PasswordHash: "hash", Role: "operator"}
	if err := st.CreateUser(ctx, user); err != nil {
		t.Fatal(err)
	}
	product := domain.Product{ID: "pg_product_" + id, SKU: "PG-" + id, Name: "Postgres product", ReorderLevel: 1, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	if _, err := st.CreateProduct(ctx, product); err != nil {
		t.Fatal(err)
	}
	tx := domain.InventoryTransaction{ID: "pg_tx_" + id, ProductID: product.ID, QuantityChange: 5, Type: domain.TxReceived, UserID: user.ID, DeviceID: "ci", Timestamp: time.Now().UTC(), IdempotencyKey: "pg_key_" + id}
	first, replay, err := st.RecordTransaction(ctx, tx)
	if err != nil || replay {
		t.Fatalf("first transaction failed: replay=%v err=%v", replay, err)
	}
	second, replay, err := st.RecordTransaction(ctx, tx)
	if err != nil || !replay || first.ID != second.ID {
		t.Fatalf("idempotency failed: replay=%v err=%v", replay, err)
	}
	quantity, err := st.Quantity(ctx, product.ID)
	if err != nil || quantity != 5 {
		t.Fatalf("want quantity 5, got %d (%v)", quantity, err)
	}
}
