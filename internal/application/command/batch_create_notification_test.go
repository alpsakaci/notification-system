package command

import (
	"context"
	"errors"
	"testing"

	"notification-system/internal/domain/notification"
)

func TestBatchCreateNotificationHandler_Handle(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockRepository{
			saveBatchFunc: func(ctx context.Context, notifications []*notification.Notification) error {
				return nil
			},
		}
		producer := &mockProducer{
			publishBatchFunc: func(ctx context.Context, ns []*notification.Notification) error {
				return nil
			},
		}
		handler := NewBatchCreateNotificationHandler(repo, producer)

		cmd := BatchCreateNotificationCommand{
			Items: []CreateNotificationCommand{
				{Recipient: "user1", Channel: "sms", Content: "A", Priority: "high"},
				{Recipient: "user2", Channel: "email", Content: "B", Priority: "normal"},
			},
		}

		ns, err := handler.Handle(context.Background(), cmd)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(ns) != 2 {
			t.Fatalf("expected 2 notifications, got %d", len(ns))
		}
		if ns[0].BatchID == nil || *ns[0].BatchID == "" {
			t.Errorf("expected batch ID to be set")
		}
		if *ns[0].BatchID != *ns[1].BatchID {
			t.Errorf("expected identical batch IDs")
		}
	})

	t.Run("empty items", func(t *testing.T) {
		handler := NewBatchCreateNotificationHandler(nil, nil)
		ns, err := handler.Handle(context.Background(), BatchCreateNotificationCommand{})
		if err != nil {
			t.Errorf("expected no error")
		}
		if ns != nil {
			t.Errorf("expected nil result")
		}
	})

	t.Run("invalid item data", func(t *testing.T) {
		handler := NewBatchCreateNotificationHandler(nil, nil)
		cmd := BatchCreateNotificationCommand{
			Items: []CreateNotificationCommand{
				{Recipient: "", Channel: "sms", Content: "A", Priority: "high"},
			},
		}
		ns, err := handler.Handle(context.Background(), cmd)
		if err == nil {
			t.Errorf("expected error")
		}
		if ns != nil {
			t.Errorf("expected nil result")
		}
	})

	t.Run("repo save error", func(t *testing.T) {
		repo := &mockRepository{
			saveBatchFunc: func(ctx context.Context, notifications []*notification.Notification) error {
				return errors.New("db error")
			},
		}
		handler := NewBatchCreateNotificationHandler(repo, &mockProducer{})
		cmd := BatchCreateNotificationCommand{
			Items: []CreateNotificationCommand{{Recipient: "user1", Channel: "sms", Content: "A", Priority: "high"}},
		}
		ns, err := handler.Handle(context.Background(), cmd)
		if err == nil || err.Error() != "failed to save batch notifications: db error" {
			t.Errorf("unexpected error: %v", err)
		}
		if ns != nil {
			t.Errorf("expected nil result")
		}
	})

	t.Run("producer publish error", func(t *testing.T) {
		repo := &mockRepository{}
		producer := &mockProducer{
			publishBatchFunc: func(ctx context.Context, ns []*notification.Notification) error {
				return errors.New("kafka error")
			},
		}
		handler := NewBatchCreateNotificationHandler(repo, producer)
		cmd := BatchCreateNotificationCommand{
			Items: []CreateNotificationCommand{{Recipient: "user1", Channel: "sms", Content: "A", Priority: "high"}},
		}
		ns, err := handler.Handle(context.Background(), cmd)
		if err == nil || err.Error() != "failed to publish batch notification events: kafka error" {
			t.Errorf("unexpected error: %v", err)
		}
		if ns != nil {
			t.Errorf("expected nil result")
		}
	})
}
