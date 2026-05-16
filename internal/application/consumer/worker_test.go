package consumer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"notification-system/internal/domain/notification"
	"notification-system/internal/infrastructure/messaging"
)

func TestWorker_ProcessMessage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
		}))
		defer ts.Close()

		n, _ := notification.NewNotification("1", "u", "email", "c", "normal", nil)

		repo := &mockRepository{
			getByIDFunc: func(ctx context.Context, id string) (*notification.Notification, error) {
				return n, nil
			},
			updateFunc: func(ctx context.Context, n *notification.Notification) error {
				return nil
			},
		}

		cache := &mockCacheClient{
			setIdempotencyKeyFunc: func(ctx context.Context, id string) (bool, error) {
				return true, nil
			},
			allowRateLimitFunc: func(ctx context.Context, channel string, max int64) (bool, error) {
				return true, nil
			},
		}

		worker := NewWorker(repo, cache, ts.URL)

		event := messaging.NotificationEvent{
			ID:        "1",
			Priority:  "normal",
			Timestamp: time.Now().Format(time.RFC3339),
		}
		msg, _ := json.Marshal(event)

		err := worker.ProcessMessage(context.Background(), msg)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if n.Status != notification.StatusSent {
			t.Errorf("expected status sent, got %v", n.Status)
		}
	})

	t.Run("invalid msg format", func(t *testing.T) {
		worker := NewWorker(nil, nil, "")
		err := worker.ProcessMessage(context.Background(), []byte("invalid"))
		if err == nil {
			t.Errorf("expected error")
		}
	})

	t.Run("already processed idempotency", func(t *testing.T) {
		cache := &mockCacheClient{
			setIdempotencyKeyFunc: func(ctx context.Context, id string) (bool, error) {
				return false, nil // not new
			},
		}
		worker := NewWorker(nil, cache, "")

		event := messaging.NotificationEvent{ID: "1"}
		msg, _ := json.Marshal(event)

		err := worker.ProcessMessage(context.Background(), msg)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	
	t.Run("rate limit exceeded", func(t *testing.T) {
		n, _ := notification.NewNotification("1", "u", "email", "c", "normal", nil)
		repo := &mockRepository{
			getByIDFunc: func(ctx context.Context, id string) (*notification.Notification, error) {
				return n, nil
			},
			updateFunc: func(ctx context.Context, n *notification.Notification) error {
				return nil
			},
		}
		cache := &mockCacheClient{
			setIdempotencyKeyFunc: func(ctx context.Context, id string) (bool, error) {
				return true, nil
			},
			allowRateLimitFunc: func(ctx context.Context, channel string, max int64) (bool, error) {
				return false, nil // exceeded
			},
		}
		worker := NewWorker(repo, cache, "")

		event := messaging.NotificationEvent{ID: "1"}
		msg, _ := json.Marshal(event)

		err := worker.ProcessMessage(context.Background(), msg)
		if err == nil || err.Error() != "rate limit exceeded for channel email" {
			t.Errorf("unexpected error: %v", err)
		}
		if n.Status != notification.StatusRetry {
			t.Errorf("expected retry status, got %v", n.Status)
		}
	})

	t.Run("webhook delivery failed", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		n, _ := notification.NewNotification("1", "u", "email", "c", "normal", nil)
		repo := &mockRepository{
			getByIDFunc: func(ctx context.Context, id string) (*notification.Notification, error) {
				return n, nil
			},
			updateFunc: func(ctx context.Context, n *notification.Notification) error {
				return nil
			},
		}
		cache := &mockCacheClient{
			setIdempotencyKeyFunc: func(ctx context.Context, id string) (bool, error) {
				return true, nil
			},
			allowRateLimitFunc: func(ctx context.Context, channel string, max int64) (bool, error) {
				return true, nil
			},
		}

		// Hack: use a very short timeout and skip retries by overriding httpClient or just letting it fail fast?
		worker := NewWorker(repo, cache, ts.URL)

		event := messaging.NotificationEvent{ID: "1"}
		msg, _ := json.Marshal(event)

		err := worker.ProcessMessage(context.Background(), msg)
		if err != nil {
			t.Errorf("expected no error, got %v", err) // because delivery failure only sets StatusFailed
		}
		if n.Status != notification.StatusFailed {
			t.Errorf("expected failed status, got %v", n.Status)
		}
	})
}
