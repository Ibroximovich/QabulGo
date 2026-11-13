package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/models"

	"github.com/google/uuid"
)

type mockDiagRepo struct {
	categories  []models.DiagnosticCategory
	equipment   *models.DiagnosticEquipment
	slots       []models.DiagnosticWeeklySlot
	overrides   []models.DiagnosticDateSlotOverride
	booked      []models.DiagnosticAppointment
	createdAppt *models.DiagnosticAppointment
}

func (m *mockDiagRepo) ListCategories(ctx context.Context) ([]models.DiagnosticCategory, error) {
	if len(m.categories) > 0 {
		return m.categories, nil
	}
	return []models.DiagnosticCategory{
		{ID: uuid.New(), Name: "MRT", Slug: "mrt"},
		{ID: uuid.New(), Name: "UZI", Slug: "uzi"},
	}, nil
}
func (m *mockDiagRepo) GetCategoryByID(ctx context.Context, id uuid.UUID) (*models.DiagnosticCategory, error) {
	return &models.DiagnosticCategory{ID: id, Name: "MRT", Slug: "mrt"}, nil
}
func (m *mockDiagRepo) GetCategoryBySlug(ctx context.Context, slug string) (*models.DiagnosticCategory, error) {
	return &models.DiagnosticCategory{ID: uuid.New(), Name: "MRT", Slug: slug}, nil
}
func (m *mockDiagRepo) CreateEquipment(ctx context.Context, eq *models.DiagnosticEquipment) error {
	eq.ID = uuid.New()
	return nil
}
func (m *mockDiagRepo) UpdateEquipment(ctx context.Context, eq *models.DiagnosticEquipment) error {
	return nil
}
func (m *mockDiagRepo) GetEquipmentByID(ctx context.Context, id uuid.UUID) (*models.DiagnosticEquipment, error) {
	if m.equipment != nil {
		return m.equipment, nil
	}
	return &models.DiagnosticEquipment{
		ID:         id,
		ClinicID:   uuid.New(),
		CategoryID: uuid.New(),
		Name:       "Siemens 1.5T",
		IsActive:   true,
		Clinic: &models.Clinic{
			Status: models.ClinicStatusApproved,
		},
	}, nil
}
func (m *mockDiagRepo) ListEquipmentsByClinic(ctx context.Context, clinicID uuid.UUID) ([]models.DiagnosticEquipment, error) {
	return []models.DiagnosticEquipment{}, nil
}
func (m *mockDiagRepo) DeleteEquipment(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockDiagRepo) CreateService(ctx context.Context, svc *models.DiagnosticService) error {
	svc.ID = uuid.New()
	return nil
}
func (m *mockDiagRepo) UpdateService(ctx context.Context, svc *models.DiagnosticService) error {
	return nil
}
func (m *mockDiagRepo) GetServiceByID(ctx context.Context, id uuid.UUID) (*models.DiagnosticService, error) {
	return &models.DiagnosticService{
		ID:               id,
		ClinicID:         m.equipment.ClinicID,
		CategoryID:       m.equipment.CategoryID,
		EquipmentID:      m.equipment.ID,
		Name:             "Bosh miya MRT",
		Price:            450000,
		DurationMinutes:  30,
		PreparationRules: "Metall buyumlarni yechish",
		IsActive:         true,
	}, nil
}
func (m *mockDiagRepo) ListServicesByClinic(ctx context.Context, clinicID uuid.UUID, categoryID *uuid.UUID) ([]models.DiagnosticService, error) {
	return []models.DiagnosticService{}, nil
}
func (m *mockDiagRepo) ListPublicServices(ctx context.Context, clinicID *uuid.UUID, categoryID *uuid.UUID) ([]models.DiagnosticService, error) {
	return []models.DiagnosticService{}, nil
}
func (m *mockDiagRepo) ListServicesByEquipment(ctx context.Context, equipmentID uuid.UUID) ([]models.DiagnosticService, error) {
	return []models.DiagnosticService{}, nil
}
func (m *mockDiagRepo) DeleteService(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockDiagRepo) GetWeeklySlots(ctx context.Context, equipmentID uuid.UUID) ([]models.DiagnosticWeeklySlot, error) {
	return m.slots, nil
}
func (m *mockDiagRepo) SaveWeeklySlots(ctx context.Context, equipmentID uuid.UUID, slots []models.DiagnosticWeeklySlot) error {
	m.slots = slots
	return nil
}
func (m *mockDiagRepo) GetDateOverrides(ctx context.Context, equipmentID uuid.UUID, date, startDate, endDate string) ([]models.DiagnosticDateSlotOverride, error) {
	return m.overrides, nil
}
func (m *mockDiagRepo) CreateDateOverride(ctx context.Context, override *models.DiagnosticDateSlotOverride) error {
	return nil
}
func (m *mockDiagRepo) UpdateDateOverride(ctx context.Context, override *models.DiagnosticDateSlotOverride) error {
	return nil
}
func (m *mockDiagRepo) DeleteDateOverride(ctx context.Context, equipmentID, overrideID uuid.UUID) error {
	return nil
}
func (m *mockDiagRepo) GetEquipmentAvailabilityData(ctx context.Context, equipmentID uuid.UUID, dateStr string, weekday int) (*models.DiagnosticEquipment, []models.DiagnosticWeeklySlot, []models.DiagnosticDateSlotOverride, error) {
	return m.equipment, m.slots, m.overrides, nil
}
func (m *mockDiagRepo) CreateAppointment(ctx context.Context, appt *models.DiagnosticAppointment) error {
	appt.ID = uuid.New()
	if m.equipment != nil {
		appt.Equipment = m.equipment
		appt.Clinic = m.equipment.Clinic
	}
	m.createdAppt = appt
	return nil
}
func (m *mockDiagRepo) GetAppointmentByID(ctx context.Context, id uuid.UUID) (*models.DiagnosticAppointment, error) {
	if m.createdAppt != nil {
		return m.createdAppt, nil
	}
	return &models.DiagnosticAppointment{
		ID:                      id,
		AppointmentDate:         "2026-12-01",
		StartTime:               "10:00",
		EndTime:                 "10:30",
		Price:                   450000,
		Status:                  models.DiagAppointmentStatusConfirmed,
		PreparationAcknowledged: true,
	}, nil
}
func (m *mockDiagRepo) FindByIdempotencyKey(ctx context.Context, patientID uuid.UUID, key string) (*models.DiagnosticAppointment, error) {
	return nil, nil
}
func (m *mockDiagRepo) CheckEquipmentBookingOverlap(ctx context.Context, equipmentID uuid.UUID, date, startTime, endTime string) (*models.DiagnosticAppointment, error) {
	for _, b := range m.booked {
		if b.AppointmentDate == date && startTime < b.EndTime && b.StartTime < endTime {
			return &b, nil
		}
	}
	return nil, nil
}
func (m *mockDiagRepo) ListForPatient(ctx context.Context, patientID uuid.UUID, filter *dto.PatientDiagnosticAppointmentFilter, todayDate, nowTime string) ([]models.DiagnosticAppointment, int64, error) {
	return nil, 0, nil
}
func (m *mockDiagRepo) ListForClinic(ctx context.Context, clinicID uuid.UUID, filter *dto.ClinicDiagnosticAppointmentFilter) ([]models.DiagnosticAppointment, int64, error) {
	return nil, 0, nil
}
func (m *mockDiagRepo) UpdateAppointmentStatus(ctx context.Context, id uuid.UUID, fromStatus, toStatus models.DiagnosticAppointmentStatus, reason *string, timestampField string, timestamp time.Time) error {
	return nil
}
func (m *mockDiagRepo) GetBookedSlotsForEquipment(ctx context.Context, equipmentID uuid.UUID, date string) ([]models.DiagnosticAppointment, error) {
	return m.booked, nil
}

