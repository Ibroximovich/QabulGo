package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
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

type TestClinicApp struct {
	Router       *gin.Engine
	UserRepo     *MockUserRepository
	ClinicRepo   *MockClinicRepository
	DoctorRepo   *MockDoctorRepository
	SpecialtyRepo *MockSpecialtyRepository
	AuditRepo    *MockAuditRepository
	EmailService *MockEmailService
	Config       *config.Config
	TokenService services.TokenService
}

func setupClinicTestApp() *TestClinicApp {
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

	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(cfg, authService, tokenService)
	patientHandler := handlers.NewPatientHandler(patientService)
	clinicHandler := handlers.NewClinicHandler(clinicService)
	doctorHandler := handlers.NewDoctorHandler(doctorService)
	adminHandler := handlers.NewAdminHandler(adminService)
	specialtyHandler := handlers.NewSpecialtyHandler(specialtyService)
	uploadHandler := handlers.NewUploadHandler(storageService)
	publicSearchHandler := handlers.NewPublicSearchHandler(publicSearchService)

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
	})

	return &TestClinicApp{
		Router:        router,
		UserRepo:      userRepo,
		ClinicRepo:    clinicRepo,
		DoctorRepo:    doctorRepo,
		SpecialtyRepo: specialtyRepo,
		AuditRepo:     auditRepo,
		EmailService:  emailService,
		Config:        cfg,
		TokenService:  tokenService,
	}
}

// 1. Clinic Owner Registration, OTP Verification & Auto Session Creation
func TestClinicAuth_RegisterAndVerify(t *testing.T) {
	app := setupClinicTestApp()

	// 1. Register Clinic Owner
	reqBody := dto.ClinicRegisterRequest{
		FirstName:       "Jamshid",
		LastName:        "Karimov",
		Email:           "jamshid@clinic.uz",
		Phone:           "+998901234500",
		Password:        "ClinicPass123!",
		ConfirmPassword: "ClinicPass123!",
		AcceptTerms:     true,
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/clinic/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Clinic registration failed with code %d: %s", w.Code, w.Body.String())
	}

	// 2. Retrieve OTP from mock email service
	code, err := app.EmailService.GetVerificationCode("jamshid@clinic.uz")
	if err != nil || code == "" {
		t.Fatalf("Verification OTP not generated or sent")
	}

	// 3. Verify Email & ensure auto-login tokens are returned!
	verifyBody := dto.VerifyEmailRequest{
		Email: "jamshid@clinic.uz",
		Code:  code,
	}
	vBytes, _ := json.Marshal(verifyBody)
	vReq, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/verify-email", bytes.NewBuffer(vBytes))
	vReq.Header.Set("Content-Type", "application/json")
	vW := httptest.NewRecorder()
	app.Router.ServeHTTP(vW, vReq)

	if vW.Code != http.StatusOK {
		t.Fatalf("Email verification failed with code %d: %s", vW.Code, vW.Body.String())
	}

	var resp dto.SuccessResponse
	_ = json.Unmarshal(vW.Body.Bytes(), &resp)

	// Check if session data returned
	if resp.Data == nil {
		t.Fatalf("Expected login session data upon email verification, got nil")
	}

	user, _ := app.UserRepo.GetByEmail(context.Background(), "jamshid@clinic.uz")
	if user.Role != models.RoleClinicOwner {
		t.Fatalf("Expected user role clinic_owner, got %s", user.Role)
	}
	if !user.IsEmailVerified() {
		t.Fatalf("User email was not verified in database")
	}
}

