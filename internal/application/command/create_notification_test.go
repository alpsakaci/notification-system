package command

import (
	"context"
	"errors"
	"testing"

	"notification-system/internal/domain/notification"
)

func TestCreateNotificationHandler_Handle(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockRepository{
			saveFunc: func(ctx context.Context, n *notification.Notification) error {
				return nil
			},
		}
		producer := &mockProducer{
			publishFunc: func(ctx context.Context, n *notification.Notification) error {
				return nil
			},
		}
		handler := NewCreateNotificationHandler(repo, producer)

		cmd := CreateNotificationCommand{
			Recipient: "user@example.com",
			Channel:   "email",
			Content:   "Hello",
			Priority:  "normal",
		}

		n, err := handler.Handle(context.Background(), cmd)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if n == nil {
			t.Fatalf("expected notification, got nil")
		}
		if n.Recipient != "user@example.com" {
			t.Errorf("expected recipient 'user@example.com', got %q", n.Recipient)
		}
	})

	t.Run("invalid domain entity", func(t *testing.T) {
		repo := &mockRepository{}
		producer := &mockProducer{}
		handler := NewCreateNotificationHandler(repo, producer)

		cmd := CreateNotificationCommand{
			Recipient: "", // invalid
			Channel:   "email",
			Content:   "Hello",
			Priority:  "normal",
		}

		n, err := handler.Handle(context.Background(), cmd)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
		if n != nil {
			t.Errorf("expected nil notification, got %v", n)
		}
	})

	t.Run("repo save failure", func(t *testing.T) {
		repo := &mockRepository{
			saveFunc: func(ctx context.Context, n *notification.Notification) error {
				return errors.New("db error")
			},
		}
		producer := &mockProducer{}
		handler := NewCreateNotificationHandler(repo, producer)

		cmd := CreateNotificationCommand{
			Recipient: "user@example.com",
			Channel:   "email",
			Content:   "Hello",
			Priority:  "normal",
		}

		n, err := handler.Handle(context.Background(), cmd)
		if err == nil || err.Error() != "failed to save notification: db error" {
			t.Errorf("unexpected error: %v", err)
		}
		if n != nil {
			t.Errorf("expected nil notification")
		}
	})

	t.Run("producer publish failure", func(t *testing.T) {
		repo := &mockRepository{}
		producer := &mockProducer{
			publishFunc: func(ctx context.Context, n *notification.Notification) error {
				return errors.New("kafka error")
			},
		}
		handler := NewCreateNotificationHandler(repo, producer)

		cmd := CreateNotificationCommand{
			Recipient: "user@example.com",
			Channel:   "email",
			Content:   "Hello",
			Priority:  "normal",
		}

		n, err := handler.Handle(context.Background(), cmd)
		if err == nil || err.Error() != "failed to publish notification event: kafka error" {
			t.Errorf("unexpected error: %v", err)
		}
		if n != nil {
			t.Errorf("expected nil notification")
		}
	})
}
