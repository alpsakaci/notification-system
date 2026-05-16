package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the go-redis client.
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates a new Redis connection.
func NewRedisClient(addr string) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &RedisClient{client: client}, nil
}

// AllowRateLimit checks if a channel has exceeded its rate limit (e.g. 100 msgs/sec).
// It uses a simple per-second counter for demonstration.
func (r *RedisClient) AllowRateLimit(ctx context.Context, channel string, max int64) (bool, error) {
	// Group keys by the current second
	key := "rate_limit:" + channel + ":" + time.Now().UTC().Format("2006-01-02T15:04:05")

	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		r.client.Expire(ctx, key, 3*time.Second) // Ensure keys are cleaned up
	}

	if count > max {
		return false, nil
	}

	return true, nil
}

// SetIdempotencyKey ensures a notification ID is processed only once.
// Returns true if the key was set (meaning it's the first time), false if it already exists.
func (r *RedisClient) SetIdempotencyKey(ctx context.Context, id string) (bool, error) {
	key := "idempotency:notification:" + id
	// Store for 24 hours
	return r.client.SetNX(ctx, key, "processed", 24*time.Hour).Result()
}