type mockClinicRepo struct{}

func (m *mockClinicRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Clinic, error) {
	return &models.Clinic{
		ID:     id,
		Status: models.ClinicStatusApproved,
	}, nil
}
func (m *mockClinicRepo) IsUserClinicAdmin(ctx context.Context, clinicID, userID uuid.UUID) (bool, error) {
	return true, nil
}
// stubs for rest of ClinicRepository if needed
func (m *mockClinicRepo) Create(ctx context.Context, clinic *models.Clinic, workingHours []models.ClinicWorkingHour) error { return nil }
func (m *mockClinicRepo) Update(ctx context.Context, clinic *models.Clinic) error { return nil }
func (m *mockClinicRepo) GetByOwnerID(ctx context.Context, ownerID uuid.UUID) (*models.Clinic, error) { return nil, nil }
func (m *mockClinicRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Clinic, error) { return nil, nil }
func (m *mockClinicRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status models.ClinicStatus, rejectionReason string) error { return nil }
func (m *mockClinicRepo) SetWorkingHours(ctx context.Context, clinicID uuid.UUID, hours []models.ClinicWorkingHour) error { return nil }
func (m *mockClinicRepo) GetWorkingHours(ctx context.Context, clinicID uuid.UUID) ([]models.ClinicWorkingHour, error) { return nil, nil }
func (m *mockClinicRepo) ListForAdmin(ctx context.Context, filter *dto.AdminClinicFilter) ([]models.Clinic, int64, error) { return nil, 0, nil }

