package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"qabulgo-backend/internal/models"
	"sync"
	"testing"
	"time"

	"qabulgo-backend/internal/config"
	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/handlers"
	"qabulgo-backend/internal/routes"
	"qabulgo-backend/internal/services"
	"qabulgo-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func setupTestApp() (*gin.Engine, *MockUserRepository, *MockEmailService, *config.Config, services.TokenService) {
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

	tokenService := services.NewTokenService(cfg, sessionRepo, userRepo)
	authService := services.NewAuthService(cfg, userRepo, tokenRepo, tokenService, emailService)
	patientService := services.NewPatientService(userRepo)

	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(cfg, authService, tokenService)
	patientHandler := handlers.NewPatientHandler(patientService)

	router := routes.SetupRouter(&routes.RouterConfig{
		Config:         cfg,
		HealthHandler:  healthHandler,
		AuthHandler:    authHandler,
		PatientHandler: patientHandler,
	})

	return router, userRepo, emailService, cfg, tokenService
}

// 1. To‘g‘ri ma’lumot bilan register
func TestRegister_Success(t *testing.T) {
	router, _, emailService, _, _ := setupTestApp()

	payload := dto.PatientRegisterRequest{
		FirstName:       "Ali",
		LastName:        "Valiyev",
		Email:           "ali@example.com",
		Phone:           "+998901234567",
		Password:        "StrongPass123!",
		ConfirmPassword: "StrongPass123!",
		AcceptTerms:     true,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/patient/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Kutilgan kod 201, lekin %d qaytdi: %s", w.Code, w.Body.String())
	}

	var res dto.SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("JSON parse xatosi: %v", err)
	}

	if !res.Success {
		t.Fatalf("Success maydoni true bo'lishi kerak")
	}

	// Emailga kod kelganligini tekshirish
	code, err := emailService.GetVerificationCode("ali@example.com")
	if err != nil || len(code) != 6 {
		t.Fatalf("Emailga 6 xonali kod kelmadi: %v, code=%s", err, code)
	}
}

// 2. Noto‘g‘ri email bilan register
func TestRegister_InvalidEmail(t *testing.T) {
	router, _, _, _, _ := setupTestApp()

	payload := dto.PatientRegisterRequest{
		FirstName:       "Ali",
		LastName:        "Valiyev",
		Email:           "not-an-email",
		Phone:           "+998901234567",
		Password:        "StrongPass123!",
		ConfirmPassword: "StrongPass123!",
		AcceptTerms:     true,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/patient/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Kutilgan kod 400, lekin %d qaytdi", w.Code)
	}
}

// 3. Takroriy email bilan register
func TestRegister_DuplicateEmail(t *testing.T) {
	router, _, _, _, _ := setupTestApp()

	payload1 := dto.PatientRegisterRequest{
		FirstName:       "Ali",
		LastName:        "Valiyev",
		Email:           "ali@example.com",
		Phone:           "+998901234567",
		Password:        "StrongPass123!",
		ConfirmPassword: "StrongPass123!",
		AcceptTerms:     true,
	}
	body1, _ := json.Marshal(payload1)
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/patient/register", bytes.NewBuffer(body1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("Birinchi register o'tmadi: %d", w1.Code)
	}

	// Takroriy email, lekin boshqa telefon
	payload2 := dto.PatientRegisterRequest{
		FirstName:       "Vali",
		LastName:        "Aliyev",
		Email:           "ALI@example.com", // Katta harfda ham tekshirish
		Phone:           "+998909876543",
		Password:        "StrongPass123!",
		ConfirmPassword: "StrongPass123!",
		AcceptTerms:     true,
	}
	body2, _ := json.Marshal(payload2)
	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/patient/register", bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("Kutilgan kod 409 Conflict, lekin %d qaytdi: %s", w2.Code, w2.Body.String())
	}
}

