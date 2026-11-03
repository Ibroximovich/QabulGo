package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"qabulgo-backend/internal/config"
	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/handlers"
	"qabulgo-backend/internal/models"
	"qabulgo-backend/internal/routes"
	"qabulgo-backend/internal/services"
	"qabulgo-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TestScheduleApp struct {
	Router          *gin.Engine
	UserRepo        *MockUserRepository
	ClinicRepo      *MockClinicRepository
	DoctorRepo      *MockDoctorRepository
	ScheduleRepo    *MockScheduleRepository
	AuditRepo       *MockAuditRepository
	Config          *config.Config
	TokenService    services.TokenService
	ScheduleService services.ScheduleService
}

func setupScheduleTestApp() *TestScheduleApp {
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

	tokenService := services.NewTokenService(cfg, sessionRepo, userRepo)
	authService := services.NewAuthService(cfg, userRepo, tokenRepo, tokenService, emailService)
	patientService := services.NewPatientService(userRepo)
	clinicService := services.NewClinicService(clinicRepo, doctorRepo, userRepo, auditRepo)
	doctorService := services.NewDoctorService(doctorRepo, clinicRepo, specialtyRepo, auditRepo)
	adminService := services.NewAdminService(clinicRepo, doctorRepo, auditRepo, nil)
	specialtyService := services.NewSpecialtyService(specialtyRepo)
	storageService := services.NewLocalStorageService("uploads_test", "/uploads")
	publicSearchRepo := NewMockPublicSearchRepository(clinicRepo, doctorRepo, specialtyRepo)
	publicSearchService := services.NewPublicSearchService(publicSearchRepo, doctorRepo)
	scheduleService := services.NewScheduleService(scheduleRepo, clinicRepo, auditRepo)

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
	})

	return &TestScheduleApp{
		Router:          router,
		UserRepo:        userRepo,
		ClinicRepo:      clinicRepo,
		DoctorRepo:      doctorRepo,
		ScheduleRepo:    scheduleRepo,
		AuditRepo:       auditRepo,
		Config:          cfg,
		TokenService:    tokenService,
		ScheduleService: scheduleService,
	}
}

// Helper to seed a clinic, doctor, and clinic_doctor
func seedClinicAndDoctor(app *TestScheduleApp, clinicOwnerID uuid.UUID, clinicStatus models.ClinicStatus, isDoctorActive bool) (*models.Clinic, *models.Doctor, *models.ClinicDoctor) {
	clinicID := uuid.New()
	clinic := &models.Clinic{
		ID:          clinicID,
		Name:        "Test Clinic Med",
		OwnerID:     clinicOwnerID,
		Status:      clinicStatus,
		Address:     "Amir Temur ko'chasi 10",
		Region:      "Toshkent shahri",
		District:    "Yunusobod",
		Phone:       "+998712001122",
		Email:       "info@testmed.uz",
	}
	_ = app.ClinicRepo.Create(context.Background(), clinic, nil)

	doctorID := uuid.New()
	doctor := &models.Doctor{
		ID:              doctorID,
		FirstName:       "Alisher",
		LastName:        "Navoiy",
		Bio:             "Tajribali kardiolog",
		ExperienceYears: 10,
	}

	specID := uuid.New()
	cd, _ := app.DoctorRepo.CreateDoctorWithClinic(
		context.Background(),
		clinic.ID,
		doctor,
		specID,
		150000,
		30,
		isDoctorActive,
	)

	return clinic, doctor, cd
}

// Helper to generate auth token
func makeAuthHeader(app *TestScheduleApp, userID uuid.UUID, role models.UserRole) string {
	now := time.Now().UTC()
	user := &models.User{
		ID:              userID,
		Role:            role,
		Status:          models.StatusActive,
		Email:           fmt.Sprintf("user-%s@test.uz", userID.String()[:8]),
		FirstName:       "Admin",
		LastName:        "User",
		Phone:           fmt.Sprintf("+99890%s", userID.String()[:7]),
		TermsAcceptedAt: now,
		TermsVersion:    "v1.0",
		EmailVerifiedAt: &now,
	}
	app.UserRepo.users[userID] = user
	token, _ := utils.GenerateAccessToken(user.ID, string(role), 15*time.Minute, app.Config.JWTSecret, app.Config.JWTIssuer, app.Config.JWTAudience)
	return "Bearer " + token
}

