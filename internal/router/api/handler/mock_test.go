package handler

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

type mockProducer struct {
	publishFunc      func(ctx context.Context, n *notification.Notification) error
	publishBatchFunc func(ctx context.Context, ns []*notification.Notification) error
}

func (m *mockProducer) Publish(ctx context.Context, n *notification.Notification) error {
	if m.publishFunc != nil {
		return m.publishFunc(ctx, n)
	}
	return nil
}

func (m *mockProducer) PublishBatch(ctx context.Context, ns []*notification.Notification) error {
	if m.publishBatchFunc != nil {
		return m.publishBatchFunc(ctx, ns)
	}
	return nil
}