// 4. Takroriy telefon bilan register
func TestRegister_DuplicatePhone(t *testing.T) {
	router, _, _, _, _ := setupTestApp()

	payload1 := dto.PatientRegisterRequest{
		FirstName:       "Ali",
		LastName:        "Valiyev",
		Email:           "ali1@example.com",
		Phone:           "+998901234567",
		Password:        "StrongPass123!",
		ConfirmPassword: "StrongPass123!",
		AcceptTerms:     true,
	}
	body1, _ := json.Marshal(payload1)
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/patient/register", bytes.NewBuffer(body1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Takroriy telefon
	payload2 := dto.PatientRegisterRequest{
		FirstName:       "Vali",
		LastName:        "Aliyev",
		Email:           "ali2@example.com",
		Phone:           "+998901234567",
		Password:        "StrongPass123!",
		ConfirmPassword: "StrongPass123!",
		AcceptTerms:     true,
	}
	body2, _ := json.Marshal(payload2)
	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/patient/register", bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("Kutilgan kod 409 Conflict, lekin %d qaytdi: %s", w2.Code, w2.Body.String())
	}
}

// 5. Zaif parol bilan register
func TestRegister_WeakPassword(t *testing.T) {
	router, _, _, _, _ := setupTestApp()

	payload := dto.PatientRegisterRequest{
		FirstName:       "Ali",
		LastName:        "Valiyev",
		Email:           "ali@example.com",
		Phone:           "+998901234567",
		Password:        "12345678", // Zaif parol
		ConfirmPassword: "12345678",
		AcceptTerms:     true,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/patient/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Zaif parol rad etilmadi, kod: %d", w.Code)
	}
}

// 6. Confirm password mos kelmasligi
func TestRegister_PasswordMismatch(t *testing.T) {
	router, _, _, _, _ := setupTestApp()

	payload := dto.PatientRegisterRequest{
		FirstName:       "Ali",
		LastName:        "Valiyev",
		Email:           "ali@example.com",
		Phone:           "+998901234567",
		Password:        "StrongPass123!",
		ConfirmPassword: "DifferentPass123!",
		AcceptTerms:     true,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/patient/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Parollar farqi rad etilmadi, kod: %d", w.Code)
	}
}

// 7. Email tasdiqlash
func TestVerifyEmail_Success(t *testing.T) {
	router, userRepo, emailService, _, _ := setupTestApp()

	// Register
	payload := dto.PatientRegisterRequest{
		FirstName:       "Ali",
		LastName:        "Valiyev",
		Email:           "ali@example.com",
		Phone:           "+998901234567",
		Password:        "StrongPass123!",
		ConfirmPassword: "StrongPass123!",
		AcceptTerms:     true,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/patient/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	code, _ := emailService.GetVerificationCode("ali@example.com")

	// Verify
	verifyPayload := dto.VerifyEmailRequest{
		Email: "ali@example.com",
		Code:  code,
	}
	vBody, _ := json.Marshal(verifyPayload)
	vReq, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/verify-email", bytes.NewBuffer(vBody))
	vReq.Header.Set("Content-Type", "application/json")
	vW := httptest.NewRecorder()
	router.ServeHTTP(vW, vReq)

	if vW.Code != http.StatusOK {
		t.Fatalf("Email tasdiqlanmadi, kod: %d: %s", vW.Code, vW.Body.String())
	}

	user, _ := userRepo.GetByEmail(context.Background(), "ali@example.com")
	if !user.IsEmailVerified() {
		t.Fatalf("Foydalanuvchi bazada email_verified bo'lmadi")
	}
}

// 8. Noto‘g‘ri yoki muddati tugagan kod
func TestVerifyEmail_InvalidCode(t *testing.T) {
	router, _, _, _, _ := setupTestApp()

	// Register
	payload := dto.PatientRegisterRequest{
		FirstName:       "Ali",
		LastName:        "Valiyev",
		Email:           "ali@example.com",
		Phone:           "+998901234567",
		Password:        "StrongPass123!",
		ConfirmPassword: "StrongPass123!",
		AcceptTerms:     true,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/patient/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify with wrong code
	verifyPayload := dto.VerifyEmailRequest{
		Email: "ali@example.com",
		Code:  "999999", // Noto'g'ri kod
	}
	vBody, _ := json.Marshal(verifyPayload)
	vReq, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/verify-email", bytes.NewBuffer(vBody))
	vReq.Header.Set("Content-Type", "application/json")
	vW := httptest.NewRecorder()
	router.ServeHTTP(vW, vReq)

	if vW.Code != http.StatusBadRequest {
		t.Fatalf("Noto'g'ri kod qabul qilindi, kod: %d", vW.Code)
	}
}

// 9. To‘g‘ri login
func TestLogin_Success(t *testing.T) {
	router, userRepo, _, _, _ := setupTestApp()

	// Register
	pReq := dto.PatientRegisterRequest{
		FirstName:       "Ali",
		LastName:        "Valiyev",
		Email:           "ali@example.com",
		Phone:           "+998901234567",
		Password:        "StrongPass123!",
		ConfirmPassword: "StrongPass123!",
		AcceptTerms:     true,
	}
	b, _ := json.Marshal(pReq)
	r, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/patient/register", bytes.NewBuffer(b))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	// Emailni tasdiqlangan deb belgilaymiz
	user, _ := userRepo.GetByEmail(context.Background(), "ali@example.com")
	_ = userRepo.UpdateEmailVerified(context.Background(), user.ID, time.Now().UTC())

	// Login
	loginReq := dto.PatientLoginRequest{
		Email:    "ALI@example.com",
		Password: "StrongPass123!",
	}
	lBody, _ := json.Marshal(loginReq)
	lReq, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/patient/login", bytes.NewBuffer(lBody))
	lReq.Header.Set("Content-Type", "application/json")
	lW := httptest.NewRecorder()
	router.ServeHTTP(lW, lReq)

	if lW.Code != http.StatusOK {
		t.Fatalf("Login muvaffaqiyatli bo'lmadi: %d, body: %s", lW.Code, lW.Body.String())
	}

	// Cookie tekshirish
	cookies := lW.Result().Cookies()
	var hasRefreshCookie bool
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" && cookie.HttpOnly && cookie.Value != "" {
			hasRefreshCookie = true
			break
		}
	}
	if !hasRefreshCookie {
		t.Fatalf("HttpOnly refresh_token cookie o'rnatilmadi")
	}
}