// 1. Automatic slot generation preview (Preview calculation with breaks, does NOT persist)
func TestSchedule_GenerateSchedulePreview(t *testing.T) {
	app := setupScheduleTestApp()
	ownerID := uuid.New()
	clinic, doctor, cd := seedClinicAndDoctor(app, ownerID, models.ClinicStatusApproved, true)
	token := makeAuthHeader(app, ownerID, models.RoleClinicOwner)

	// 09:00 - 17:00, 60 min slots, break 13:00 - 14:00
	reqBody := dto.GenerateScheduleRequest{
		Weekdays:            []int{1, 2}, // Dushanba, Seshanba
		StartTime:           "09:00",
		EndTime:             "17:00",
		SlotDurationMinutes: 60,
		Breaks: []dto.TimeBreakItem{
			{StartTime: "13:00", EndTime: "14:00"},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("/api/v1/clinics/%s/doctors/%s/schedule/generate", clinic.ID, doctor.ID)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for preview, got %d: %s", w.Code, w.Body.String())
	}

	var resp dto.SuccessResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	dataBytes, _ := json.Marshal(resp.Data)
	var genResp dto.GenerateScheduleResponse
	_ = json.Unmarshal(dataBytes, &genResp)

	// 09:00-10:00, 10:00-11:00, 11:00-12:00, 12:00-13:00, 14:00-15:00, 15:00-16:00, 16:00-17:00 => 7 slots per day * 2 days = 14 slots
	if genResp.TotalSlots != 14 {
		t.Errorf("Expected 14 generated slots, got %d", genResp.TotalSlots)
	}

	// Verify preview does NOT persist into DB
	slotsInDb, _ := app.ScheduleRepo.GetWeeklySlots(context.Background(), cd.ID)
	if len(slotsInDb) != 0 {
		t.Errorf("Generate preview must not persist into DB, but found %d slots", len(slotsInDb))
	}
}

// 2. Manual and atomic save of weekly schedule (PUT) & Get Weekly Schedule
func TestSchedule_SaveAndGetWeeklySchedule(t *testing.T) {
	app := setupScheduleTestApp()
	ownerID := uuid.New()
	clinic, doctor, cd := seedClinicAndDoctor(app, ownerID, models.ClinicStatusApproved, true)
	token := makeAuthHeader(app, ownerID, models.RoleClinicOwner)

	// Save schedule manually: Monday with 3 slots, Wednesday with 2 slots
	saveReq := dto.SaveWeeklyScheduleRequest{
		Slots: []dto.WeeklySlotItemRequest{
			{Weekday: 1, StartTime: "09:00", EndTime: "10:00", Status: "active"},
			{Weekday: 1, StartTime: "11:30", EndTime: "12:30", Status: "active"},
			{Weekday: 1, StartTime: "15:00", EndTime: "16:00", Status: "active"},
			{Weekday: 3, StartTime: "14:00", EndTime: "15:00", Status: "active"},
			{Weekday: 3, StartTime: "15:00", EndTime: "16:00", Status: "active"},
		},
	}
	bodyBytes, _ := json.Marshal(saveReq)

	putURL := fmt.Sprintf("/api/v1/clinics/%s/doctors/%s/schedule", clinic.ID, doctor.ID)
	req, _ := http.NewRequest(http.MethodPut, putURL, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for save schedule, got %d: %s", w.Code, w.Body.String())
	}

	// Verify persisted
	slotsInDb, _ := app.ScheduleRepo.GetWeeklySlots(context.Background(), cd.ID)
	if len(slotsInDb) != 5 {
		t.Fatalf("Expected 5 slots in DB, got %d", len(slotsInDb))
	}

	// GET schedule
	getReq, _ := http.NewRequest(http.MethodGet, putURL, nil)
	getReq.Header.Set("Authorization", token)
	getW := httptest.NewRecorder()
	app.Router.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for get schedule, got %d: %s", getW.Code, getW.Body.String())
	}

	var resp dto.SuccessResponse
	_ = json.Unmarshal(getW.Body.Bytes(), &resp)
	dataBytes, _ := json.Marshal(resp.Data)
	var schedResp dto.DoctorWeeklyScheduleDTO
	_ = json.Unmarshal(dataBytes, &schedResp)

	if len(schedResp.Slots) != 5 {
		t.Errorf("Expected 5 total slots in schedule DTO, got %d", len(schedResp.Slots))
	}

	// Atomic update: now replace with only 1 slot on Friday
	updateReq := dto.SaveWeeklyScheduleRequest{
		Slots: []dto.WeeklySlotItemRequest{
			{Weekday: 5, StartTime: "10:00", EndTime: "11:00", Status: "active"},
		},
	}
	uBytes, _ := json.Marshal(updateReq)
	uReq, _ := http.NewRequest(http.MethodPut, putURL, bytes.NewBuffer(uBytes))
	uReq.Header.Set("Content-Type", "application/json")
	uReq.Header.Set("Authorization", token)
	uW := httptest.NewRecorder()
	app.Router.ServeHTTP(uW, uReq)

	if uW.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for atomic update, got %d: %s", uW.Code, uW.Body.String())
	}

	slotsAfterUpdate, _ := app.ScheduleRepo.GetWeeklySlots(context.Background(), cd.ID)
	if len(slotsAfterUpdate) != 1 {
		t.Errorf("Expected 1 slot after atomic replacement, got %d", len(slotsAfterUpdate))
	}
}

