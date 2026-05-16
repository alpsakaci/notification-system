package command

import (
	"context"
	"fmt"

	"notification-system/internal/domain/notification"
)

type CancelNotificationCommand struct {
	ID string
}

type CancelNotificationHandler struct {
	repo notification.Repository
}

func NewCancelNotificationHandler(repo notification.Repository) *CancelNotificationHandler {
	return &CancelNotificationHandler{repo: repo}
}

func (h *CancelNotificationHandler) Handle(ctx context.Context, cmd CancelNotificationCommand) error {
	n, err := h.repo.GetByID(ctx, cmd.ID)
	if err != nil {
		return fmt.Errorf("notification not found: %w", err)
	}

	if err := n.Cancel(); err != nil {
		return fmt.Errorf("failed to cancel notification: %w", err)
	}

	if err := h.repo.Update(ctx, n); err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	return nil
}
