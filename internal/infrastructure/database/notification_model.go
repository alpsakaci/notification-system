package database

import (
	"time"

	"notification-system/internal/domain/notification"
)

// NotificationModel represents the database schema for a notification.
type NotificationModel struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)"`
	BatchID   *string   `gorm:"index;type:varchar(36)"`
	Recipient string    `gorm:"not null;index"`
	Channel   string    `gorm:"not null"`
	Content   string    `gorm:"type:text;not null"`
	Priority  string    `gorm:"not null"`
	Status    string    `gorm:"not null;index"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// TableName overrides the default table name for GORM.
func (NotificationModel) TableName() string {
	return "notifications"
}

// toDomain converts the DB model to the domain entity.
func (m *NotificationModel) toDomain() *notification.Notification {
	return &notification.Notification{
		ID:        m.ID,
		BatchID:   m.BatchID,
		Recipient: m.Recipient,
		Channel:   notification.Channel(m.Channel),
		Content:   m.Content,
		Priority:  notification.Priority(m.Priority),
		Status:    notification.Status(m.Status),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// fromDomain populates the DB model from the domain entity.
func fromDomain(n *notification.Notification) *NotificationModel {
	return &NotificationModel{
		ID:        n.ID,
		BatchID:   n.BatchID,
		Recipient: n.Recipient,
		Channel:   string(n.Channel),
		Content:   n.Content,
		Priority:  string(n.Priority),
		Status:    string(n.Status),
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
}