// 10. Noto‘g‘ri parol bilan login
func TestLogin_InvalidPassword(t *testing.T) {
	router, userRepo, _, _, _ := setupTestApp()

	// Register
	pReq := dto.PatientRegisterRequest{
		FirstName:       "Ali",
		LastName:        "Valiyev",
		Email:           "ali@example.com",
		Phone:           "+998901234567",
		Password:        "StrongPass123!",
		ConfirmPassword: "StrongPass123!",
		AcceptTerms:     true,
	}
	b, _ := json.Marshal(pReq)
	r, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/patient/register", bytes.NewBuffer(b))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	user, _ := userRepo.GetByEmail(context.Background(), "ali@example.com")
	_ = userRepo.UpdateEmailVerified(context.Background(), user.ID, time.Now().UTC())

	// Noto'g'ri parol
	loginReq := dto.PatientLoginRequest{
		Email:    "ali@example.com",
		Password: "WrongPassword999!",
	}
	lBody, _ := json.Marshal(loginReq)
	lReq, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/patient/login", bytes.NewBuffer(lBody))
	lReq.Header.Set("Content-Type", "application/json")
	lW := httptest.NewRecorder()
	router.ServeHTTP(lW, lReq)

	if lW.Code != http.StatusUnauthorized {
		t.Fatalf("Kutilgan 401 Unauthorized, lekin %d qaytdi", lW.Code)
	}
}

// 11. Tasdiqlanmagan email bilan login
func TestLogin_UnverifiedEmail(t *testing.T) {
	router, _, _, _, _ := setupTestApp()

	pReq := dto.PatientRegisterRequest{
		FirstName:       "Ali",
		LastName:        "Valiyev",
		Email:           "ali@example.com",
		Phone:           "+998901234567",
		Password:        "StrongPass123!",
		ConfirmPassword: "StrongPass123!",
		AcceptTerms:     true,
	}
	b, _ := json.Marshal(pReq)
	r, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/patient/register", bytes.NewBuffer(b))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	// Tasdiqlanmasdan login qilish
	loginReq := dto.PatientLoginRequest{
		Email:    "ali@example.com",
		Password: "StrongPass123!",
	}
	lBody, _ := json.Marshal(loginReq)
	lReq, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/patient/login", bytes.NewBuffer(lBody))
	lReq.Header.Set("Content-Type", "application/json")
	lW := httptest.NewRecorder()
	router.ServeHTTP(lW, lReq)

	if lW.Code != http.StatusForbidden {
		t.Fatalf("Tasdiqlanmagan login rad etilmadi, kod: %d", lW.Code)
	}
}

