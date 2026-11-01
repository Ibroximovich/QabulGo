package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"qabulgo-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrClinicDoctorNotFound  = errors.New("shifokor va klinika bog‘lanishi topilmadi")
	ErrDateOverrideNotFound  = errors.New("istisno sana yozuvi topilmadi")
	ErrScheduleSlotOverlap   = errors.New("qabul vaqtlari bir-birini qoplamasligi shart")
	ErrCrossClinicConflict   = errors.New("shifokor bu vaqtda boshqa klinikada band")
	ErrInvalidTimeRange      = errors.New("boshlanish vaqti tugash vaqtidan kichik bo‘lishi shart")
)

type ScheduleRepository interface {
	FindClinicDoctor(ctx context.Context, clinicID, doctorIDOrClinicDoctorID uuid.UUID) (*models.ClinicDoctor, error)
	GetWeeklySlots(ctx context.Context, clinicDoctorID uuid.UUID) ([]models.DoctorWeeklySlot, error)
	ReplaceWeeklySlots(ctx context.Context, clinicDoctorID uuid.UUID, slots []models.DoctorWeeklySlot) error
	CheckDoctorCrossClinicConflict(ctx context.Context, doctorID, currentClinicDoctorID uuid.UUID, weekday int, startTime, endTime string) (*models.DoctorWeeklySlot, *models.Clinic, error)
	GetDateOverrides(ctx context.Context, clinicDoctorID uuid.UUID, date string, startDate, endDate string) ([]models.DoctorDateSlotOverride, error)
	GetDateOverrideByID(ctx context.Context, overrideID uuid.UUID) (*models.DoctorDateSlotOverride, error)
	CreateDateOverride(ctx context.Context, override *models.DoctorDateSlotOverride) error
	UpdateDateOverride(ctx context.Context, override *models.DoctorDateSlotOverride) error
	DeleteDateOverride(ctx context.Context, overrideID uuid.UUID) error
	GetAvailabilityData(ctx context.Context, clinicID, doctorIDOrCDID uuid.UUID, dateStr string, weekday int) (*models.Clinic, *models.ClinicDoctor, []models.DoctorWeeklySlot, []models.DoctorDateSlotOverride, error)
}

type scheduleRepository struct {
	db *gorm.DB
}

func NewScheduleRepository(db *gorm.DB) ScheduleRepository {
	return &scheduleRepository{db: db}
}

// FindClinicDoctor finds clinic_doctor by id or doctor_id for this clinic
func (r *scheduleRepository) FindClinicDoctor(ctx context.Context, clinicID, doctorIDOrCDID uuid.UUID) (*models.ClinicDoctor, error) {
	var cd models.ClinicDoctor

	// 1. First attempt: match by primary key id
	err := r.db.WithContext(ctx).
		Preload("Doctor").
		Preload("Clinic").
		Preload("Specialty").
		Where("clinic_id = ? AND id = ?", clinicID, doctorIDOrCDID).
		First(&cd).Error
	if err == nil {
		return &cd, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 2. Second attempt: match by doctor_id
	err = r.db.WithContext(ctx).
		Preload("Doctor").
		Preload("Clinic").
		Preload("Specialty").
		Where("clinic_id = ? AND doctor_id = ?", clinicID, doctorIDOrCDID).
		First(&cd).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrClinicDoctorNotFound
		}
		return nil, err
	}

	return &cd, nil
}

// GetWeeklySlots returns all active weekly slots for a clinic doctor
func (r *scheduleRepository) GetWeeklySlots(ctx context.Context, clinicDoctorID uuid.UUID) ([]models.DoctorWeeklySlot, error) {
	var slots []models.DoctorWeeklySlot
	err := r.db.WithContext(ctx).
		Where("clinic_doctor_id = ?", clinicDoctorID).
		Order("weekday ASC, start_time ASC").
		Find(&slots).Error
	return slots, err
}

// ReplaceWeeklySlots atomically replaces weekly slots in a database transaction
func (r *scheduleRepository) ReplaceWeeklySlots(ctx context.Context, clinicDoctorID uuid.UUID, slots []models.DoctorWeeklySlot) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Remove existing slots
		if err := tx.Where("clinic_doctor_id = ?", clinicDoctorID).Delete(&models.DoctorWeeklySlot{}).Error; err != nil {
			return fmt.Errorf("failed to delete old slots: %w", err)
		}

		if len(slots) == 0 {
			return nil
		}

		now := time.Now().UTC()
		for i := range slots {
			slots[i].ID = uuid.New()
			slots[i].ClinicDoctorID = clinicDoctorID
			if slots[i].Status == "" {
				slots[i].Status = models.SlotStatusActive
			}
			slots[i].CreatedAt = now
			slots[i].UpdatedAt = now
		}

		if err := tx.Create(&slots).Error; err != nil {
			return fmt.Errorf("failed to insert new slots: %w", err)
		}

		return nil
	})
}

