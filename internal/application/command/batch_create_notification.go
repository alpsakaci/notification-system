package command

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"notification-system/internal/domain/notification"
)

type BatchCreateNotificationCommand struct {
	Items []CreateNotificationCommand
}

type BatchCreateNotificationHandler struct {
	repo     notification.Repository
	producer EventProducer
}

func NewBatchCreateNotificationHandler(repo notification.Repository, producer EventProducer) *BatchCreateNotificationHandler {
	return &BatchCreateNotificationHandler{
		repo:     repo,
		producer: producer,
	}
}

func (h *BatchCreateNotificationHandler) Handle(ctx context.Context, cmd BatchCreateNotificationCommand) ([]*notification.Notification, error) {
	if len(cmd.Items) == 0 {
		return nil, nil
	}

	batchID := uuid.New().String()
	var notifications []*notification.Notification

	for _, item := range cmd.Items {
		id := uuid.New().String()

		n, err := notification.NewNotification(
			id,
			item.Recipient,
			notification.Channel(item.Channel),
			item.Content,
			notification.Priority(item.Priority),
			&batchID,
		)
		if err != nil {
			return nil, fmt.Errorf("invalid notification data for recipient %s: %w", item.Recipient, err)
		}
		notifications = append(notifications, n)
	}

	// Persist batch to database
	if err := h.repo.SaveBatch(ctx, notifications); err != nil {
		return nil, fmt.Errorf("failed to save batch notifications: %w", err)
	}

	// Publish batch to Kafka
	if err := h.producer.PublishBatch(ctx, notifications); err != nil {
		return nil, fmt.Errorf("failed to publish batch notification events: %w", err)
	}

	return notifications, nil
}
