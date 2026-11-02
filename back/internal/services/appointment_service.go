package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/models"
	"qabulgo-backend/internal/repositories"

	"github.com/google/uuid"
)

var (
	ErrPatientOnly                   = errors.New("faqat bemorlar qabulga yozilishi mumkin")
	ErrDoctorNotAssignedToClinic     = errors.New("shifokor ushbu klinikaga biriktirilmagan yoki faol emas")
	ErrSlotNotInWeeklySchedule       = errors.New("tanlangan vaqt shifokor qabul jadvalida mavjud emas")
	ErrSlotDateOverrideClosed         = errors.New("tanlangan vaqt shifokor tomonidan yopilgan yoki dam olish kuni")
	ErrPastTimeBooking               = errors.New("o‘tib ketgan sana va vaqtga qabulga yozilib bo‘lmaydi")
	ErrPastAppointmentCancellation   = errors.New("o‘tib ketgan qabulni bekor qilib bo‘lmaydi")
	ErrOnlyConfirmedCancellable      = errors.New("faqat tasdiqlangan (confirmed) qabulni bekor qilish mumkin")
	ErrPrematureCompletion           = errors.New("qabul vaqti hali kelmagan, qabulni faqat boshlanish vaqtidan so‘ng yakunlash mumkin")
	ErrPrematureNoShow               = errors.New("qabul vaqti hali kelmagan, bemor kelmaganligini faqat qabul vaqtidan so‘ng belgilash mumkin")
	ErrUnauthorizedAppointmentAccess = errors.New("ushbu qabul ma’lumotlarini ko‘rish huquqiga ega emassiz")
	ErrScheduleBookingConflict       = errors.New("jadvalni o‘zgartirib bo‘lmaydi: ushbu vaqtga tasdiqlangan qabul mavjud")
)

type AppointmentService interface {
	BookAppointment(ctx context.Context, patientID uuid.UUID, idempotencyKey string, req *dto.CreateAppointmentRequest) (*dto.AppointmentResponseDTO, error)
	GetPatientAppointments(ctx context.Context, patientID uuid.UUID, filter *dto.PatientAppointmentFilter) (*dto.AppointmentListResponseDTO, error)
	GetPatientAppointmentByID(ctx context.Context, patientID, appointmentID uuid.UUID) (*dto.AppointmentResponseDTO, error)
	CancelPatientAppointment(ctx context.Context, patientID, appointmentID uuid.UUID, req *dto.CancelAppointmentRequest) (*dto.AppointmentResponseDTO, error)

	GetClinicAppointments(ctx context.Context, clinicID, userID uuid.UUID, filter *dto.ClinicAppointmentFilter) (*dto.AppointmentListResponseDTO, error)
	GetClinicAppointmentByID(ctx context.Context, clinicID, userID, appointmentID uuid.UUID) (*dto.AppointmentResponseDTO, error)
	CancelClinicAppointment(ctx context.Context, clinicID, userID, appointmentID uuid.UUID, req *dto.CancelAppointmentRequest) (*dto.AppointmentResponseDTO, error)
	CompleteClinicAppointment(ctx context.Context, clinicID, userID, appointmentID uuid.UUID) (*dto.AppointmentResponseDTO, error)
	NoShowClinicAppointment(ctx context.Context, clinicID, userID, appointmentID uuid.UUID) (*dto.AppointmentResponseDTO, error)
}

type appointmentService struct {
	appointmentRepo repositories.AppointmentRepository
	scheduleRepo    repositories.ScheduleRepository
	clinicRepo      repositories.ClinicRepository
	auditRepo       repositories.AuditRepository
	outboxRepo      repositories.OutboxRepository
}

func NewAppointmentService(
	appointmentRepo repositories.AppointmentRepository,
	scheduleRepo repositories.ScheduleRepository,
	clinicRepo repositories.ClinicRepository,
	auditRepo repositories.AuditRepository,
	outboxRepo ...repositories.OutboxRepository,
) AppointmentService {
	svc := &appointmentService{
		appointmentRepo: appointmentRepo,
		scheduleRepo:    scheduleRepo,
		clinicRepo:      clinicRepo,
		auditRepo:       auditRepo,
	}
	if len(outboxRepo) > 0 {
		svc.outboxRepo = outboxRepo[0]
	}
	return svc
}

