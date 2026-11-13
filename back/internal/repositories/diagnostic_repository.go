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
	ErrDiagnosticCategoryNotFound  = errors.New("diagnostika kategoriyasi topilmadi")
	ErrDiagnosticEquipmentNotFound = errors.New("diagnostika uskunasi topilmadi")
	ErrDiagnosticServiceNotFound   = errors.New("diagnostika xizmati topilmadi")
	ErrDiagnosticAppointmentNotFound = errors.New("diagnostika qabuli topilmadi")
)

type DiagnosticRepository interface {
	// Categories
	ListCategories(ctx context.Context) ([]models.DiagnosticCategory, error)
	GetCategoryByID(ctx context.Context, id uuid.UUID) (*models.DiagnosticCategory, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*models.DiagnosticCategory, error)

	// Equipments
	CreateEquipment(ctx context.Context, eq *models.DiagnosticEquipment) error
	UpdateEquipment(ctx context.Context, eq *models.DiagnosticEquipment) error
	GetEquipmentByID(ctx context.Context, id uuid.UUID) (*models.DiagnosticEquipment, error)
	ListEquipmentsByClinic(ctx context.Context, clinicID uuid.UUID) ([]models.DiagnosticEquipment, error)
	DeleteEquipment(ctx context.Context, id uuid.UUID) error

	// Services
	CreateService(ctx context.Context, svc *models.DiagnosticService) error
	UpdateService(ctx context.Context, svc *models.DiagnosticService) error
	GetServiceByID(ctx context.Context, id uuid.UUID) (*models.DiagnosticService, error)
	ListPublicServices(ctx context.Context, clinicID *uuid.UUID, categoryID *uuid.UUID) ([]models.DiagnosticService, error)
	ListServicesByClinic(ctx context.Context, clinicID uuid.UUID, categoryID *uuid.UUID) ([]models.DiagnosticService, error)
	ListServicesByEquipment(ctx context.Context, equipmentID uuid.UUID) ([]models.DiagnosticService, error)
	DeleteService(ctx context.Context, id uuid.UUID) error

	// Schedule & Slots
	GetWeeklySlots(ctx context.Context, equipmentID uuid.UUID) ([]models.DiagnosticWeeklySlot, error)
	SaveWeeklySlots(ctx context.Context, equipmentID uuid.UUID, slots []models.DiagnosticWeeklySlot) error
	GetDateOverrides(ctx context.Context, equipmentID uuid.UUID, date, startDate, endDate string) ([]models.DiagnosticDateSlotOverride, error)
	CreateDateOverride(ctx context.Context, override *models.DiagnosticDateSlotOverride) error
	UpdateDateOverride(ctx context.Context, override *models.DiagnosticDateSlotOverride) error
	DeleteDateOverride(ctx context.Context, equipmentID, overrideID uuid.UUID) error
	GetEquipmentAvailabilityData(ctx context.Context, equipmentID uuid.UUID, dateStr string, weekday int) (*models.DiagnosticEquipment, []models.DiagnosticWeeklySlot, []models.DiagnosticDateSlotOverride, error)

	// Appointments
	CreateAppointment(ctx context.Context, appt *models.DiagnosticAppointment) error
	GetAppointmentByID(ctx context.Context, id uuid.UUID) (*models.DiagnosticAppointment, error)
	FindByIdempotencyKey(ctx context.Context, patientID uuid.UUID, key string) (*models.DiagnosticAppointment, error)
	CheckEquipmentBookingOverlap(ctx context.Context, equipmentID uuid.UUID, date, startTime, endTime string) (*models.DiagnosticAppointment, error)
	ListForPatient(ctx context.Context, patientID uuid.UUID, filter *dto.PatientDiagnosticAppointmentFilter, todayDate, nowTime string) ([]models.DiagnosticAppointment, int64, error)
	ListForClinic(ctx context.Context, clinicID uuid.UUID, filter *dto.ClinicDiagnosticAppointmentFilter) ([]models.DiagnosticAppointment, int64, error)
	UpdateAppointmentStatus(ctx context.Context, id uuid.UUID, fromStatus, toStatus models.DiagnosticAppointmentStatus, reason *string, timestampField string, timestamp time.Time) error
	GetBookedSlotsForEquipment(ctx context.Context, equipmentID uuid.UUID, date string) ([]models.DiagnosticAppointment, error)
}

