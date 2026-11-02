package repositories

import (
	"context"
	"time"

	"qabulgo-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OutboxRepository interface {
	Enqueue(ctx context.Context, item *models.NotificationOutbox) error
	GetPending(ctx context.Context, limit int) ([]models.NotificationOutbox, error)
	MarkDelivered(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID, lastError string, canRetry bool, nextRetryDelay time.Duration) error
	MarkSkipped(ctx context.Context, id uuid.UUID, reason string) error
}

type outboxRepository struct {
	db *gorm.DB
}

func NewNotificationOutboxRepository(db *gorm.DB) OutboxRepository {
	return &outboxRepository{db: db}
}

func (r *outboxRepository) Enqueue(ctx context.Context, item *models.NotificationOutbox) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	item.NextRetryAt = time.Now()
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *outboxRepository) GetPending(ctx context.Context, limit int) ([]models.NotificationOutbox, error) {
	var list []models.NotificationOutbox
	err := r.db.WithContext(ctx).
		Where("status = ? AND next_retry_at <= ?", "pending", time.Now()).
		Order("created_at ASC").
		Limit(limit).
		Find(&list).Error
	return list, err
}

func (r *outboxRepository) MarkDelivered(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.NotificationOutbox{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":       "delivered",
			"delivered_at": now,
			"updated_at":   now,
		}).Error
}

func (r *outboxRepository) MarkFailed(ctx context.Context, id uuid.UUID, lastError string, canRetry bool, nextRetryDelay time.Duration) error {
	now := time.Now()
	updates := map[string]any{
		"last_error": lastError,
		"updated_at": now,
	}

	if canRetry {
		updates["next_retry_at"] = now.Add(nextRetryDelay)
		return r.db.WithContext(ctx).
			Model(&models.NotificationOutbox{}).
			Where("id = ?", id).
			UpdateColumn("retry_count", gorm.Expr("retry_count + 1")).
			Updates(updates).Error
	}

	updates["status"] = "failed"
	return r.db.WithContext(ctx).
		Model(&models.NotificationOutbox{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *outboxRepository) MarkSkipped(ctx context.Context, id uuid.UUID, reason string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.NotificationOutbox{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":     "skipped",
			"last_error": reason,
			"updated_at": now,
		}).Error
}