var (
	apptDateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	apptTimeRegex = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)
)

// BookAppointment automatically confirms and creates a booking
func (s *appointmentService) BookAppointment(
	ctx context.Context,
	patientID uuid.UUID,
	idempotencyKey string,
	req *dto.CreateAppointmentRequest,
) (*dto.AppointmentResponseDTO, error) {
	if patientID == uuid.Nil {
		return nil, ErrPatientOnly
	}

	// 1. Validate request payload formats
	if !apptDateRegex.MatchString(req.Date) {
		return nil, ErrInvalidDateFormat
	}
	if !apptTimeRegex.MatchString(req.StartTime) || !apptTimeRegex.MatchString(req.EndTime) {
		return nil, ErrInvalidTimeFormat
	}
	if req.StartTime >= req.EndTime {
		return nil, ErrStartMustBeBeforeEnd
	}

	// 2. Idempotency Key validation
	cleanKey := strings.TrimSpace(idempotencyKey)
	var keyPtr *string
	if cleanKey != "" {
		keyPtr = &cleanKey
		existingAppt, err := s.appointmentRepo.FindByIdempotencyKey(ctx, patientID, cleanKey)
		if err != nil {
			return nil, fmt.Errorf("idempotency tekshiruvida xatolik: %w", err)
		}
		if existingAppt != nil {
			// Check if payload matches
			if existingAppt.ClinicID == req.ClinicID &&
				existingAppt.DoctorID == req.DoctorID &&
				existingAppt.AppointmentDate == req.Date &&
				existingAppt.StartTime == req.StartTime &&
				existingAppt.EndTime == req.EndTime {
				return mapAppointmentToDTO(existingAppt, true), nil
			}
			return nil, repositories.ErrIdempotencyConflict
		}
	}

	// 3. Check Clinic exists and is Approved
	clinic, err := s.clinicRepo.GetByID(ctx, req.ClinicID)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicNotFound) {
			return nil, repositories.ErrClinicNotFound
		}
		return nil, fmt.Errorf("klinikani tekshirishda xatolik: %w", err)
	}
	if clinic.Status != models.ClinicStatusApproved {
		return nil, ErrClinicNotApproved
	}

	// 4. Check ClinicDoctor active link
	cd, err := s.scheduleRepo.FindClinicDoctor(ctx, req.ClinicID, req.DoctorID)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicDoctorNotFound) {
			return nil, ErrDoctorNotAssignedToClinic
		}
		return nil, fmt.Errorf("shifokor bog‘lanishini tekshirishda xatolik: %w", err)
	}
	if !cd.IsActive {
		return nil, ErrDoctorNotAssignedToClinic
	}
	if cd.Doctor == nil {
		// doctor record inactive check
		return nil, ErrDoctorInactive
	}

	// 5. Timezone validation (Asia/Tashkent): check past date/time
	loc := getTashkentLocation()
	now := time.Now().In(loc)

	slotDateTime, err := time.ParseInLocation("2006-01-02 15:04", req.Date+" "+req.StartTime, loc)
	if err != nil {
		return nil, ErrInvalidDateFormat
	}
	if slotDateTime.Before(now) {
		return nil, ErrPastTimeBooking
	}

	// 6. Check Schedule: must be in weekly schedule or override
	weekday := int(slotDateTime.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday
	}

	weeklySlots, err := s.scheduleRepo.GetWeeklySlots(ctx, cd.ID)
	if err != nil {
		return nil, fmt.Errorf("shifokor haftalik jadvalini olishda xatolik: %w", err)
	}

	dateOverrides, err := s.scheduleRepo.GetDateOverrides(ctx, cd.ID, req.Date, "", "")
	if err != nil {
		return nil, fmt.Errorf("istisno sanalarni olishda xatolik: %w", err)
	}

	// Check if entire day is off
	for _, o := range dateOverrides {
		if o.Status == models.OverrideStatusDayOff {
			return nil, ErrSlotDateOverrideClosed
		}
	}

	// Check if closed by an override
	for _, o := range dateOverrides {
		if o.Status == models.OverrideStatusClosed && o.StartTime != nil && o.EndTime != nil {
			if req.StartTime < *o.EndTime && *o.StartTime < req.EndTime {
				return nil, ErrSlotDateOverrideClosed
			}
		}
	}

	// Check if slot exists in active weekly slots
	slotInWeekly := false
	for _, ws := range weeklySlots {
		if ws.Status == models.SlotStatusActive && ws.Weekday == weekday {
			// Slot must exactly match or fit within schedule
			if ws.StartTime <= req.StartTime && ws.EndTime >= req.EndTime {
				slotInWeekly = true
				break
			}
		}
	}

	// Or added as an available override
	slotInOverride := false
	for _, o := range dateOverrides {
		if o.Status == models.OverrideStatusAvailable && o.StartTime != nil && o.EndTime != nil {
			if *o.StartTime <= req.StartTime && *o.EndTime >= req.EndTime {
				slotInOverride = true
				break
			}
		}
	}

	if !slotInWeekly && !slotInOverride {
		return nil, ErrSlotNotInWeeklySchedule
	}

	// 7. Check Cross-Clinic Doctor Booking Overlap
	conflictAppt, err := s.appointmentRepo.CheckDoctorBookingOverlap(ctx, cd.DoctorID, req.Date, req.StartTime, req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("bandlikni tekshirishda xatolik: %w", err)
	}
	if conflictAppt != nil {
		return nil, repositories.ErrDoubleBookingConflict
	}

	// 8. Capture authoritative price from ClinicDoctor record (never trust frontend)
	consultationPrice := cd.ConsultationPrice

	// 9. Create Appointment model and execute atomic transaction
	newAppt := &models.Appointment{
		ID:                uuid.New(),
		PatientID:         patientID,
		ClinicID:          req.ClinicID,
		DoctorID:          cd.DoctorID,
		ClinicDoctorID:    cd.ID,
		AppointmentDate:   req.Date,
		StartTime:         req.StartTime,
		EndTime:           req.EndTime,
		Status:            models.AppointmentStatusConfirmed,
		ConsultationPrice: consultationPrice,
		IdempotencyKey:    keyPtr,
	}

	if err := s.appointmentRepo.Create(ctx, newAppt); err != nil {
		return nil, err
	}

	// 10. Audit log
	_ = s.auditRepo.CreateAuditLog(ctx, &models.AuditLog{
		UserID:     &patientID,
		Action:     "BOOK_APPOINTMENT",
		EntityType: "appointments",
		EntityID:   &newAppt.ID,
		Details:    fmt.Sprintf(`{"clinic_id":"%s","doctor_id":"%s","date":"%s","start_time":"%s"}`, req.ClinicID, cd.DoctorID, req.Date, req.StartTime),
	})

	// 11. Retrieve fully populated appointment
	createdAppt, err := s.appointmentRepo.GetByID(ctx, newAppt.ID)
	if err != nil {
		return mapAppointmentToDTO(newAppt, true), nil
	}

	// 12. Enqueue Outbox Notification for Clinic Telegram Bot
	if s.outboxRepo != nil {
		clinicName := ""
		if createdAppt.Clinic != nil {
			clinicName = createdAppt.Clinic.Name
		}
		patientName := "Bemor"
		patientPhone := ""
		if createdAppt.Patient != nil {
			patientName = fmt.Sprintf("%s %s", createdAppt.Patient.FirstName, createdAppt.Patient.LastName)
			patientPhone = createdAppt.Patient.Phone
		}
		doctorName := "Shifokor"
		specialtyName := ""
		if createdAppt.Doctor != nil {
			doctorName = fmt.Sprintf("%s %s", createdAppt.Doctor.FirstName, createdAppt.Doctor.LastName)
		}
		if cd.Specialty != nil {
			specialtyName = cd.Specialty.Name
		}

		payloadBytes, _ := json.Marshal(map[string]any{
			"appointment_id": createdAppt.ID.String(),
			"clinic_id":      createdAppt.ClinicID.String(),
			"clinic_name":    clinicName,
			"patient_name":   patientName,
			"patient_phone":  patientPhone,
			"doctor_name":    doctorName,
			"specialty":      specialtyName,
			"date":           createdAppt.AppointmentDate,
			"start_time":     createdAppt.StartTime,
			"end_time":       createdAppt.EndTime,
			"price":          createdAppt.ConsultationPrice,
		})

		_ = s.outboxRepo.Enqueue(ctx, &models.NotificationOutbox{
			ID:             uuid.New(),
			EventType:      "appointment_created",
			EntityType:     "appointment",
			EntityID:       createdAppt.ID,
			TargetClinicID: &createdAppt.ClinicID,
			TargetRole:     "clinic",
			Payload:        string(payloadBytes),
			Status:         "pending",
		})
	}

	return mapAppointmentToDTO(createdAppt, true), nil
}