type diagnosticRepository struct {
	db *gorm.DB
}

func NewDiagnosticRepository(db *gorm.DB) DiagnosticRepository {
	return &diagnosticRepository{db: db}
}

// --- Categories ---

func (r *diagnosticRepository) ListCategories(ctx context.Context) ([]models.DiagnosticCategory, error) {
	var categories []models.DiagnosticCategory
	err := r.db.WithContext(ctx).Order("name ASC").Find(&categories).Error
	return categories, err
}

func (r *diagnosticRepository) GetCategoryByID(ctx context.Context, id uuid.UUID) (*models.DiagnosticCategory, error) {
	var cat models.DiagnosticCategory
	err := r.db.WithContext(ctx).First(&cat, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDiagnosticCategoryNotFound
		}
		return nil, err
	}
	return &cat, nil
}

func (r *diagnosticRepository) GetCategoryBySlug(ctx context.Context, slug string) (*models.DiagnosticCategory, error) {
	var cat models.DiagnosticCategory
	err := r.db.WithContext(ctx).First(&cat, "slug = ?", slug).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDiagnosticCategoryNotFound
		}
		return nil, err
	}
	return &cat, nil
}

// --- Equipments ---

func (r *diagnosticRepository) CreateEquipment(ctx context.Context, eq *models.DiagnosticEquipment) error {
	return r.db.WithContext(ctx).Create(eq).Error
}

func (r *diagnosticRepository) UpdateEquipment(ctx context.Context, eq *models.DiagnosticEquipment) error {
	return r.db.WithContext(ctx).Save(eq).Error
}

func (r *diagnosticRepository) GetEquipmentByID(ctx context.Context, id uuid.UUID) (*models.DiagnosticEquipment, error) {
	var eq models.DiagnosticEquipment
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Clinic").
		First(&eq, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDiagnosticEquipmentNotFound
		}
		return nil, err
	}
	return &eq, nil
}

