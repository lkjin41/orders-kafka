package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lkjin41/orders-kafka/internal/kafka"
	"github.com/lkjin41/orders-kafka/internal/orders"
)

type Worker struct {
	pool      *pgxpool.Pool
	consumer  kafka.Consumer
	orderRepo *orders.Repo
}

func NewWorker(pool *pgxpool.Pool, consumer kafka.Consumer, orderRepo *orders.Repo) *Worker {
	return &Worker{
		pool:      pool,
		consumer:  consumer,
		orderRepo: orderRepo,
	}
}

func (w *Worker) Start(ctx context.Context) {
	for {
		msg, err := w.consumer.Consume(ctx)
		if err != nil {
			log.Printf("Error consuming message: %v", err)
			continue
		}

		log.Printf("Received event: %s", msg.EventType)

		tx, err := w.pool.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			log.Printf("Failed to begin transaction: %v", err)
			continue
		}

		err = w.processEvent(ctx, tx, msg)
		if err != nil {
			_ = tx.Rollback(ctx)
			log.Printf("Failed to process event: %v", err)
			continue
		}

		if err := tx.Commit(ctx); err != nil {
			_ = tx.Rollback(ctx)
			log.Printf("Failed to commit transaction: %v", err)
			continue
		}

		log.Printf("Processed event: %s", msg.EventType)
	}
}

func (w *Worker) processEvent(ctx context.Context, tx pgx.Tx, msg kafka.Message) error {
	if err := w.checkProcessedEvent(ctx, tx, msg); err != nil {
		return err
	}

	switch msg.EventType {
	case "OrderCreated":
		return w.handleOrderCreated(ctx, tx, msg)
	case "OrderPaid":
		return w.handleOrderPaid(ctx, tx, msg)
	case "OrderShipped":
		return w.handleOrderShipped(ctx, tx, msg)
	case "OrderCanceled":
		return w.handleOrderCanceled(ctx, tx, msg)
	default:
		return fmt.Errorf("unknown event type: %s", msg.EventType)
	}
}

func (w *Worker) checkProcessedEvent(ctx context.Context, tx pgx.Tx, msg kafka.Message) error {
	var exists bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM processed_events WHERE event_id = $1)`,
		msg.EventID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check processed events: %w", err)
	}

	if exists {
		log.Printf("Event %s already processed", msg.EventID)
		return nil
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO processed_events (event_id) VALUES ($1)`,
		msg.EventID)
	if err != nil {
		return fmt.Errorf("insert error: %v", err)
	}

	return nil
}

func (w *Worker) handleOrderCreated(ctx context.Context, tx pgx.Tx, msg kafka.Message) error {
	var order orders.Order
	if err := json.Unmarshal(msg.Payload, &order); err != nil {
		return fmt.Errorf("unmarshal OrderCreated: %w", err)
	}

	_, err := tx.Exec(ctx, `
		INSERT INTO order_activity (event_id, order_id, event_type, payload, received_at)
		VALUES ($1, $2, $3, $4, now())`,
		msg.EventID, order.ID, "OrderCreated", msg.Payload)
	return err
}

func (w *Worker) handleOrderPaid(ctx context.Context, tx pgx.Tx, msg kafka.Message) error {
	var order orders.Order
	if err := json.Unmarshal(msg.Payload, &order); err != nil {
		return fmt.Errorf("unmarshal OrderPaid: %w", err)
	}

	_, err := tx.Exec(ctx, `
		INSERT INTO order_activity (event_id, order_id, event_type, payload, received_at)
		VALUES ($1, $2, $3, $4, now())`,
		msg.EventID, order.ID, "OrderPaid", msg.Payload)
	return err
}

func (w *Worker) handleOrderShipped(ctx context.Context, tx pgx.Tx, msg kafka.Message) error {
	var order orders.Order
	if err := json.Unmarshal(msg.Payload, &order); err != nil {
		return fmt.Errorf("unmarshal OrderShipped: %w", err)
	}

	_, err := tx.Exec(ctx, `
		INSERT INTO order_activity (event_id, order_id, event_type, payload, received_at)
		VALUES ($1, $2, $3, $4, now())`,
		msg.EventID, order.ID, "OrderShipped", msg.Payload)
	return err
}

func (w *Worker) handleOrderCanceled(ctx context.Context, tx pgx.Tx, msg kafka.Message) error {
	var order orders.Order
	if err := json.Unmarshal(msg.Payload, &order); err != nil {
		return fmt.Errorf("unmarshal OrderCanceled: %w", err)
	}

	_, err := tx.Exec(ctx, `
		INSERT INTO order_activity (event_id, order_id, event_type, payload, received_at)
		VALUES ($1, $2, $3, $4, now())`,
		msg.EventID, order.ID, "OrderCanceled", msg.Payload)
	return err
}
