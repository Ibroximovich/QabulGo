package repositories

import (
	"context"
	"errors"
	"time"

	"qabulgo-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrSessionNotFound = errors.New("refresh sessiyasi topilmadi")
)

type RefreshSessionRepository interface {
	Create(ctx context.Context, session *models.RefreshSession) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshSession, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
}

type refreshSessionRepository struct {
	db *gorm.DB
}

func NewRefreshSessionRepository(db *gorm.DB) RefreshSessionRepository {
	return &refreshSessionRepository{db: db}
}

func (r *refreshSessionRepository) Create(ctx context.Context, session *models.RefreshSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *refreshSessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshSession, error) {
	var session models.RefreshSession
	err := r.db.WithContext(ctx).Preload("User").Where("token_hash = ?", tokenHash).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	return &session, nil
}

func (r *refreshSessionRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&models.RefreshSession{}).
		Where("id = ?", id).
		Update("revoked_at", &now).Error
}

func (r *refreshSessionRepository) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&models.RefreshSession{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", &now).Error
}

func (r *refreshSessionRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.RefreshSession{}).
		Where("id = ?", id).
		Update("last_used_at", time.Now().UTC()).Error
}
