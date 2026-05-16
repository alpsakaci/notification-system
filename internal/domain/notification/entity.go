package notification

import (
	"errors"
	"time"
)

// Channel defines the communication channel for the notification.
type Channel string

const (
	ChannelSMS   Channel = "sms"
	ChannelEmail Channel = "email"
	ChannelPush  Channel = "push"
)

// Priority defines the processing priority.
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityNormal Priority = "normal"
	PriorityLow    Priority = "low"
)

// Status defines the current state of the notification.
type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusSent       Status = "sent"
	StatusFailed     Status = "failed"
	StatusCanceled   Status = "canceled"
	StatusRetry      Status = "retry"
)

// Notification represents the core domain entity.
type Notification struct {
	ID        string
	BatchID   *string
	Recipient string
	Channel   Channel
	Content   string
	Priority  Priority
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewNotification is a factory for creating a valid Notification.
func NewNotification(id string, recipient string, channel Channel, content string, priority Priority, batchID *string) (*Notification, error) {
	if recipient == "" {
		return nil, errors.New("recipient cannot be empty")
	}
	if content == "" {
		return nil, errors.New("content cannot be empty")
	}
	if channel != ChannelSMS && channel != ChannelEmail && channel != ChannelPush {
		return nil, errors.New("invalid channel")
	}
	if priority != PriorityHigh && priority != PriorityNormal && priority != PriorityLow {
		return nil, errors.New("invalid priority")
	}

	now := time.Now().UTC()
	return &Notification{
		ID:        id,
		BatchID:   batchID,
		Recipient: recipient,
		Channel:   channel,
		Content:   content,
		Priority:  priority,
		Status:    StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Cancel updates the status to canceled if it is still pending.
func (n *Notification) Cancel() error {
	if n.Status != StatusPending {
		return errors.New("only pending notifications can be canceled")
	}
	n.Status = StatusCanceled
	n.UpdatedAt = time.Now().UTC()
	return nil
}