// GetPatientAppointments retrieves paginated appointments for the patient
func (s *appointmentService) GetPatientAppointments(
	ctx context.Context,
	patientID uuid.UUID,
	filter *dto.PatientAppointmentFilter,
) (*dto.AppointmentListResponseDTO, error) {
	loc := getTashkentLocation()
	now := time.Now().In(loc)
	todayDate := now.Format("2006-01-02")
	nowTime := now.Format("15:04")

	list, total, err := s.appointmentRepo.ListForPatient(ctx, patientID, filter, todayDate, nowTime)
	if err != nil {
		return nil, fmt.Errorf("bemor qabullarini olishda xatolik: %w", err)
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

	dtos := make([]dto.AppointmentResponseDTO, 0, len(list))
	for i := range list {
		dtos = append(dtos, *mapAppointmentToDTO(&list[i], false))
	}

	return &dto.AppointmentListResponseDTO{
		Appointments: dtos,
		Total:        total,
		Page:         page,
		Limit:        limit,
	}, nil
}

// GetPatientAppointmentByID retrieves specific appointment for the patient with IDOR protection
func (s *appointmentService) GetPatientAppointmentByID(ctx context.Context, patientID, appointmentID uuid.UUID) (*dto.AppointmentResponseDTO, error) {
	appt, err := s.appointmentRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return nil, err
	}
	if appt.PatientID != patientID {
		return nil, ErrUnauthorizedAppointmentAccess
	}
	return mapAppointmentToDTO(appt, false), nil
}