func TestDiagnosticService_PreparationAcknowledgedRequired(t *testing.T) {
	clinicID := uuid.New()
	eqID := uuid.New()
	svcID := uuid.New()
	patientID := uuid.New()

	diagRepo := &mockDiagRepo{
		equipment: &models.DiagnosticEquipment{
			ID:         eqID,
			ClinicID:   clinicID,
			CategoryID: uuid.New(),
			Name:       "Siemens 1.5T",
			IsActive:   true,
			Clinic:     &models.Clinic{ID: clinicID, Status: models.ClinicStatusApproved},
		},
	}
	clinicRepo := &mockClinicRepo{}

	service := NewDiagnosticService(diagRepo, clinicRepo, nil)

	// Attempt booking without acknowledging preparation rules
	req := &dto.CreateDiagnosticAppointmentRequest{
		ClinicID:                clinicID,
		ServiceID:               svcID,
		EquipmentID:             eqID,
		Date:                    "2026-12-01",
		StartTime:               "10:00",
		EndTime:                 "10:30",
		PreparationAcknowledged: false, // NOT acknowledged
	}

	_, err := service.BookDiagnosticAppointment(context.Background(), patientID, "", req)
	if err == nil {
		t.Fatal("Expected error when preparation rules are not acknowledged, got nil")
	}
	if err != ErrPreparationNotAcknowledged {
		t.Fatalf("Expected ErrPreparationNotAcknowledged, got %v", err)
	}
}

func TestDiagnosticService_EquipmentAvailabilityCalculations(t *testing.T) {
	clinicID := uuid.New()
	eqID := uuid.New()

	diagRepo := &mockDiagRepo{
		equipment: &models.DiagnosticEquipment{
			ID:         eqID,
			ClinicID:   clinicID,
			CategoryID: uuid.New(),
			Name:       "Siemens 1.5T",
			IsActive:   true,
			Clinic:     &models.Clinic{ID: clinicID, Status: models.ClinicStatusApproved},
		},
		slots: []models.DiagnosticWeeklySlot{
			{EquipmentID: eqID, Weekday: 2, StartTime: "09:00", EndTime: "09:30", Status: models.SlotStatusActive},
			{EquipmentID: eqID, Weekday: 2, StartTime: "09:30", EndTime: "10:00", Status: models.SlotStatusActive},
			{EquipmentID: eqID, Weekday: 2, StartTime: "10:00", EndTime: "10:30", Status: models.SlotStatusActive},
			{EquipmentID: eqID, Weekday: 2, StartTime: "10:30", EndTime: "11:00", Status: models.SlotStatusActive},
		},
		overrides: []models.DiagnosticDateSlotOverride{
			// 09:30-10:00 is closed for maintenance
			{
				EquipmentID: eqID,
				Date:        "2026-12-01",
				StartTime:   ptr("09:30"),
				EndTime:     ptr("10:00"),
				Status:      models.OverrideStatusClosed,
				Reason:      "Profilaktika",
			},
		},
		booked: []models.DiagnosticAppointment{
			// 10:00-10:30 is already booked by another patient
			{
				EquipmentID:     eqID,
				AppointmentDate: "2026-12-01",
				StartTime:       "10:00",
				EndTime:         "10:30",
				Status:          models.DiagAppointmentStatusConfirmed,
			},
		},
	}
	clinicRepo := &mockClinicRepo{}

	service := NewDiagnosticService(diagRepo, clinicRepo, nil)

	// 2026-12-01 is Tuesday (Weekday 2)
	avail, err := service.GetDiagnosticEquipmentAvailability(context.Background(), eqID, "2026-12-01")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Out of 4 slots:
	// - 09:00-09:30: available
	// - 09:30-10:00: closed by override
	// - 10:00-10:30: booked
	// - 10:30-11:00: available
	// Expected exactly 2 slots: 09:00 and 10:30
	if len(avail.Slots) != 2 {
		t.Fatalf("Expected 2 available slots, got %d", len(avail.Slots))
	}
	if avail.Slots[0].StartTime != "09:00" || avail.Slots[0].EndTime != "09:30" {
		t.Errorf("Expected first slot 09:00-09:30, got %s-%s", avail.Slots[0].StartTime, avail.Slots[0].EndTime)
	}
	if avail.Slots[1].StartTime != "10:30" || avail.Slots[1].EndTime != "11:00" {
		t.Errorf("Expected second slot 10:30-11:00, got %s-%s", avail.Slots[1].StartTime, avail.Slots[1].EndTime)
	}
}

