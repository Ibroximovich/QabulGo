package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"qabulgo-backend/internal/config"
	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/models"
	"qabulgo-backend/internal/repositories"
	"qabulgo-backend/internal/utils"

	"github.com/google/uuid"
)

var (
	ErrPasswordsDoNotMatch  = errors.New("parol va tasdiqlash paroli mos kelmadi")
	ErrTermsNotAccepted     = errors.New("foydalanish shartlariga rozilik bildirilishi shart")
	ErrInvalidCredentials   = errors.New("email yoki parol noto‘g‘ri")
	ErrAccountBlocked       = errors.New("akkauntingiz bloklangan, iltimos qo‘llab-quvvatlash xizmatiga murojaat qiling")
	ErrEmailNotVerified     = errors.New("email tasdiqlanmagan, iltimos emailingizga yuborilgan kod orqali tasdiqlang")
	ErrRoleNotAllowed       = errors.New("ushbu portal orqali kirish huquqiga ega emassiz")
	ErrInvalidOrExpiredCode = errors.New("tasdiqlash kodi noto‘g‘ri yoki muddati tugagan")
	ErrTooManyAttempts      = errors.New("noto‘g‘ri urinishlar soni oshib ketdi, iltimos yangi kod so‘rang")
	ErrNameLength           = errors.New("ism va familiya 2 dan 50 tagacha belgidan iborat bo‘lishi kerak")
	ErrInvalidEmailFormat   = errors.New("email formati noto‘g‘ri")
	ErrInvalidPhoneFormat   = errors.New("telefon raqami noto‘g‘ri (+998XXXXXXXXX formatida bo‘lishi kerak)")
	ErrEmailAlreadyVerified = errors.New("ushbu email allaqachon tasdiqlangan")
)

