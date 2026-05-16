package notification

import (
	"context"
)

// ListFilter represents filtering options for the List method.
type ListFilter struct {
	Status    *Status
	Channel   *Channel
	StartDate *string
	EndDate   *string
	Limit     int
	Offset    int
}

// Repository defines the interface for database operations.
type Repository interface {
	Save(ctx context.Context, n *Notification) error
	SaveBatch(ctx context.Context, notifications []*Notification) error
	GetByID(ctx context.Context, id string) (*Notification, error)
	GetByBatchID(ctx context.Context, batchID string) ([]*Notification, error)
	Update(ctx context.Context, n *Notification) error
	List(ctx context.Context, filter ListFilter) ([]*Notification, error)
}