// 2. Clinic Login with role protection
func TestClinicAuth_Login(t *testing.T) {
	app := setupClinicTestApp()

	// Create clinic owner
	now := time.Now().UTC()
	hash, _ := utils.HashPassword("ClinicPass123!")
	owner := &models.User{
		ID:              uuid.New(),
		FirstName:       "Jamshid",
		LastName:        "Karimov",
		Email:           "owner@clinic.uz",
		Phone:           "+998901234501",
		PasswordHash:    hash,
		Role:            models.RoleClinicOwner,
		Status:          models.StatusActive,
		EmailVerifiedAt: &now,
		TermsAcceptedAt: now,
		TermsVersion:    "v1.0",
	}
	_ = app.UserRepo.Create(context.Background(), owner)

	// Create regular patient
	patient := &models.User{
		ID:              uuid.New(),
		FirstName:       "Ali",
		LastName:        "Valiyev",
		Email:           "ali@patient.uz",
		Phone:           "+998901234502",
		PasswordHash:    hash,
		Role:            models.RolePatient,
		Status:          models.StatusActive,
		EmailVerifiedAt: &now,
		TermsAcceptedAt: now,
		TermsVersion:    "v1.0",
	}
	_ = app.UserRepo.Create(context.Background(), patient)

	// Test 1: Successful Clinic Owner login
	loginBody := dto.ClinicLoginRequest{
		Email:    "owner@clinic.uz",
		Password: "ClinicPass123!",
	}
	lBytes, _ := json.Marshal(loginBody)
	lReq, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/clinic/login", bytes.NewBuffer(lBytes))
	lReq.Header.Set("Content-Type", "application/json")
	lW := httptest.NewRecorder()
	app.Router.ServeHTTP(lW, lReq)

	if lW.Code != http.StatusOK {
		t.Fatalf("Clinic login failed: %d: %s", lW.Code, lW.Body.String())
	}

	// Test 2: Patient attempting to login via Clinic Login route must be rejected (403)
	pLoginBody := dto.ClinicLoginRequest{
		Email:    "ali@patient.uz",
		Password: "ClinicPass123!",
	}
	pBytes, _ := json.Marshal(pLoginBody)
	pReq, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/clinic/login", bytes.NewBuffer(pBytes))
	pReq.Header.Set("Content-Type", "application/json")
	pW := httptest.NewRecorder()
	app.Router.ServeHTTP(pW, pReq)

	if pW.Code != http.StatusForbidden {
		t.Fatalf("Expected 403 Forbidden for patient on clinic login, got %d", pW.Code)
	}
}

// 3. Clinic Profile Lifecycle: Create (draft) -> Update -> Submit (pending)
func TestClinicProfile_Lifecycle(t *testing.T) {
	app := setupClinicTestApp()

	// Create verified clinic owner
	now := time.Now().UTC()
	hash, _ := utils.HashPassword("ClinicPass123!")
	owner := &models.User{
		ID:              uuid.New(),
		FirstName:       "Sarvar",
		LastName:        "Azamov",
		Email:           "sarvar@shifonur.uz",
		Phone:           "+998901234503",
		PasswordHash:    hash,
		Role:            models.RoleClinicOwner,
		Status:          models.StatusActive,
		EmailVerifiedAt: &now,
	}
	_ = app.UserRepo.Create(context.Background(), owner)

	token, _ := utils.GenerateAccessToken(owner.ID, string(models.RoleClinicOwner), 15*time.Minute, app.Config.JWTSecret, app.Config.JWTIssuer, app.Config.JWTAudience)

	// Step A: Create Clinic (status should be draft)
	createReq := dto.CreateClinicRequest{
		Name:        "Shifo Nur Med",
		Description: "Zamonaviy klinika",
		Phone:       "+998712001122",
		Email:       "info@shifonur.uz",
		Region:      "Toshkent shahri",
		District:    "Yunusobod",
		Address:     "Amir Temur ko‘chasi, 12",
		WorkingHours: []dto.WorkingHourItem{
			{Weekday: 1, OpeningTime: "08:00", ClosingTime: "18:00", IsClosed: false},
			{Weekday: 2, OpeningTime: "08:00", ClosingTime: "18:00", IsClosed: false},
		},
	}
	cBytes, _ := json.Marshal(createReq)
	cReq, _ := http.NewRequest(http.MethodPost, "/api/v1/clinics", bytes.NewBuffer(cBytes))
	cReq.Header.Set("Authorization", "Bearer "+token)
	cReq.Header.Set("Content-Type", "application/json")
	cW := httptest.NewRecorder()
	app.Router.ServeHTTP(cW, cReq)

	if cW.Code != http.StatusCreated {
		t.Fatalf("Create clinic failed with code %d: %s", cW.Code, cW.Body.String())
	}

	var createResp struct {
		Data dto.ClinicDetailDTO `json:"data"`
	}
	_ = json.Unmarshal(cW.Body.Bytes(), &createResp)
	clinicID := createResp.Data.ID

	if createResp.Data.Status != "draft" {
		t.Fatalf("Expected clinic initial status 'draft', got %s", createResp.Data.Status)
	}

	// Step B: Get My Clinic
	mReq, _ := http.NewRequest(http.MethodGet, "/api/v1/clinics/mine", nil)
	mReq.Header.Set("Authorization", "Bearer "+token)
	mW := httptest.NewRecorder()
	app.Router.ServeHTTP(mW, mReq)

	if mW.Code != http.StatusOK {
		t.Fatalf("Get my clinic failed: %d: %s", mW.Code, mW.Body.String())
	}

	// Step C: Update Clinic
	newName := "Shifo Nur Premium Clinic"
	updateReq := dto.UpdateClinicRequest{
		Name: &newName,
	}
	uBytes, _ := json.Marshal(updateReq)
	uReq, _ := http.NewRequest(http.MethodPatch, "/api/v1/clinics/"+clinicID.String(), bytes.NewBuffer(uBytes))
	uReq.Header.Set("Authorization", "Bearer "+token)
	uReq.Header.Set("Content-Type", "application/json")
	uW := httptest.NewRecorder()
	app.Router.ServeHTTP(uW, uReq)

	if uW.Code != http.StatusOK {
		t.Fatalf("Update clinic failed: %d: %s", uW.Code, uW.Body.String())
	}

	// Step D: Submit Clinic for Review (draft -> pending)
	sReq, _ := http.NewRequest(http.MethodPost, "/api/v1/clinics/"+clinicID.String()+"/submit", nil)
	sReq.Header.Set("Authorization", "Bearer "+token)
	sW := httptest.NewRecorder()
	app.Router.ServeHTTP(sW, sReq)

	if sW.Code != http.StatusOK {
		t.Fatalf("Submit clinic failed: %d: %s", sW.Code, sW.Body.String())
	}

	var submitResp struct {
		Data dto.ClinicDetailDTO `json:"data"`
	}
	_ = json.Unmarshal(sW.Body.Bytes(), &submitResp)
	if submitResp.Data.Status != "pending" {
		t.Fatalf("Expected clinic status 'pending' after submit, got %s", submitResp.Data.Status)
	}

	// Step E: Duplicate submit should be rejected
	sReq2, _ := http.NewRequest(http.MethodPost, "/api/v1/clinics/"+clinicID.String()+"/submit", nil)
	sReq2.Header.Set("Authorization", "Bearer "+token)
	sW2 := httptest.NewRecorder()
	app.Router.ServeHTTP(sW2, sReq2)

	if sW2.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 Bad Request on duplicate submit, got %d", sW2.Code)
	}
}