// 12. Token muddati tugashi
func TestAccessToken_Expired(t *testing.T) {
	_, _, _, cfg, _ := setupTestApp()

	userID := uuid.New()
	// TTL -1 soniya qilib eskirgan token yaratish
	expiredToken, err := utils.GenerateAccessToken(userID, "patient", -1*time.Second, cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience)
	if err != nil {
		t.Fatalf("Token generatsiya xatosi: %v", err)
	}

	_, err = utils.ValidateAccessToken(expiredToken, cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience)
	if err == nil {
		t.Fatalf("Muddati o'tgan token yaroqli deb topildi!")
	}
}

// 13. Refresh token rotation
func TestRefreshToken_Rotation(t *testing.T) {
	router, userRepo, _, _, _ := setupTestApp()

	// Register & verify
	pReq := dto.PatientRegisterRequest{
		FirstName:       "Ali",
		LastName:        "Valiyev",
		Email:           "ali@example.com",
		Phone:           "+998901234567",
		Password:        "StrongPass123!",
		ConfirmPassword: "StrongPass123!",
		AcceptTerms:     true,
	}
	b, _ := json.Marshal(pReq)
	r, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/patient/register", bytes.NewBuffer(b))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	user, _ := userRepo.GetByEmail(context.Background(), "ali@example.com")
	_ = userRepo.UpdateEmailVerified(context.Background(), user.ID, time.Now().UTC())

	// Login
	loginReq := dto.PatientLoginRequest{
		Email:    "ali@example.com",
		Password: "StrongPass123!",
	}
	lBody, _ := json.Marshal(loginReq)
	lReq, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/patient/login", bytes.NewBuffer(lBody))
	lReq.Header.Set("Content-Type", "application/json")
	lW := httptest.NewRecorder()
	router.ServeHTTP(lW, lReq)

	var initialCookie *http.Cookie
	for _, c := range lW.Result().Cookies() {
		if c.Name == "refresh_token" {
			initialCookie = c
			break
		}
	}
	if initialCookie == nil {
		t.Fatalf("Refresh cookie topilmadi")
	}

	// Refresh call
	refReq, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	refReq.AddCookie(initialCookie)
	refW := httptest.NewRecorder()
	router.ServeHTTP(refW, refReq)

	if refW.Code != http.StatusOK {
		t.Fatalf("Refresh muvaffaqiyatli bo'lmadi: %d, body: %s", refW.Code, refW.Body.String())
	}

	// Yangi cookie olinganini tekshirish
	var newCookie *http.Cookie
	for _, c := range refW.Result().Cookies() {
		if c.Name == "refresh_token" {
			newCookie = c
			break
		}
	}
	if newCookie == nil || newCookie.Value == initialCookie.Value {
		t.Fatalf("Token rotation ishlamadi, yangi cookie olinmadi yoki eski bilan bir xil")
	}

	// Eski token bilan qayta urinish (re-use attack)
	attackReq, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	attackReq.AddCookie(initialCookie)
	attackW := httptest.NewRecorder()
	router.ServeHTTP(attackW, attackReq)

	if attackW.Code != http.StatusUnauthorized {
		t.Fatalf("Eski refresh token ishlatildi, lekin rad etilmadi: %d", attackW.Code)
	}
}

