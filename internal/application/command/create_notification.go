package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"notification-system/internal/domain/notification"
)

// EventProducer defines the interface for producing notification events.
type EventProducer interface {
	Publish(ctx context.Context, n *notification.Notification) error
}

type CreateNotificationCommand struct {
	Recipient string
	Channel   string
	Content   string
	Priority  string
	BatchID   *string
}

type CreateNotificationHandler struct {
	repo     notification.Repository
	producer EventProducer
}

func NewCreateNotificationHandler(repo notification.Repository, producer EventProducer) *CreateNotificationHandler {
	return &CreateNotificationHandler{
		repo:     repo,
		producer: producer,
	}
}

func (h *CreateNotificationHandler) Handle(ctx context.Context, cmd CreateNotificationCommand) (*notification.Notification, error) {
	id := uuid.New().String()

	n, err := notification.NewNotification(
		id,
		cmd.Recipient,
		notification.Channel(cmd.Channel),
		cmd.Content,
		notification.Priority(cmd.Priority),
		cmd.BatchID,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid notification data: %w", err)
	}

	// Persist to database
	if err := h.repo.Save(ctx, n); err != nil {
		return nil, fmt.Errorf("failed to save notification: %w", err)
	}

	// Publish to Kafka
	if err := h.producer.Publish(ctx, n); err != nil {
		// Even if publishing fails, we've saved it as Pending. A background job could retry unpublished events in a real system.
		return nil, fmt.Errorf("failed to publish notification event: %w", err)
	}

	return n, nil
}
