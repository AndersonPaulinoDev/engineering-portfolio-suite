package service

import (
	"context"
	"testing"

	"github.com/AndersonPaulinoDev/scanstock/api/internal/domain"
	"github.com/AndersonPaulinoDev/scanstock/api/internal/store"
)

func testService(t *testing.T) *Service {
	t.Helper()
	svc := New(store.NewMemoryStore())
	if err := svc.SeedDemo(context.Background()); err != nil {
		t.Fatal(err)
	}
	return svc
}

func TestLedgerQuantity(t *testing.T) {
	svc := testService(t)
	rows, err := svc.Products(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var quantity int
	for _, row := range rows {
		if row.Product.ID == "prd_1001" {
			quantity = row.Quantity
		}
	}
	if quantity != 35 {
		t.Fatalf("want 35, got %d", quantity)
	}
}

func TestIdempotency(t *testing.T) {
	ctx := context.Background()
	svc := testService(t)
	first, err := svc.Record(ctx, "u", "prd_1001", 5, domain.TxReceived, "d", "r", "same-key")
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.Record(ctx, "u", "prd_1001", 5, domain.TxReceived, "d", "r", "same-key")
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID {
		t.Fatal("duplicate key must return original transaction")
	}
	stock, _ := svc.FindByCode(ctx, "HVAC-FLT-20")
	if stock.Quantity != 40 {
		t.Fatalf("expected one +5 application, got %d", stock.Quantity)
	}
}

func TestPreventNegativeInventory(t *testing.T) {
	svc := testService(t)
	if _, err := svc.Record(context.Background(), "u", "prd_1003", 100, domain.TxShipped, "d", "r", "ship-too-many"); err == nil {
		t.Fatal("expected negative stock rejection")
	}
}

func TestSyncDuplicateOperationIsSafe(t *testing.T) {
	ctx := context.Background()
	svc := testService(t)
	op := domain.SyncOperation{ID: "offline-1", ProductID: "prd_1002", QuantityChange: 2, Type: domain.TxReceived, DeviceID: "phone"}
	first := svc.Sync(ctx, "u", []domain.SyncOperation{op})
	second := svc.Sync(ctx, "u", []domain.SyncOperation{op})
	if first[0].Transaction == nil || second[0].Transaction == nil || first[0].Transaction.ID != second[0].Transaction.ID {
		t.Fatal("sync retry should return same transaction")
	}
}

func TestFindByUPC(t *testing.T) {
	svc := testService(t)
	stock, err := svc.FindByCode(context.Background(), "840000000001")
	if err != nil {
		t.Fatal(err)
	}
	if stock.Product.SKU != "HVAC-FLT-20" {
		t.Fatalf("unexpected product: %#v", stock.Product)
	}
}