// 3. Overlap validation within the same day
func TestSchedule_OverlapValidation(t *testing.T) {
	app := setupScheduleTestApp()
	ownerID := uuid.New()
	clinic, doctor, _ := seedClinicAndDoctor(app, ownerID, models.ClinicStatusApproved, true)
	token := makeAuthHeader(app, ownerID, models.RoleClinicOwner)

	// Two overlapping slots on Monday: 09:00-10:00 and 09:30-10:30
	saveReq := dto.SaveWeeklyScheduleRequest{
		Slots: []dto.WeeklySlotItemRequest{
			{Weekday: 1, StartTime: "09:00", EndTime: "10:00", Status: "active"},
			{Weekday: 1, StartTime: "09:30", EndTime: "10:30", Status: "active"},
		},
	}
	bodyBytes, _ := json.Marshal(saveReq)

	putURL := fmt.Sprintf("/api/v1/clinics/%s/doctors/%s/schedule", clinic.ID, doctor.ID)
	req, _ := http.NewRequest(http.MethodPut, putURL, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for overlapping slots, got %d: %s", w.Code, w.Body.String())
	}
}

// 4. Cross-clinic doctor schedule conflict
func TestSchedule_CrossClinicConflict(t *testing.T) {
	app := setupScheduleTestApp()
	ownerA := uuid.New()
	ownerB := uuid.New()

	clinicA, doctor, cdA := seedClinicAndDoctor(app, ownerA, models.ClinicStatusApproved, true)
	_ = clinicA

	// Create Clinic B with ownerB
	clinicB := &models.Clinic{
		ID:       uuid.New(),
		Name:     "Second Clinic Med",
		OwnerID:  ownerB,
		Status:   models.ClinicStatusApproved,
	}
	_ = app.ClinicRepo.Create(context.Background(), clinicB, nil)

	// Doctor also works in Clinic B
	specID := uuid.New()
	cdB, _ := app.DoctorRepo.CreateDoctorWithClinic(context.Background(), clinicB.ID, doctor, specID, 120000, 30, true)

	// Set Clinic A schedule: Monday 10:00 - 12:00
	_ = app.ScheduleRepo.ReplaceWeeklySlots(context.Background(), cdA.ID, []models.DoctorWeeklySlot{
		{
			ClinicDoctorID: cdA.ID,
			Weekday:        1,
			StartTime:      "10:00",
			EndTime:        "12:00",
			Status:         models.SlotStatusActive,
		},
	})

	// Try to set Clinic B schedule overlapping with Clinic A on Monday: 11:00 - 13:00
	tokenB := makeAuthHeader(app, ownerB, models.RoleClinicOwner)
	saveReq := dto.SaveWeeklyScheduleRequest{
		Slots: []dto.WeeklySlotItemRequest{
			{Weekday: 1, StartTime: "11:00", EndTime: "13:00", Status: "active"},
		},
	}
	bodyBytes, _ := json.Marshal(saveReq)

	putURL := fmt.Sprintf("/api/v1/clinics/%s/doctors/%s/schedule", clinicB.ID, cdB.ID)
	req, _ := http.NewRequest(http.MethodPut, putURL, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", tokenB)
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected 409 Conflict for cross-clinic schedule collision, got %d: %s", w.Code, w.Body.String())
	}
}

