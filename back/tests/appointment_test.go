package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"qabulgo-backend/internal/config"
	"qabulgo-backend/internal/database"
	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/handlers"
	"qabulgo-backend/internal/models"
	"qabulgo-backend/internal/repositories"
	"qabulgo-backend/internal/routes"
	"qabulgo-backend/internal/services"
	"qabulgo-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TestAppointmentApp struct {
	Router             *gin.Engine
	UserRepo           *MockUserRepository
	ClinicRepo         *MockClinicRepository
	DoctorRepo         *MockDoctorRepository
	SpecialtyRepo      *MockSpecialtyRepository
	ScheduleRepo       *MockScheduleRepository
	AppointmentRepo    *MockAppointmentRepository
	AuditRepo          *MockAuditRepository
	Config             *config.Config
	TokenService       services.TokenService
	ScheduleService    services.ScheduleService
	AppointmentService services.AppointmentService
}

func setupAppointmentTestApp() *TestAppointmentApp {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		AppEnv:          "testing",
		Port:            "8080",
		JWTSecret:       "super_secret_test_key_minimum_32_characters_long!",
		JWTIssuer:       "qabulgo-api",
		JWTAudience:     "qabulgo-web",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		FrontendURL:     "http://localhost:5173",
		TermsVersion:    "v1.0",
		EmailMock:       true,
		RateLimitRPS:    1000,
		RateLimitBurst:  2000,
	}

	userRepo := NewMockUserRepository()
	sessionRepo := NewMockRefreshSessionRepository()
	tokenRepo := NewMockVerificationTokenRepository()
	emailService := NewMockEmailService()
	clinicRepo := NewMockClinicRepository()
	doctorRepo := NewMockDoctorRepository()
	specialtyRepo := NewMockSpecialtyRepository()
	auditRepo := NewMockAuditRepository()
	scheduleRepo := NewMockScheduleRepository(clinicRepo, doctorRepo)
	appointmentRepo := NewMockAppointmentRepository(userRepo, clinicRepo, doctorRepo)

	tokenService := services.NewTokenService(cfg, sessionRepo, userRepo)
	authService := services.NewAuthService(cfg, userRepo, tokenRepo, tokenService, emailService)
	patientService := services.NewPatientService(userRepo)
	clinicService := services.NewClinicService(clinicRepo, doctorRepo, userRepo, auditRepo)
	doctorService := services.NewDoctorService(doctorRepo, clinicRepo, specialtyRepo, auditRepo)
	adminService := services.NewAdminService(clinicRepo, doctorRepo, auditRepo)
	specialtyService := services.NewSpecialtyService(specialtyRepo)
	storageService := services.NewLocalStorageService("uploads_test", "/uploads")
	publicSearchRepo := NewMockPublicSearchRepository(clinicRepo, doctorRepo, specialtyRepo)
	publicSearchService := services.NewPublicSearchService(publicSearchRepo, doctorRepo)

	scheduleService := services.NewScheduleService(scheduleRepo, clinicRepo, auditRepo, appointmentRepo)
	appointmentService := services.NewAppointmentService(appointmentRepo, scheduleRepo, clinicRepo, auditRepo)

	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(cfg, authService, tokenService)
	patientHandler := handlers.NewPatientHandler(patientService)
	clinicHandler := handlers.NewClinicHandler(clinicService)
	doctorHandler := handlers.NewDoctorHandler(doctorService)
	adminHandler := handlers.NewAdminHandler(adminService)
	specialtyHandler := handlers.NewSpecialtyHandler(specialtyService)
	uploadHandler := handlers.NewUploadHandler(storageService)
	publicSearchHandler := handlers.NewPublicSearchHandler(publicSearchService)
	scheduleHandler := handlers.NewScheduleHandler(scheduleService)
	appointmentHandler := handlers.NewAppointmentHandler(appointmentService)

	router := routes.SetupRouter(&routes.RouterConfig{
		Config:              cfg,
		HealthHandler:       healthHandler,
		AuthHandler:         authHandler,
		PatientHandler:      patientHandler,
		ClinicHandler:       clinicHandler,
		DoctorHandler:       doctorHandler,
		AdminHandler:        adminHandler,
		SpecialtyHandler:    specialtyHandler,
		UploadHandler:       uploadHandler,
		PublicSearchHandler: publicSearchHandler,
		ScheduleHandler:     scheduleHandler,
		AppointmentHandler:  appointmentHandler,
	})

	return &TestAppointmentApp{
		Router:             router,
		UserRepo:           userRepo,
		ClinicRepo:         clinicRepo,
		DoctorRepo:         doctorRepo,
		SpecialtyRepo:      specialtyRepo,
		ScheduleRepo:       scheduleRepo,
		AppointmentRepo:    appointmentRepo,
		AuditRepo:          auditRepo,
		Config:             cfg,
		TokenService:       tokenService,
		ScheduleService:    scheduleService,
		AppointmentService: appointmentService,
	}
}

// Helper to seed a standard environment with clinic, doctor, schedule, and patient
type testFixture struct {
	Patient     *models.User
	PatientTok  string
	Owner       *models.User
	OwnerTok    string
	Clinic      *models.Clinic
	Doctor      *models.Doctor
	ClinicDoc   *models.ClinicDoctor
	Specialty   *models.Specialty
	TargetDate  string // 2026-11-09 (Monday, Weekday=1)
	StartTime   string // 10:00
	EndTime     string // 11:00
}

