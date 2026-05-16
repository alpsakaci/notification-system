package messaging

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

// MessageHandler is a callback function for processing a Kafka message.
type MessageHandler func(ctx context.Context, msg []byte) error

// KafkaConsumer is responsible for consuming messages from a specific Kafka topic.
type KafkaConsumer struct {
	reader *kafka.Reader
}

// NewKafkaConsumer initializes a new Kafka consumer for a specific topic.
func NewKafkaConsumer(brokers []string, groupID string, topic string) *KafkaConsumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: groupID,
		Topic:   topic,
	})
	return &KafkaConsumer{reader: r}
}

// Start begins consuming messages in a blocking loop until the context is canceled.
func (c *KafkaConsumer) Start(ctx context.Context, handler MessageHandler) {
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return // Context canceled, gracefully exit
			}
			log.Printf("error while fetching message from topic %s: %v", c.reader.Config().Topic, err)
			continue
		}

		// Execute the handler logic
		if err := handler(ctx, m.Value); err != nil {
			log.Printf("failed to handle message %s: %v", string(m.Key), err)
			// In a more complex system, if this fails due to a transient issue, we might NOT commit
			// or we might push to a Dead Letter Queue. For now, we commit to avoid blocking.
		}

		// Acknowledge the message to Kafka
		if err := c.reader.CommitMessages(ctx, m); err != nil {
			log.Printf("failed to commit message: %v", err)
		}
	}
}

// Close gracefully shuts down the Kafka reader.
func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}