func (r *diagnosticRepository) ListEquipmentsByClinic(ctx context.Context, clinicID uuid.UUID) ([]models.DiagnosticEquipment, error) {
	var list []models.DiagnosticEquipment
	err := r.db.WithContext(ctx).
		Preload("Category").
		Where("clinic_id = ? AND is_active = true", clinicID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

func (r *diagnosticRepository) DeleteEquipment(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Faol (tasdiqlangan) qabullar borligini tekshirish
		var activeCount int64
		if err := tx.Model(&models.DiagnosticAppointment{}).
			Where("equipment_id = ? AND status = ?", id, models.DiagAppointmentStatusConfirmed).
			Count(&activeCount).Error; err != nil {
			return err
		}
		if activeCount > 0 {
			return errors.New("ushbu uskunada tasdiqlangan faol qabullar mavjud. Avval ularni yakunlang yoki bekor qiling")
		}

		// 2. Biriktirilgan xizmatlarda faol qabullar borligini tekshirish
		var serviceActiveCount int64
		if err := tx.Model(&models.DiagnosticAppointment{}).
			Joins("JOIN diagnostic_services ON diagnostic_services.id = diagnostic_appointments.service_id").
			Where("diagnostic_services.equipment_id = ? AND diagnostic_appointments.status = ?", id, models.DiagAppointmentStatusConfirmed).
			Count(&serviceActiveCount).Error; err != nil {
			return err
		}
		if serviceActiveCount > 0 {
			return errors.New("ushbu uskunaga biriktirilgan xizmatlarda faol qabullar mavjud. Avval ularni yakunlang yoki bekor qiling")
		}

		// 3. Tarixiy qabullar (yakunlangan/bekor qilingan) mavjudligini tekshirish
		var totalAppts int64
		if err := tx.Model(&models.DiagnosticAppointment{}).
			Where("equipment_id = ?", id).
			Count(&totalAppts).Error; err != nil {
			return err
		}
		var totalServiceAppts int64
		if err := tx.Model(&models.DiagnosticAppointment{}).
			Joins("JOIN diagnostic_services ON diagnostic_services.id = diagnostic_appointments.service_id").
			Where("diagnostic_services.equipment_id = ?", id).
			Count(&totalServiceAppts).Error; err != nil {
			return err
		}

		// Agar tarixiy qabullar bo'lsa, xorijiy kalit (FK) yaxlitligini buzmaslik uchun xavfsiz soft-delete (nofaol) qilamiz
		if totalAppts > 0 || totalServiceAppts > 0 {
			if err := tx.Model(&models.DiagnosticEquipment{}).Where("id = ?", id).Update("is_active", false).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.DiagnosticService{}).Where("equipment_id = ?", id).Update("is_active", false).Error; err != nil {
				return err
			}
			if err := tx.Where("equipment_id = ?", id).Delete(&models.DiagnosticWeeklySlot{}).Error; err != nil {
				return err
			}
			if err := tx.Where("equipment_id = ?", id).Delete(&models.DiagnosticDateSlotOverride{}).Error; err != nil {
				return err
			}
			return nil
		}

		// Agar hech qanday qabul bo'lmasa, to'liq kaskadli o'chirish
		if err := tx.Where("equipment_id = ?", id).Delete(&models.DiagnosticDateSlotOverride{}).Error; err != nil {
			return err
		}
		if err := tx.Where("equipment_id = ?", id).Delete(&models.DiagnosticWeeklySlot{}).Error; err != nil {
			return err
		}
		if err := tx.Where("equipment_id = ?", id).Delete(&models.DiagnosticService{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", id).Delete(&models.DiagnosticEquipment{}).Error; err != nil {
			return err
		}
		return nil
	})
}

// --- Services ---

func (r *diagnosticRepository) CreateService(ctx context.Context, svc *models.DiagnosticService) error {
	return r.db.WithContext(ctx).Create(svc).Error
}

func (r *diagnosticRepository) UpdateService(ctx context.Context, svc *models.DiagnosticService) error {
	return r.db.WithContext(ctx).Save(svc).Error
}

func (r *diagnosticRepository) GetServiceByID(ctx context.Context, id uuid.UUID) (*models.DiagnosticService, error) {
	var svc models.DiagnosticService
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Equipment").
		Preload("Clinic").
		First(&svc, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDiagnosticServiceNotFound
		}
		return nil, err
	}
	return &svc, nil
}

func (r *diagnosticRepository) ListPublicServices(ctx context.Context, clinicID *uuid.UUID, categoryID *uuid.UUID) ([]models.DiagnosticService, error) {
	query := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Equipment").
		Preload("Clinic").
		Joins("JOIN clinics ON clinics.id = diagnostic_services.clinic_id").
		Joins("JOIN diagnostic_equipments ON diagnostic_equipments.id = diagnostic_services.equipment_id").
		Where("clinics.status = ?", "approved").
		Where("diagnostic_equipments.is_active = ?", true).
		Where("diagnostic_services.is_active = ?", true)

	if clinicID != nil {
		query = query.Where("diagnostic_services.clinic_id = ?", *clinicID)
	}

	if categoryID != nil {
		query = query.Where("diagnostic_services.category_id = ?", *categoryID)
	}

	var list []models.DiagnosticService
	err := query.Order("diagnostic_services.name ASC").Find(&list).Error
	return list, err
}