func setupFixture(app *TestAppointmentApp) *testFixture {
	ctx := context.Background()

	// 1. Patient
	patient := &models.User{
		ID:        uuid.New(),
		FirstName: "Alisher",
		LastName:  "Navoiy",
		Email:     fmt.Sprintf("patient_%s@example.com", uuid.New().String()[:8]),
		Phone:     fmt.Sprintf("+99890%07d", time.Now().UnixNano()%10000000),
		Role:      models.RolePatient,
		Status:    models.StatusActive,
	}
	_ = app.UserRepo.Create(ctx, patient)
	pTok, _ := utils.GenerateAccessToken(patient.ID, string(patient.Role), 15*time.Minute, app.Config.JWTSecret, app.Config.JWTIssuer, app.Config.JWTAudience)

	// 2. Clinic Owner
	owner := &models.User{
		ID:        uuid.New(),
		FirstName: "Jamshid",
		LastName:  "Karimov",
		Email:     fmt.Sprintf("owner_%s@example.com", uuid.New().String()[:8]),
		Phone:     fmt.Sprintf("+99891%07d", time.Now().UnixNano()%10000000),
		Role:      models.RoleClinicOwner,
		Status:    models.StatusActive,
	}
	_ = app.UserRepo.Create(ctx, owner)
	oTok, _ := utils.GenerateAccessToken(owner.ID, string(owner.Role), 15*time.Minute, app.Config.JWTSecret, app.Config.JWTIssuer, app.Config.JWTAudience)

	// 3. Approved Clinic
	clinic := &models.Clinic{
		ID:       uuid.New(),
		Name:     "Shifo Nur Markazi",
		Region:   "Toshkent shahri",
		District: "Yunusobod",
		Address:  "Amir Temur ko‘chasi, 45",
		Phone:    "+998712001122",
		Status:   models.ClinicStatusApproved,
		OwnerID:  owner.ID,
	}
	_ = app.ClinicRepo.Create(ctx, clinic, []models.ClinicWorkingHour{
		{ClinicID: clinic.ID, Weekday: 1, OpeningTime: "08:00", ClosingTime: "20:00"},
	})

	// 4. Specialty
	specialty := &models.Specialty{
		ID:   uuid.New(),
		Name: "Kardiologiya",
		Slug: "kardiologiya",
	}
	_ = app.SpecialtyRepo.Create(ctx, specialty)

	// 5. Doctor & ClinicDoctor link
	doctor := &models.Doctor{
		ID:              uuid.New(),
		FirstName:       "Jasur",
		LastName:        "Toshev",
		ExperienceYears: 12,
	}
	cd, _ := app.DoctorRepo.CreateDoctorWithClinic(ctx, clinic.ID, doctor, specialty.ID, 250000, 60, true)
	cd.Clinic = clinic
	cd.Doctor = doctor
	cd.Specialty = specialty

	// 7. Weekly Slots (Monday=1: 09:00-10:00, 10:00-11:00, 11:00-12:00)
	slots := []models.DoctorWeeklySlot{
		{ClinicDoctorID: cd.ID, Weekday: 1, StartTime: "09:00", EndTime: "10:00", Status: models.SlotStatusActive},
		{ClinicDoctorID: cd.ID, Weekday: 1, StartTime: "10:00", EndTime: "11:00", Status: models.SlotStatusActive},
		{ClinicDoctorID: cd.ID, Weekday: 1, StartTime: "11:00", EndTime: "12:00", Status: models.SlotStatusActive},
	}
	_ = app.ScheduleRepo.ReplaceWeeklySlots(ctx, cd.ID, slots)

	return &testFixture{
		Patient:    patient,
		PatientTok: pTok,
		Owner:      owner,
		OwnerTok:   oTok,
		Clinic:     clinic,
		Doctor:     doctor,
		ClinicDoc:  cd,
		Specialty:  specialty,
		TargetDate: "2026-11-09", // Monday
		StartTime:  "10:00",
		EndTime:    "11:00",
	}
}

