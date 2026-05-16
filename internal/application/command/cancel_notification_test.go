package command

import (
	"context"
	"errors"
	"testing"

	"notification-system/internal/domain/notification"
)

func TestCancelNotificationHandler_Handle(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		n, _ := notification.NewNotification("1", "u", "email", "c", "normal", nil)
		repo := &mockRepository{
			getByIDFunc: func(ctx context.Context, id string) (*notification.Notification, error) {
				return n, nil
			},
			updateFunc: func(ctx context.Context, n *notification.Notification) error {
				return nil
			},
		}
		handler := NewCancelNotificationHandler(repo)
		
		err := handler.Handle(context.Background(), CancelNotificationCommand{ID: "1"})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if n.Status != notification.StatusCanceled {
			t.Errorf("expected status canceled, got %v", n.Status)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := &mockRepository{
			getByIDFunc: func(ctx context.Context, id string) (*notification.Notification, error) {
				return nil, errors.New("not found")
			},
		}
		handler := NewCancelNotificationHandler(repo)
		
		err := handler.Handle(context.Background(), CancelNotificationCommand{ID: "1"})
		if err == nil || err.Error() != "notification not found: not found" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("cancel domain error", func(t *testing.T) {
		n, _ := notification.NewNotification("1", "u", "email", "c", "normal", nil)
		n.Status = notification.StatusSent // cannot cancel
		repo := &mockRepository{
			getByIDFunc: func(ctx context.Context, id string) (*notification.Notification, error) {
				return n, nil
			},
		}
		handler := NewCancelNotificationHandler(repo)
		
		err := handler.Handle(context.Background(), CancelNotificationCommand{ID: "1"})
		if err == nil || err.Error() != "failed to cancel notification: only pending notifications can be canceled" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("update failure", func(t *testing.T) {
		n, _ := notification.NewNotification("1", "u", "email", "c", "normal", nil)
		repo := &mockRepository{
			getByIDFunc: func(ctx context.Context, id string) (*notification.Notification, error) {
				return n, nil
			},
			updateFunc: func(ctx context.Context, n *notification.Notification) error {
				return errors.New("db error")
			},
		}
		handler := NewCancelNotificationHandler(repo)
		
		err := handler.Handle(context.Background(), CancelNotificationCommand{ID: "1"})
		if err == nil || err.Error() != "failed to update notification status: db error" {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