type AuthService interface {
	RegisterPatient(ctx context.Context, req *dto.PatientRegisterRequest) (*dto.PatientRegisterData, error)
	RegisterClinic(ctx context.Context, req *dto.ClinicRegisterRequest) (*dto.PatientRegisterData, error)
	VerifyEmail(ctx context.Context, req *dto.VerifyEmailRequest) (*dto.LoginData, string, error)
	ResendVerification(ctx context.Context, req *dto.ResendVerificationRequest) error
	LoginPatient(ctx context.Context, req *dto.PatientLoginRequest) (*dto.LoginData, string, error)
	LoginClinic(ctx context.Context, req *dto.ClinicLoginRequest) (*dto.LoginData, string, error)
	LoginSuperAdmin(ctx context.Context, req *dto.SuperAdminLoginRequest) (*dto.LoginData, string, error)
	ForgotPassword(ctx context.Context, req *dto.ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error
}

type authService struct {
	cfg          *config.Config
	userRepo     repositories.UserRepository
	tokenRepo    repositories.VerificationTokenRepository
	tokenService TokenService
	emailService EmailService
}

func NewAuthService(
	cfg *config.Config,
	userRepo repositories.UserRepository,
	tokenRepo repositories.VerificationTokenRepository,
	tokenService TokenService,
	emailService EmailService,
) AuthService {
	return &authService{
		cfg:          cfg,
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		tokenService: tokenService,
		emailService: emailService,
	}
}

func (s *authService) RegisterPatient(ctx context.Context, req *dto.PatientRegisterRequest) (*dto.PatientRegisterData, error) {
	return s.registerUser(ctx, req.FirstName, req.LastName, req.Email, req.Phone, req.Password, req.ConfirmPassword, req.AcceptTerms, models.RolePatient)
}

func (s *authService) RegisterClinic(ctx context.Context, req *dto.ClinicRegisterRequest) (*dto.PatientRegisterData, error) {
	return s.registerUser(ctx, req.FirstName, req.LastName, req.Email, req.Phone, req.Password, req.ConfirmPassword, req.AcceptTerms, models.RoleClinicOwner)
}

func (s *authService) registerUser(
	ctx context.Context,
	firstNameRaw, lastNameRaw, emailRaw, phoneRaw, password, confirmPassword string,
	acceptTerms bool,
	role models.UserRole,
) (*dto.PatientRegisterData, error) {
	// 1. Shartlarga rozilik
	if !acceptTerms {
		return nil, ErrTermsNotAccepted
	}

	// 2. Parollarni solishtirish
	if password != confirmPassword {
		return nil, ErrPasswordsDoNotMatch
	}

	// 3. Parol murakkabligi
	if err := utils.ValidatePassword(password); err != nil {
		return nil, err
	}

	// 4. Ism va familiya tozalash va tekshirish
	firstName := utils.SanitizeString(firstNameRaw)
	lastName := utils.SanitizeString(lastNameRaw)
	if len(firstName) < 2 || len(firstName) > 50 || len(lastName) < 2 || len(lastName) > 50 {
		return nil, ErrNameLength
	}

	// 5. Email normalizatsiya va validatsiya
	email := utils.NormalizeEmail(emailRaw)
	if !utils.ValidateEmail(email) {
		return nil, ErrInvalidEmailFormat
	}

	// 6. Telefon raqami tekshirish
	phone := utils.SanitizeString(phoneRaw)
	if !utils.ValidateUzbekistanPhone(phone) {
		return nil, ErrInvalidPhoneFormat
	}

	// 7. Bazada mavjudlikni tekshirish
	if _, err := s.userRepo.GetByEmail(ctx, email); err == nil {
		return nil, repositories.ErrEmailAlreadyExist
	}
	if _, err := s.userRepo.GetByPhone(ctx, phone); err == nil {
		return nil, repositories.ErrPhoneAlreadyExist
	}

	// 8. Parolni Argon2id orqali xeshlash
	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// 9. Foydalanuvchini yaratish
	now := time.Now().UTC()
	user := &models.User{
		ID:              uuid.New(),
		FirstName:       firstName,
		LastName:        lastName,
		Email:           email,
		Phone:           phone,
		PasswordHash:    passwordHash,
		Role:            role,
		Status:          models.StatusActive,
		TermsAcceptedAt: now,
		TermsVersion:    s.cfg.TermsVersion,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if strings.Contains(err.Error(), "uq_users_email") || strings.Contains(strings.ToLower(err.Error()), "email") {
			return nil, repositories.ErrEmailAlreadyExist
		}
		if strings.Contains(err.Error(), "uq_users_phone") || strings.Contains(strings.ToLower(err.Error()), "phone") {
			return nil, repositories.ErrPhoneAlreadyExist
		}
		return nil, err
	}

	// 10. Email tasdiqlash uchun 6 xonali kod yaratish
	code, err := utils.Generate6DigitCode()
	if err != nil {
		return nil, err
	}

	tokenHash := utils.HashToken(code)
	verToken := &models.VerificationToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Purpose:   models.PurposeEmailVerification,
		TokenHash: tokenHash,
		ExpiresAt: now.Add(10 * time.Minute), // 10 daqiqa amal qiladi
		Attempts:  0,
		CreatedAt: now,
	}

	if err := s.tokenRepo.Create(ctx, verToken); err != nil {
		return nil, fmt.Errorf("tasdiqlash tokenini saqlashda xatolik: %w", err)
	}

	// 11. Email orqali yuborish
	if err := s.emailService.SendVerificationEmail(user.Email, code); err != nil {
		// Log error, lekin tranzaksiyani bekor qilmaymiz (resend mavjud)
	}

	return &dto.PatientRegisterData{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Phone:     user.Phone,
		Role:      string(user.Role),
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *authService) VerifyEmail(ctx context.Context, req *dto.VerifyEmailRequest) (*dto.LoginData, string, error) {
	email := utils.NormalizeEmail(req.Email)
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", ErrInvalidOrExpiredCode
	}

	if user.IsEmailVerified() {
		return nil, "", ErrEmailAlreadyVerified
	}

	token, err := s.tokenRepo.GetLatestActive(ctx, user.ID, models.PurposeEmailVerification)
	if err != nil {
		return nil, "", ErrInvalidOrExpiredCode
	}

	if token.Attempts >= 5 {
		return nil, "", ErrTooManyAttempts
	}

	inputHash := utils.HashToken(strings.TrimSpace(req.Code))
	if inputHash != token.TokenHash {
		_ = s.tokenRepo.IncrementAttempts(ctx, token.ID)
		return nil, "", ErrInvalidOrExpiredCode
	}

	// Kod to'g'ri! Mark consumed
	if err := s.tokenRepo.MarkConsumed(ctx, token.ID); err != nil {
		return nil, "", err
	}

	// Update user verified_at
	now := time.Now().UTC()
	if err := s.userRepo.UpdateEmailVerified(ctx, user.ID, now); err != nil {
		return nil, "", err
	}

	// Automatic Session generation so user doesn't have to re-login!
	accessToken, rawRefreshToken, _, err := s.tokenService.GenerateTokenPair(ctx, user.ID, string(user.Role))
	if err != nil {
		return nil, "", err
	}

	loginData := &dto.LoginData{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.cfg.AccessTokenTTL.Seconds()),
		User: dto.UserSummary{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Phone:     user.Phone,
			Role:      string(user.Role),
		},
	}

	return loginData, rawRefreshToken, nil
}

func (s *authService) ResendVerification(ctx context.Context, req *dto.ResendVerificationRequest) error {
	email := utils.NormalizeEmail(req.Email)
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil
	}

	if user.IsEmailVerified() {
		return ErrEmailAlreadyVerified
	}

	// Eski faol kodlarni bekor qilish
	_ = s.tokenRepo.RevokeAllActiveByUserID(ctx, user.ID, models.PurposeEmailVerification)

	// Yangi 6 xonali kod yaratish
	code, err := utils.Generate6DigitCode()
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	verToken := &models.VerificationToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Purpose:   models.PurposeEmailVerification,
		TokenHash: utils.HashToken(code),
		ExpiresAt: now.Add(10 * time.Minute),
		Attempts:  0,
		CreatedAt: now,
	}

	if err := s.tokenRepo.Create(ctx, verToken); err != nil {
		return err
	}

	return s.emailService.SendVerificationEmail(user.Email, code)
}