// 4. IDOR Protection: Owner A cannot update Clinic B; Patient cannot access clinic management
func TestClinic_AuthorizationAndIDOR(t *testing.T) {
	app := setupClinicTestApp()

	now := time.Now().UTC()
	hash, _ := utils.HashPassword("Pass123!")

	// Owner A
	ownerA := &models.User{
		ID:              uuid.New(),
		FirstName:       "Owner",
		LastName:        "A",
		Email:           "ownerA@clinic.uz",
		Phone:           "+998901111111",
		PasswordHash:    hash,
		Role:            models.RoleClinicOwner,
		Status:          models.StatusActive,
		EmailVerifiedAt: &now,
	}
	_ = app.UserRepo.Create(context.Background(), ownerA)
	tokenA, _ := utils.GenerateAccessToken(ownerA.ID, string(models.RoleClinicOwner), 15*time.Minute, app.Config.JWTSecret, app.Config.JWTIssuer, app.Config.JWTAudience)

	// Owner B
	ownerB := &models.User{
		ID:              uuid.New(),
		FirstName:       "Owner",
		LastName:        "B",
		Email:           "ownerB@clinic.uz",
		Phone:           "+998902222222",
		PasswordHash:    hash,
		Role:            models.RoleClinicOwner,
		Status:          models.StatusActive,
		EmailVerifiedAt: &now,
	}
	_ = app.UserRepo.Create(context.Background(), ownerB)
	tokenB, _ := utils.GenerateAccessToken(ownerB.ID, string(models.RoleClinicOwner), 15*time.Minute, app.Config.JWTSecret, app.Config.JWTIssuer, app.Config.JWTAudience)

	// Patient
	patient := &models.User{
		ID:              uuid.New(),
		FirstName:       "Patient",
		LastName:        "User",
		Email:           "patient@user.uz",
		Phone:           "+998903333333",
		PasswordHash:    hash,
		Role:            models.RolePatient,
		Status:          models.StatusActive,
		EmailVerifiedAt: &now,
	}
	_ = app.UserRepo.Create(context.Background(), patient)
	tokenPatient, _ := utils.GenerateAccessToken(patient.ID, string(models.RolePatient), 15*time.Minute, app.Config.JWTSecret, app.Config.JWTIssuer, app.Config.JWTAudience)

	// Create Clinic for Owner A
	clinicA := &models.Clinic{
		ID:       uuid.New(),
		Name:     "Clinic A",
		Phone:    "+998711111111",
		Region:   "Toshkent",
		District: "Chilonzor",
		Address:  "Qatortol",
		Status:   models.ClinicStatusDraft,
		OwnerID:  ownerA.ID,
	}
	_ = app.ClinicRepo.Create(context.Background(), clinicA, nil)

	// Test 0: Owner A updates their own Clinic A -> Must succeed (200 OK)
	validName := "Valid Clinic Update"
	vReq := dto.UpdateClinicRequest{Name: &validName}
	vBytes, _ := json.Marshal(vReq)
	req0, _ := http.NewRequest(http.MethodPatch, "/api/v1/clinics/"+clinicA.ID.String(), bytes.NewBuffer(vBytes))
	req0.Header.Set("Authorization", "Bearer "+tokenA)
	req0.Header.Set("Content-Type", "application/json")
	w0 := httptest.NewRecorder()
	app.Router.ServeHTTP(w0, req0)

	if w0.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for owner updating own clinic, got %d", w0.Code)
	}

	// Test 1: Owner B tries to update Clinic A -> Must fail with 403 Forbidden
	newName := "Hacked Clinic"
	uReq := dto.UpdateClinicRequest{Name: &newName}
	uBytes, _ := json.Marshal(uReq)
	req1, _ := http.NewRequest(http.MethodPatch, "/api/v1/clinics/"+clinicA.ID.String(), bytes.NewBuffer(uBytes))
	req1.Header.Set("Authorization", "Bearer "+tokenB)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	app.Router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusForbidden {
		t.Fatalf("Expected 403 Forbidden for cross-clinic update, got %d", w1.Code)
	}

	// Test 2: Patient tries to update Clinic A -> Must fail with 403 Forbidden
	req2, _ := http.NewRequest(http.MethodPatch, "/api/v1/clinics/"+clinicA.ID.String(), bytes.NewBuffer(uBytes))
	req2.Header.Set("Authorization", "Bearer "+tokenPatient)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	app.Router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusForbidden {
		t.Fatalf("Expected 403 Forbidden for patient modifying clinic, got %d", w2.Code)
	}
}

