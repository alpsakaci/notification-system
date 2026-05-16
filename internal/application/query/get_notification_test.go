package query

import (
	"context"
	"errors"
	"testing"

	"notification-system/internal/domain/notification"
)

func TestGetNotificationHandler_Handle(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		n, _ := notification.NewNotification("1", "u", "email", "c", "normal", nil)
		repo := &mockRepository{
			getByIDFunc: func(ctx context.Context, id string) (*notification.Notification, error) {
				return n, nil
			},
		}
		handler := NewGetNotificationHandler(repo)

		result, err := handler.Handle(context.Background(), GetNotificationQuery{ID: "1"})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result == nil || result.ID != "1" {
			t.Errorf("unexpected result: %v", result)
		}
	})

	t.Run("error", func(t *testing.T) {
		repo := &mockRepository{
			getByIDFunc: func(ctx context.Context, id string) (*notification.Notification, error) {
				return nil, errors.New("db error")
			},
		}
		handler := NewGetNotificationHandler(repo)

		result, err := handler.Handle(context.Background(), GetNotificationQuery{ID: "1"})
		if err == nil || err.Error() != "failed to retrieve notification: db error" {
			t.Errorf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})
}
