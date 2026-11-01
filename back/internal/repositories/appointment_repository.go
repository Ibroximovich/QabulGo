package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrAppointmentNotFound    = errors.New("qabul yozuvi topilmadi")
	ErrDoubleBookingConflict  = errors.New("bu qabul vaqti allaqachon band qilingan. Iltimos, boshqa vaqtni tanlang")
	ErrIdempotencyConflict    = errors.New("idempotency-key boshqa bron so‘rovi uchun allaqachon ishlatilgan")
	ErrAppointmentNotEditable = errors.New("ushbu holatdagi qabulni o‘zgartirib bo‘lmaydi")
)

type AppointmentRepository interface {
	Create(ctx context.Context, appt *models.Appointment) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Appointment, error)
	FindByIdempotencyKey(ctx context.Context, patientID uuid.UUID, key string) (*models.Appointment, error)
	CheckDoctorBookingOverlap(ctx context.Context, doctorID uuid.UUID, date, startTime, endTime string) (*models.Appointment, error)
	ListForPatient(ctx context.Context, patientID uuid.UUID, filter *dto.PatientAppointmentFilter, todayDate, nowTime string) ([]models.Appointment, int64, error)
	ListForClinic(ctx context.Context, clinicID uuid.UUID, filter *dto.ClinicAppointmentFilter) ([]models.Appointment, int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, fromStatus, toStatus models.AppointmentStatus, reason *string, timestampField string, timestamp time.Time) error
	GetBookedSlotsForDoctor(ctx context.Context, doctorID uuid.UUID, date string) ([]models.Appointment, error)
	FindFutureConfirmedAppointmentsForDoctor(ctx context.Context, clinicDoctorID uuid.UUID, fromDate string) ([]models.Appointment, error)
	FindFutureConfirmedAppointmentsOnDate(ctx context.Context, clinicDoctorID uuid.UUID, date string, startTime, endTime *string) ([]models.Appointment, error)
}

type appointmentRepository struct {
	db *gorm.DB
}

func NewAppointmentRepository(db *gorm.DB) AppointmentRepository {
	return &appointmentRepository{db: db}
}

// Create creates an appointment inside an isolated transaction with row locking
func (r *appointmentRepository) Create(ctx context.Context, appt *models.Appointment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Lock doctor row for update to serialize concurrent booking attempts for this doctor
		var lockedDoctorID string
		err := tx.Raw("SELECT id::text FROM doctors WHERE id = ? FOR UPDATE", appt.DoctorID).Scan(&lockedDoctorID).Error
		if err != nil {
			return fmt.Errorf("shifokor qatorini bloklashda xatolik: %w", err)
		}

		// 2. Check if doctor already has an active overlapping booking on that date
		var existing models.Appointment
		err = tx.Where("doctor_id = ? AND appointment_date = ?", appt.DoctorID, appt.AppointmentDate).
			Where("status IN ('confirmed', 'completed')").
			Where("start_time < ? AND end_time > ?", appt.EndTime, appt.StartTime).
			First(&existing).Error
		if err == nil {
			return ErrDoubleBookingConflict
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("bandlikni tekshirishda xatolik: %w", err)
		}

		// 3. Check if patient already has an active booking at the same date and start_time
		var patientBooking models.Appointment
		err = tx.Where("patient_id = ? AND appointment_date = ? AND start_time = ?", appt.PatientID, appt.AppointmentDate, appt.StartTime).
			Where("status IN ('confirmed', 'completed')").
			First(&patientBooking).Error
		if err == nil {
			return fmt.Errorf("siz allaqachon ushbu vaqtda boshqa qabulga yozilgansiz")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("bemor bandligini tekshirishda xatolik: %w", err)
		}

		// 4. Insert appointment (DB partial unique index also provides hard guarantee)
		if appt.ID == uuid.Nil {
			appt.ID = uuid.New()
		}
		if err := tx.Create(appt).Error; err != nil {
			if strings.Contains(err.Error(), "uq_doctor_active_slot") || strings.Contains(err.Error(), "duplicate key") {
				return ErrDoubleBookingConflict
			}
			return fmt.Errorf("qabulni saqlashda xatolik: %w", err)
		}

		return nil
	})
}

// GetByID retrieves appointment with full preloads
func (r *appointmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Appointment, error) {
	var appt models.Appointment
	err := r.db.WithContext(ctx).
		Preload("Patient").
		Preload("Clinic").
		Preload("Doctor").
		Preload("ClinicDoctor.Specialty").
		First(&appt, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAppointmentNotFound
		}
		return nil, err
	}
	return &appt, nil
}

// FindByIdempotencyKey finds an appointment by patient and idempotency key
func (r *appointmentRepository) FindByIdempotencyKey(ctx context.Context, patientID uuid.UUID, key string) (*models.Appointment, error) {
	var appt models.Appointment
	err := r.db.WithContext(ctx).
		Preload("Clinic").
		Preload("Doctor").
		Preload("ClinicDoctor.Specialty").
		Where("patient_id = ? AND idempotency_key = ?", patientID, key).
		First(&appt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // not found
		}
		return nil, err
	}
	return &appt, nil
}

