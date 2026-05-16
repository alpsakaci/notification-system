package database

import (
	"context"

	"gorm.io/gorm"

	"notification-system/internal/domain/notification"
)

type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository returns a new instance of a PostgreSQL repository.
func NewNotificationRepository(db *gorm.DB) notification.Repository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Save(ctx context.Context, n *notification.Notification) error {
	model := fromDomain(n)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *notificationRepository) SaveBatch(ctx context.Context, notifications []*notification.Notification) error {
	var models []*NotificationModel
	for _, n := range notifications {
		models = append(models, fromDomain(n))
	}
	// Create records in batches of 100 to optimize performance and prevent SQL syntax errors for too many placeholders
	return r.db.WithContext(ctx).CreateInBatches(models, 100).Error
}

func (r *notificationRepository) GetByID(ctx context.Context, id string) (*notification.Notification, error) {
	var model NotificationModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		return nil, err
	}
	return model.toDomain(), nil
}

func (r *notificationRepository) GetByBatchID(ctx context.Context, batchID string) ([]*notification.Notification, error) {
	var models []*NotificationModel
	if err := r.db.WithContext(ctx).Where("batch_id = ?", batchID).Find(&models).Error; err != nil {
		return nil, err
	}

	var results []*notification.Notification
	for _, m := range models {
		results = append(results, m.toDomain())
	}
	return results, nil
}

func (r *notificationRepository) Update(ctx context.Context, n *notification.Notification) error {
	model := fromDomain(n)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *notificationRepository) List(ctx context.Context, filter notification.ListFilter) ([]*notification.Notification, error) {
	query := r.db.WithContext(ctx).Model(&NotificationModel{})

	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.Channel != nil {
		query = query.Where("channel = ?", *filter.Channel)
	}
	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", *filter.EndDate)
	}

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// Order by most recent first
	query = query.Order("created_at DESC")

	var models []*NotificationModel
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	var results []*notification.Notification
	for _, m := range models {
		results = append(results, m.toDomain())
	}
	return results, nil
}