// 1. Patient successful booking → confirmed (HTTP 201)
func TestAppointment_01_PatientSuccessfulBooking(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	reqBody := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: f.StartTime,
		EndTime:   f.EndTime,
	}
	b, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+f.PatientTok)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool                       `json:"success"`
		Message string                     `json:"message"`
		Data    dto.AppointmentResponseDTO `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Errorf("expected success true")
	}
	if resp.Data.Status != "confirmed" {
		t.Errorf("expected status confirmed, got %s", resp.Data.Status)
	}
	if resp.Data.ConsultationPrice != 250000 {
		t.Errorf("expected price 250000, got %f", resp.Data.ConsultationPrice)
	}
}

// 2. Booking without authentication → 401
func TestAppointment_02_BookingWithoutAuthentication(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	reqBody := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: f.StartTime,
		EndTime:   f.EndTime,
	}
	b, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized, got %d", w.Code)
	}
}

// 3. Non-patient role booking → 403
func TestAppointment_03_NonPatientRoleBooking(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	reqBody := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: f.StartTime,
		EndTime:   f.EndTime,
	}
	b, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+f.OwnerTok) // Clinic owner token
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for non-patient, got %d", w.Code)
	}
}

// 4. Invalid clinic → 404
func TestAppointment_04_InvalidClinic(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	reqBody := dto.CreateAppointmentRequest{
		ClinicID:  uuid.New(), // random invalid clinic
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: f.StartTime,
		EndTime:   f.EndTime,
	}
	b, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+f.PatientTok)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found, got %d", w.Code)
	}
}

// 5. Unapproved clinic → error (400)
func TestAppointment_05_UnapprovedClinic(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	// Set clinic status to draft
	f.Clinic.Status = models.ClinicStatusDraft
	_ = app.ClinicRepo.Update(context.Background(), f.Clinic)

	reqBody := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: f.StartTime,
		EndTime:   f.EndTime,
	}
	b, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+f.PatientTok)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for unapproved clinic, got %d", w.Code)
	}
}

// 6. Inactive doctor in clinic → error
func TestAppointment_06_InactiveDoctor(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	f.ClinicDoc.IsActive = false

	reqBody := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: f.StartTime,
		EndTime:   f.EndTime,
	}
	b, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+f.PatientTok)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for inactive doctor, got %d", w.Code)
	}
}

// 7. Doctor not assigned to clinic → error
func TestAppointment_07_DoctorNotAssignedToClinic(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	reqBody := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  uuid.New(), // unassigned doctor
		Date:      f.TargetDate,
		StartTime: f.StartTime,
		EndTime:   f.EndTime,
	}
	b, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+f.PatientTok)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for unassigned doctor, got %d", w.Code)
	}
}

// 8. Slot not in weekly schedule → error
func TestAppointment_08_SlotNotInWeeklySchedule(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	// Doctor only works 09:00-12:00. Trying to book 15:00-16:00
	reqBody := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: "15:00",
		EndTime:   "16:00",
	}
	b, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+f.PatientTok)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for slot not in weekly schedule, got %d: %s", w.Code, w.Body.String())
	}
}

// 9. Closed date override → error
func TestAppointment_09_ClosedDateOverride(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	// Admin closes 10:00-11:00 on target date
	s10 := "10:00"
	e11 := "11:00"
	_ = app.ScheduleRepo.CreateDateOverride(context.Background(), &models.DoctorDateSlotOverride{
		ClinicDoctorID: f.ClinicDoc.ID,
		Date:           f.TargetDate,
		StartTime:      &s10,
		EndTime:        &e11,
		Status:         models.OverrideStatusClosed,
	})

	reqBody := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: "10:00",
		EndTime:   "11:00",
	}
	b, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+f.PatientTok)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for closed override, got %d", w.Code)
	}
}

// 10. Past date/time booking → error
func TestAppointment_10_PastDateTimeBooking(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	reqBody := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  f.Doctor.ID,
		Date:      "2020-01-06", // Past Monday
		StartTime: "10:00",
		EndTime:   "11:00",
	}
	b, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+f.PatientTok)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for past date/time booking, got %d", w.Code)
	}
}

// 11. Double booking → 409 Conflict
func TestAppointment_11_DoubleBookingConflict(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	// First booking
	reqBody := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: "10:00",
		EndTime:   "11:00",
	}
	b, _ := json.Marshal(reqBody)

	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req1.Header.Set("Authorization", "Bearer "+f.PatientTok)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	app.Router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("first booking failed: %d", w1.Code)
	}

	// Create second patient
	patient2 := &models.User{
		ID:        uuid.New(),
		FirstName: "Bobur",
		LastName:  "Mirzo",
		Email:     "bobur@example.com",
		Phone:     "+998901234599",
		Role:      models.RolePatient,
		Status:    models.StatusActive,
	}
	_ = app.UserRepo.Create(context.Background(), patient2)
	p2Tok, _ := utils.GenerateAccessToken(patient2.ID, string(patient2.Role), 15*time.Minute, app.Config.JWTSecret, app.Config.JWTIssuer, app.Config.JWTAudience)

	// Second booking for identical slot
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req2.Header.Set("Authorization", "Bearer "+p2Tok)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	app.Router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict on double-booking, got %d: %s", w2.Code, w2.Body.String())
	}
}

// 12. Parallel booking requests → exactly 1 succeeds
func TestAppointment_12_ParallelBookingRequests(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	concurrentUsers := 10
	var successCount int32
	var conflictCount int32

	var wg sync.WaitGroup
	wg.Add(concurrentUsers)

	for i := 0; i < concurrentUsers; i++ {
		go func(idx int) {
			defer wg.Done()

			patient := &models.User{
				ID:        uuid.New(),
				FirstName: fmt.Sprintf("Patient%d", idx),
				LastName:  "Test",
				Email:     fmt.Sprintf("patient_%d_%s@example.com", idx, uuid.New().String()[:6]),
				Phone:     fmt.Sprintf("+99899%07d", (idx*100000)+12345),
				Role:      models.RolePatient,
				Status:    models.StatusActive,
			}
			_ = app.UserRepo.Create(context.Background(), patient)
			tok, _ := utils.GenerateAccessToken(patient.ID, string(patient.Role), 15*time.Minute, app.Config.JWTSecret, app.Config.JWTIssuer, app.Config.JWTAudience)

			reqBody := dto.CreateAppointmentRequest{
				ClinicID:  f.Clinic.ID,
				DoctorID:  f.Doctor.ID,
				Date:      f.TargetDate,
				StartTime: "09:00",
				EndTime:   "10:00",
			}
			b, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
			req.Header.Set("Authorization", "Bearer "+tok)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			app.Router.ServeHTTP(w, req)

			if w.Code == http.StatusCreated {
				atomic.AddInt32(&successCount, 1)
			} else if w.Code == http.StatusConflict {
				atomic.AddInt32(&conflictCount, 1)
			}
		}(i)
	}

	wg.Wait()

	if successCount != 1 {
		t.Errorf("expected exactly 1 successful booking, got %d", successCount)
	}
	if conflictCount != int32(concurrentUsers-1) {
		t.Errorf("expected %d conflicts, got %d", concurrentUsers-1, conflictCount)
	}
}

// 13. Duplicate Idempotency-Key → return existing appointment without duplicate insert
func TestAppointment_13_DuplicateIdempotencyKey(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	idempotencyKey := "idem-key-12345"

	reqBody := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: "10:00",
		EndTime:   "11:00",
	}
	b, _ := json.Marshal(reqBody)

	// First request with idempotency key
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req1.Header.Set("Authorization", "Bearer "+f.PatientTok)
	req1.Header.Set("Idempotency-Key", idempotencyKey)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	app.Router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("first request failed: %d", w1.Code)
	}

	// Second request with exact same idempotency key and payload
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req2.Header.Set("Authorization", "Bearer "+f.PatientTok)
	req2.Header.Set("Idempotency-Key", idempotencyKey)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	app.Router.ServeHTTP(w2, req2)

	// Must succeed idempotently
	if w2.Code != http.StatusCreated && w2.Code != http.StatusOK {
		t.Fatalf("expected 201/200 idempotent response, got %d", w2.Code)
	}

	// Ensure only 1 booking was created in total
	if len(app.AppointmentRepo.appointments) != 1 {
		t.Errorf("expected 1 appointment in repo, got %d", len(app.AppointmentRepo.appointments))
	}
}

// 14. Same Idempotency-Key, different payload → 409 Conflict
func TestAppointment_14_SameKeyDifferentPayload(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	idempotencyKey := "idem-conflict-key"

	reqBody1 := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: "10:00",
		EndTime:   "11:00",
	}
	b1, _ := json.Marshal(reqBody1)

	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b1))
	req1.Header.Set("Authorization", "Bearer "+f.PatientTok)
	req1.Header.Set("Idempotency-Key", idempotencyKey)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	app.Router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("first request failed: %d", w1.Code)
	}

	// Second request with different time slot but same key
	reqBody2 := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: "11:00",
		EndTime:   "12:00",
	}
	b2, _ := json.Marshal(reqBody2)

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b2))
	req2.Header.Set("Authorization", "Bearer "+f.PatientTok)
	req2.Header.Set("Idempotency-Key", idempotencyKey)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	app.Router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict for same key with different payload, got %d", w2.Code)
	}
}

// 15. Patient can view own appointments (GET /my)
func TestAppointment_15_PatientCanViewOwnAppointments(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	// Create booking
	reqBody := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: "10:00",
		EndTime:   "11:00",
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+f.PatientTok)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("booking failed: %d", w.Code)
	}

	// List patient appointments
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/appointments/my?timeframe=upcoming", nil)
	listReq.Header.Set("Authorization", "Bearer "+f.PatientTok)
	listW := httptest.NewRecorder()
	app.Router.ServeHTTP(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", listW.Code)
	}

	var resp struct {
		Success bool                           `json:"success"`
		Data    dto.AppointmentListResponseDTO `json:"data"`
	}
	_ = json.Unmarshal(listW.Body.Bytes(), &resp)
	if resp.Data.Total != 1 {
		t.Errorf("expected 1 appointment, got %d", resp.Data.Total)
	}
}

// 16. Patient cannot view others' appointments (IDOR 403)
func TestAppointment_16_PatientCannotViewOthersAppointments(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	// Patient 1 books
	reqBody := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: "10:00",
		EndTime:   "11:00",
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+f.PatientTok)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	var resp struct {
		Data dto.AppointmentResponseDTO `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	apptID := resp.Data.ID

	// Create Patient 2
	p2 := &models.User{
		ID:        uuid.New(),
		FirstName: "Stranger",
		LastName:  "Danger",
		Email:     "stranger@example.com",
		Phone:     "+998909998877",
		Role:      models.RolePatient,
		Status:    models.StatusActive,
	}
	_ = app.UserRepo.Create(context.Background(), p2)
	p2Tok, _ := utils.GenerateAccessToken(p2.ID, string(p2.Role), 15*time.Minute, app.Config.JWTSecret, app.Config.JWTIssuer, app.Config.JWTAudience)

	// Patient 2 attempts to fetch Patient 1's appointment
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/appointments/"+apptID.String(), nil)
	getReq.Header.Set("Authorization", "Bearer "+p2Tok)
	getW := httptest.NewRecorder()
	app.Router.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden on IDOR check, got %d", getW.Code)
	}
}

