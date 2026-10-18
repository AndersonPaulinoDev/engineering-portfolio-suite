package store

import (
	"context"
	"errors"

	"github.com/AndersonPaulinoDev/opstrack/api/internal/domain"
)

var ErrNotFound = errors.New("not found")

type Store interface {
	CreateUser(context.Context, domain.User) error
	UserByEmail(context.Context, string) (domain.User, error)
	ListWork(context.Context, string, domain.Role) ([]domain.WorkOrder, error)
	GetWork(context.Context, string) (domain.WorkOrder, error)
	SaveWork(context.Context, domain.WorkOrder) error
	ByEquipment(context.Context, string) ([]domain.WorkOrder, error)
	AppendAudit(context.Context, domain.AuditEntry) error
	ListAudit(context.Context) ([]domain.AuditEntry, error)
	SyncResult(context.Context, string) (domain.SyncResult, bool, error)
	SaveSyncResult(context.Context, domain.SyncResult, string) error
}