func (r *diagnosticRepository) ListServicesByClinic(ctx context.Context, clinicID uuid.UUID, categoryID *uuid.UUID) ([]models.DiagnosticService, error) {
	query := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Equipment").
		Where("clinic_id = ? AND is_active = true", clinicID)

	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}

	var list []models.DiagnosticService
	err := query.Order("name ASC").Find(&list).Error
	return list, err
}

func (r *diagnosticRepository) ListServicesByEquipment(ctx context.Context, equipmentID uuid.UUID) ([]models.DiagnosticService, error) {
	var list []models.DiagnosticService
	err := r.db.WithContext(ctx).
		Preload("Category").
		Where("equipment_id = ? AND is_active = true", equipmentID).
		Order("name ASC").
		Find(&list).Error
	return list, err
}

func (r *diagnosticRepository) DeleteService(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Faol (tasdiqlangan) qabullar borligini tekshirish
		var activeCount int64
		if err := tx.Model(&models.DiagnosticAppointment{}).
			Where("service_id = ? AND status = ?", id, models.DiagAppointmentStatusConfirmed).
			Count(&activeCount).Error; err != nil {
			return err
		}
		if activeCount > 0 {
			return errors.New("ushbu xizmatda tasdiqlangan faol qabullar mavjud. Avval ularni yakunlang yoki bekor qiling")
		}

		// 2. Tarixiy qabullar borligini tekshirish
		var totalAppts int64
		if err := tx.Model(&models.DiagnosticAppointment{}).
			Where("service_id = ?", id).
			Count(&totalAppts).Error; err != nil {
			return err
		}

		if totalAppts > 0 {
			// Soft delete (nofaol qilish)
			return tx.Model(&models.DiagnosticService{}).Where("id = ?", id).Update("is_active", false).Error
		}

		// Qabul bo'lmagan xizmatni to'liq o'chirish
		return tx.Where("id = ?", id).Delete(&models.DiagnosticService{}).Error
	})
}

// --- Schedule & Slots ---

func (r *diagnosticRepository) GetWeeklySlots(ctx context.Context, equipmentID uuid.UUID) ([]models.DiagnosticWeeklySlot, error) {
	var slots []models.DiagnosticWeeklySlot
	err := r.db.WithContext(ctx).
		Where("equipment_id = ?", equipmentID).
		Order("weekday ASC, start_time ASC").
		Find(&slots).Error
	return slots, err
}

func (r *diagnosticRepository) SaveWeeklySlots(ctx context.Context, equipmentID uuid.UUID, slots []models.DiagnosticWeeklySlot) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete existing slots
		if err := tx.Where("equipment_id = ?", equipmentID).Delete(&models.DiagnosticWeeklySlot{}).Error; err != nil {
			return err
		}
		if len(slots) > 0 {
			for i := range slots {
				if slots[i].ID == uuid.Nil {
					slots[i].ID = uuid.New()
				}
			}
			if err := tx.Create(&slots).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *diagnosticRepository) GetDateOverrides(ctx context.Context, equipmentID uuid.UUID, date, startDate, endDate string) ([]models.DiagnosticDateSlotOverride, error) {
	query := r.db.WithContext(ctx).Where("equipment_id = ?", equipmentID)
	if date != "" {
		query = query.Where("date = ?", date)
	}
	if startDate != "" && endDate != "" {
		query = query.Where("date >= ? AND date <= ?", startDate, endDate)
	}
	var overrides []models.DiagnosticDateSlotOverride
	err := query.Order("date ASC, start_time ASC").Find(&overrides).Error
	return overrides, err
}

func (r *diagnosticRepository) CreateDateOverride(ctx context.Context, override *models.DiagnosticDateSlotOverride) error {
	return r.db.WithContext(ctx).Create(override).Error
}

func (r *diagnosticRepository) UpdateDateOverride(ctx context.Context, override *models.DiagnosticDateSlotOverride) error {
	return r.db.WithContext(ctx).Save(override).Error
}