// 5. Superadmin Review Flow: List pending -> Reject with reason -> Resubmit -> Approve -> Suspend
func TestSuperadmin_ReviewFlow(t *testing.T) {
	app := setupClinicTestApp()

	now := time.Now().UTC()
	hash, _ := utils.HashPassword("AdminPass123!")

	// Create Superadmin
	admin := &models.User{
		ID:              uuid.New(),
		FirstName:       "Super",
		LastName:        "Admin",
		Email:           "superadmin@qabulgo.uz",
		Phone:           "+998909999999",
		PasswordHash:    hash,
		Role:            models.RoleSuperAdmin,
		Status:          models.StatusActive,
		EmailVerifiedAt: &now,
	}
	_ = app.UserRepo.Create(context.Background(), admin)
	adminToken, _ := utils.GenerateAccessToken(admin.ID, string(models.RoleSuperAdmin), 15*time.Minute, app.Config.JWTSecret, app.Config.JWTIssuer, app.Config.JWTAudience)

	// Create Clinic in pending status
	clinic := &models.Clinic{
		ID:       uuid.New(),
		Name:     "Akfa Medline Test",
		Phone:    "+998712033003",
		Region:   "Toshkent shahri",
		District: "Olmazor",
		Address:  "Kichik Halka Yo'li",
		Status:   models.ClinicStatusPending,
		OwnerID:  uuid.New(),
	}
	_ = app.ClinicRepo.Create(context.Background(), clinic, []models.ClinicWorkingHour{
		{Weekday: 1, OpeningTime: "08:00", ClosingTime: "20:00"},
	})

	// Step 1: Superadmin lists pending clinics
	listReq, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/clinics?status=pending", nil)
	listReq.Header.Set("Authorization", "Bearer "+adminToken)
	listW := httptest.NewRecorder()
	app.Router.ServeHTTP(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Fatalf("Superadmin list clinics failed: %d", listW.Code)
	}

	// Step 2: Superadmin views clinic details
	getReq, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/clinics/"+clinic.ID.String(), nil)
	getReq.Header.Set("Authorization", "Bearer "+adminToken)
	getW := httptest.NewRecorder()
	app.Router.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("Superadmin get clinic details failed: %d", getW.Code)
	}

	// Step 3: Superadmin rejects clinic with reason
	rejectBody := dto.AdminReviewClinicRequest{
		Reason: "Tibbiy litsenziya muddati o‘tgan, yangi litsenziya yuklang",
	}
	rBytes, _ := json.Marshal(rejectBody)
	rReq, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/clinics/"+clinic.ID.String()+"/reject", bytes.NewBuffer(rBytes))
	rReq.Header.Set("Authorization", "Bearer "+adminToken)
	rReq.Header.Set("Content-Type", "application/json")
	rW := httptest.NewRecorder()
	app.Router.ServeHTTP(rW, rReq)

	if rW.Code != http.StatusOK {
		t.Fatalf("Reject clinic failed: %d: %s", rW.Code, rW.Body.String())
	}

	cAfterReject, _ := app.ClinicRepo.GetByID(context.Background(), clinic.ID)
	if cAfterReject.Status != models.ClinicStatusRejected {
		t.Fatalf("Expected status 'rejected', got %s", cAfterReject.Status)
	}
	if cAfterReject.RejectionReason != rejectBody.Reason {
		t.Fatalf("Expected rejection reason saved, got %s", cAfterReject.RejectionReason)
	}

	// Step 4: Superadmin approves clinic
	appReq, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/clinics/"+clinic.ID.String()+"/approve", nil)
	appReq.Header.Set("Authorization", "Bearer "+adminToken)
	appW := httptest.NewRecorder()
	app.Router.ServeHTTP(appW, appReq)

	if appW.Code != http.StatusOK {
		t.Fatalf("Approve clinic failed: %d: %s", appW.Code, appW.Body.String())
	}

	cAfterApprove, _ := app.ClinicRepo.GetByID(context.Background(), clinic.ID)
	if cAfterApprove.Status != models.ClinicStatusApproved {
		t.Fatalf("Expected status 'approved', got %s", cAfterApprove.Status)
	}

	// Step 5: Superadmin suspends clinic
	susBody := dto.AdminReviewClinicRequest{
		Reason: "Qoidabuzarlik tekshiruvi sababli to‘xtatildi",
	}
	sBytes, _ := json.Marshal(susBody)
	susReq, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/clinics/"+clinic.ID.String()+"/suspend", bytes.NewBuffer(sBytes))
	susReq.Header.Set("Authorization", "Bearer "+adminToken)
	susReq.Header.Set("Content-Type", "application/json")
	susW := httptest.NewRecorder()
	app.Router.ServeHTTP(susW, susReq)

	if susW.Code != http.StatusOK {
		t.Fatalf("Suspend clinic failed: %d: %s", susW.Code, susW.Body.String())
	}

	cAfterSuspend, _ := app.ClinicRepo.GetByID(context.Background(), clinic.ID)
	if cAfterSuspend.Status != models.ClinicStatusSuspended {
		t.Fatalf("Expected status 'suspended', got %s", cAfterSuspend.Status)
	}
}