// 17. Clinic can view its appointments
func TestAppointment_17_ClinicCanViewItsAppointments(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	// Patient books
	reqBody := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: "10:00",
		EndTime:   "11:00",
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+f.PatientTok)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	// Clinic views appointments
	cReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/clinics/%s/appointments", f.Clinic.ID), nil)
	cReq.Header.Set("Authorization", "Bearer "+f.OwnerTok)
	cW := httptest.NewRecorder()
	app.Router.ServeHTTP(cW, cReq)

	if cW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", cW.Code, cW.Body.String())
	}

	var resp struct {
		Data dto.AppointmentListResponseDTO `json:"data"`
	}
	_ = json.Unmarshal(cW.Body.Bytes(), &resp)
	if resp.Data.Total != 1 {
		t.Errorf("expected 1 appointment for clinic, got %d", resp.Data.Total)
	}
}

// 18. Clinic cannot view another clinic's appointments
func TestAppointment_18_ClinicCannotViewAnotherClinicsAppointments(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	// Create Clinic B with Owner B
	ownerB := &models.User{
		ID:        uuid.New(),
		FirstName: "OwnerB",
		LastName:  "User",
		Email:     "ownerb@example.com",
		Phone:     "+998905554433",
		Role:      models.RoleClinicOwner,
		Status:    models.StatusActive,
	}
	_ = app.UserRepo.Create(context.Background(), ownerB)
	ownerBTok, _ := utils.GenerateAccessToken(ownerB.ID, string(ownerB.Role), 15*time.Minute, app.Config.JWTSecret, app.Config.JWTIssuer, app.Config.JWTAudience)

	// Owner B attempts to view Clinic A's appointments
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/clinics/%s/appointments", f.Clinic.ID), nil)
	req.Header.Set("Authorization", "Bearer "+ownerBTok)
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for cross-clinic view, got %d", w.Code)
	}
}

