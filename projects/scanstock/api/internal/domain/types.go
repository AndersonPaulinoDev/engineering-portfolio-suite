package domain

import "time"

type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"`
}
type Product struct {
	ID           string    `json:"id"`
	SKU          string    `json:"sku"`
	UPC          string    `json:"upc"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Location     string    `json:"location"`
	ReorderLevel int       `json:"reorder_level"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
type TransactionType string

const (
	TxReceived    TransactionType = "received"
	TxShipped     TransactionType = "shipped"
	TxDamaged     TransactionType = "damaged"
	TxCorrection  TransactionType = "correction"
	TxCount       TransactionType = "count_adjustment"
	TxTransferIn  TransactionType = "transfer_in"
	TxTransferOut TransactionType = "transfer_out"
)

type InventoryTransaction struct {
	ID             string          `json:"transaction_id"`
	ProductID      string          `json:"product_id"`
	QuantityChange int             `json:"quantity_change"`
	Type           TransactionType `json:"transaction_type"`
	UserID         string          `json:"user_id"`
	DeviceID       string          `json:"device_id"`
	Reference      string          `json:"reference"`
	Timestamp      time.Time       `json:"timestamp"`
	IdempotencyKey string          `json:"idempotency_key"`
}
type StockView struct {
	Product      Product `json:"product"`
	Quantity     int     `json:"quantity"`
	NeedsReorder bool    `json:"needs_reorder"`
}
type SyncOperation struct {
	ID             string          `json:"id"`
	ProductID      string          `json:"product_id"`
	QuantityChange int             `json:"quantity_change"`
	Type           TransactionType `json:"transaction_type"`
	DeviceID       string          `json:"device_id"`
	Reference      string          `json:"reference"`
}
type SyncResult struct {
	ID          string                `json:"id"`
	State       string                `json:"state"`
	Transaction *InventoryTransaction `json:"transaction,omitempty"`
	Error       string                `json:"error,omitempty"`
}
