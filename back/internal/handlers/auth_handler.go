package handlers

import (
	"errors"
	"net/http"
	"time"

	"qabulgo-backend/internal/config"
	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/repositories"
	"qabulgo-backend/internal/services"

	"github.com/gin-gonic/gin"
)

const RefreshTokenCookieName = "refresh_token"

type AuthHandler struct {
	cfg          *config.Config
	authService  services.AuthService
	tokenService services.TokenService
}

func NewAuthHandler(cfg *config.Config, authService services.AuthService, tokenService services.TokenService) *AuthHandler {
	return &AuthHandler{
		cfg:          cfg,
		authService:  authService,
		tokenService: tokenService,
	}
}

func (h *AuthHandler) setRefreshTokenCookie(c *gin.Context, rawRefreshToken string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}

	cookieDomain := ""
	if h.cfg.CookieDomain != "" && h.cfg.CookieDomain != "localhost" {
		cookieDomain = h.cfg.CookieDomain
	}

	if h.cfg.CookieSecure {
		c.SetSameSite(http.SameSiteNoneMode)
	} else {
		c.SetSameSite(http.SameSiteLaxMode)
	}

	c.SetCookie(
		RefreshTokenCookieName,
		rawRefreshToken,
		maxAge,
		"/",
		cookieDomain,
		h.cfg.CookieSecure,
		true, // HttpOnly
	)
}

func (h *AuthHandler) clearRefreshTokenCookie(c *gin.Context) {
	cookieDomain := ""
	if h.cfg.CookieDomain != "" && h.cfg.CookieDomain != "localhost" {
		cookieDomain = h.cfg.CookieDomain
	}

	if h.cfg.CookieSecure {
		c.SetSameSite(http.SameSiteNoneMode)
	} else {
		c.SetSameSite(http.SameSiteLaxMode)
	}

	c.SetCookie(
		RefreshTokenCookieName,
		"",
		-1,
		"/",
		cookieDomain,
		h.cfg.CookieSecure,
		true,
	)
}

// Register godoc
// @Summary Bemor ro‘yxatdan o‘tishi
// @Description Yangi bemor foydalanuvchini ro‘yxatdan o‘tkazadi va emailga tasdiqlash kodi yuboradi
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.PatientRegisterRequest true "Ro‘yxatdan o‘tish ma’lumotlari"
// @Success 201 {object} dto.SuccessResponse{data=dto.PatientRegisterData} "Muvaffaqiyatli ro‘yxatdan o‘tildi"
// @Failure 400 {object} dto.ErrorResponse "Noto‘g‘ri ma’lumot kiritilgan"
// @Failure 409 {object} dto.ErrorResponse "Email yoki telefon allaqachon mavjud"
// @Failure 422 {object} dto.ErrorResponse "Validatsiya xatoligi"
// @Router /api/v1/auth/patient/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.PatientRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse(
			"Kiritilgan ma'lumotlar formati noto‘g‘ri",
			map[string]string{"body": err.Error()},
		))
		return
	}

	data, err := h.authService.RegisterPatient(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, repositories.ErrEmailAlreadyExist) {
			c.JSON(http.StatusConflict, dto.NewErrorResponse(err.Error(), map[string]string{"email": err.Error()}))
			return
		}
		if errors.Is(err, repositories.ErrPhoneAlreadyExist) {
			c.JSON(http.StatusConflict, dto.NewErrorResponse(err.Error(), map[string]string{"phone": err.Error()}))
			return
		}
		if errors.Is(err, services.ErrTermsNotAccepted) ||
			errors.Is(err, services.ErrPasswordsDoNotMatch) ||
			errors.Is(err, services.ErrNameLength) ||
			errors.Is(err, services.ErrInvalidEmailFormat) ||
			errors.Is(err, services.ErrInvalidPhoneFormat) {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
			return
		}

		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusCreated, dto.NewSuccessResponse(
		"Muvaffaqiyatli ro‘yxatdan o‘tdingiz! Emailingizga tasdiqlash kodi yuborildi.",
		data,
	))
}

