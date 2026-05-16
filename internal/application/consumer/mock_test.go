package consumer

import (
	"context"

	"notification-system/internal/domain/notification"
)

type mockRepository struct {
	saveFunc         func(ctx context.Context, n *notification.Notification) error
	saveBatchFunc    func(ctx context.Context, notifications []*notification.Notification) error
	getByIDFunc      func(ctx context.Context, id string) (*notification.Notification, error)
	getByBatchIDFunc func(ctx context.Context, batchID string) ([]*notification.Notification, error)
	updateFunc       func(ctx context.Context, n *notification.Notification) error
	listFunc         func(ctx context.Context, filter notification.ListFilter) ([]*notification.Notification, error)
}

func (m *mockRepository) Save(ctx context.Context, n *notification.Notification) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, n)
	}
	return nil
}

func (m *mockRepository) SaveBatch(ctx context.Context, notifications []*notification.Notification) error {
	if m.saveBatchFunc != nil {
		return m.saveBatchFunc(ctx, notifications)
	}
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id string) (*notification.Notification, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockRepository) GetByBatchID(ctx context.Context, batchID string) ([]*notification.Notification, error) {
	if m.getByBatchIDFunc != nil {
		return m.getByBatchIDFunc(ctx, batchID)
	}
	return nil, nil
}

func (m *mockRepository) Update(ctx context.Context, n *notification.Notification) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, n)
	}
	return nil
}

func (m *mockRepository) List(ctx context.Context, filter notification.ListFilter) ([]*notification.Notification, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	return nil, nil
}

type mockCacheClient struct {
	setIdempotencyKeyFunc func(ctx context.Context, id string) (bool, error)
	allowRateLimitFunc    func(ctx context.Context, channel string, max int64) (bool, error)
}

func (m *mockCacheClient) SetIdempotencyKey(ctx context.Context, id string) (bool, error) {
	if m.setIdempotencyKeyFunc != nil {
		return m.setIdempotencyKeyFunc(ctx, id)
	}
	return true, nil
}

func (m *mockCacheClient) AllowRateLimit(ctx context.Context, channel string, max int64) (bool, error) {
	if m.allowRateLimitFunc != nil {
		return m.allowRateLimitFunc(ctx, channel, max)
	}
	return true, nil
}
