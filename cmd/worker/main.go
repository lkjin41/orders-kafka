package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lkjin41/orders-kafka/internal/config"
	"github.com/lkjin41/orders-kafka/internal/db"
	"github.com/lkjin41/orders-kafka/internal/kafka"
	"github.com/lkjin41/orders-kafka/internal/orders"
	"github.com/lkjin41/orders-kafka/internal/worker"
)

func main() {
	cfg := config.MustLoad()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	producer := kafka.NewProducer("localhost:9093")

	pool := db.New(ctx, cfg.DatabaseURL)
	defer pool.Close()

	log.Println("worker started (db connected)")

	consumer := kafka.NewConsumer(cfg.KafkaBrokers)

	ordersRepo := orders.NewRepo(pool, producer)
	wkr := worker.NewWorker(pool, consumer, ordersRepo)

	go wkr.Start(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down worker...")
	time.Sleep(1 * time.Second)
	log.Println("worker stopped")
}