// 19. Patient cancellation → slot available again
func TestAppointment_19_PatientCancellationRestoresSlot(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	// 1. Book appointment
	reqBody := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: "10:00",
		EndTime:   "11:00",
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+f.PatientTok)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	var bookResp struct {
		Data dto.AppointmentResponseDTO `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &bookResp)
	apptID := bookResp.Data.ID

	// 2. Cancel appointment
	cancelBody := dto.CancelAppointmentRequest{Reason: "Rejalarim o‘zgarib qoldi"}
	cb, _ := json.Marshal(cancelBody)
	cReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/appointments/%s/cancel", apptID), bytes.NewReader(cb))
	cReq.Header.Set("Authorization", "Bearer "+f.PatientTok)
	cReq.Header.Set("Content-Type", "application/json")
	cW := httptest.NewRecorder()
	app.Router.ServeHTTP(cW, cReq)

	if cW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on cancel, got %d", cW.Code)
	}

	// 3. Verify availability has the slot again
	availReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/public/clinics/%s/doctors/%s/availability?date=%s", f.Clinic.ID, f.Doctor.ID, f.TargetDate), nil)
	availW := httptest.NewRecorder()
	app.Router.ServeHTTP(availW, availReq)

	var availResp struct {
		Data dto.PublicAvailabilityResponse `json:"data"`
	}
	_ = json.Unmarshal(availW.Body.Bytes(), &availResp)

	found := false
	for _, s := range availResp.Data.Slots {
		if s.StartTime == "10:00" && s.EndTime == "11:00" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("cancelled slot was not restored in availability")
	}
}

// 20. Clinic cancellation → slot available again
func TestAppointment_20_ClinicCancellationRestoresSlot(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	// Book
	reqBody := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: "10:00",
		EndTime:   "11:00",
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+f.PatientTok)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	var bookResp struct {
		Data dto.AppointmentResponseDTO `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &bookResp)
	apptID := bookResp.Data.ID

	// Clinic cancels
	cancelBody := dto.CancelAppointmentRequest{Reason: "Shifokor kasal bo‘lib qoldi"}
	cb, _ := json.Marshal(cancelBody)
	cReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/clinics/%s/appointments/%s/cancel", f.Clinic.ID, apptID), bytes.NewReader(cb))
	cReq.Header.Set("Authorization", "Bearer "+f.OwnerTok)
	cReq.Header.Set("Content-Type", "application/json")
	cW := httptest.NewRecorder()
	app.Router.ServeHTTP(cW, cReq)

	if cW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on clinic cancel, got %d: %s", cW.Code, cW.Body.String())
	}
}

// 21. Completed appointment status
func TestAppointment_21_CompletedAppointmentStatus(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	// Directly insert a past confirmed appointment for testing complete transition
	appt := &models.Appointment{
		ID:                uuid.New(),
		PatientID:         f.Patient.ID,
		ClinicID:          f.Clinic.ID,
		DoctorID:          f.Doctor.ID,
		ClinicDoctorID:    f.ClinicDoc.ID,
		AppointmentDate:   "2026-10-01", // Past date
		StartTime:         "10:00",
		EndTime:           "11:00",
		Status:            models.AppointmentStatusConfirmed,
		ConsultationPrice: 250000,
	}
	_ = app.AppointmentRepo.Create(context.Background(), appt)

	// Clinic completes appointment
	cReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/clinics/%s/appointments/%s/complete", f.Clinic.ID, appt.ID), nil)
	cReq.Header.Set("Authorization", "Bearer "+f.OwnerTok)
	cW := httptest.NewRecorder()
	app.Router.ServeHTTP(cW, cReq)

	if cW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", cW.Code, cW.Body.String())
	}

	var resp struct {
		Data dto.AppointmentResponseDTO `json:"data"`
	}
	_ = json.Unmarshal(cW.Body.Bytes(), &resp)
	if resp.Data.Status != "completed" {
		t.Errorf("expected completed status, got %s", resp.Data.Status)
	}
}