// 5. Unauthorized access & IDOR prevention
func TestSchedule_UnauthorizedAccess(t *testing.T) {
	app := setupScheduleTestApp()
	ownerA := uuid.New()
	unauthorizedOwner := uuid.New()

	clinicA, doctor, _ := seedClinicAndDoctor(app, ownerA, models.ClinicStatusApproved, true)

	// User who is not owner/admin of Clinic A
	tokenUnauthorized := makeAuthHeader(app, unauthorizedOwner, models.RoleClinicOwner)

	saveReq := dto.SaveWeeklyScheduleRequest{
		Slots: []dto.WeeklySlotItemRequest{
			{Weekday: 1, StartTime: "09:00", EndTime: "10:00", Status: "active"},
		},
	}
	bodyBytes, _ := json.Marshal(saveReq)

	putURL := fmt.Sprintf("/api/v1/clinics/%s/doctors/%s/schedule", clinicA.ID, doctor.ID)
	req, _ := http.NewRequest(http.MethodPut, putURL, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", tokenUnauthorized)
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for unauthorized clinic access, got %d: %s", w.Code, w.Body.String())
	}

	// Patient role cannot access management endpoint
	patientID := uuid.New()
	tokenPatient := makeAuthHeader(app, patientID, models.RolePatient)
	pReq, _ := http.NewRequest(http.MethodPut, putURL, bytes.NewBuffer(bodyBytes))
	pReq.Header.Set("Content-Type", "application/json")
	pReq.Header.Set("Authorization", tokenPatient)
	pW := httptest.NewRecorder()
	app.Router.ServeHTTP(pW, pReq)

	if pW.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for patient role, got %d: %s", pW.Code, pW.Body.String())
	}
}

