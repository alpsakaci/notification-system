package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"notification-system/internal/domain/notification"
	"notification-system/internal/infrastructure/cache"
	"notification-system/internal/infrastructure/messaging"
)

type Worker struct {
	repo       notification.Repository
	redis      *cache.RedisClient
	webhookURL string
	httpClient *http.Client
}

func NewWorker(repo notification.Repository, redisClient *cache.RedisClient, webhookURL string) *Worker {
	return &Worker{
		repo:       repo,
		redis:      redisClient,
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (w *Worker) ProcessMessage(ctx context.Context, msg []byte) error {
	var event messaging.NotificationEvent
	if err := json.Unmarshal(msg, &event); err != nil {
		return fmt.Errorf("invalid message format: %w", err)
	}

	// 1. Check Idempotency
	isNew, err := w.redis.SetIdempotencyKey(ctx, event.ID)
	if err != nil {
		return fmt.Errorf("failed to check idempotency: %w", err)
	}
	if !isNew {
		log.Printf("Notification %s was already processed, skipping", event.ID)
		return nil
	}

	// 2. Fetch Notification Details from DB
	n, err := w.repo.GetByID(ctx, event.ID)
	if err != nil {
		return fmt.Errorf("failed to get notification from db: %w", err)
	}

	if n.Status == notification.StatusCanceled {
		log.Printf("Notification %s is canceled, skipping", event.ID)
		return nil
	}

	// 3. Rate Limiting Check
	allowed, err := w.redis.AllowRateLimit(ctx, string(n.Channel), 100)
	if err != nil {
		return fmt.Errorf("failed to check rate limit: %w", err)
	}
	if !allowed {
		// Rate limit exceeded, we should ideally push back to queue or retry later.
		// Setting status to Retry so it can be handled by a cron/retry mechanism.
		n.Status = notification.StatusRetry
		_ = w.repo.Update(ctx, n)
		return fmt.Errorf("rate limit exceeded for channel %s, notification %s set to retry", n.Channel, n.ID)
	}

	// Update status to processing
	n.Status = notification.StatusProcessing
	_ = w.repo.Update(ctx, n)

	// 4. Send to Webhook (Simulate external provider)
	payload := map[string]string{
		"to":      n.Recipient,
		"channel": string(n.Channel),
		"content": n.Content,
	}
	jsonPayload, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, w.webhookURL, bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")

	// Retry logic variables
	maxRetries := 3
	var resp *http.Response
	var deliveryErr error

	for i := 0; i < maxRetries; i++ {
		resp, deliveryErr = w.httpClient.Do(req)
		if deliveryErr == nil && (resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusOK) {
			break // Success
		}
		log.Printf("Webhook attempt %d failed for notification %s. Retrying...", i+1, n.ID)
		time.Sleep(time.Duration(2<<i) * time.Second) // Exponential backoff
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	// Determine final status
	if deliveryErr == nil && resp != nil && (resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusOK) {
		n.Status = notification.StatusSent
		log.Printf("Notification %s delivered successfully", n.ID)
	} else {
		n.Status = notification.StatusFailed
		log.Printf("Notification %s delivery failed after %d retries", n.ID, maxRetries)
	}

	if err := w.repo.Update(ctx, n); err != nil {
		return fmt.Errorf("failed to update final status: %w", err)
	}

	return nil
}