// 22. No-show appointment status
func TestAppointment_22_NoShowAppointmentStatus(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	appt := &models.Appointment{
		ID:                uuid.New(),
		PatientID:         f.Patient.ID,
		ClinicID:          f.Clinic.ID,
		DoctorID:          f.Doctor.ID,
		ClinicDoctorID:    f.ClinicDoc.ID,
		AppointmentDate:   "2026-10-01",
		StartTime:         "10:00",
		EndTime:           "11:00",
		Status:            models.AppointmentStatusConfirmed,
		ConsultationPrice: 250000,
	}
	_ = app.AppointmentRepo.Create(context.Background(), appt)

	// Clinic marks no-show
	cReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/clinics/%s/appointments/%s/no-show", f.Clinic.ID, appt.ID), nil)
	cReq.Header.Set("Authorization", "Bearer "+f.OwnerTok)
	cW := httptest.NewRecorder()
	app.Router.ServeHTTP(cW, cReq)

	if cW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", cW.Code, cW.Body.String())
	}

	var resp struct {
		Data dto.AppointmentResponseDTO `json:"data"`
	}
	_ = json.Unmarshal(cW.Body.Bytes(), &resp)
	if resp.Data.Status != "no_show" {
		t.Errorf("expected no_show status, got %s", resp.Data.Status)
	}
}

// 23. Invalid status transitions & premature complete/no-show
func TestAppointment_23_InvalidStatusTransitions(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	// Future appointment cannot be completed before its start time
	reqBody := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: "10:00",
		EndTime:   "11:00",
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+f.PatientTok)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	var bookResp struct {
		Data dto.AppointmentResponseDTO `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &bookResp)
	apptID := bookResp.Data.ID

	// Premature complete attempt
	cReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/clinics/%s/appointments/%s/complete", f.Clinic.ID, apptID), nil)
	cReq.Header.Set("Authorization", "Bearer "+f.OwnerTok)
	cW := httptest.NewRecorder()
	app.Router.ServeHTTP(cW, cReq)

	if cW.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for premature completion, got %d", cW.Code)
	}
}

// 24. Availability excludes booked slots
func TestAppointment_24_AvailabilityExcludesBookedSlots(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	// Book 10:00-11:00
	reqBody := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: "10:00",
		EndTime:   "11:00",
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+f.PatientTok)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	// Check availability: slots 09:00-10:00 and 11:00-12:00 must exist, 10:00-11:00 must NOT
	availReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/public/clinics/%s/doctors/%s/availability?date=%s", f.Clinic.ID, f.Doctor.ID, f.TargetDate), nil)
	availW := httptest.NewRecorder()
	app.Router.ServeHTTP(availW, availReq)

	var availResp struct {
		Data dto.PublicAvailabilityResponse `json:"data"`
	}
	_ = json.Unmarshal(availW.Body.Bytes(), &availResp)

	for _, s := range availResp.Data.Slots {
		if s.StartTime == "10:00" && s.EndTime == "11:00" {
			t.Errorf("booked slot 10:00-11:00 should NOT be present in availability")
		}
	}
	if len(availResp.Data.Slots) != 2 {
		t.Errorf("expected 2 remaining slots, got %d", len(availResp.Data.Slots))
	}
}

// 25. Availability restores cancelled slots
func TestAppointment_25_AvailabilityRestoresCancelledSlots(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	// Book
	reqBody := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: "10:00",
		EndTime:   "11:00",
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+f.PatientTok)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	var bookResp struct {
		Data dto.AppointmentResponseDTO `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &bookResp)
	apptID := bookResp.Data.ID

	// Cancel
	cReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/appointments/%s/cancel", apptID), nil)
	cReq.Header.Set("Authorization", "Bearer "+f.PatientTok)
	cW := httptest.NewRecorder()
	app.Router.ServeHTTP(cW, cReq)

	// Availability must now show all 3 slots
	availReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/public/clinics/%s/doctors/%s/availability?date=%s", f.Clinic.ID, f.Doctor.ID, f.TargetDate), nil)
	availW := httptest.NewRecorder()
	app.Router.ServeHTTP(availW, availReq)

	var availResp struct {
		Data dto.PublicAvailabilityResponse `json:"data"`
	}
	_ = json.Unmarshal(availW.Body.Bytes(), &availResp)
	if len(availResp.Data.Slots) != 3 {
		t.Errorf("expected 3 available slots after cancel, got %d", len(availResp.Data.Slots))
	}
}

// 26. Schedule changes cannot invalidate confirmed bookings (409 Conflict)
func TestAppointment_26_ScheduleChangesCannotInvalidateConfirmedBookings(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	// Book 10:00-11:00 on Monday
	reqBody := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: "10:00",
		EndTime:   "11:00",
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+f.PatientTok)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	// Clinic admin tries to save a weekly schedule that REMOVES 10:00-11:00 (keeps only 09:00-10:00)
	putBody := dto.SaveWeeklyScheduleRequest{
		Slots: []dto.WeeklySlotItemRequest{
			{Weekday: 1, StartTime: "09:00", EndTime: "10:00", Status: "active"},
		},
	}
	pb, _ := json.Marshal(putBody)

	putReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/clinics/%s/doctors/%s/schedule", f.Clinic.ID, f.Doctor.ID), bytes.NewReader(pb))
	putReq.Header.Set("Authorization", "Bearer "+f.OwnerTok)
	putReq.Header.Set("Content-Type", "application/json")
	putW := httptest.NewRecorder()
	app.Router.ServeHTTP(putW, putReq)

	if putW.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict when removing slot with active confirmed booking, got %d: %s", putW.Code, putW.Body.String())
	}
}