// 6. Exceptions / Overrides (CRUD: day off, closed slot, extra slot)
func TestSchedule_Exceptions(t *testing.T) {
	app := setupScheduleTestApp()
	ownerID := uuid.New()
	clinic, doctor, _ := seedClinicAndDoctor(app, ownerID, models.ClinicStatusApproved, true)
	token := makeAuthHeader(app, ownerID, models.RoleClinicOwner)

	exceptionsURL := fmt.Sprintf("/api/v1/clinics/%s/doctors/%s/schedule/exceptions", clinic.ID, doctor.ID)

	// A) Create Day Off Exception for 2026-11-20
	dayOffReq := dto.CreateDateOverrideRequest{
		Date:   "2026-11-20",
		Status: "day_off",
		Reason: "Shifokor konferensiyada",
	}
	doBytes, _ := json.Marshal(dayOffReq)
	req, _ := http.NewRequest(http.MethodPost, exceptionsURL, bytes.NewBuffer(doBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created for day off exception, got %d: %s", w.Code, w.Body.String())
	}

	// B) Create single slot closed exception for 2026-11-25 10:00-11:00
	st10 := "10:00"
	et11 := "11:00"
	slotClosedReq := dto.CreateDateOverrideRequest{
		Date:      "2026-11-25",
		StartTime: &st10,
		EndTime:   &et11,
		Status:    "closed",
		Reason:    "Konsilium yig'ilishi",
	}
	scBytes, _ := json.Marshal(slotClosedReq)
	req2, _ := http.NewRequest(http.MethodPost, exceptionsURL, bytes.NewBuffer(scBytes))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", token)
	w2 := httptest.NewRecorder()
	app.Router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created for slot closed exception, got %d: %s", w2.Code, w2.Body.String())
	}

	var resp2 dto.SuccessResponse
	_ = json.Unmarshal(w2.Body.Bytes(), &resp2)
	data2Bytes, _ := json.Marshal(resp2.Data)
	var createdEx dto.DoctorDateOverrideDTO
	_ = json.Unmarshal(data2Bytes, &createdEx)

	// C) GET exceptions
	getReq, _ := http.NewRequest(http.MethodGet, exceptionsURL, nil)
	getReq.Header.Set("Authorization", token)
	getW := httptest.NewRecorder()
	app.Router.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for GET exceptions, got %d: %s", getW.Code, getW.Body.String())
	}

	// D) PATCH / Update exception
	patchURL := fmt.Sprintf("%s/%s", exceptionsURL, createdEx.ID)
	newReason := "Konsilium bekor bo'ldi, qayta ochildi"
	newStatus := "available"
	updateReq := dto.UpdateDateOverrideRequest{
		Reason: &newReason,
		Status: &newStatus,
	}
	upBytes, _ := json.Marshal(updateReq)
	patchReq, _ := http.NewRequest(http.MethodPatch, patchURL, bytes.NewBuffer(upBytes))
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.Header.Set("Authorization", token)
	patchW := httptest.NewRecorder()
	app.Router.ServeHTTP(patchW, patchReq)

	if patchW.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for PATCH exception, got %d: %s", patchW.Code, patchW.Body.String())
	}

	// E) DELETE exception
	delReq, _ := http.NewRequest(http.MethodDelete, patchURL, nil)
	delReq.Header.Set("Authorization", token)
	delW := httptest.NewRecorder()
	app.Router.ServeHTTP(delW, delReq)

	if delW.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for DELETE exception, got %d: %s", delW.Code, delW.Body.String())
	}
}

