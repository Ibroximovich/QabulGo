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
	ErrDoctorNotFound = errors.New("shifokor topilmadi")
)

type DoctorRepository interface {
	CreateDoctorWithClinic(ctx context.Context, clinicID uuid.UUID, doc *models.Doctor, specialtyID uuid.UUID, price float64, duration int, isActive bool) (*models.ClinicDoctor, error)
	GetClinicDoctor(ctx context.Context, clinicID, clinicDoctorID uuid.UUID) (*models.ClinicDoctor, error)
	UpdateClinicDoctor(ctx context.Context, cd *models.ClinicDoctor, doc *models.Doctor) error
	SoftDeleteClinicDoctor(ctx context.Context, clinicID, clinicDoctorID uuid.UUID) error
	ListClinicDoctors(ctx context.Context, clinicID uuid.UUID, filter *dto.DoctorFilter) ([]models.ClinicDoctor, int64, error)
	CountDoctorsByClinic(ctx context.Context, clinicID uuid.UUID) (int, error)
}

type doctorRepository struct {
	db *gorm.DB
}

func NewDoctorRepository(db *gorm.DB) DoctorRepository {
	return &doctorRepository{db: db}
}

func (r *doctorRepository) CreateDoctorWithClinic(
	ctx context.Context,
	clinicID uuid.UUID,
	doc *models.Doctor,
	specialtyID uuid.UUID,
	price float64,
	duration int,
	isActive bool,
) (*models.ClinicDoctor, error) {
	var clinicDoctor models.ClinicDoctor

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Create doctor
		if doc.ID == uuid.Nil {
			doc.ID = uuid.New()
		}
		if err := tx.Create(doc).Error; err != nil {
			return err
		}

		// 2. Create clinic_doctor association
		clinicDoctor = models.ClinicDoctor{
			ID:                         uuid.New(),
			ClinicID:                   clinicID,
			DoctorID:                   doc.ID,
			SpecialtyID:                specialtyID,
			ConsultationPrice:          price,
			AppointmentDurationMinutes: duration,
			IsActive:                   isActive,
		}
		if err := tx.Create(&clinicDoctor).Error; err != nil {
			return err
		}

		// Reload with relations
		return tx.Preload("Doctor").Preload("Specialty").First(&clinicDoctor, "id = ?", clinicDoctor.ID).Error
	})

	if err != nil {
		return nil, err
	}
	return &clinicDoctor, nil
}

func (r *doctorRepository) GetClinicDoctor(ctx context.Context, clinicID, clinicDoctorID uuid.UUID) (*models.ClinicDoctor, error) {
	var cd models.ClinicDoctor
	err := r.db.WithContext(ctx).
		Preload("Doctor").
		Preload("Specialty").
		First(&cd, "(id = ? OR doctor_id = ?) AND clinic_id = ?", clinicDoctorID, clinicDoctorID, clinicID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDoctorNotFound
		}
		return nil, err
	}
	return &cd, nil
}

func (r *doctorRepository) UpdateClinicDoctor(ctx context.Context, cd *models.ClinicDoctor, doc *models.Doctor) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if doc != nil {
			if err := tx.Save(doc).Error; err != nil {
				return err
			}
		}
		return tx.Save(cd).Error
	})
}

func (r *doctorRepository) SoftDeleteClinicDoctor(ctx context.Context, clinicID, clinicDoctorID uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Where("(id = ? OR doctor_id = ?) AND clinic_id = ?", clinicDoctorID, clinicDoctorID, clinicID).
		Delete(&models.ClinicDoctor{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrDoctorNotFound
	}
	return nil
}

func (r *doctorRepository) ListClinicDoctors(ctx context.Context, clinicID uuid.UUID, filter *dto.DoctorFilter) ([]models.ClinicDoctor, int64, error) {
	var items []models.ClinicDoctor
	var total int64

	q := r.db.WithContext(ctx).
		Model(&models.ClinicDoctor{}).
		Where("clinic_id = ?", clinicID)

	if filter != nil {
		if filter.SpecialtyID != nil {
			q = q.Where("specialty_id = ?", *filter.SpecialtyID)
		}
		if filter.IsActive != nil {
			q = q.Where("is_active = ?", *filter.IsActive)
		}
		if filter.Search != "" {
			searchTerm := "%" + filter.Search + "%"
			q = q.Joins("JOIN doctors ON doctors.id = clinic_doctors.doctor_id").
				Where("doctors.first_name ILIKE ? OR doctors.last_name ILIKE ?", searchTerm, searchTerm)
		}
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := 1
	limit := 20
	if filter != nil {
		if filter.Page > 0 {
			page = filter.Page
		}
		if filter.Limit > 0 && filter.Limit <= 100 {
			limit = filter.Limit
		}
	}
	offset := (page - 1) * limit

	err := q.Order("clinic_doctors.created_at DESC").
		Offset(offset).
		Limit(limit).
		Preload("Doctor").
		Preload("Specialty").
		Find(&items).Error

	return items, total, err
}

func (r *doctorRepository) CountDoctorsByClinic(ctx context.Context, clinicID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ClinicDoctor{}).
		Where("clinic_id = ?", clinicID).
		Count(&count).Error
	return int(count), err
}