// VerifyEmail godoc
// @Summary Emailni tasdiqlash
// @Description 6 xonali tasdiqlash kodi orqali emailni tasdiqlaydi va avtomatik sessiya yaratadi
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.VerifyEmailRequest true "Email va tasdiqlash kodi"
// @Success 200 {object} dto.SuccessResponse{data=dto.LoginData} "Email muvaffaqiyatli tasdiqlandi"
// @Failure 400 {object} dto.ErrorResponse "Kod noto‘g‘ri yoki muddati tugagan"
// @Failure 429 {object} dto.ErrorResponse "Ko‘p noto‘g‘ri urinish"
// @Router /api/v1/auth/verify-email [post]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req dto.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Kerakli maydonlar to‘ldirilmagan", nil))
		return
	}

	loginData, rawRefreshToken, expiresAt, err := h.authService.VerifyEmail(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, services.ErrTooManyAttempts) {
			c.JSON(http.StatusTooManyRequests, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	if rawRefreshToken != "" {
		h.setRefreshTokenCookie(c, rawRefreshToken, expiresAt)
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Emailingiz muvaffaqiyatli tasdiqlandi.", loginData))
}

// RegisterClinic godoc
// @Summary Klinika egasi ro‘yxatdan o‘tishi
// @Description Yangi klinika egasini ro‘yxatdan o‘tkazadi va emailga tasdiqlash kodi yuboradi
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ClinicRegisterRequest true "Klinika egasi ro‘yxatdan o‘tish ma’lumotlari"
// @Success 201 {object} dto.SuccessResponse{data=dto.PatientRegisterData} "Muvaffaqiyatli ro‘yxatdan o‘tildi"
// @Failure 400 {object} dto.ErrorResponse "Noto‘g‘ri ma’lumot kiritilgan"
// @Failure 409 {object} dto.ErrorResponse "Email yoki telefon allaqachon mavjud"
// @Failure 422 {object} dto.ErrorResponse "Validatsiya xatoligi"
// @Router /api/v1/auth/clinic/register [post]
func (h *AuthHandler) RegisterClinic(c *gin.Context) {
	var req dto.ClinicRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse(
			"Kiritilgan ma'lumotlar formati noto‘g‘ri",
			map[string]string{"body": err.Error()},
		))
		return
	}

	data, err := h.authService.RegisterClinic(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, repositories.ErrEmailAlreadyExist) {
			c.JSON(http.StatusConflict, dto.NewErrorResponse(err.Error(), map[string]string{"email": err.Error()}))
			return
		}
		if errors.Is(err, repositories.ErrPhoneAlreadyExist) {
			c.JSON(http.StatusConflict, dto.NewErrorResponse(err.Error(), map[string]string{"phone": err.Error()}))
			return
		}
		if errors.Is(err, services.ErrPasswordsDoNotMatch) {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), map[string]string{"confirm_password": err.Error()}))
			return
		}
		if errors.Is(err, services.ErrTermsNotAccepted) {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), map[string]string{"accept_terms": err.Error()}))
			return
		}

		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusCreated, dto.NewSuccessResponse(
		"Muvaffaqiyatli ro‘yxatdan o‘tdingiz! Emailingizga tasdiqlash kodi yuborildi.",
		data,
	))
}

