package query

import (
	"context"
	"fmt"

	"notification-system/internal/domain/notification"
)

type ListNotificationsQuery struct {
	Filter notification.ListFilter
}

type ListNotificationsHandler struct {
	repo notification.Repository
}

func NewListNotificationsHandler(repo notification.Repository) *ListNotificationsHandler {
	return &ListNotificationsHandler{repo: repo}
}

func (h *ListNotificationsHandler) Handle(ctx context.Context, query ListNotificationsQuery) ([]*notification.Notification, error) {
	results, err := h.repo.List(ctx, query.Filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}
	return results, nil
}
