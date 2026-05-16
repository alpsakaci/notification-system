package notification

import (
	"testing"
)

func TestNewNotification(t *testing.T) {
	batchID := "batch-123"

	tests := []struct {
		name      string
		id        string
		recipient string
		channel   Channel
		content   string
		priority  Priority
		batchID   *string
		wantErr   string
	}{
		{
			name:      "valid notification without batch ID",
			id:        "1",
			recipient: "user@example.com",
			channel:   ChannelEmail,
			content:   "Hello",
			priority:  PriorityNormal,
			batchID:   nil,
			wantErr:   "",
		},
		{
			name:      "valid notification with batch ID",
			id:        "2",
			recipient: "user@example.com",
			channel:   ChannelSMS,
			content:   "Hello",
			priority:  PriorityHigh,
			batchID:   &batchID,
			wantErr:   "",
		},
		{
			name:      "empty recipient",
			id:        "3",
			recipient: "",
			channel:   ChannelEmail,
			content:   "Hello",
			priority:  PriorityNormal,
			batchID:   nil,
			wantErr:   "recipient cannot be empty",
		},
		{
			name:      "empty content",
			id:        "4",
			recipient: "user@example.com",
			channel:   ChannelEmail,
			content:   "",
			priority:  PriorityNormal,
			batchID:   nil,
			wantErr:   "content cannot be empty",
		},
		{
			name:      "invalid channel",
			id:        "5",
			recipient: "user@example.com",
			channel:   Channel("invalid"),
			content:   "Hello",
			priority:  PriorityNormal,
			batchID:   nil,
			wantErr:   "invalid channel",
		},
		{
			name:      "invalid priority",
			id:        "6",
			recipient: "user@example.com",
			channel:   ChannelEmail,
			content:   "Hello",
			priority:  Priority("invalid"),
			batchID:   nil,
			wantErr:   "invalid priority",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := NewNotification(tt.id, tt.recipient, tt.channel, tt.content, tt.priority, tt.batchID)

			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.wantErr)
				} else if err.Error() != tt.wantErr {
					t.Errorf("expected error %q, got %q", tt.wantErr, err.Error())
				}
				if n != nil {
					t.Errorf("expected nil notification, got %+v", n)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if n == nil {
					t.Fatalf("expected valid notification, got nil")
				}
				if n.ID != tt.id {
					t.Errorf("expected ID %q, got %q", tt.id, n.ID)
				}
				if n.Recipient != tt.recipient {
					t.Errorf("expected Recipient %q, got %q", tt.recipient, n.Recipient)
				}
				if n.Channel != tt.channel {
					t.Errorf("expected Channel %q, got %q", tt.channel, n.Channel)
				}
				if n.Content != tt.content {
					t.Errorf("expected Content %q, got %q", tt.content, n.Content)
				}
				if n.Priority != tt.priority {
					t.Errorf("expected Priority %q, got %q", tt.priority, n.Priority)
				}
				if n.BatchID != tt.batchID {
					t.Errorf("expected BatchID %v, got %v", tt.batchID, n.BatchID)
				}
				if n.Status != StatusPending {
					t.Errorf("expected Status %q, got %q", StatusPending, n.Status)
				}
			}
		})
	}
}

func TestCancel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		n, _ := NewNotification("1", "user", ChannelEmail, "content", PriorityNormal, nil)
		err := n.Cancel()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if n.Status != StatusCanceled {
			t.Errorf("expected status %q, got %q", StatusCanceled, n.Status)
		}
	})

	t.Run("failure - not pending", func(t *testing.T) {
		n, _ := NewNotification("1", "user", ChannelEmail, "content", PriorityNormal, nil)
		n.Status = StatusProcessing // change state
		
		err := n.Cancel()
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if err.Error() != "only pending notifications can be canceled" {
			t.Errorf("expected specific error, got %q", err.Error())
		}
	})
}