// 7. Public Availability (Timezone Asia/Tashkent, Day Off, Overrides, Past Slots Filter)
func TestSchedule_PublicAvailability(t *testing.T) {
	app := setupScheduleTestApp()
	ownerID := uuid.New()
	clinic, doctor, cd := seedClinicAndDoctor(app, ownerID, models.ClinicStatusApproved, true)

	// Setup weekly slots for Thursday (Weekday 4): 09:00-10:00, 10:00-11:00, 11:00-12:00
	_ = app.ScheduleRepo.ReplaceWeeklySlots(context.Background(), cd.ID, []models.DoctorWeeklySlot{
		{ClinicDoctorID: cd.ID, Weekday: 4, StartTime: "09:00", EndTime: "10:00", Status: models.SlotStatusActive},
		{ClinicDoctorID: cd.ID, Weekday: 4, StartTime: "10:00", EndTime: "11:00", Status: models.SlotStatusActive},
		{ClinicDoctorID: cd.ID, Weekday: 4, StartTime: "11:00", EndTime: "12:00", Status: models.SlotStatusActive},
	})

	// Target future Thursday: 2026-11-05 (Thursday)
	availURL := fmt.Sprintf("/api/v1/public/clinics/%s/doctors/%s/availability?date=2026-11-05", clinic.ID, doctor.ID)
	req, _ := http.NewRequest(http.MethodGet, availURL, nil)
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for public availability, got %d: %s", w.Code, w.Body.String())
	}

	var resp dto.SuccessResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	dataBytes, _ := json.Marshal(resp.Data)
	var availResp dto.PublicAvailabilityResponse
	_ = json.Unmarshal(dataBytes, &availResp)

	if len(availResp.Slots) != 3 {
		t.Errorf("Expected 3 available slots on future date, got %d", len(availResp.Slots))
	}

	// Apply an exception: slot 10:00-11:00 is closed on 2026-11-05
	st10 := "10:00"
	et11 := "11:00"
	_ = app.ScheduleRepo.CreateDateOverride(context.Background(), &models.DoctorDateSlotOverride{
		ClinicDoctorID: cd.ID,
		Date:           "2026-11-05",
		StartTime:      &st10,
		EndTime:        &et11,
		Status:         models.OverrideStatusClosed,
	})

	// Query again
	req2, _ := http.NewRequest(http.MethodGet, availURL, nil)
	w2 := httptest.NewRecorder()
	app.Router.ServeHTTP(w2, req2)

	var resp2 dto.SuccessResponse
	_ = json.Unmarshal(w2.Body.Bytes(), &resp2)
	data2Bytes, _ := json.Marshal(resp2.Data)
	var availResp2 dto.PublicAvailabilityResponse
	_ = json.Unmarshal(data2Bytes, &availResp2)

	if len(availResp2.Slots) != 2 {
		t.Errorf("Expected 2 available slots after closing 10:00-11:00, got %d", len(availResp2.Slots))
	}

	// Add Day Off override on 2026-11-05
	_ = app.ScheduleRepo.CreateDateOverride(context.Background(), &models.DoctorDateSlotOverride{
		ClinicDoctorID: cd.ID,
		Date:           "2026-11-05",
		Status:         models.OverrideStatusDayOff,
		Reason:         "Bayram kuni",
	})

	// Query again
	req3, _ := http.NewRequest(http.MethodGet, availURL, nil)
	w3 := httptest.NewRecorder()
	app.Router.ServeHTTP(w3, req3)

	var resp3 dto.SuccessResponse
	_ = json.Unmarshal(w3.Body.Bytes(), &resp3)
	data3Bytes, _ := json.Marshal(resp3.Data)
	var availResp3 dto.PublicAvailabilityResponse
	_ = json.Unmarshal(data3Bytes, &availResp3)

	if len(availResp3.Slots) != 0 {
		t.Errorf("Expected 0 slots for day off, got count=%d", len(availResp3.Slots))
	}
}

// 8. Public availability rejects unapproved clinic or inactive doctor
func TestSchedule_PublicAvailability_InactiveDoctorOrClinic(t *testing.T) {
	app := setupScheduleTestApp()
	ownerID := uuid.New()

	// Seed pending clinic
	pendingClinic, doctor, _ := seedClinicAndDoctor(app, ownerID, models.ClinicStatusPending, true)

	url := fmt.Sprintf("/api/v1/public/clinics/%s/doctors/%s/availability?date=2026-11-05", pendingClinic.ID, doctor.ID)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Errorf("Expected 400 or 404 for unapproved clinic, got %d: %s", w.Code, w.Body.String())
	}
}

