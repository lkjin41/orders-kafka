package orders

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	StatusNew = "new"
)

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) Create(ctx context.Context, req CreateOrderRequest) (Order, error) {
	if req.CustomerID == "" {
		return Order{}, errors.New("customer_id is required")
	}
	if req.TotalAmount <= 0 {
		return Order{}, errors.New("total_amount must be > 0")
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Order{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var o Order
	o.CustomerID = req.CustomerID
	o.TotalAmount = req.TotalAmount
	o.Status = StatusNew

	err = tx.QueryRow(ctx, `
		INSERT INTO orders (customer_id, status, total_amount)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`, o.CustomerID, o.Status, o.TotalAmount).Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return Order{}, fmt.Errorf("insert order: %w", err)
	}

	payload := map[string]any{
		"order_id":     o.ID,
		"customer_id":  o.CustomerID,
		"status":       o.Status,
		"total_amount": o.TotalAmount,
		"created_at":   o.CreatedAt,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return Order{}, fmt.Errorf("marshal payload: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO outbox (aggregate_type, aggregate_id, event_type, payload)
		VALUES ($1, $2, $3, $4)
	`, "order", o.ID, "OrderCreated", payloadJSON)
	if err != nil {
		return Order{}, fmt.Errorf("insert outbox: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Order{}, fmt.Errorf("commit: %w", err)
	}

	return o, nil
}
