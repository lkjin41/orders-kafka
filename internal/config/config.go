package config

import (
	"log"
	"os"
)

type Config struct {
	HTTPPort     string
	DatabaseURL  string
	KafkaBrokers string
}

func MustLoad() Config {
	cfg := Config{
		HTTPPort:     getenv("HTTP_PORT", "8080"),
		DatabaseURL:  getenv("DATABASE_URL", "postgres://app:app@localhost:5432/app?sslmode=disable"),
		KafkaBrokers: getenv("KAFKA_BROKERS", "localhost:9092"),
	}

	if cfg.HTTPPort == "" {
		log.Fatal("HTTP_PORT is empty")
	}
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is empty")
	}
	if cfg.KafkaBrokers == "" {
		log.Fatal("KAFKA_BROKERS is empty")
	}

	return cfg
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}