// CancelPatientAppointment cancels a patient's own confirmed appointment
func (s *appointmentService) CancelPatientAppointment(
	ctx context.Context,
	patientID, appointmentID uuid.UUID,
	req *dto.CancelAppointmentRequest,
) (*dto.AppointmentResponseDTO, error) {
	appt, err := s.appointmentRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return nil, err
	}
	if appt.PatientID != patientID {
		return nil, ErrUnauthorizedAppointmentAccess
	}
	if appt.Status != models.AppointmentStatusConfirmed {
		return nil, ErrOnlyConfirmedCancellable
	}

	loc := getTashkentLocation()
	now := time.Now().In(loc)
	slotDateTime, err := time.ParseInLocation("2006-01-02 15:04", appt.AppointmentDate+" "+appt.StartTime, loc)
	if err == nil && slotDateTime.Before(now) {
		return nil, ErrPastAppointmentCancellation
	}

	var reason *string
	if req != nil && strings.TrimSpace(req.Reason) != "" {
		trimmed := strings.TrimSpace(req.Reason)
		reason = &trimmed
	}

	if err := s.appointmentRepo.UpdateStatus(
		ctx,
		appointmentID,
		models.AppointmentStatusConfirmed,
		models.AppointmentStatusCancelled,
		reason,
		"cancelled_at",
		time.Now(),
	); err != nil {
		return nil, err
	}

	_ = s.auditRepo.CreateAuditLog(ctx, &models.AuditLog{
		UserID:     &patientID,
		Action:     "CANCEL_APPOINTMENT_BY_PATIENT",
		EntityType: "appointments",
		EntityID:   &appointmentID,
		Details:    fmt.Sprintf(`{"reason":"%s"}`, req.Reason),
	})

	updated, err := s.appointmentRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return nil, err
	}

	if s.outboxRepo != nil {
		clinicName := ""
		if updated.Clinic != nil {
			clinicName = updated.Clinic.Name
		}
		patientName := "Bemor"
		if updated.Patient != nil {
			patientName = fmt.Sprintf("%s %s", updated.Patient.FirstName, updated.Patient.LastName)
		}
		docName := "Shifokor"
		if updated.Doctor != nil {
			docName = fmt.Sprintf("%s %s", updated.Doctor.FirstName, updated.Doctor.LastName)
		}
		reasonStr := "Bemor tomonidan bekor qilindi"
		if req != nil && strings.TrimSpace(req.Reason) != "" {
			reasonStr = strings.TrimSpace(req.Reason)
		}

		payloadBytes, _ := json.Marshal(map[string]any{
			"appointment_id": updated.ID.String(),
			"clinic_id":      updated.ClinicID.String(),
			"clinic_name":    clinicName,
			"patient_name":   patientName,
			"doctor_name":    docName,
			"date":           updated.AppointmentDate,
			"start_time":     updated.StartTime,
			"end_time":       updated.EndTime,
			"cancelled_by":   "Bemor",
			"reason":         reasonStr,
		})

		_ = s.outboxRepo.Enqueue(ctx, &models.NotificationOutbox{
			ID:             uuid.New(),
			EventType:      "appointment_cancelled",
			EntityType:     "appointment",
			EntityID:       updated.ID,
			TargetClinicID: &updated.ClinicID,
			TargetRole:     "clinic",
			Payload:        string(payloadBytes),
			Status:         "pending",
		})
	}

	return mapAppointmentToDTO(updated, false), nil
}