// 6. Doctor Management: Create, List, Update, Soft-delete
func TestDoctorManagement_CRUD(t *testing.T) {
	app := setupClinicTestApp()

	now := time.Now().UTC()
	hash, _ := utils.HashPassword("ClinicPass123!")
	owner := &models.User{
		ID:              uuid.New(),
		FirstName:       "Jamshid",
		LastName:        "Karimov",
		Email:           "clinicowner@shifo.uz",
		Phone:           "+998901239999",
		PasswordHash:    hash,
		Role:            models.RoleClinicOwner,
		Status:          models.StatusActive,
		EmailVerifiedAt: &now,
	}
	_ = app.UserRepo.Create(context.Background(), owner)
	token, _ := utils.GenerateAccessToken(owner.ID, string(models.RoleClinicOwner), 15*time.Minute, app.Config.JWTSecret, app.Config.JWTIssuer, app.Config.JWTAudience)

	clinic := &models.Clinic{
		ID:       uuid.New(),
		Name:     "Test Medical Center",
		Phone:    "+998712223344",
		Region:   "Toshkent",
		District: "Yunusobod",
		Address:  "Bog'ishamol",
		Status:   models.ClinicStatusApproved,
		OwnerID:  owner.ID,
	}
	_ = app.ClinicRepo.Create(context.Background(), clinic, nil)

	specialties, _ := app.SpecialtyRepo.ListAll(context.Background())
	if len(specialties) == 0 {
		t.Fatalf("Specialties seed empty")
	}
	specialtyID := specialties[0].ID

	// Step 1: Add Doctor to Clinic
	createDocReq := dto.CreateDoctorRequest{
		FirstName:                  "Anvar",
		LastName:                   "Saidov",
		SpecialtyID:                specialtyID,
		Bio:                        "Oliy toifali shifokor",
		ExperienceYears:            12,
		ConsultationPrice:          150000,
		AppointmentDurationMinutes: 30,
	}
	dBytes, _ := json.Marshal(createDocReq)
	dReq, _ := http.NewRequest(http.MethodPost, "/api/v1/clinics/"+clinic.ID.String()+"/doctors", bytes.NewBuffer(dBytes))
	dReq.Header.Set("Authorization", "Bearer "+token)
	dReq.Header.Set("Content-Type", "application/json")
	dW := httptest.NewRecorder()
	app.Router.ServeHTTP(dW, dReq)

	if dW.Code != http.StatusCreated {
		t.Fatalf("Add doctor failed: %d: %s", dW.Code, dW.Body.String())
	}

	var dResp struct {
		Data dto.ClinicDoctorItemDTO `json:"data"`
	}
	_ = json.Unmarshal(dW.Body.Bytes(), &dResp)
	clinicDoctorID := dResp.Data.ID

	// Step 2: List Doctors
	listReq, _ := http.NewRequest(http.MethodGet, "/api/v1/clinics/"+clinic.ID.String()+"/doctors", nil)
	listW := httptest.NewRecorder()
	app.Router.ServeHTTP(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Fatalf("List doctors failed: %d", listW.Code)
	}

	// Step 3: Update Doctor Price
	newPrice := 200000.0
	upReq := dto.UpdateDoctorRequest{
		ConsultationPrice: &newPrice,
	}
	upBytes, _ := json.Marshal(upReq)
	patchReq, _ := http.NewRequest(http.MethodPatch, "/api/v1/clinics/"+clinic.ID.String()+"/doctors/"+clinicDoctorID.String(), bytes.NewBuffer(upBytes))
	patchReq.Header.Set("Authorization", "Bearer "+token)
	patchReq.Header.Set("Content-Type", "application/json")
	patchW := httptest.NewRecorder()
	app.Router.ServeHTTP(patchW, patchReq)

	if patchW.Code != http.StatusOK {
		t.Fatalf("Update doctor failed: %d: %s", patchW.Code, patchW.Body.String())
	}

	// Step 4: Deactivate / Delete Doctor
	delReq, _ := http.NewRequest(http.MethodDelete, "/api/v1/clinics/"+clinic.ID.String()+"/doctors/"+clinicDoctorID.String(), nil)
	delReq.Header.Set("Authorization", "Bearer "+token)
	delW := httptest.NewRecorder()
	app.Router.ServeHTTP(delW, delReq)

	if delW.Code != http.StatusOK {
		t.Fatalf("Deactivate doctor failed: %d: %s", delW.Code, delW.Body.String())
	}
}