func TestDiagnosticService_EquipmentAvailabilityShiftChunking(t *testing.T) {
	clinicID := uuid.New()
	eqID := uuid.New()

	// Clinic admin configured Monday (Weekday 1) shift as 09:00 - 12:00 (3 hours = 6 slots of 30 mins)
	diagRepo := &mockDiagRepo{
		equipment: &models.DiagnosticEquipment{
			ID:         eqID,
			ClinicID:   clinicID,
			CategoryID: uuid.New(),
			Name:       "General Electric MRI",
			IsActive:   true,
			Clinic:     &models.Clinic{ID: clinicID, Status: models.ClinicStatusApproved},
		},
		slots: []models.DiagnosticWeeklySlot{
			{EquipmentID: eqID, Weekday: 1, StartTime: "09:00", EndTime: "12:00", Status: models.SlotStatusActive},
		},
		overrides: []models.DiagnosticDateSlotOverride{
			// 10:00 - 10:30 is closed for calibration
			{
				EquipmentID: eqID,
				Date:        "2026-11-30", // Monday
				StartTime:   ptr("10:00"),
				EndTime:     ptr("10:30"),
				Status:      models.OverrideStatusClosed,
				Reason:      "Kalibrovka",
			},
		},
		booked: []models.DiagnosticAppointment{
			// 11:00 - 11:30 is booked
			{
				EquipmentID:     eqID,
				AppointmentDate: "2026-11-30",
				StartTime:       "11:00",
				EndTime:         "11:30",
				Status:          models.DiagAppointmentStatusConfirmed,
			},
		},
	}
	clinicRepo := &mockClinicRepo{}
	service := NewDiagnosticService(diagRepo, clinicRepo, nil)

	avail, err := service.GetDiagnosticEquipmentAvailability(context.Background(), eqID, "2026-11-30")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 09:00-12:00 has 6 slots:
	// 09:00-09:30 (avail)
	// 09:30-10:00 (avail)
	// 10:00-10:30 (closed by override)
	// 10:30-11:00 (avail)
	// 11:00-11:30 (booked)
	// 11:30-12:00 (avail)
	// Total available = 4
	if len(avail.Slots) != 4 {
		t.Fatalf("Expected 4 available slots from 3-hour shift window, got %d", len(avail.Slots))
	}
	expectedSlots := []struct{ start, end string }{
		{"09:00", "09:30"},
		{"09:30", "10:00"},
		{"10:30", "11:00"},
		{"11:30", "12:00"},
	}
	for i, exp := range expectedSlots {
		if avail.Slots[i].StartTime != exp.start || avail.Slots[i].EndTime != exp.end {
			t.Errorf("Slot %d: expected %s-%s, got %s-%s", i, exp.start, exp.end, avail.Slots[i].StartTime, avail.Slots[i].EndTime)
		}
	}
}

func ptr(s string) *string {
	return &s
}

type mockAuditRepo struct {
	logs []*models.AuditLog
}

func (m *mockAuditRepo) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	m.logs = append(m.logs, log)
	return nil
}

func (m *mockAuditRepo) CreateVerificationReview(ctx context.Context, review *models.ClinicVerificationReview) error {
	return nil
}

func (m *mockAuditRepo) GetReviewsByClinic(ctx context.Context, clinicID uuid.UUID) ([]models.ClinicVerificationReview, error) {
	return nil, nil
}

type mockOutboxRepo struct {
	items []*models.NotificationOutbox
}

