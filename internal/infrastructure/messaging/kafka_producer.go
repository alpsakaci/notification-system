package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"

	"notification-system/internal/domain/notification"
)

// NotificationEvent defines the message structure published to Kafka.
type NotificationEvent struct {
	ID        string `json:"id"`
	Priority  string `json:"priority"`
	Timestamp string `json:"timestamp"`
}

// KafkaProducer is responsible for sending messages to Kafka topics.
type KafkaProducer struct {
	writer *kafka.Writer
}

// NewKafkaProducer initializes a new Kafka producer.
func NewKafkaProducer(brokers []string) *KafkaProducer {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		AllowAutoTopicCreation: true, // Auto-creates topics if they don't exist
		Balancer:               &kafka.LeastBytes{},
	}
	return &KafkaProducer{writer: w}
}

// Publish routes the notification to the correct topic based on its priority.
func (p *KafkaProducer) Publish(ctx context.Context, n *notification.Notification) error {
	var topic string
	switch n.Priority {
	case notification.PriorityHigh:
		topic = "notifications.high"
	case notification.PriorityLow:
		topic = "notifications.low"
	default:
		topic = "notifications.normal"
	}

	event := NotificationEvent{
		ID:        n.ID,
		Priority:  string(n.Priority),
		Timestamp: n.CreatedAt.Format(time.RFC3339),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(n.ID), // Use ID as partition key
		Value: payload,
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to publish message to topic %s: %w", topic, err)
	}

	return nil
}

// Close gracefully shuts down the Kafka writer.
func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}
