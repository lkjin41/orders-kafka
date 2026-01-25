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
)

func main() {
	cfg := config.MustLoad()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := db.New(ctx, cfg.DatabaseURL)
	defer pool.Close()

	log.Println("worker started (db connected)")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down worker...")
	time.Sleep(1 * time.Second)
	log.Println("worker stopped")
}