// GetClinicAppointments retrieves paginated appointments for the clinic
func (s *appointmentService) GetClinicAppointments(
	ctx context.Context,
	clinicID, userID uuid.UUID,
	filter *dto.ClinicAppointmentFilter,
) (*dto.AppointmentListResponseDTO, error) {
	isAdmin, err := s.clinicRepo.IsUserClinicAdmin(ctx, clinicID, userID)
	if err != nil || !isAdmin {
		return nil, repositories.ErrClinicAccessDenied
	}

	list, total, err := s.appointmentRepo.ListForClinic(ctx, clinicID, filter)
	if err != nil {
		return nil, fmt.Errorf("klinika qabullarini olishda xatolik: %w", err)
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

	dtos := make([]dto.AppointmentResponseDTO, 0, len(list))
	for i := range list {
		dtos = append(dtos, *mapAppointmentToDTO(&list[i], true))
	}

	return &dto.AppointmentListResponseDTO{
		Appointments: dtos,
		Total:        total,
		Page:         page,
		Limit:        limit,
	}, nil
}

// GetClinicAppointmentByID retrieves appointment detail for the clinic
func (s *appointmentService) GetClinicAppointmentByID(ctx context.Context, clinicID, userID, appointmentID uuid.UUID) (*dto.AppointmentResponseDTO, error) {
	isAdmin, err := s.clinicRepo.IsUserClinicAdmin(ctx, clinicID, userID)
	if err != nil || !isAdmin {
		return nil, repositories.ErrClinicAccessDenied
	}

	appt, err := s.appointmentRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return nil, err
	}
	if appt.ClinicID != clinicID {
		return nil, ErrUnauthorizedAppointmentAccess
	}

	return mapAppointmentToDTO(appt, true), nil
}

