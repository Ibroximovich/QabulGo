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
	ErrTokenNotFound = errors.New("tasdiqlash kodi topilmadi yoki eskirgan")
)

type VerificationTokenRepository interface {
	Create(ctx context.Context, token *models.VerificationToken) error
	GetLatestActive(ctx context.Context, userID uuid.UUID, purpose models.TokenPurpose) (*models.VerificationToken, error)
	IncrementAttempts(ctx context.Context, id uuid.UUID) error
	MarkConsumed(ctx context.Context, id uuid.UUID) error
	RevokeAllActiveByUserID(ctx context.Context, userID uuid.UUID, purpose models.TokenPurpose) error
}

type verificationTokenRepository struct {
	db *gorm.DB
}

func NewVerificationTokenRepository(db *gorm.DB) VerificationTokenRepository {
	return &verificationTokenRepository{db: db}
}

func (r *verificationTokenRepository) Create(ctx context.Context, token *models.VerificationToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *verificationTokenRepository) GetLatestActive(ctx context.Context, userID uuid.UUID, purpose models.TokenPurpose) (*models.VerificationToken, error) {
	var token models.VerificationToken
	now := time.Now().UTC()
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND purpose = ? AND consumed_at IS NULL AND expires_at > ?", userID, purpose, now).
		Order("created_at DESC").
		First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}
	return &token, nil
}

func (r *verificationTokenRepository) IncrementAttempts(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.VerificationToken{}).
		Where("id = ?", id).
		Update("attempts", gorm.Expr("attempts + 1")).Error
}

func (r *verificationTokenRepository) MarkConsumed(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&models.VerificationToken{}).
		Where("id = ?", id).
		Update("consumed_at", &now).Error
}

func (r *verificationTokenRepository) RevokeAllActiveByUserID(ctx context.Context, userID uuid.UUID, purpose models.TokenPurpose) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&models.VerificationToken{}).
		Where("user_id = ? AND purpose = ? AND consumed_at IS NULL", userID, purpose).
		Update("consumed_at", &now).Error
}
