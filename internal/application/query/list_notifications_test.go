package query

import (
	"context"
	"errors"
	"testing"

	"notification-system/internal/domain/notification"
)

func TestListNotificationsHandler_Handle(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		n1, _ := notification.NewNotification("1", "u", "email", "c", "normal", nil)
		n2, _ := notification.NewNotification("2", "u", "sms", "c", "normal", nil)
		repo := &mockRepository{
			listFunc: func(ctx context.Context, filter notification.ListFilter) ([]*notification.Notification, error) {
				return []*notification.Notification{n1, n2}, nil
			},
		}
		handler := NewListNotificationsHandler(repo)

		result, err := handler.Handle(context.Background(), ListNotificationsQuery{})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 notifications, got %d", len(result))
		}
	})

	t.Run("error", func(t *testing.T) {
		repo := &mockRepository{
			listFunc: func(ctx context.Context, filter notification.ListFilter) ([]*notification.Notification, error) {
				return nil, errors.New("db error")
			},
		}
		handler := NewListNotificationsHandler(repo)

		result, err := handler.Handle(context.Background(), ListNotificationsQuery{})
		if err == nil || err.Error() != "failed to list notifications: db error" {
			t.Errorf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})
}