// CancelClinicAppointment cancels appointment by clinic staff
func (s *appointmentService) CancelClinicAppointment(
	ctx context.Context,
	clinicID, userID, appointmentID uuid.UUID,
	req *dto.CancelAppointmentRequest,
) (*dto.AppointmentResponseDTO, error) {
	isAdmin, err := s.clinicRepo.IsUserClinicAdmin(ctx, clinicID, userID)
	if err != nil || !isAdmin {
		return nil, repositories.ErrClinicAccessDenied
	}

	appt, err := s.appointmentRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return nil, err
	}
	if appt.ClinicID != clinicID {
		return nil, ErrUnauthorizedAppointmentAccess
	}
	if appt.Status != models.AppointmentStatusConfirmed {
		return nil, ErrOnlyConfirmedCancellable
	}

	var reason *string
	if req != nil && strings.TrimSpace(req.Reason) != "" {
		trimmed := strings.TrimSpace(req.Reason)
		reason = &trimmed
	}

	if err := s.appointmentRepo.UpdateStatus(
		ctx,
		appointmentID,
		models.AppointmentStatusConfirmed,
		models.AppointmentStatusCancelled,
		reason,
		"cancelled_at",
		time.Now(),
	); err != nil {
		return nil, err
	}

	_ = s.auditRepo.CreateAuditLog(ctx, &models.AuditLog{
		UserID:     &userID,
		Action:     "CANCEL_APPOINTMENT_BY_CLINIC",
		EntityType: "appointments",
		EntityID:   &appointmentID,
	})

	updated, err := s.appointmentRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return nil, err
	}

	if s.outboxRepo != nil {
		clinicName := ""
		if updated.Clinic != nil {
			clinicName = updated.Clinic.Name
		}
		patientName := "Bemor"
		if updated.Patient != nil {
			patientName = fmt.Sprintf("%s %s", updated.Patient.FirstName, updated.Patient.LastName)
		}
		docName := "Shifokor"
		if updated.Doctor != nil {
			docName = fmt.Sprintf("%s %s", updated.Doctor.FirstName, updated.Doctor.LastName)
		}
		reasonStr := "Klinika tomonidan bekor qilindi"
		if req != nil && strings.TrimSpace(req.Reason) != "" {
			reasonStr = strings.TrimSpace(req.Reason)
		}

		payloadBytes, _ := json.Marshal(map[string]any{
			"appointment_id": updated.ID.String(),
			"clinic_id":      updated.ClinicID.String(),
			"clinic_name":    clinicName,
			"patient_name":   patientName,
			"doctor_name":    docName,
			"date":           updated.AppointmentDate,
			"start_time":     updated.StartTime,
			"end_time":       updated.EndTime,
			"cancelled_by":   "Klinika ma’muriyati",
			"reason":         reasonStr,
		})

		_ = s.outboxRepo.Enqueue(ctx, &models.NotificationOutbox{
			ID:             uuid.New(),
			EventType:      "appointment_cancelled",
			EntityType:     "appointment",
			EntityID:       updated.ID,
			TargetClinicID: &updated.ClinicID,
			TargetRole:     "clinic",
			Payload:        string(payloadBytes),
			Status:         "pending",
		})
	}

	return mapAppointmentToDTO(updated, true), nil
}

// CompleteClinicAppointment marks appointment as completed (only after slot start time)
func (s *appointmentService) CompleteClinicAppointment(ctx context.Context, clinicID, userID, appointmentID uuid.UUID) (*dto.AppointmentResponseDTO, error) {
	isAdmin, err := s.clinicRepo.IsUserClinicAdmin(ctx, clinicID, userID)
	if err != nil || !isAdmin {
		return nil, repositories.ErrClinicAccessDenied
	}

	appt, err := s.appointmentRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return nil, err
	}
	if appt.ClinicID != clinicID {
		return nil, ErrUnauthorizedAppointmentAccess
	}
	if appt.Status != models.AppointmentStatusConfirmed {
		return nil, fmt.Errorf("faqat tasdiqlangan (confirmed) qabulni yakunlash mumkin: %w", repositories.ErrAppointmentNotEditable)
	}

	loc := getTashkentLocation()
	now := time.Now().In(loc)
	slotStart, err := time.ParseInLocation("2006-01-02 15:04", appt.AppointmentDate+" "+appt.StartTime, loc)
	if err == nil && now.Before(slotStart) {
		return nil, ErrPrematureCompletion
	}

	if err := s.appointmentRepo.UpdateStatus(
		ctx,
		appointmentID,
		models.AppointmentStatusConfirmed,
		models.AppointmentStatusCompleted,
		nil,
		"completed_at",
		time.Now(),
	); err != nil {
		return nil, err
	}

	_ = s.auditRepo.CreateAuditLog(ctx, &models.AuditLog{
		UserID:     &userID,
		Action:     "COMPLETE_APPOINTMENT",
		EntityType: "appointments",
		EntityID:   &appointmentID,
	})

	updated, err := s.appointmentRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return nil, err
	}
	return mapAppointmentToDTO(updated, true), nil
}