// 14. Logoutdan keyin refresh ishlamasligi
func TestLogout_InvalidatesSession(t *testing.T) {
	router, userRepo, _, _, _ := setupTestApp()

	// Register & verify
	pReq := dto.PatientRegisterRequest{
		FirstName:       "Ali",
		LastName:        "Valiyev",
		Email:           "ali@example.com",
		Phone:           "+998901234567",
		Password:        "StrongPass123!",
		ConfirmPassword: "StrongPass123!",
		AcceptTerms:     true,
	}
	b, _ := json.Marshal(pReq)
	r, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/patient/register", bytes.NewBuffer(b))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	user, _ := userRepo.GetByEmail(context.Background(), "ali@example.com")
	_ = userRepo.UpdateEmailVerified(context.Background(), user.ID, time.Now().UTC())

	// Login
	loginReq := dto.PatientLoginRequest{
		Email:    "ali@example.com",
		Password: "StrongPass123!",
	}
	lBody, _ := json.Marshal(loginReq)
	lReq, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/patient/login", bytes.NewBuffer(lBody))
	lReq.Header.Set("Content-Type", "application/json")
	lW := httptest.NewRecorder()
	router.ServeHTTP(lW, lReq)

	var refreshCookie *http.Cookie
	for _, c := range lW.Result().Cookies() {
		if c.Name == "refresh_token" {
			refreshCookie = c
			break
		}
	}

	// Logout
	logoutReq, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	logoutReq.AddCookie(refreshCookie)
	logoutW := httptest.NewRecorder()
	router.ServeHTTP(logoutW, logoutReq)

	if logoutW.Code != http.StatusOK {
		t.Fatalf("Logout kodi 200 bo'lmadi: %d", logoutW.Code)
	}

	// Refresh call after logout
	refReq, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	refReq.AddCookie(refreshCookie)
	refW := httptest.NewRecorder()
	router.ServeHTTP(refW, refReq)

	if refW.Code != http.StatusUnauthorized {
		t.Fatalf("Logout qilingan token bilan refresh muvaffaqiyatli bo'lmasligi kerak: %d", refW.Code)
	}
}

// 15. Login qilmasdan /patient/me ga kirish
func TestPatientMe_Unauthorized(t *testing.T) {
	router, _, _, _, _ := setupTestApp()

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/patient/me", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Kutilgan 401 Unauthorized, lekin %d qaytdi", w.Code)
	}
}

// 16. Boshqa foydalanuvchi ma’lumotlariga kira olmaslik
func TestPatientMe_Isolation(t *testing.T) {
	router, userRepo, _, cfg, _ := setupTestApp()

	// 2 ta foydalanuvchi yaratamiz
	user1 := &models.User{
		ID:              uuid.New(),
		FirstName:       "User1",
		LastName:        "Test",
		Email:           "user1@example.com",
		Phone:           "+998901111111",
		PasswordHash:    "dummy",
		Role:            models.RolePatient,
		Status:          models.StatusActive,
		TermsAcceptedAt: time.Now().UTC(),
		TermsVersion:    "v1.0",
	}
	user2 := &models.User{
		ID:              uuid.New(),
		FirstName:       "User2",
		LastName:        "Test",
		Email:           "user2@example.com",
		Phone:           "+998902222222",
		PasswordHash:    "dummy",
		Role:            models.RolePatient,
		Status:          models.StatusActive,
		TermsAcceptedAt: time.Now().UTC(),
		TermsVersion:    "v1.0",
	}
	_ = userRepo.Create(context.Background(), user1)
	_ = userRepo.Create(context.Background(), user2)

	// User1 uchun token yaratamiz
	token1, _ := utils.GenerateAccessToken(user1.ID, "patient", 15*time.Minute, cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience)

	// User1 tokenni yuborib /patient/me chaqiradi
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/patient/me", nil)
	req.Header.Set("Authorization", "Bearer "+token1)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Profil olinmadi: %d", w.Code)
	}

	var res dto.SuccessResponse
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	dataMap, _ := res.Data.(map[string]interface{})

	// Faqat User1 ma'lumotlari bo'lishi shart
	if dataMap["email"] != "user1@example.com" {
		t.Fatalf("Kutilgan user1 email, lekin %v qaytdi", dataMap["email"])
	}
	if dataMap["email"] == "user2@example.com" {
		t.Fatalf("Boshqa foydalanuvchi ma'lumotlari sizib chiqdi!")
	}
}