// 7. Public Specialties API
func TestSpecialties_List(t *testing.T) {
	app := setupClinicTestApp()

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/specialties", nil)
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/v1/specialties failed: %d", w.Code)
	}

	var resp struct {
		Data []dto.SpecialtyDTO `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if len(resp.Data) == 0 {
		t.Fatalf("Expected non-empty specialties list")
	}
}

// 8. Image Upload API with JPEG validation
func TestUpload_Image(t *testing.T) {
	app := setupClinicTestApp()

	// Authenticated owner
	now := time.Now().UTC()
	owner := &models.User{
		ID:              uuid.New(),
		Email:           "uploader@clinic.uz",
		Role:            models.RoleClinicOwner,
		Status:          models.StatusActive,
		EmailVerifiedAt: &now,
	}
	_ = app.UserRepo.Create(context.Background(), owner)
	token, _ := utils.GenerateAccessToken(owner.ID, string(models.RoleClinicOwner), 15*time.Minute, app.Config.JWTSecret, app.Config.JWTIssuer, app.Config.JWTAudience)

	// Create multipart buffer with valid 1x1 JPEG magic bytes
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.jpg")
	// Minimal valid JPEG header bytes: FF D8 FF E0 ...
	jpegBytes := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01, 0x01, 0x01, 0x00, 0x48, 0x00, 0x48, 0x00, 0x00, 0xFF, 0xDB, 0x00, 0x43, 0x00, 0xFF, 0xD9}
	_, _ = part.Write(jpegBytes)
	_ = writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/upload?folder=clinics", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Upload failed with code %d: %s", w.Code, w.Body.String())
	}
}