func (m *mockOutboxRepo) Enqueue(ctx context.Context, item *models.NotificationOutbox) error {
	m.items = append(m.items, item)
	return nil
}

func (m *mockOutboxRepo) GetPending(ctx context.Context, limit int) ([]models.NotificationOutbox, error) {
	return nil, nil
}

func (m *mockOutboxRepo) MarkDelivered(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockOutboxRepo) MarkFailed(ctx context.Context, id uuid.UUID, lastError string, canRetry bool, nextRetryDelay time.Duration) error {
	return nil
}

func (m *mockOutboxRepo) MarkSkipped(ctx context.Context, id uuid.UUID, reason string) error {
	return nil
}

func TestDiagnosticService_BookingAndCancelEnqueuesOutboxNotification(t *testing.T) {
	clinicID := uuid.New()
	eqID := uuid.New()
	svcID := uuid.New()
	patientID := uuid.New()

	diagRepo := &mockDiagRepo{
		equipment: &models.DiagnosticEquipment{
			ID:         eqID,
			ClinicID:   clinicID,
			CategoryID: uuid.New(),
			Name:       "Tesla MRI 3.0T",
			IsActive:   true,
			Clinic: &models.Clinic{
				ID:     clinicID,
				Name:   "Shox Med Center",
				Status: models.ClinicStatusApproved,
			},
		},
		slots: []models.DiagnosticWeeklySlot{
			{EquipmentID: eqID, Weekday: 2, StartTime: "09:00", EndTime: "09:30", Status: models.SlotStatusActive},
		},
	}
	clinicRepo := &mockClinicRepo{}
	auditRepo := &mockAuditRepo{}
	outboxRepo := &mockOutboxRepo{}

	service := NewDiagnosticService(diagRepo, clinicRepo, auditRepo, outboxRepo)

	req := &dto.CreateDiagnosticAppointmentRequest{
		ClinicID:                clinicID,
		ServiceID:               svcID,
		EquipmentID:             eqID,
		Date:                    "2026-12-01",
		StartTime:               "09:00",
		EndTime:                 "09:30",
		PreparationAcknowledged: true,
	}

	res, err := service.BookDiagnosticAppointment(context.Background(), patientID, "", req)
	if err != nil {
		t.Fatalf("BookDiagnosticAppointment failed: %v", err)
	}

	if res.ID == uuid.Nil {
		t.Errorf("Expected created appointment to have non-nil UUID, got nil")
	}

	if len(auditRepo.logs) != 1 {
		t.Fatalf("Expected 1 audit log created, got %d", len(auditRepo.logs))
	}
	if auditRepo.logs[0].Action != "BOOK_DIAGNOSTIC_APPOINTMENT" {
		t.Errorf("Expected action BOOK_DIAGNOSTIC_APPOINTMENT, got %s", auditRepo.logs[0].Action)
	}

	if len(outboxRepo.items) != 1 {
		t.Fatalf("Expected 1 outbox notification enqueued, got %d", len(outboxRepo.items))
	}

	createdNotification := outboxRepo.items[0]
	if createdNotification.EventType != "diagnostic_appointment_created" {
		t.Errorf("Expected EventType diagnostic_appointment_created, got %s", createdNotification.EventType)
	}
	if createdNotification.TargetRole != "clinic" {
		t.Errorf("Expected TargetRole clinic, got %s", createdNotification.TargetRole)
	}
	if createdNotification.TargetClinicID == nil || *createdNotification.TargetClinicID != clinicID {
		t.Errorf("Expected TargetClinicID %s, got %v", clinicID, createdNotification.TargetClinicID)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(createdNotification.Payload), &payload); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}
	if payload["clinic_name"] != "Shox Med Center" {
		t.Errorf("Expected clinic_name 'Shox Med Center', got %v", payload["clinic_name"])
	}
	if payload["equipment_name"] != "Tesla MRI 3.0T" {
		t.Errorf("Expected equipment_name 'Tesla MRI 3.0T', got %v", payload["equipment_name"])
	}

	// Now cancel the appointment
	cancelReq := &dto.CancelDiagnosticAppointmentRequest{
		Reason: "Vaqtim o'zgardi",
	}
	_, err = service.CancelPatientDiagnosticAppointment(context.Background(), patientID, res.ID, cancelReq)
	if err != nil {
		t.Fatalf("CancelPatientDiagnosticAppointment failed: %v", err)
	}

	if len(outboxRepo.items) != 2 {
		t.Fatalf("Expected 2 outbox notifications enqueued, got %d", len(outboxRepo.items))
	}

	cancelNotification := outboxRepo.items[1]
	if cancelNotification.EventType != "diagnostic_appointment_cancelled" {
		t.Errorf("Expected EventType diagnostic_appointment_cancelled, got %s", cancelNotification.EventType)
	}
	var cancelPayload map[string]any
	if err := json.Unmarshal([]byte(cancelNotification.Payload), &cancelPayload); err != nil {
		t.Fatalf("Failed to unmarshal cancel payload: %v", err)
	}
	if cancelPayload["cancelled_by"] != "Bemor" {
		t.Errorf("Expected cancelled_by 'Bemor', got %v", cancelPayload["cancelled_by"])
	}
	if cancelPayload["reason"] != "Vaqtim o'zgardi" {
		t.Errorf("Expected reason 'Vaqtim o'zgardi', got %v", cancelPayload["reason"])
	}
}

