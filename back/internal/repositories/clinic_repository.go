package repositories

import (
	"context"
	"errors"

	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrClinicNotFound      = errors.New("klinika topilmadi")
	ErrClinicAlreadyExists = errors.New("siz allaqachon klinika yaratgansiz")
	ErrClinicAccessDenied  = errors.New("ushbu klinikani boshqarish huquqiga ega emassiz")
)

type ClinicRepository interface {
	Create(ctx context.Context, clinic *models.Clinic, workingHours []models.ClinicWorkingHour) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Clinic, error)
	GetByOwnerID(ctx context.Context, ownerID uuid.UUID) (*models.Clinic, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Clinic, error)
	Update(ctx context.Context, clinic *models.Clinic) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.ClinicStatus, rejectionReason string) error
	SetWorkingHours(ctx context.Context, clinicID uuid.UUID, workingHours []models.ClinicWorkingHour) error
	GetWorkingHours(ctx context.Context, clinicID uuid.UUID) ([]models.ClinicWorkingHour, error)
	IsUserClinicAdmin(ctx context.Context, clinicID, userID uuid.UUID) (bool, error)
	ListForAdmin(ctx context.Context, filter *dto.AdminClinicFilter) ([]models.Clinic, int64, error)
}

type clinicRepository struct {
	db *gorm.DB
}

func NewClinicRepository(db *gorm.DB) ClinicRepository {
	return &clinicRepository{db: db}
}

func (r *clinicRepository) Create(ctx context.Context, clinic *models.Clinic, workingHours []models.ClinicWorkingHour) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Create clinic
		if err := tx.Create(clinic).Error; err != nil {
			return err
		}

		// 2. Add owner to clinic_members
		member := &models.ClinicMember{
			ID:       uuid.New(),
			ClinicID: clinic.ID,
			UserID:   clinic.OwnerID,
			Role:     models.MemberRoleOwner,
		}
		if err := tx.Create(member).Error; err != nil {
			return err
		}

		// 3. Add working hours if provided
		if len(workingHours) > 0 {
			for i := range workingHours {
				workingHours[i].ClinicID = clinic.ID
				if workingHours[i].ID == uuid.Nil {
					workingHours[i].ID = uuid.New()
				}
			}
			if err := tx.Create(&workingHours).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *clinicRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Clinic, error) {
	var clinic models.Clinic
	err := r.db.WithContext(ctx).
		Preload("WorkingHours").
		Preload("Members.User").
		First(&clinic, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrClinicNotFound
		}
		return nil, err
	}
	return &clinic, nil
}

func (r *clinicRepository) GetByOwnerID(ctx context.Context, ownerID uuid.UUID) (*models.Clinic, error) {
	var clinic models.Clinic
	err := r.db.WithContext(ctx).
		Preload("WorkingHours").
		First(&clinic, "owner_id = ?", ownerID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrClinicNotFound
		}
		return nil, err
	}
	return &clinic, nil
}

func (r *clinicRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Clinic, error) {
	// First check clinic_members
	var member models.ClinicMember
	err := r.db.WithContext(ctx).
		First(&member, "user_id = ?", userID).Error
	if err == nil {
		return r.GetByID(ctx, member.ClinicID)
	}

	// Fallback to owner_id
	return r.GetByOwnerID(ctx, userID)
}

func (r *clinicRepository) Update(ctx context.Context, clinic *models.Clinic) error {
	return r.db.WithContext(ctx).Save(clinic).Error
}

func (r *clinicRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.ClinicStatus, rejectionReason string) error {
	updates := map[string]interface{}{
		"status":           status,
		"rejection_reason": rejectionReason,
		"updated_at":       gorm.Expr("CURRENT_TIMESTAMP"),
	}
	res := r.db.WithContext(ctx).Model(&models.Clinic{}).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrClinicNotFound
	}
	return nil
}

func (r *clinicRepository) SetWorkingHours(ctx context.Context, clinicID uuid.UUID, workingHours []models.ClinicWorkingHour) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete existing
		if err := tx.Where("clinic_id = ?", clinicID).Delete(&models.ClinicWorkingHour{}).Error; err != nil {
			return err
		}
		// Insert new
		if len(workingHours) > 0 {
			for i := range workingHours {
				workingHours[i].ClinicID = clinicID
				if workingHours[i].ID == uuid.Nil {
					workingHours[i].ID = uuid.New()
				}
			}
			if err := tx.Create(&workingHours).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *clinicRepository) GetWorkingHours(ctx context.Context, clinicID uuid.UUID) ([]models.ClinicWorkingHour, error) {
	var hours []models.ClinicWorkingHour
	err := r.db.WithContext(ctx).
		Where("clinic_id = ?", clinicID).
		Order("weekday ASC").
		Find(&hours).Error
	return hours, err
}

func (r *clinicRepository) IsUserClinicAdmin(ctx context.Context, clinicID, userID uuid.UUID) (bool, error) {
	var count int64
	// Check if owner or member
	err := r.db.WithContext(ctx).
		Model(&models.Clinic{}).
		Where("id = ? AND owner_id = ?", clinicID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	// Check clinic_members
	err = r.db.WithContext(ctx).
		Model(&models.ClinicMember{}).
		Where("clinic_id = ? AND user_id = ?", clinicID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *clinicRepository) ListForAdmin(ctx context.Context, filter *dto.AdminClinicFilter) ([]models.Clinic, int64, error) {
	var clinics []models.Clinic
	var total int64

	q := r.db.WithContext(ctx).Model(&models.Clinic{})

	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}

	if filter.Search != "" {
		searchTerm := "%" + filter.Search + "%"
		q = q.Where("name ILIKE ? OR phone ILIKE ? OR region ILIKE ? OR district ILIKE ?", searchTerm, searchTerm, searchTerm, searchTerm)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	err := q.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Preload("WorkingHours").
		Find(&clinics).Error

	return clinics, total, err
}