// 9. Comprehensive test checking all 7 user items for 35 slots flow
func TestSchedule_FullFlow_Generate35Slots_PutAndAvailability(t *testing.T) {
	app := setupScheduleTestApp()
	ownerID := uuid.New()
	clinic, doctor, cd := seedClinicAndDoctor(app, ownerID, models.ClinicStatusApproved, true)
	token := makeAuthHeader(app, ownerID, models.RoleClinicOwner)

	// Item 1 & 2: Generate preview for weekdays [1, 2, 3, 4, 5], 09:00 - 17:00, 60 min, break 13:00 - 14:00
	genReq := dto.GenerateScheduleRequest{
		Weekdays:            []int{1, 2, 3, 4, 5},
		StartTime:           "09:00",
		EndTime:             "17:00",
		SlotDurationMinutes: 60,
		Breaks: []dto.TimeBreakItem{
			{StartTime: "13:00", EndTime: "14:00"},
		},
	}
	genBytes, _ := json.Marshal(genReq)

	genURL := fmt.Sprintf("/api/v1/clinics/%s/doctors/%s/schedule/generate", clinic.ID, doctor.ID)
	reqGen, _ := http.NewRequest(http.MethodPost, genURL, bytes.NewBuffer(genBytes))
	reqGen.Header.Set("Content-Type", "application/json")
	reqGen.Header.Set("Authorization", token)
	wGen := httptest.NewRecorder()
	app.Router.ServeHTTP(wGen, reqGen)

	if wGen.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for preview, got %d: %s", wGen.Code, wGen.Body.String())
	}

	var respGen dto.SuccessResponse
	_ = json.Unmarshal(wGen.Body.Bytes(), &respGen)
	dBytes, _ := json.Marshal(respGen.Data)
	var genData dto.GenerateScheduleResponse
	_ = json.Unmarshal(dBytes, &genData)

	// 1. Verify all selected weekdays have slots
	weekdayCount := make(map[int]int)
	for _, sl := range genData.Slots {
		weekdayCount[sl.Weekday]++
	}
	for wd := 1; wd <= 5; wd++ {
		if weekdayCount[wd] != 7 {
			t.Errorf("Weekday %d expected 7 slots, got %d", wd, weekdayCount[wd])
		}
	}

	// 2. Verify total_slots == 35
	if genData.TotalSlots != 35 {
		t.Fatalf("Expected total_slots: 35, got %d", genData.TotalSlots)
	}

	// 3. Verify preview does NOT write to database
	slotsInDb, _ := app.ScheduleRepo.GetWeeklySlots(context.Background(), cd.ID)
	if len(slotsInDb) != 0 {
		t.Fatalf("Preview MUST NOT write to database, found %d slots", len(slotsInDb))
	}

	// 4. Save previewed 35 slots via PUT
	var putSlots []dto.WeeklySlotItemRequest
	for _, sl := range genData.Slots {
		putSlots = append(putSlots, dto.WeeklySlotItemRequest{
			Weekday:   sl.Weekday,
			StartTime: sl.StartTime,
			EndTime:   sl.EndTime,
			Status:    sl.Status,
		})
	}
	putReq := dto.SaveWeeklyScheduleRequest{Slots: putSlots}
	putBytes, _ := json.Marshal(putReq)

	schedURL := fmt.Sprintf("/api/v1/clinics/%s/doctors/%s/schedule", clinic.ID, doctor.ID)
	reqPut, _ := http.NewRequest(http.MethodPut, schedURL, bytes.NewBuffer(putBytes))
	reqPut.Header.Set("Content-Type", "application/json")
	reqPut.Header.Set("Authorization", token)
	wPut := httptest.NewRecorder()
	app.Router.ServeHTTP(wPut, reqPut)

	if wPut.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for saving 35 slots, got %d: %s", wPut.Code, wPut.Body.String())
	}

	slotsInDbAfterPut, _ := app.ScheduleRepo.GetWeeklySlots(context.Background(), cd.ID)
	if len(slotsInDbAfterPut) != 35 {
		t.Fatalf("Expected 35 slots saved in DB, found %d", len(slotsInDbAfterPut))
	}

	// 5. Test Public Availability for each day:
	// Monday 2026-11-09 to Friday 2026-11-13 should each return 7 slots
	// Saturday 2026-11-14 should return 0 slots
	testDates := []struct {
		date  string
		slots int
	}{
		{"2026-11-09", 7}, // Monday
		{"2026-11-10", 7}, // Tuesday
		{"2026-11-11", 7}, // Wednesday
		{"2026-11-12", 7}, // Thursday
		{"2026-11-13", 7}, // Friday
		{"2026-11-14", 0}, // Saturday
	}

	for _, tc := range testDates {
		availURL := fmt.Sprintf("/api/v1/public/clinics/%s/doctors/%s/availability?date=%s", clinic.ID, doctor.ID, tc.date)
		reqAvail, _ := http.NewRequest(http.MethodGet, availURL, nil)
		wAvail := httptest.NewRecorder()
		app.Router.ServeHTTP(wAvail, reqAvail)

		if wAvail.Code != http.StatusOK {
			t.Fatalf("Availability for %s failed with code %d: %s", tc.date, wAvail.Code, wAvail.Body.String())
		}

		var respAvail dto.SuccessResponse
		_ = json.Unmarshal(wAvail.Body.Bytes(), &respAvail)
		aBytes, _ := json.Marshal(respAvail.Data)
		var availData dto.PublicAvailabilityResponse
		_ = json.Unmarshal(aBytes, &availData)

		if len(availData.Slots) != tc.slots {
			t.Errorf("Date %s expected %d slots, got %d", tc.date, tc.slots, len(availData.Slots))
		}
	}

	// 6. Test Cross-Clinic conflict check:
	// Another clinic tries to assign this same doctor to Monday 10:00-11:00
	owner2 := uuid.New()
	clinic2 := &models.Clinic{
		ID:      uuid.New(),
		Name:    "Markaziy Shifo Klinikasi",
		OwnerID: owner2,
		Status:  models.ClinicStatusApproved,
	}
	_ = app.ClinicRepo.Create(context.Background(), clinic2, nil)
	cd2, _ := app.DoctorRepo.CreateDoctorWithClinic(context.Background(), clinic2.ID, doctor, uuid.New(), 200000, 30, true)

	token2 := makeAuthHeader(app, owner2, models.RoleClinicOwner)
	conflictPutReq := dto.SaveWeeklyScheduleRequest{
		Slots: []dto.WeeklySlotItemRequest{
			{Weekday: 1, StartTime: "10:00", EndTime: "11:00", Status: "active"}, // Overlaps with Monday slot in Clinic 1!
		},
	}
	cpBytes, _ := json.Marshal(conflictPutReq)
	reqConflict, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/clinics/%s/doctors/%s/schedule", clinic2.ID, cd2.ID), bytes.NewBuffer(cpBytes))
	reqConflict.Header.Set("Content-Type", "application/json")
	reqConflict.Header.Set("Authorization", token2)
	wConflict := httptest.NewRecorder()
	app.Router.ServeHTTP(wConflict, reqConflict)

	if wConflict.Code != http.StatusConflict {
		t.Errorf("Expected 409 Conflict for cross-clinic schedule collision, got %d: %s", wConflict.Code, wConflict.Body.String())
	}

	// 7. Test Exception Override: close 10:00-11:00 slot on Monday 2026-11-09
	st10 := "10:00"
	et11 := "11:00"
	_ = app.ScheduleRepo.CreateDateOverride(context.Background(), &models.DoctorDateSlotOverride{
		ClinicDoctorID: cd.ID,
		Date:           "2026-11-09",
		StartTime:      &st10,
		EndTime:        &et11,
		Status:         models.OverrideStatusClosed,
		Reason:         "Vrach bemor ko'rikida",
	})

	// Monday 2026-11-09 should now have 6 slots (10:00-11:00 removed)
	reqAfterEx, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/public/clinics/%s/doctors/%s/availability?date=2026-11-09", clinic.ID, doctor.ID), nil)
	wAfterEx := httptest.NewRecorder()
	app.Router.ServeHTTP(wAfterEx, reqAfterEx)

	var respAfterEx dto.SuccessResponse
	_ = json.Unmarshal(wAfterEx.Body.Bytes(), &respAfterEx)
	aBytes, _ := json.Marshal(respAfterEx.Data)
	var availAfterEx dto.PublicAvailabilityResponse
	_ = json.Unmarshal(aBytes, &availAfterEx)

	if len(availAfterEx.Slots) != 6 {
		t.Errorf("Expected 6 slots after closing one slot, got %d", len(availAfterEx.Slots))
	}
	for _, sl := range availAfterEx.Slots {
		if sl.StartTime == "10:00" && sl.EndTime == "11:00" {
			t.Errorf("Closed slot 10:00-11:00 should NOT be visible to public bemor!")
		}
	}
}