// LoginClinic godoc
// @Summary Klinika kabinetiga kirish
// @Description Email va parol orqali tizimga kiradi (faqat clinic_owner yoki clinic_admin)
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ClinicLoginRequest true "Klinika login ma’lumotlari"
// @Success 200 {object} dto.SuccessResponse{data=dto.LoginData} "Muvaffaqiyatli kirildi"
// @Failure 401 {object} dto.ErrorResponse "Email yoki parol noto‘g‘ri"
// @Failure 403 {object} dto.ErrorResponse "Email tasdiqlanmagan yoki ruxsat yo‘q"
// @Router /api/v1/auth/clinic/login [post]
func (h *AuthHandler) LoginClinic(c *gin.Context) {
	var req dto.ClinicLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Email va parol kiritilishi shart", nil))
		return
	}

	data, rawRefreshToken, expiresAt, err := h.authService.LoginClinic(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, services.ErrEmailNotVerified) || errors.Is(err, services.ErrAccountBlocked) || errors.Is(err, services.ErrRoleNotAllowed) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	h.setRefreshTokenCookie(c, rawRefreshToken, expiresAt)
	c.JSON(http.StatusOK, dto.NewSuccessResponse("Muvaffaqiyatli tizimga kirdingiz", data))
}

// LoginSuperAdmin godoc
// @Summary Superadmin paneliga kirish
// @Description Faqat superadmin huquqiga ega foydalanuvchilar kirishi uchun
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.SuperAdminLoginRequest true "Superadmin login ma’lumotlari"
// @Success 200 {object} dto.SuccessResponse{data=dto.LoginData} "Muvaffaqiyatli kirildi"
// @Failure 401 {object} dto.ErrorResponse "Email yoki parol noto‘g‘ri"
// @Failure 403 {object} dto.ErrorResponse "Ruxsat berilmagan"
// @Router /api/v1/auth/admin/login [post]
func (h *AuthHandler) LoginSuperAdmin(c *gin.Context) {
	var req dto.SuperAdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Email va parol kiritilishi shart", nil))
		return
	}

	data, rawRefreshToken, expiresAt, err := h.authService.LoginSuperAdmin(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, services.ErrEmailNotVerified) || errors.Is(err, services.ErrAccountBlocked) || errors.Is(err, services.ErrRoleNotAllowed) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	h.setRefreshTokenCookie(c, rawRefreshToken, expiresAt)
	c.JSON(http.StatusOK, dto.NewSuccessResponse("Superadmin sifatida muvaffaqiyatli tizimga kirdingiz", data))
}

// ResendVerification godoc
// @Summary Tasdiqlash kodini qayta yuborish
// @Description Emailga yangi bir martalik tasdiqlash kodini yuboradi
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ResendVerificationRequest true "Email"
// @Success 200 {object} dto.SuccessResponse "Kod qayta yuborildi"
// @Failure 400 {object} dto.ErrorResponse "Xatolik"
// @Router /api/v1/auth/resend-verification [post]
func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var req dto.ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Email kiritilishi shart", nil))
		return
	}

	if err := h.authService.ResendVerification(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Agar email to‘g‘ri bo‘lsa, yangi tasdiqlash kodi yuborildi.", nil))
}

// Login godoc
// @Summary Bemor kabinetiga kirish
// @Description Email va parol orqali tizimga kiradi, access token va HttpOnly refresh token cookie qaytaradi
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.PatientLoginRequest true "Login ma’lumotlari"
// @Success 200 {object} dto.SuccessResponse{data=dto.LoginData} "Muvaffaqiyatli kirildi"
// @Failure 401 {object} dto.ErrorResponse "Email yoki parol noto‘g‘ri"
// @Failure 403 {object} dto.ErrorResponse "Email tasdiqlanmagan yoki akkaunt bloklangan"
// @Router /api/v1/auth/patient/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.PatientLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Email va parol kiritilishi shart", nil))
		return
	}

	data, rawRefreshToken, expiresAt, err := h.authService.LoginPatient(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, services.ErrEmailNotVerified) || errors.Is(err, services.ErrAccountBlocked) || errors.Is(err, services.ErrRoleNotAllowed) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	// HttpOnly Refresh Token Cookie o'rnatish
	h.setRefreshTokenCookie(c, rawRefreshToken, expiresAt)

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Muvaffaqiyatli tizimga kirdingiz", data))
}

