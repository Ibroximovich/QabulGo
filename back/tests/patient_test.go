package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/models"
	"qabulgo-backend/internal/utils"

	"github.com/google/uuid"
)

// Test Patient Profile GET va PATCH
func TestPatientProfile_Update(t *testing.T) {
	router, userRepo, _, cfg, _ := setupTestApp()

	// Patient foydalanuvchi yaratish
	user := &models.User{
		ID:              uuid.New(),
		FirstName:       "EskiIsm",
		LastName:        "EskiFamiliya",
		Email:           "bemor@example.com",
		Phone:           "+998901112233",
		PasswordHash:    "dummy",
		Role:            models.RolePatient,
		Status:          models.StatusActive,
		TermsAcceptedAt: time.Now().UTC(),
		TermsVersion:    "v1.0",
	}
	_ = userRepo.Create(context.Background(), user)

	token, _ := utils.GenerateAccessToken(user.ID, "patient", 15*time.Minute, cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAudience)

	// Profilni olish
	getReq, _ := http.NewRequest(http.MethodGet, "/api/v1/patient/me", nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("Profilni olishda xatolik: %d", getW.Code)
	}

	// Profilni tahrirlash (PATCH)
	updateReq := dto.UpdatePatientProfileRequest{
		FirstName: "YangiIsm",
		LastName:  "YangiFamiliya",
		Phone:     "+998909998811",
	}
	uBody, _ := json.Marshal(updateReq)
	patchReq, _ := http.NewRequest(http.MethodPatch, "/api/v1/patient/me", bytes.NewBuffer(uBody))
	patchReq.Header.Set("Authorization", "Bearer "+token)
	patchReq.Header.Set("Content-Type", "application/json")
	patchW := httptest.NewRecorder()
	router.ServeHTTP(patchW, patchReq)

	if patchW.Code != http.StatusOK {
		t.Fatalf("Profilni tahrirlash kodi 200 bo'lmadi: %d, body: %s", patchW.Code, patchW.Body.String())
	}

	// Yangilangan profilni tekshirish
	updatedUser, _ := userRepo.GetByID(context.Background(), user.ID)
	if updatedUser.FirstName != "YangiIsm" || updatedUser.Phone != "+998909998811" {
		t.Fatalf("Profil ma'lumotlari to'g'ri yangilanmadi: %+v", updatedUser)
	}
	// Email o'zgarmaganligini tekshirish
	if updatedUser.Email != "bemor@example.com" {
		t.Fatalf("Email o'zgarib ketdi, bu ruxsat etilmagan!")
	}
}
