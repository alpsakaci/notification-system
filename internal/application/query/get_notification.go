package query

import (
	"context"
	"fmt"

	"notification-system/internal/domain/notification"
)

type GetNotificationQuery struct {
	ID string
}

type GetNotificationHandler struct {
	repo notification.Repository
}

func NewGetNotificationHandler(repo notification.Repository) *GetNotificationHandler {
	return &GetNotificationHandler{repo: repo}
}

func (h *GetNotificationHandler) Handle(ctx context.Context, query GetNotificationQuery) (*notification.Notification, error) {
	n, err := h.repo.GetByID(ctx, query.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve notification: %w", err)
	}
	return n, nil
}
