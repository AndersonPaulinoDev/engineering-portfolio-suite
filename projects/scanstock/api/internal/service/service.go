package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/AndersonPaulinoDev/scanstock/api/internal/domain"
	"github.com/AndersonPaulinoDev/scanstock/api/internal/store"
)

type Service struct{ store store.Store }

func New(st store.Store) *Service { return &Service{store: st} }

func (s *Service) SeedDemo(ctx context.Context) error {
	now := time.Now().UTC()
	products := []domain.Product{
		{ID: "prd_1001", SKU: "HVAC-FLT-20", UPC: "840000000001", Name: "20x20 HVAC Filter", Description: "MERV 8 pleated filter", Location: "A-01-02", ReorderLevel: 12, CreatedAt: now, UpdatedAt: now},
		{ID: "prd_1002", SKU: "LED-T8-4FT", UPC: "840000000002", Name: "4ft LED Tube", Description: "18W neutral white", Location: "B-03-01", ReorderLevel: 20, CreatedAt: now, UpdatedAt: now},
		{ID: "prd_1003", SKU: "GLOVE-L-NIT", UPC: "840000000003", Name: "Nitrile Gloves L", Description: "Box of 100", Location: "C-02-04", ReorderLevel: 10, CreatedAt: now, UpdatedAt: now},
	}
	for _, p := range products {
		if _, err := s.store.CreateProduct(ctx, p); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return err
		}
	}
	seeds := []domain.InventoryTransaction{
		{ID: "tx_seed_1", ProductID: "prd_1001", QuantityChange: 40, Type: domain.TxReceived, UserID: "usr_stock", DeviceID: "seed", Reference: "initial stock", Timestamp: now, IdempotencyKey: "seed_1"},
		{ID: "tx_seed_2", ProductID: "prd_1001", QuantityChange: -5, Type: domain.TxShipped, UserID: "usr_stock", DeviceID: "seed", Reference: "demo shipment", Timestamp: now, IdempotencyKey: "seed_2"},
		{ID: "tx_seed_3", ProductID: "prd_1002", QuantityChange: 18, Type: domain.TxReceived, UserID: "usr_stock", DeviceID: "seed", Reference: "initial stock", Timestamp: now, IdempotencyKey: "seed_3"},
		{ID: "tx_seed_4", ProductID: "prd_1003", QuantityChange: 8, Type: domain.TxReceived, UserID: "usr_stock", DeviceID: "seed", Reference: "initial stock", Timestamp: now, IdempotencyKey: "seed_4"},
	}
	for _, tx := range seeds {
		_, _, _ = s.store.RecordTransaction(ctx, tx)
	}
	return nil
}
func (s *Service) Products(ctx context.Context) ([]domain.StockView, error) {
	return s.store.ListStock(ctx)
}
func (s *Service) FindByCode(ctx context.Context, code string) (domain.StockView, error) {
	v, err := s.store.FindByCode(ctx, strings.TrimSpace(code))
	if errors.Is(err, store.ErrNotFound) {
		return v, errors.New("product not found")
	}
	return v, err
}
func (s *Service) CreateProduct(ctx context.Context, p domain.Product) (domain.Product, error) {
	if strings.TrimSpace(p.SKU) == "" || strings.TrimSpace(p.Name) == "" {
		return p, errors.New("sku and name required")
	}
	p.ID = newID("prd")
	p.CreatedAt = time.Now().UTC()
	p.UpdatedAt = p.CreatedAt
	return s.store.CreateProduct(ctx, p)
}
func (s *Service) Record(ctx context.Context, user, product string, qty int, typ domain.TransactionType, device, ref, key string) (domain.InventoryTransaction, error) {
	if strings.TrimSpace(key) == "" {
		return domain.InventoryTransaction{}, errors.New("idempotency key required")
	}
	if qty == 0 {
		return domain.InventoryTransaction{}, errors.New("quantity change cannot be zero")
	}
	if typ == domain.TxShipped || typ == domain.TxDamaged || typ == domain.TxTransferOut {
		if qty > 0 {
			qty = -qty
		}
	}
	if typ == domain.TxReceived || typ == domain.TxTransferIn {
		if qty < 0 {
			qty = -qty
		}
	}
	tx := domain.InventoryTransaction{ID: newID("tx"), ProductID: product, QuantityChange: qty, Type: typ, UserID: user, DeviceID: device, Reference: ref, Timestamp: time.Now().UTC(), IdempotencyKey: key}
	saved, _, err := s.store.RecordTransaction(ctx, tx)
	if errors.Is(err, store.ErrNotFound) {
		return domain.InventoryTransaction{}, errors.New("product not found")
	}
	return saved, err
}
func (s *Service) Transactions(ctx context.Context, product string) ([]domain.InventoryTransaction, error) {
	return s.store.ListTransactions(ctx, product)
}
func (s *Service) Sync(ctx context.Context, user string, ops []domain.SyncOperation) []domain.SyncResult {
	out := make([]domain.SyncResult, 0, len(ops))
	for _, op := range ops {
		tx, err := s.Record(ctx, user, op.ProductID, op.QuantityChange, op.Type, op.DeviceID, op.Reference, op.ID)
		if err != nil {
			out = append(out, domain.SyncResult{ID: op.ID, State: "conflict", Error: err.Error()})
			continue
		}
		copy := tx
		out = append(out, domain.SyncResult{ID: op.ID, State: "synced", Transaction: &copy})
	}
	return out
}
func newID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return prefix + "_" + hex.EncodeToString(b)
}