// 27. Date override cannot invalidate confirmed bookings (409 Conflict)
func TestAppointment_27_DateOverrideCannotInvalidateConfirmedBookings(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	// Book 10:00-11:00
	reqBody := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: "10:00",
		EndTime:   "11:00",
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+f.PatientTok)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	// Clinic admin tries to mark that date as DayOff
	overrideReq := dto.CreateDateOverrideRequest{
		Date:   f.TargetDate,
		Status: "day_off",
		Reason: "Shifokor seminarda",
	}
	ob, _ := json.Marshal(overrideReq)

	oReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/clinics/%s/doctors/%s/schedule/exceptions", f.Clinic.ID, f.Doctor.ID), bytes.NewReader(ob))
	oReq.Header.Set("Authorization", "Bearer "+f.OwnerTok)
	oReq.Header.Set("Content-Type", "application/json")
	oW := httptest.NewRecorder()
	app.Router.ServeHTTP(oW, oReq)

	if oW.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict when closing date with active confirmed booking, got %d: %s", oW.Code, oW.Body.String())
	}
}

// 28. Cross-clinic doctor booking conflicts (409 Conflict)
func TestAppointment_28_CrossClinicDoctorBookingConflict(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	// Create Clinic 2 where the same doctor also works
	clinic2 := &models.Clinic{
		ID:       uuid.New(),
		Name:     "Akfa Medline",
		Region:   "Toshkent",
		District: "Olmazor",
		Address:  "Kichik Halka Yo‘li, 5",
		Phone:    "+998711234567",
		Status:   models.ClinicStatusApproved,
		OwnerID:  f.Owner.ID,
	}
	_ = app.ClinicRepo.Create(context.Background(), clinic2, []models.ClinicWorkingHour{
		{ClinicID: clinic2.ID, Weekday: 1, OpeningTime: "08:00", ClosingTime: "20:00"},
	})

	cd2, _ := app.DoctorRepo.CreateDoctorWithClinic(context.Background(), clinic2.ID, f.Doctor, f.Specialty.ID, 300000, 60, true)
	cd2.Clinic = clinic2
	cd2.Doctor = f.Doctor
	cd2.Specialty = f.Specialty

	// Add weekly schedule at Clinic 2 for 10:00-11:00 on Monday
	_ = app.ScheduleRepo.ReplaceWeeklySlots(context.Background(), cd2.ID, []models.DoctorWeeklySlot{
		{ClinicDoctorID: cd2.ID, Weekday: 1, StartTime: "10:00", EndTime: "11:00", Status: models.SlotStatusActive},
	})

	// Patient 1 books at Clinic 1 for 10:00-11:00
	req1 := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: "10:00",
		EndTime:   "11:00",
	}
	b1, _ := json.Marshal(req1)
	httpReq1 := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b1))
	httpReq1.Header.Set("Authorization", "Bearer "+f.PatientTok)
	httpReq1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	app.Router.ServeHTTP(w1, httpReq1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("first booking failed: %d", w1.Code)
	}

	// Create Patient 2 to book at Clinic 2 for same doctor and same slot
	p2 := &models.User{
		ID:        uuid.New(),
		FirstName: "Sardor",
		LastName:  "Rasulov",
		Email:     "sardor@example.com",
		Phone:     "+998904443322",
		Role:      models.RolePatient,
		Status:    models.StatusActive,
	}
	_ = app.UserRepo.Create(context.Background(), p2)
	p2Tok, _ := utils.GenerateAccessToken(p2.ID, string(p2.Role), 15*time.Minute, app.Config.JWTSecret, app.Config.JWTIssuer, app.Config.JWTAudience)

	req2 := dto.CreateAppointmentRequest{
		ClinicID:  clinic2.ID,
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: "10:00",
		EndTime:   "11:00",
	}
	b2, _ := json.Marshal(req2)
	httpReq2 := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b2))
	httpReq2.Header.Set("Authorization", "Bearer "+p2Tok)
	httpReq2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	app.Router.ServeHTTP(w2, httpReq2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict for cross-clinic doctor booking overlap, got %d: %s", w2.Code, w2.Body.String())
	}
}