// Refresh godoc
// @Summary Access tokenni yangilash
// @Description HttpOnly cookie yoki body'dagi refresh token orqali yangi access token va yangi refresh token beradi
// @Tags Auth
// @Produce json
// @Success 200 {object} dto.SuccessResponse "Token yangilandi"
// @Failure 401 {object} dto.ErrorResponse "Yaroqsiz yoki muddati tugagan refresh token"
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var rawRefreshToken string

	// 1. Avval Cookie'dan qidiramiz
	cookieToken, err := c.Cookie(RefreshTokenCookieName)
	if err == nil && cookieToken != "" {
		rawRefreshToken = cookieToken
	} else {
		// 2. Body yoki Headerdan tekshiramiz
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := c.ShouldBindJSON(&body); err == nil && body.RefreshToken != "" {
			rawRefreshToken = body.RefreshToken
		}
	}

	if rawRefreshToken == "" {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Refresh token topilmadi", nil))
		return
	}

	newAccessToken, newRefreshToken, expiresAt, err := h.tokenService.RotateRefreshToken(c.Request.Context(), rawRefreshToken)
	if err != nil {
		// Token yaroqsiz bo'lsa cookie'ni tozalaymiz
		h.clearRefreshTokenCookie(c)
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	// Yangi refresh token cookie
	h.setRefreshTokenCookie(c, newRefreshToken, expiresAt)

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Token muvaffaqiyatli yangilandi", gin.H{
		"access_token": newAccessToken,
		"token_type":   "Bearer",
		"expires_in":   int64(h.cfg.AccessTokenTTL.Seconds()),
	}))
}

// Logout godoc
// @Summary Kabinetdan chiqish
// @Description Refresh token sessiyasini bekor qiladi va cookie'ni tozalaydi
// @Tags Auth
// @Produce json
// @Success 200 {object} dto.SuccessResponse "Muvaffaqiyatli chiqildi"
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	cookieToken, _ := c.Cookie(RefreshTokenCookieName)
	if cookieToken != "" {
		_ = h.tokenService.RevokeSession(c.Request.Context(), cookieToken)
	}

	// Cookieni tozalash
	h.clearRefreshTokenCookie(c)

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Muvaffaqiyatli tizimdan chiqildi", nil))
}

// ForgotPassword godoc
// @Summary Parolni unutganda tiklash so‘rovi
// @Description Email kiritiladi, agar mavjud bo‘lsa parolni tiklash kodi yuboriladi
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ForgotPasswordRequest true "Email"
// @Success 200 {object} dto.SuccessResponse "Tiklash kodi yuborildi"
// @Router /api/v1/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Email kiritilishi shart", nil))
		return
	}

	_ = h.authService.ForgotPassword(c.Request.Context(), &req)

	// Har doim bir xil javob qaytaramiz (email mavjudligini oshkor qilmaslik)
	c.JSON(http.StatusOK, dto.NewSuccessResponse(
		"Agar ushbu email tizimda mavjud bo‘lsa, parolni tiklash kodi yuborildi.",
		nil,
	))
}

// ResetPassword godoc
// @Summary Parolni tiklash
// @Description Kod va yangi parol orqali parolni o‘zgartiradi va mavjud sessiyalarni to‘xtatadi
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ResetPasswordRequest true "Tiklash ma’lumotlari"
// @Success 200 {object} dto.SuccessResponse "Parol muvaffaqiyatli o‘zgartirildi"
// @Failure 400 {object} dto.ErrorResponse "Kod yoki parol noto‘g‘ri"
// @Router /api/v1/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Barcha maydonlar to‘ldirilishi shart", nil))
		return
	}

	if err := h.authService.ResetPassword(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse(
		"Parolingiz muvaffaqiyatli yangilandi. Endi yangi parol bilan tizimga kirishingiz mumkin.",
		nil,
	))
}
