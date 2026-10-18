package store

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"

	"github.com/AndersonPaulinoDev/scanstock/api/internal/domain"
)

type MemoryStore struct {
	mu       sync.RWMutex
	users    map[string]domain.User
	products map[string]domain.Product
	txs      []domain.InventoryTransaction
	byKey    map[string]domain.InventoryTransaction
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{users: map[string]domain.User{}, products: map[string]domain.Product{}, txs: []domain.InventoryTransaction{}, byKey: map[string]domain.InventoryTransaction{}}
}

func (m *MemoryStore) CreateUser(_ context.Context, u domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, existing := range m.users {
		if strings.EqualFold(existing.Email, u.Email) {
			return errors.New("user already exists")
		}
	}
	m.users[u.ID] = u
	return nil
}
func (m *MemoryStore) UserByEmail(_ context.Context, email string) (domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, u := range m.users {
		if strings.EqualFold(u.Email, email) {
			return u, nil
		}
	}
	return domain.User{}, ErrNotFound
}

func (m *MemoryStore) ListStock(_ context.Context) ([]domain.StockView, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]domain.StockView, 0, len(m.products))
	for _, p := range m.products {
		q := m.quantityLocked(p.ID)
		out = append(out, domain.StockView{Product: p, Quantity: q, NeedsReorder: q <= p.ReorderLevel})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Product.Name < out[j].Product.Name })
	return out, nil
}
func (m *MemoryStore) FindByCode(_ context.Context, code string) (domain.StockView, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	code = strings.TrimSpace(code)
	for _, p := range m.products {
		if p.UPC == code || strings.EqualFold(p.SKU, code) {
			q := m.quantityLocked(p.ID)
			return domain.StockView{Product: p, Quantity: q, NeedsReorder: q <= p.ReorderLevel}, nil
		}
	}
	return domain.StockView{}, ErrNotFound
}
func (m *MemoryStore) CreateProduct(_ context.Context, p domain.Product) (domain.Product, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, x := range m.products {
		if strings.EqualFold(x.SKU, p.SKU) || (p.UPC != "" && x.UPC == p.UPC) {
			return p, errors.New("sku or upc already exists")
		}
	}
	m.products[p.ID] = p
	return p, nil
}
func (m *MemoryStore) RecordTransaction(_ context.Context, tx domain.InventoryTransaction) (domain.InventoryTransaction, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if old, ok := m.byKey[tx.IdempotencyKey]; ok {
		return old, true, nil
	}
	if _, ok := m.products[tx.ProductID]; !ok {
		return tx, false, ErrNotFound
	}
	if m.quantityLocked(tx.ProductID)+tx.QuantityChange < 0 {
		return tx, false, errors.New("transaction would make stock negative")
	}
	m.txs = append(m.txs, tx)
	m.byKey[tx.IdempotencyKey] = tx
	return tx, false, nil
}
func (m *MemoryStore) ListTransactions(_ context.Context, product string) ([]domain.InventoryTransaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := []domain.InventoryTransaction{}
	for i := len(m.txs) - 1; i >= 0; i-- {
		if product == "" || m.txs[i].ProductID == product {
			out = append(out, m.txs[i])
		}
	}
	return out, nil
}
func (m *MemoryStore) Quantity(_ context.Context, product string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, ok := m.products[product]; !ok {
		return 0, ErrNotFound
	}
	return m.quantityLocked(product), nil
}
func (m *MemoryStore) quantityLocked(product string) int {
	q := 0
	for _, tx := range m.txs {
		if tx.ProductID == product {
			q += tx.QuantityChange
		}
	}
	return q
}