func (s *authService) LoginPatient(ctx context.Context, req *dto.PatientLoginRequest) (*dto.LoginData, string, error) {
	return s.loginWithRoles(ctx, req.Email, req.Password, []models.UserRole{models.RolePatient})
}

func (s *authService) LoginClinic(ctx context.Context, req *dto.ClinicLoginRequest) (*dto.LoginData, string, error) {
	return s.loginWithRoles(ctx, req.Email, req.Password, []models.UserRole{models.RoleClinicOwner, models.RoleClinicAdmin})
}

func (s *authService) LoginSuperAdmin(ctx context.Context, req *dto.SuperAdminLoginRequest) (*dto.LoginData, string, error) {
	return s.loginWithRoles(ctx, req.Email, req.Password, []models.UserRole{models.RoleSuperAdmin})
}

func (s *authService) loginWithRoles(ctx context.Context, emailRaw, password string, allowedRoles []models.UserRole) (*dto.LoginData, string, error) {
	email := utils.NormalizeEmail(emailRaw)
	user, err := s.userRepo.GetByEmail(ctx, email)

	// Timing attack oldini olish
	if err != nil {
		dummyHash := "$argon2id$v=19$m=65536,t=3,p=2$c29tZXNhbHQxMjM0NTY3OA$c29tZWhhc2gxMjM0NTY3OA"
		_, _ = utils.VerifyPassword(password, dummyHash)
		return nil, "", ErrInvalidCredentials
	}

	// Parolni tekshirish
	match, err := utils.VerifyPassword(password, user.PasswordHash)
	if err != nil || !match {
		return nil, "", ErrInvalidCredentials
	}

	// Akkaunt holati tekshiruvi
	if !user.IsActive() {
		return nil, "", ErrAccountBlocked
	}

	// Email tasdiqlanganligini tekshirish
	if !user.IsEmailVerified() {
		return nil, "", ErrEmailNotVerified
	}

	// Role tekshiruvi
	roleAllowed := false
	for _, r := range allowedRoles {
		if user.Role == r {
			roleAllowed = true
			break
		}
	}
	if !roleAllowed {
		return nil, "", ErrRoleNotAllowed
	}

	// Token juftligini yaratish
	accessToken, rawRefreshToken, _, err := s.tokenService.GenerateTokenPair(ctx, user.ID, string(user.Role))
	if err != nil {
		return nil, "", err
	}

	loginData := &dto.LoginData{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.cfg.AccessTokenTTL.Seconds()),
		User: dto.UserSummary{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Phone:     user.Phone,
			Role:      string(user.Role),
		},
	}

	return loginData, rawRefreshToken, nil
}

func (s *authService) ForgotPassword(ctx context.Context, req *dto.ForgotPasswordRequest) error {
	email := utils.NormalizeEmail(req.Email)
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil
	}

	_ = s.tokenRepo.RevokeAllActiveByUserID(ctx, user.ID, models.PurposePasswordReset)

	code, err := utils.Generate6DigitCode()
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	token := &models.VerificationToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Purpose:   models.PurposePasswordReset,
		TokenHash: utils.HashToken(code),
		ExpiresAt: now.Add(15 * time.Minute),
		Attempts:  0,
		CreatedAt: now,
	}

	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return err
	}

	return s.emailService.SendPasswordResetEmail(user.Email, code)
}

func (s *authService) ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error {
	if req.NewPassword != req.ConfirmPassword {
		return ErrPasswordsDoNotMatch
	}

	if err := utils.ValidatePassword(req.NewPassword); err != nil {
		return err
	}

	email := utils.NormalizeEmail(req.Email)
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return ErrInvalidOrExpiredCode
	}

	token, err := s.tokenRepo.GetLatestActive(ctx, user.ID, models.PurposePasswordReset)
	if err != nil {
		return ErrInvalidOrExpiredCode
	}

	if token.Attempts >= 5 {
		return ErrTooManyAttempts
	}

	inputHash := utils.HashToken(strings.TrimSpace(req.Code))
	if inputHash != token.TokenHash {
		_ = s.tokenRepo.IncrementAttempts(ctx, token.ID)
		return ErrInvalidOrExpiredCode
	}

	// Yangi parolni xeshlash
	newHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	// Bazada parolni yangilash
	if err := s.userRepo.UpdatePassword(ctx, user.ID, newHash); err != nil {
		return err
	}

	// Agar email hali tasdiqlanmagan bo'lsa, parolni tiklash orqali email egaligi isbotlangani uchun tasdiqlaymiz
	if !user.IsEmailVerified() {
		now := time.Now().UTC()
		_ = s.userRepo.UpdateEmailVerified(ctx, user.ID, now)
	}

	_ = s.tokenRepo.MarkConsumed(ctx, token.ID)
	_ = s.tokenService.RevokeAllUserSessions(ctx, user.ID)

	return nil
}
