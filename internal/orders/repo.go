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
func (r *Repo) GetByID(ctx context.Context, id int64) (Order, error) {
	var o Order
	err := r.pool.QueryRow(ctx, `
		SELECT id, customer_id, status, total_amount, created_at, updated_at
		FROM orders
		WHERE id = $1
	`, id).Scan(&o.ID, &o.CustomerID, &o.Status, &o.TotalAmount, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Order{}, fmt.Errorf("not found")
		}
		return Order{}, fmt.Errorf("select order: %w", err)
	}
	return o, nil
}

func (r *Repo) MarkPaid(ctx context.Context, id int64) (Order, error) {
	return r.transition(ctx, id, "OrderPaid", []string{StatusNew}, "paid")
}

func (r *Repo) MarkShipped(ctx context.Context, id int64) (Order, error) {
	return r.transition(ctx, id, "OrderShipped", []string{"paid"}, "shipped")
}

func (r *Repo) Cancel(ctx context.Context, id int64) (Order, error) {
	return r.transition(ctx, id, "OrderCanceled", []string{StatusNew, "paid"}, "canceled")
}

func (r *Repo) transition(ctx context.Context, id int64, eventType string, allowedFrom []string, toStatus string) (Order, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Order{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var o Order
	err = tx.QueryRow(ctx, `
		SELECT id, customer_id, status, total_amount, created_at, updated_at
		FROM orders
		WHERE id = $1
		FOR UPDATE
	`, id).Scan(&o.ID, &o.CustomerID, &o.Status, &o.TotalAmount, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Order{}, fmt.Errorf("not found")
		}
		return Order{}, fmt.Errorf("select for update: %w", err)
	}

	ok := false
	for _, s := range allowedFrom {
		if o.Status == s {
			ok = true
			break
		}
	}
	if !ok {
		return Order{}, fmt.Errorf("invalid transition: %s -> %s", o.Status, toStatus)
	}

	err = tx.QueryRow(ctx, `
		UPDATE orders
		SET status = $2, updated_at = now()
		WHERE id = $1
		RETURNING updated_at
	`, o.ID, toStatus).Scan(&o.UpdatedAt)
	if err != nil {
		return Order{}, fmt.Errorf("update order: %w", err)
	}
	o.Status = toStatus

	payload := map[string]any{
		"order_id":     o.ID,
		"customer_id":  o.CustomerID,
		"status":       o.Status,
		"total_amount": o.TotalAmount,
		"updated_at":   o.UpdatedAt,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return Order{}, fmt.Errorf("marshal payload: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO outbox (aggregate_type, aggregate_id, event_type, payload)
		VALUES ($1, $2, $3, $4)
	`, "order", o.ID, eventType, payloadJSON)
	if err != nil {
		return Order{}, fmt.Errorf("insert outbox: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Order{}, fmt.Errorf("commit: %w", err)
	}

	return o, nil
}