// 29. Price captured from server, not client
func TestAppointment_29_PriceCapturedFromServerNotClient(t *testing.T) {
	app := setupAppointmentTestApp()
	f := setupFixture(app)

	// Client sends malicious request (even if client attempts to pass a price)
	reqBody := dto.CreateAppointmentRequest{
		ClinicID:  f.Clinic.ID,
		DoctorID:  f.Doctor.ID,
		Date:      f.TargetDate,
		StartTime: "10:00",
		EndTime:   "11:00",
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/appointments", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+f.PatientTok)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	var resp struct {
		Data dto.AppointmentResponseDTO `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	// Price must strictly match ClinicDoctor.ConsultationPrice (250000)
	if resp.Data.ConsultationPrice != 250000 {
		t.Errorf("expected price 250000 from server, got %f", resp.Data.ConsultationPrice)
	}
}

// 30. Real PostgreSQL Parallel Concurrent Booking Test (Neon DB or fallback)
func TestAppointment_30_RealPostgresParallelBooking(t *testing.T) {
	cfg, err := config.LoadConfig()
	if err != nil || cfg.DatabaseURL == "" {
		t.Skip("PostgreSQL database connection string not available, skipping live DB test")
	}

	db, err := database.ConnectDB(cfg)
	if err != nil {
		t.Skipf("Cannot reach live PostgreSQL database (%v), skipping live DB test", err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		_ = database.RunMigrationsUp(sqlDB, "../migrations")
	}

	// Create test entities in live DB
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	ownerID := uuid.New()
	testOwner := models.User{
		ID:              ownerID,
		FirstName:       "LiveDB",
		LastName:        "Owner",
		Email:           fmt.Sprintf("livedb_owner_%s@example.com", uuid.New().String()[:8]),
		Phone:           fmt.Sprintf("+99888%07d", time.Now().UnixNano()%10000000),
		PasswordHash:    "$2a$10$xyz",
		Role:            models.RoleClinicOwner,
		Status:          models.StatusActive,
		TermsAcceptedAt: time.Now(),
		TermsVersion:    "v1.0",
	}
	if err := db.WithContext(ctx).Create(&testOwner).Error; err != nil {
		t.Skipf("Live DB owner creation skipped: %v", err)
	}

	clinicID := uuid.New()
	testClinic := models.Clinic{
		ID:       clinicID,
		Name:     "Live DB Test Clinic",
		Region:   "Toshkent",
		District: "Yunusobod",
		Address:  "Test Address",
		Phone:    "+998719999999",
		Status:   models.ClinicStatusApproved,
		OwnerID:  ownerID,
	}
	if err := db.WithContext(ctx).Create(&testClinic).Error; err != nil {
		t.Skipf("Live DB clinic creation skipped: %v", err)
	}

	doctorID := uuid.New()
	testDoctor := models.Doctor{
		ID:        doctorID,
		FirstName: "LiveDB",
		LastName:  "Doctor",
	}
	if err := db.WithContext(ctx).Create(&testDoctor).Error; err != nil {
		t.Skipf("Live DB doctor creation skipped: %v", err)
	}

	var specialty models.Specialty
	if err := db.WithContext(ctx).First(&specialty).Error; err != nil {
		specialty = models.Specialty{
			ID:   uuid.New(),
			Name: "LiveDB Specialty",
			Slug: fmt.Sprintf("spec-%s", uuid.New().String()[:8]),
		}
		_ = db.WithContext(ctx).Create(&specialty).Error
	}

	clinicDoctorID := uuid.New()
	testCD := models.ClinicDoctor{
		ID:                         clinicDoctorID,
		ClinicID:                   clinicID,
		DoctorID:                   doctorID,
		SpecialtyID:                specialty.ID,
		ConsultationPrice:          150000,
		AppointmentDurationMinutes: 60,
		IsActive:                   true,
	}
	if err := db.WithContext(ctx).Create(&testCD).Error; err != nil {
		t.Skipf("Live DB clinic_doctor creation skipped: %v", err)
	}

	// Create 5 patient users
	concurrency := 5
	patientIDs := make([]uuid.UUID, concurrency)
	for i := 0; i < concurrency; i++ {
		pID := uuid.New()
		patientIDs[i] = pID
		pUser := models.User{
			ID:              pID,
			FirstName:       fmt.Sprintf("LivePatient%d", i),
			LastName:        "Tester",
			Email:           fmt.Sprintf("live_patient_%d_%s@example.com", i, uuid.New().String()[:8]),
			Phone:           fmt.Sprintf("+99877%07d", (i*100000)+(int(time.Now().UnixNano()%100000))),
			PasswordHash:    "$2a$10$xyz",
			Role:            models.RolePatient,
			Status:          models.StatusActive,
			TermsAcceptedAt: time.Now(),
			TermsVersion:    "v1.0",
		}
		_ = db.WithContext(ctx).Create(&pUser).Error
	}

	defer func() {
		_ = db.Exec("DELETE FROM appointments WHERE doctor_id = ?", doctorID).Error
		_ = db.Exec("DELETE FROM clinic_doctors WHERE id = ?", clinicDoctorID).Error
		_ = db.Exec("DELETE FROM doctors WHERE id = ?", doctorID).Error
		_ = db.Exec("DELETE FROM clinics WHERE id = ?", clinicID).Error
		_ = db.Exec("DELETE FROM users WHERE id = ?", ownerID).Error
		for _, pid := range patientIDs {
			_ = db.Exec("DELETE FROM users WHERE id = ?", pid).Error
		}
	}()

	apptRepo := repositories.NewAppointmentRepository(db)

	date := "2026-11-20"
	sTime := "10:00"
	eTime := "11:00"

	var successCount int32
	var conflictCount int32

	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			defer wg.Done()

			appt := &models.Appointment{
				ID:                uuid.New(),
				PatientID:         patientIDs[idx],
				ClinicID:          clinicID,
				DoctorID:          doctorID,
				ClinicDoctorID:    clinicDoctorID,
				AppointmentDate:   date,
				StartTime:         sTime,
				EndTime:           eTime,
				Status:            models.AppointmentStatusConfirmed,
				ConsultationPrice: 150000,
			}

			err := apptRepo.Create(context.Background(), appt)
			if err == nil {
				atomic.AddInt32(&successCount, 1)
			} else {
				atomic.AddInt32(&conflictCount, 1)
			}
		}(i)
	}

	wg.Wait()

	t.Logf("Live PostgreSQL parallel booking result: %d succeeded, %d conflicts", successCount, conflictCount)
	if successCount != 1 {
		t.Errorf("expected exactly 1 booking to succeed in live DB, got %d", successCount)
	}
}