func TestDiagnosticService_ListCategories_21Categories(t *testing.T) {
	all21Categories := []models.DiagnosticCategory{
		// 7 mavjud
		{ID: uuid.New(), Name: "MRT", Slug: "mrt"},
		{ID: uuid.New(), Name: "MSKT", Slug: "mskt"},
		{ID: uuid.New(), Name: "UZI", Slug: "uzi"},
		{ID: uuid.New(), Name: "Rentgen", Slug: "rentgen"},
		{ID: uuid.New(), Name: "EKG va EXO-KG", Slug: "ekg-exo-kg"},
		{ID: uuid.New(), Name: "Laboratoriya", Slug: "laboratoriya"},
		{ID: uuid.New(), Name: "Endoskopiya", Slug: "endoskopiya"},
		// 14 yangi
		{ID: uuid.New(), Name: "Mammografiya", Slug: "mammografiya"},
		{ID: uuid.New(), Name: "Densitometriya", Slug: "densitometriya"},
		{ID: uuid.New(), Name: "Fluorografiya", Slug: "fluorografiya"},
		{ID: uuid.New(), Name: "EEG", Slug: "eeg"},
		{ID: uuid.New(), Name: "ENMG", Slug: "enmg"},
		{ID: uuid.New(), Name: "Dopplerografiya", Slug: "dopplerografiya"},
		{ID: uuid.New(), Name: "Holter monitoring", Slug: "holter-monitoring"},
		{ID: uuid.New(), Name: "Spirometriya", Slug: "spirometriya"},
		{ID: uuid.New(), Name: "Kolposkopiya", Slug: "kolposkopiya"},
		{ID: uuid.New(), Name: "Audiometriya", Slug: "audiometriya"},
		{ID: uuid.New(), Name: "Oftalmologik diagnostika", Slug: "oftalmologik-diagnostika"},
		{ID: uuid.New(), Name: "PET/KT", Slug: "pet-kt"},
		{ID: uuid.New(), Name: "Polisomnografiya", Slug: "polisomnografiya"},
		{ID: uuid.New(), Name: "Scintigrafiya", Slug: "scintigrafiya"},
	}

	diagRepo := &mockDiagRepo{
		categories: all21Categories,
	}
	clinicRepo := &mockClinicRepo{}
	svc := NewDiagnosticService(diagRepo, clinicRepo, nil)

	res, err := svc.ListCategories(context.Background())
	if err != nil {
		t.Fatalf("ListCategories failed: %v", err)
	}

	if len(res.Categories) != 21 {
		t.Fatalf("Expected 21 categories, got %d", len(res.Categories))
	}

	catSlugs := make(map[string]bool)
	for _, c := range res.Categories {
		catSlugs[c.Slug] = true
	}

	expectedSlugs := []string{
		"mrt", "mskt", "uzi", "rentgen", "ekg-exo-kg", "laboratoriya", "endoskopiya",
		"mammografiya", "densitometriya", "fluorografiya", "eeg", "enmg",
		"dopplerografiya", "holter-monitoring", "spirometriya", "kolposkopiya",
		"audiometriya", "oftalmologik-diagnostika", "pet-kt", "polisomnografiya", "scintigrafiya",
	}

	for _, slug := range expectedSlugs {
		if !catSlugs[slug] {
			t.Errorf("Expected category slug '%s' not found", slug)
		}
	}
}
