package kafka

import (
	"context"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

type Producer interface {
	Publish(ctx context.Context, message Message) error
}

type kafkaProducer struct {
	writer *kafka.Writer
}

func NewProducer(broker string) Producer {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{broker},
		Topic:   "orders-events",
	})
	return &kafkaProducer{writer: writer}

}

func (p *kafkaProducer) Publish(ctx context.Context, message Message) error {
	err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(message.EventType),
		Value: message.Payload,
		Headers: []kafka.Header{
			{Key: "EventType", Value: []byte(message.EventType)},
		},
	})
	if err != nil {
		log.Printf("failed to write message to Kafka: %v", err)
		return fmt.Errorf("failed to write message: %w", err)
	}
	return nil
}

func (p *kafkaProducer) Close() error {
	return p.writer.Close()
}
