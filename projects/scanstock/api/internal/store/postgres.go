//go:build postgres

package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/AndersonPaulinoDev/scanstock/api/internal/domain"
	_ "github.com/lib/pq"
)

type PostgresStore struct{ db *sql.DB }

func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &PostgresStore{db: db}, nil
}
func (p *PostgresStore) Close() error { return p.db.Close() }

func (p *PostgresStore) CreateUser(ctx context.Context, u domain.User) error {
	_, err := p.db.ExecContext(ctx, `INSERT INTO users(id,email,password_hash,role) VALUES($1,$2,$3,$4) ON CONFLICT(email) DO NOTHING`, u.ID, u.Email, u.PasswordHash, u.Role)
	return err
}

func (p *PostgresStore) UserByEmail(ctx context.Context, email string) (domain.User, error) {
	var u domain.User
	err := p.db.QueryRowContext(ctx, `SELECT id,email,password_hash,role FROM users WHERE lower(email)=lower($1)`, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}
func (p *PostgresStore) ListStock(ctx context.Context) ([]domain.StockView, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT p.id,p.sku,COALESCE(p.upc,''),p.name,p.description,p.location,p.reorder_level,p.created_at,p.updated_at,COALESCE(SUM(t.quantity_change),0)::bigint FROM products p LEFT JOIN inventory_transactions t ON t.product_id=p.id GROUP BY p.id ORDER BY p.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.StockView{}
	for rows.Next() {
		var v domain.StockView
		if err := rows.Scan(&v.Product.ID, &v.Product.SKU, &v.Product.UPC, &v.Product.Name, &v.Product.Description, &v.Product.Location, &v.Product.ReorderLevel, &v.Product.CreatedAt, &v.Product.UpdatedAt, &v.Quantity); err != nil {
			return nil, err
		}
		v.NeedsReorder = v.Quantity <= v.Product.ReorderLevel
		out = append(out, v)
	}
	return out, rows.Err()
}
func (p *PostgresStore) FindByCode(ctx context.Context, code string) (domain.StockView, error) {
	var v domain.StockView
	err := p.db.QueryRowContext(ctx, `SELECT p.id,p.sku,COALESCE(p.upc,''),p.name,p.description,p.location,p.reorder_level,p.created_at,p.updated_at,COALESCE(SUM(t.quantity_change),0)::bigint FROM products p LEFT JOIN inventory_transactions t ON t.product_id=p.id WHERE p.upc=$1 OR lower(p.sku)=lower($1) GROUP BY p.id`, code).Scan(&v.Product.ID, &v.Product.SKU, &v.Product.UPC, &v.Product.Name, &v.Product.Description, &v.Product.Location, &v.Product.ReorderLevel, &v.Product.CreatedAt, &v.Product.UpdatedAt, &v.Quantity)
	if errors.Is(err, sql.ErrNoRows) {
		return v, ErrNotFound
	}
	v.NeedsReorder = v.Quantity <= v.Product.ReorderLevel
	return v, err
}
func (p *PostgresStore) CreateProduct(ctx context.Context, v domain.Product) (domain.Product, error) {
	_, err := p.db.ExecContext(ctx, `INSERT INTO products(id,sku,upc,name,description,location,reorder_level,created_at,updated_at) VALUES($1,$2,NULLIF($3,''),$4,$5,$6,$7,$8,$9)`, v.ID, v.SKU, v.UPC, v.Name, v.Description, v.Location, v.ReorderLevel, v.CreatedAt, v.UpdatedAt)
	return v, err
}
func (p *PostgresStore) RecordTransaction(ctx context.Context, v domain.InventoryTransaction) (domain.InventoryTransaction, bool, error) {
	tx, err := p.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return v, false, err
	}
	defer tx.Rollback()
	var existing domain.InventoryTransaction
	err = tx.QueryRowContext(ctx, `SELECT transaction_id,product_id,quantity_change,transaction_type,user_id,device_id,reference,timestamp,idempotency_key FROM inventory_transactions WHERE idempotency_key=$1`, v.IdempotencyKey).Scan(&existing.ID, &existing.ProductID, &existing.QuantityChange, &existing.Type, &existing.UserID, &existing.DeviceID, &existing.Reference, &existing.Timestamp, &existing.IdempotencyKey)
	if err == nil {
		return existing, true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return v, false, err
	}
	var qty int
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(quantity_change),0)::bigint FROM inventory_transactions WHERE product_id=$1`, v.ProductID).Scan(&qty)
	if err != nil {
		return v, false, err
	}
	if qty+v.QuantityChange < 0 {
		return v, false, errors.New("transaction would make stock negative")
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO inventory_transactions(transaction_id,product_id,quantity_change,transaction_type,user_id,device_id,reference,timestamp,idempotency_key) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, v.ID, v.ProductID, v.QuantityChange, v.Type, v.UserID, v.DeviceID, v.Reference, v.Timestamp, v.IdempotencyKey)
	if err != nil {
		return v, false, err
	}
	if err := tx.Commit(); err != nil {
		return v, false, err
	}
	return v, false, nil
}
func (p *PostgresStore) ListTransactions(ctx context.Context, product string) ([]domain.InventoryTransaction, error) {
	q := `SELECT transaction_id,product_id,quantity_change,transaction_type,user_id,device_id,reference,timestamp,idempotency_key FROM inventory_transactions`
	args := []any{}
	if product != "" {
		q += ` WHERE product_id=$1`
		args = append(args, product)
	}
	q += ` ORDER BY timestamp DESC`
	rows, err := p.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.InventoryTransaction{}
	for rows.Next() {
		var v domain.InventoryTransaction
		if err := rows.Scan(&v.ID, &v.ProductID, &v.QuantityChange, &v.Type, &v.UserID, &v.DeviceID, &v.Reference, &v.Timestamp, &v.IdempotencyKey); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
func (p *PostgresStore) Quantity(ctx context.Context, product string) (int, error) {
	var q int
	err := p.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(quantity_change),0)::bigint FROM inventory_transactions WHERE product_id=$1`, product).Scan(&q)
	return q, err
}
