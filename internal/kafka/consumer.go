package kafka

import (
	"context"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

type Consumer interface {
	Consume(ctx context.Context) (Message, error)
}

type kafkaConsumer struct {
	reader *kafka.Reader
}

type Message struct {
	EventID   string
	EventType string
	Payload   []byte
}

func NewConsumer(broker string) Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   "orders-events",
		GroupID: "order-service-group",
	})
	return &kafkaConsumer{reader: reader}
}

func (c *kafkaConsumer) Consume(ctx context.Context) (Message, error) {
	m, err := c.reader.ReadMessage(ctx)
	if err != nil {
		log.Printf("failed to read message: %v", err)
		return Message{}, fmt.Errorf("failed to read message: %w", err)
	}
	var eventType string

	if len(m.Headers) > 0 {
		eventType = string(m.Headers[0].Value)
	} else {
		eventType = "unknown"
	}

	return Message{
		EventID:   string(m.Key),
		EventType: eventType,
		Payload:   m.Value,
	}, nil
}

func (c *kafkaConsumer) Close() error {
	return c.reader.Close()
}