func (r *diagnosticRepository) DeleteDateOverride(ctx context.Context, equipmentID, overrideID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND equipment_id = ?", overrideID, equipmentID).
		Delete(&models.DiagnosticDateSlotOverride{}).Error
}

func (r *diagnosticRepository) GetEquipmentAvailabilityData(
	ctx context.Context,
	equipmentID uuid.UUID,
	dateStr string,
	weekday int,
) (*models.DiagnosticEquipment, []models.DiagnosticWeeklySlot, []models.DiagnosticDateSlotOverride, error) {
	var eq models.DiagnosticEquipment
	err := r.db.WithContext(ctx).
		Preload("Clinic").
		First(&eq, "id = ?", equipmentID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, ErrDiagnosticEquipmentNotFound
		}
		return nil, nil, nil, err
	}

	var weeklySlots []models.DiagnosticWeeklySlot
	err = r.db.WithContext(ctx).
		Where("equipment_id = ? AND weekday = ? AND (LOWER(status) = 'active' OR status = '' OR status IS NULL)", equipmentID, weekday).
		Order("start_time ASC").
		Find(&weeklySlots).Error
	if err != nil {
		return nil, nil, nil, err
	}

	var overrides []models.DiagnosticDateSlotOverride
	err = r.db.WithContext(ctx).
		Where("equipment_id = ? AND date = ?", equipmentID, dateStr).
		Find(&overrides).Error
	if err != nil {
		return nil, nil, nil, err
	}

	return &eq, weeklySlots, overrides, nil
}

// --- Appointments ---

func (r *diagnosticRepository) CreateAppointment(ctx context.Context, appt *models.DiagnosticAppointment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Row-level lock on equipment to serialize concurrent booking attempts
		var lockedID string
		err := tx.Raw("SELECT id::text FROM diagnostic_equipments WHERE id = ? FOR UPDATE", appt.EquipmentID).Scan(&lockedID).Error
		if err != nil {
			return fmt.Errorf("uskuna qatorini bloklashda xatolik: %w", err)
		}

		// 2. Double-booking check
		var existing models.DiagnosticAppointment
		err = tx.Where("equipment_id = ? AND appointment_date = ?", appt.EquipmentID, appt.AppointmentDate).
			Where("status IN ('confirmed', 'completed')").
			Where("start_time < ? AND end_time > ?", appt.EndTime, appt.StartTime).
			First(&existing).Error
		if err == nil {
			return ErrDoubleBookingConflict
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("qabul to‘qnashuvini tekshirishda xatolik: %w", err)
		}

		// 3. Create appointment
		if appt.ID == uuid.Nil {
			appt.ID = uuid.New()
		}
		if err := tx.Create(appt).Error; err != nil {
			if strings.Contains(err.Error(), "uq_equipment_active_slot") ||
				strings.Contains(err.Error(), "uq_diag_patient_active_slot") {
				return ErrDoubleBookingConflict
			}
			if strings.Contains(err.Error(), "uq_diag_appointments_patient_idempotency") {
				return ErrIdempotencyConflict
			}
			return err
		}
		return nil
	})
}

func (r *diagnosticRepository) GetAppointmentByID(ctx context.Context, id uuid.UUID) (*models.DiagnosticAppointment, error) {
	var appt models.DiagnosticAppointment
	err := r.db.WithContext(ctx).
		Preload("Patient").
		Preload("Clinic").
		Preload("Service").
		Preload("Service.Category").
		Preload("Equipment").
		First(&appt, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDiagnosticAppointmentNotFound
		}
		return nil, err
	}
	return &appt, nil
}

func (r *diagnosticRepository) FindByIdempotencyKey(ctx context.Context, patientID uuid.UUID, key string) (*models.DiagnosticAppointment, error) {
	var appt models.DiagnosticAppointment
	err := r.db.WithContext(ctx).
		Preload("Patient").
		Preload("Clinic").
		Preload("Service").
		Preload("Service.Category").
		Preload("Equipment").
		First(&appt, "patient_id = ? AND idempotency_key = ?", patientID, key).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &appt, nil
}

