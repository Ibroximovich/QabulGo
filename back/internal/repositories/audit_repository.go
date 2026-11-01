package repositories

import (
	"context"

	"qabulgo-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditRepository interface {
	CreateAuditLog(ctx context.Context, log *models.AuditLog) error
	CreateVerificationReview(ctx context.Context, review *models.ClinicVerificationReview) error
	GetReviewsByClinic(ctx context.Context, clinicID uuid.UUID) ([]models.ClinicVerificationReview, error)
}

type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *auditRepository) CreateVerificationReview(ctx context.Context, review *models.ClinicVerificationReview) error {
	if review.ID == uuid.Nil {
		review.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(review).Error
}

func (r *auditRepository) GetReviewsByClinic(ctx context.Context, clinicID uuid.UUID) ([]models.ClinicVerificationReview, error) {
	var reviews []models.ClinicVerificationReview
	err := r.db.WithContext(ctx).
		Where("clinic_id = ?", clinicID).
		Order("reviewed_at DESC").
		Preload("Reviewer").
		Find(&reviews).Error
	return reviews, err
}

