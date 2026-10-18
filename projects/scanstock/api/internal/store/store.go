package store

import (
	"context"
	"errors"

	"github.com/AndersonPaulinoDev/scanstock/api/internal/domain"
)

var ErrNotFound = errors.New("not found")

type Store interface {
	CreateUser(context.Context, domain.User) error
	UserByEmail(context.Context, string) (domain.User, error)
	ListStock(context.Context) ([]domain.StockView, error)
	FindByCode(context.Context, string) (domain.StockView, error)
	CreateProduct(context.Context, domain.Product) (domain.Product, error)
	RecordTransaction(context.Context, domain.InventoryTransaction) (domain.InventoryTransaction, bool, error)
	ListTransactions(context.Context, string) ([]domain.InventoryTransaction, error)
	Quantity(context.Context, string) (int, error)
}