func (r *diagnosticRepository) CheckEquipmentBookingOverlap(ctx context.Context, equipmentID uuid.UUID, date, startTime, endTime string) (*models.DiagnosticAppointment, error) {
	var appt models.DiagnosticAppointment
	err := r.db.WithContext(ctx).
		Where("equipment_id = ? AND appointment_date = ?", equipmentID, date).
		Where("status IN ('confirmed', 'completed')").
		Where("start_time < ? AND end_time > ?", endTime, startTime).
		First(&appt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &appt, nil
}

func (r *diagnosticRepository) ListForPatient(
	ctx context.Context,
	patientID uuid.UUID,
	filter *dto.PatientDiagnosticAppointmentFilter,
	todayDate, nowTime string,
) ([]models.DiagnosticAppointment, int64, error) {
	query := r.db.WithContext(ctx).
		Model(&models.DiagnosticAppointment{}).
		Where("patient_id = ?", patientID)

	if filter != nil {
		switch filter.Timeframe {
		case "upcoming":
			query = query.Where("status = 'confirmed' AND (appointment_date > ? OR (appointment_date = ? AND start_time >= ?))", todayDate, todayDate, nowTime)
		case "past":
			query = query.Where("(status IN ('completed', 'no_show')) OR (status = 'confirmed' AND (appointment_date < ? OR (appointment_date = ? AND start_time < ?)))", todayDate, todayDate, nowTime)
		case "cancelled":
			query = query.Where("status = 'cancelled'")
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

	var list []models.DiagnosticAppointment
	err := query.
		Preload("Clinic").
		Preload("Service").
		Preload("Equipment").
		Order("appointment_date DESC, start_time DESC").
		Offset(offset).
		Limit(limit).
		Find(&list).Error

	return list, total, err
}

func (r *diagnosticRepository) ListForClinic(
	ctx context.Context,
	clinicID uuid.UUID,
	filter *dto.ClinicDiagnosticAppointmentFilter,
) ([]models.DiagnosticAppointment, int64, error) {
	query := r.db.WithContext(ctx).
		Model(&models.DiagnosticAppointment{}).
		Where("clinic_id = ?", clinicID)

	if filter != nil {
		if filter.EquipmentID != nil {
			query = query.Where("equipment_id = ?", *filter.EquipmentID)
		}
		if filter.ServiceID != nil {
			query = query.Where("service_id = ?", *filter.ServiceID)
		}
		if filter.Date != "" {
			query = query.Where("appointment_date = ?", filter.Date)
		}
		if filter.DateFrom != "" {
			query = query.Where("appointment_date >= ?", filter.DateFrom)
		}
		if filter.DateTo != "" {
			query = query.Where("appointment_date <= ?", filter.DateTo)
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

	var list []models.DiagnosticAppointment
	err := query.
		Preload("Patient").
		Preload("Service").
		Preload("Equipment").
		Order("appointment_date DESC, start_time DESC").
		Offset(offset).
		Limit(limit).
		Find(&list).Error

	return list, total, err
}

func (r *diagnosticRepository) UpdateAppointmentStatus(
	ctx context.Context,
	id uuid.UUID,
	fromStatus, toStatus models.DiagnosticAppointmentStatus,
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

	res := r.db.WithContext(ctx).
		Model(&models.DiagnosticAppointment{}).
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

func (r *diagnosticRepository) GetBookedSlotsForEquipment(ctx context.Context, equipmentID uuid.UUID, date string) ([]models.DiagnosticAppointment, error) {
	var list []models.DiagnosticAppointment
	err := r.db.WithContext(ctx).
		Where("equipment_id = ? AND appointment_date = ?", equipmentID, date).
		Where("status IN ('confirmed', 'completed')").
		Order("start_time ASC").
		Find(&list).Error
	return list, err
}

