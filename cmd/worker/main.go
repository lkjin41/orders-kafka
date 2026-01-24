package main

import (
	"log"
	"time"

	"github.com/lkjin41/orders-kafka/internal/config"
)

func main() {
	cfg := config.MustLoad()
	log.Printf("worker started (noop). brokers=%s db=%s", cfg.KafkaBrokers, cfg.DatabaseURL)

	for {
		time.Sleep(10 * time.Second)
		log.Println("worker heartbeat")
	}
}