// CheckDoctorCrossClinicConflict verifies if a doctor has an overlapping weekly slot in ANOTHER clinic
func (r *scheduleRepository) CheckDoctorCrossClinicConflict(
	ctx context.Context,
	doctorID, currentClinicDoctorID uuid.UUID,
	weekday int,
	startTime, endTime string,
) (*models.DoctorWeeklySlot, *models.Clinic, error) {
	var slot models.DoctorWeeklySlot

	// Find any other clinic_doctors for this same doctor
	err := r.db.WithContext(ctx).
		Table("doctor_weekly_slots").
		Joins("JOIN clinic_doctors ON clinic_doctors.id = doctor_weekly_slots.clinic_doctor_id").
		Where("clinic_doctors.doctor_id = ? AND clinic_doctors.id != ?", doctorID, currentClinicDoctorID).
		Where("clinic_doctors.deleted_at IS NULL AND clinic_doctors.is_active = true").
		Where("doctor_weekly_slots.weekday = ? AND doctor_weekly_slots.status = 'active'", weekday).
		Where("doctor_weekly_slots.start_time < ? AND doctor_weekly_slots.end_time > ?", endTime, startTime).
		Preload("ClinicDoctor.Clinic").
		First(&slot).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil // No conflict
		}
		return nil, nil, err
	}

	var clinic models.Clinic
	if slot.ClinicDoctor != nil && slot.ClinicDoctor.Clinic != nil {
		clinic = *slot.ClinicDoctor.Clinic
	} else {
		// fetch clinic directly
		var otherCD models.ClinicDoctor
		if err := r.db.WithContext(ctx).Preload("Clinic").First(&otherCD, slot.ClinicDoctorID).Error; err == nil && otherCD.Clinic != nil {
			clinic = *otherCD.Clinic
		}
	}

	return &slot, &clinic, ErrCrossClinicConflict
}

// GetDateOverrides retrieves exceptions for a doctor on a specific date or date range
func (r *scheduleRepository) GetDateOverrides(ctx context.Context, clinicDoctorID uuid.UUID, date string, startDate, endDate string) ([]models.DoctorDateSlotOverride, error) {
	query := r.db.WithContext(ctx).Where("clinic_doctor_id = ?", clinicDoctorID)

	if date != "" {
		query = query.Where("date = ?", date)
	}
	if startDate != "" {
		query = query.Where("date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("date <= ?", endDate)
	}

	var overrides []models.DoctorDateSlotOverride
	err := query.Order("date ASC, start_time ASC").Find(&overrides).Error
	return overrides, err
}

// GetDateOverrideByID finds an override by ID
func (r *scheduleRepository) GetDateOverrideByID(ctx context.Context, overrideID uuid.UUID) (*models.DoctorDateSlotOverride, error) {
	var override models.DoctorDateSlotOverride
	err := r.db.WithContext(ctx).First(&override, "id = ?", overrideID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDateOverrideNotFound
		}
		return nil, err
	}
	return &override, nil
}

// CreateDateOverride inserts a date override
func (r *scheduleRepository) CreateDateOverride(ctx context.Context, override *models.DoctorDateSlotOverride) error {
	override.ID = uuid.New()
	override.CreatedAt = time.Now().UTC()
	override.UpdatedAt = override.CreatedAt
	return r.db.WithContext(ctx).Create(override).Error
}

// UpdateDateOverride updates an existing date override
func (r *scheduleRepository) UpdateDateOverride(ctx context.Context, override *models.DoctorDateSlotOverride) error {
	override.UpdatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).Save(override).Error
}

// DeleteDateOverride removes a date override
func (r *scheduleRepository) DeleteDateOverride(ctx context.Context, overrideID uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&models.DoctorDateSlotOverride{}, "id = ?", overrideID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrDateOverrideNotFound
	}
	return nil
}

// GetAvailabilityData retrieves all entities needed to compute public availability for a date
func (r *scheduleRepository) GetAvailabilityData(
	ctx context.Context,
	clinicID, doctorIDOrCDID uuid.UUID,
	dateStr string,
	weekday int,
) (*models.Clinic, *models.ClinicDoctor, []models.DoctorWeeklySlot, []models.DoctorDateSlotOverride, error) {
	// 1. Fetch clinic
	var clinic models.Clinic
	if err := r.db.WithContext(ctx).First(&clinic, "id = ?", clinicID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, nil, ErrClinicNotFound
		}
		return nil, nil, nil, nil, err
	}

	// 2. Fetch active clinic_doctor
	cd, err := r.FindClinicDoctor(ctx, clinicID, doctorIDOrCDID)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// 3. Fetch weekly slots for this weekday
	var weeklySlots []models.DoctorWeeklySlot
	if err := r.db.WithContext(ctx).
		Where("clinic_doctor_id = ? AND weekday = ? AND status = 'active'", cd.ID, weekday).
		Order("start_time ASC").
		Find(&weeklySlots).Error; err != nil {
		return nil, nil, nil, nil, err
	}

	// 4. Fetch date overrides for this date
	var dateOverrides []models.DoctorDateSlotOverride
	if err := r.db.WithContext(ctx).
		Where("clinic_doctor_id = ? AND date = ?", cd.ID, dateStr).
		Order("start_time ASC").
		Find(&dateOverrides).Error; err != nil {
		return nil, nil, nil, nil, err
	}

	return &clinic, cd, weeklySlots, dateOverrides, nil
}

