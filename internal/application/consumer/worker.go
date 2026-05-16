package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"notification-system/internal/domain/notification"
	"notification-system/internal/infrastructure/messaging"
	"notification-system/internal/infrastructure/observability"
)

// CacheClient abstracts the caching mechanisms for idempotency and rate limiting.
type CacheClient interface {
	SetIdempotencyKey(ctx context.Context, id string) (bool, error)
	AllowRateLimit(ctx context.Context, channel string, max int64) (bool, error)
}

type Worker struct {
	repo       notification.Repository
	cache      CacheClient
	webhookURL string
	httpClient *http.Client
}

func NewWorker(repo notification.Repository, cacheClient CacheClient, webhookURL string) *Worker {
	return &Worker{
		repo:       repo,
		cache:      cacheClient,
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (w *Worker) ProcessMessage(ctx context.Context, msg []byte) error {
	start := time.Now()

	var event messaging.NotificationEvent
	if err := json.Unmarshal(msg, &event); err != nil {
		slog.Error("Invalid message format", "error", err)
		return fmt.Errorf("invalid message format: %w", err)
	}

	// For structured logging, we attach the event ID
	logger := slog.With("notification_id", event.ID, "priority", event.Priority)

	// 1. Check Idempotency
	isNew, err := w.cache.SetIdempotencyKey(ctx, event.ID)
	if err != nil {
		logger.Error("Failed to check idempotency", "error", err)
		return fmt.Errorf("failed to check idempotency: %w", err)
	}
	if !isNew {
		logger.Info("Notification was already processed, skipping")
		return nil
	}

	// 2. Fetch Notification Details from DB
	n, err := w.repo.GetByID(ctx, event.ID)
	if err != nil {
		logger.Error("Failed to get notification from DB", "error", err)
		return fmt.Errorf("failed to get notification from db: %w", err)
	}

	// Instrument latency at the end
	defer func() {
		observability.NotificationLatency.WithLabelValues(string(n.Channel)).Observe(time.Since(start).Seconds())
	}()

	if n.Status == notification.StatusCanceled {
		logger.Info("Notification is canceled, skipping")
		return nil
	}

	// 3. Rate Limiting Check
	allowed, err := w.cache.AllowRateLimit(ctx, string(n.Channel), 100)
	if err != nil {
		logger.Error("Failed to check rate limit", "error", err)
		return fmt.Errorf("failed to check rate limit: %w", err)
	}
	if !allowed {
		observability.RateLimitHits.WithLabelValues(string(n.Channel)).Inc()
		logger.Warn("Rate limit exceeded", "channel", n.Channel)

		n.Status = notification.StatusRetry
		_ = w.repo.Update(ctx, n)
		return fmt.Errorf("rate limit exceeded for channel %s", n.Channel)
	}

	n.Status = notification.StatusProcessing
	_ = w.repo.Update(ctx, n)

	// 4. Send to Webhook
	payload := map[string]string{
		"to":      n.Recipient,
		"channel": string(n.Channel),
		"content": n.Content,
	}
	jsonPayload, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, w.webhookURL, bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")

	maxRetries := 3
	var resp *http.Response
	var deliveryErr error

	for i := 0; i < maxRetries; i++ {
		resp, deliveryErr = w.httpClient.Do(req)
		if deliveryErr == nil && (resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusOK) {
			break
		}
		logger.Warn("Webhook attempt failed", "attempt", i+1, "error", deliveryErr)
		time.Sleep(time.Duration(2<<i) * time.Second) // Exponential backoff
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	// Determine final status
	if deliveryErr == nil && resp != nil && (resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusOK) {
		n.Status = notification.StatusSent
		logger.Info("Notification delivered successfully")
		observability.NotificationsProcessed.WithLabelValues(string(n.Channel), "success").Inc()
	} else {
		n.Status = notification.StatusFailed
		logger.Error("Notification delivery failed after retries", "max_retries", maxRetries)
		observability.NotificationsProcessed.WithLabelValues(string(n.Channel), "failed").Inc()
	}

	if err := w.repo.Update(ctx, n); err != nil {
		logger.Error("Failed to update final status", "error", err)
		return fmt.Errorf("failed to update final status: %w", err)
	}

	return nil
}