// CheckDoctorBookingOverlap checks if doctor has active overlapping bookings on that date
func (r *appointmentRepository) CheckDoctorBookingOverlap(ctx context.Context, doctorID uuid.UUID, date, startTime, endTime string) (*models.Appointment, error) {
	var appt models.Appointment
	err := r.db.WithContext(ctx).
		Where("doctor_id = ? AND appointment_date = ?", doctorID, date).
		Where("status IN ('confirmed', 'completed')").
		Where("start_time < ? AND end_time > ?", endTime, startTime).
		First(&appt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No overlap
		}
		return nil, err
	}
	return &appt, nil
}

// ListForPatient lists patient's appointments with filters and pagination
func (r *appointmentRepository) ListForPatient(
	ctx context.Context,
	patientID uuid.UUID,
	filter *dto.PatientAppointmentFilter,
	todayDate, nowTime string,
) ([]models.Appointment, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Appointment{}).
		Where("patient_id = ?", patientID)

	if filter != nil {
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.Timeframe != "" {
			switch filter.Timeframe {
			case "upcoming":
				query = query.Where("status = ?", models.AppointmentStatusConfirmed).
					Where("(appointment_date > ? OR (appointment_date = ? AND end_time > ?))", todayDate, todayDate, nowTime)
			case "past":
				query = query.Where("(status IN ('completed', 'no_show') OR (appointment_date < ? OR (appointment_date = ? AND end_time <= ?)))", todayDate, todayDate, nowTime)
			case "cancelled":
				query = query.Where("status = ?", models.AppointmentStatusCancelled)
			}
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := 1
	limit := 10
	if filter != nil {
		if filter.Page > 0 {
			page = filter.Page
		}
		if filter.Limit > 0 && filter.Limit <= 100 {
			limit = filter.Limit
		}
	}
	offset := (page - 1) * limit

	var list []models.Appointment
	err := query.
		Preload("Clinic").
		Preload("Doctor").
		Preload("ClinicDoctor.Specialty").
		Order("appointment_date DESC, start_time DESC").
		Offset(offset).
		Limit(limit).
		Find(&list).Error
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// ListForClinic lists clinic's appointments with filters and pagination
func (r *appointmentRepository) ListForClinic(
	ctx context.Context,
	clinicID uuid.UUID,
	filter *dto.ClinicAppointmentFilter,
) ([]models.Appointment, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Appointment{}).
		Where("clinic_id = ?", clinicID)

	if filter != nil {
		if filter.Date != "" {
			query = query.Where("appointment_date = ?", filter.Date)
		}
		if filter.DateFrom != "" {
			query = query.Where("appointment_date >= ?", filter.DateFrom)
		}
		if filter.DateTo != "" {
			query = query.Where("appointment_date <= ?", filter.DateTo)
		}
		if filter.DoctorID != nil && *filter.DoctorID != uuid.Nil {
			query = query.Where("doctor_id = ?", *filter.DoctorID)
		}
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
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

	var list []models.Appointment
	err := query.
		Preload("Patient").
		Preload("Doctor").
		Preload("ClinicDoctor.Specialty").
		Order("appointment_date DESC, start_time DESC").
		Offset(offset).
		Limit(limit).
		Find(&list).Error
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// UpdateStatus atomically updates appointment status with validation
func (r *appointmentRepository) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	fromStatus, toStatus models.AppointmentStatus,
	reason *string,
	timestampField string,
	timestamp time.Time,
) error {
	updates := map[string]interface{}{
		"status":     toStatus,
		"updated_at": time.Now(),
	}
	if reason != nil {
		updates["cancellation_reason"] = *reason
	}
	if timestampField != "" {
		updates[timestampField] = timestamp
	}

	res := r.db.WithContext(ctx).Model(&models.Appointment{}).
		Where("id = ? AND status = ?", id, fromStatus).
		Updates(updates)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrAppointmentNotEditable
	}
	return nil
}

// GetBookedSlotsForDoctor gets all confirmed/completed appointments for doctor on date
func (r *appointmentRepository) GetBookedSlotsForDoctor(ctx context.Context, doctorID uuid.UUID, date string) ([]models.Appointment, error) {
	var list []models.Appointment
	err := r.db.WithContext(ctx).
		Where("doctor_id = ? AND appointment_date = ?", doctorID, date).
		Where("status IN ('confirmed', 'completed')").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// FindFutureConfirmedAppointmentsForDoctor gets all future confirmed bookings for a clinic_doctor
func (r *appointmentRepository) FindFutureConfirmedAppointmentsForDoctor(
	ctx context.Context,
	clinicDoctorID uuid.UUID,
	fromDate string,
) ([]models.Appointment, error) {
	var list []models.Appointment
	err := r.db.WithContext(ctx).
		Preload("Patient").
		Where("clinic_doctor_id = ? AND status = 'confirmed'", clinicDoctorID).
		Where("appointment_date >= ?", fromDate).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// FindFutureConfirmedAppointmentsOnDate gets confirmed bookings for a clinic_doctor on a specific date
func (r *appointmentRepository) FindFutureConfirmedAppointmentsOnDate(
	ctx context.Context,
	clinicDoctorID uuid.UUID,
	date string,
	startTime, endTime *string,
) ([]models.Appointment, error) {
	query := r.db.WithContext(ctx).
		Preload("Patient").
		Where("clinic_doctor_id = ? AND status = 'confirmed' AND appointment_date = ?", clinicDoctorID, date)

	if startTime != nil && endTime != nil {
		query = query.Where("start_time < ? AND end_time > ?", *endTime, *startTime)
	}

	var list []models.Appointment
	if err := query.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