// NoShowClinicAppointment marks patient as no-show (only after slot start time)
func (s *appointmentService) NoShowClinicAppointment(ctx context.Context, clinicID, userID, appointmentID uuid.UUID) (*dto.AppointmentResponseDTO, error) {
	isAdmin, err := s.clinicRepo.IsUserClinicAdmin(ctx, clinicID, userID)
	if err != nil || !isAdmin {
		return nil, repositories.ErrClinicAccessDenied
	}

	appt, err := s.appointmentRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return nil, err
	}
	if appt.ClinicID != clinicID {
		return nil, ErrUnauthorizedAppointmentAccess
	}
	if appt.Status != models.AppointmentStatusConfirmed {
		return nil, fmt.Errorf("faqat tasdiqlangan (confirmed) qabulni no-show holatiga o‘tkazish mumkin: %w", repositories.ErrAppointmentNotEditable)
	}

	loc := getTashkentLocation()
	now := time.Now().In(loc)
	slotStart, err := time.ParseInLocation("2006-01-02 15:04", appt.AppointmentDate+" "+appt.StartTime, loc)
	if err == nil && now.Before(slotStart) {
		return nil, ErrPrematureNoShow
	}

	if err := s.appointmentRepo.UpdateStatus(
		ctx,
		appointmentID,
		models.AppointmentStatusConfirmed,
		models.AppointmentStatusNoShow,
		nil,
		"no_show_at",
		time.Now(),
	); err != nil {
		return nil, err
	}

	_ = s.auditRepo.CreateAuditLog(ctx, &models.AuditLog{
		UserID:     &userID,
		Action:     "NO_SHOW_APPOINTMENT",
		EntityType: "appointments",
		EntityID:   &appointmentID,
	})

	updated, err := s.appointmentRepo.GetByID(ctx, appointmentID)
	if err != nil {
		return nil, err
	}
	return mapAppointmentToDTO(updated, true), nil
}

// mapAppointmentToDTO converts models.Appointment to sanitized DTO
func mapAppointmentToDTO(appt *models.Appointment, includePatient bool) *dto.AppointmentResponseDTO {
	res := &dto.AppointmentResponseDTO{
		ID:                 appt.ID,
		PatientID:          appt.PatientID,
		ClinicID:           appt.ClinicID,
		DoctorID:           appt.DoctorID,
		ClinicDoctorID:     appt.ClinicDoctorID,
		AppointmentDate:    appt.AppointmentDate,
		StartTime:          appt.StartTime,
		EndTime:            appt.EndTime,
		Status:             string(appt.Status),
		ConsultationPrice:  appt.ConsultationPrice,
		CancellationReason: appt.CancellationReason,
		CancelledAt:        appt.CancelledAt,
		CompletedAt:        appt.CompletedAt,
		NoShowAt:           appt.NoShowAt,
		CreatedAt:          appt.CreatedAt,
		UpdatedAt:          appt.UpdatedAt,
	}

	if appt.Doctor != nil {
		specialtyName := ""
		if appt.ClinicDoctor != nil && appt.ClinicDoctor.Specialty != nil {
			specialtyName = appt.ClinicDoctor.Specialty.Name
		}
		res.Doctor = &dto.AppointmentDoctorDTO{
			ID:        appt.Doctor.ID,
			FirstName: appt.Doctor.FirstName,
			LastName:  appt.Doctor.LastName,
			Specialty: specialtyName,
			PhotoURL:  appt.Doctor.PhotoURL,
		}
	}

	if appt.Clinic != nil {
		res.Clinic = &dto.AppointmentClinicDTO{
			ID:       appt.Clinic.ID,
			Name:     appt.Clinic.Name,
			Address:  appt.Clinic.Address,
			Region:   appt.Clinic.Region,
			District: appt.Clinic.District,
			Phone:    appt.Clinic.Phone,
		}
	}

	if includePatient && appt.Patient != nil {
		res.Patient = &dto.AppointmentPatientDTO{
			ID:        appt.Patient.ID,
			FirstName: appt.Patient.FirstName,
			LastName:  appt.Patient.LastName,
			Phone:     appt.Patient.Phone,
		}
	}

	return res
}