// 17. Parolni tiklash
func TestForgotPassword_And_ResetPassword(t *testing.T) {
	router, userRepo, emailService, _, _ := setupTestApp()

	// Register & verify
	pReq := dto.PatientRegisterRequest{
		FirstName:       "Ali",
		LastName:        "Valiyev",
		Email:           "ali@example.com",
		Phone:           "+998901234567",
		Password:        "OldPassword123!",
		ConfirmPassword: "OldPassword123!",
		AcceptTerms:     true,
	}
	b, _ := json.Marshal(pReq)
	r, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/patient/register", bytes.NewBuffer(b))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	user, _ := userRepo.GetByEmail(context.Background(), "ali@example.com")
	_ = userRepo.UpdateEmailVerified(context.Background(), user.ID, time.Now().UTC())

	// Forgot password
	fReq := dto.ForgotPasswordRequest{Email: "ali@example.com"}
	fBody, _ := json.Marshal(fReq)
	fR, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewBuffer(fBody))
	fR.Header.Set("Content-Type", "application/json")
	fW := httptest.NewRecorder()
	router.ServeHTTP(fW, fR)

	if fW.Code != http.StatusOK {
		t.Fatalf("Forgot password so'rovi xato: %d", fW.Code)
	}

	resetCode, err := emailService.GetResetCode("ali@example.com")
	if err != nil || len(resetCode) != 6 {
		t.Fatalf("Reset kodi emailga kelmadi: %v", err)
	}

	// Reset password
	resetReq := dto.ResetPasswordRequest{
		Email:           "ali@example.com",
		Code:            resetCode,
		NewPassword:     "NewBrandPass123!",
		ConfirmPassword: "NewBrandPass123!",
	}
	rBody, _ := json.Marshal(resetReq)
	resetR, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewBuffer(rBody))
	resetR.Header.Set("Content-Type", "application/json")
	resetW := httptest.NewRecorder()
	router.ServeHTTP(resetW, resetR)

	if resetW.Code != http.StatusOK {
		t.Fatalf("Reset password kodi 200 bo'lmadi: %d, body: %s", resetW.Code, resetW.Body.String())
	}

	// Yangi parol bilan login qilib ko'ramiz
	loginReq := dto.PatientLoginRequest{
		Email:    "ali@example.com",
		Password: "NewBrandPass123!",
	}
	lBody, _ := json.Marshal(loginReq)
	lR, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/patient/login", bytes.NewBuffer(lBody))
	lR.Header.Set("Content-Type", "application/json")
	lW := httptest.NewRecorder()
	router.ServeHTTP(lW, lR)

	if lW.Code != http.StatusOK {
		t.Fatalf("Yangi parol bilan login bo'lmadi: %d", lW.Code)
	}
}

// 18. Bir vaqtdagi register so‘rovlarida UNIQUE constraint ishlashi
func TestRegister_ConcurrentUniqueConstraint(t *testing.T) {
	router, _, _, _, _ := setupTestApp()

	concurrentCount := 10
	var wg sync.WaitGroup
	statusCodeChan := make(chan int, concurrentCount)

	payload := dto.PatientRegisterRequest{
		FirstName:       "Ali",
		LastName:        "Valiyev",
		Email:           "concurrent@example.com",
		Phone:           "+998909998877",
		Password:        "StrongPass123!",
		ConfirmPassword: "StrongPass123!",
		AcceptTerms:     true,
	}
	body, _ := json.Marshal(payload)

	for i := 0; i < concurrentCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/patient/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			statusCodeChan <- w.Code
		}()
	}

	wg.Wait()
	close(statusCodeChan)

	var successCount, conflictCount int
	for code := range statusCodeChan {
		if code == http.StatusCreated {
			successCount++
		} else if code == http.StatusConflict {
			conflictCount++
		}
	}

	if successCount != 1 {
		t.Fatalf("Bir vaqtdagi bir xil so'rovlarda faqat 1 tasi 201 Created bo'lishi kerak, lekin %d ta bo'ldi", successCount)
	}
	if conflictCount != concurrentCount-1 {
		t.Fatalf("Qolgan %d tasi 409 Conflict bo'lishi kerak edi, lekin %d ta bo'ldi", concurrentCount-1, conflictCount)
	}
}
