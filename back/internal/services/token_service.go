package services

import (
	"context"
	"errors"
	"time"

	"qabulgo-backend/internal/config"
	"qabulgo-backend/internal/models"
	"qabulgo-backend/internal/repositories"
	"qabulgo-backend/internal/utils"

	"github.com/google/uuid"
)

var (
	ErrSessionExpired = errors.New("refresh sessiyasining muddati tugagan")
	ErrSessionRevoked = errors.New("refresh sessiyasi allaqachon bekor qilingan")
	ErrTokenReused    = errors.New("bekor qilingan refresh token qayta ishlatildi, barcha sessiyalar xavfsizlik maqsadida to‘xtatildi")
)

type TokenService interface {
	GenerateTokenPair(ctx context.Context, userID uuid.UUID, role string) (accessToken, rawRefreshToken string, expiresAt time.Time, err error)
	RotateRefreshToken(ctx context.Context, rawRefreshToken string) (newAccessToken, newRefreshToken string, expiresAt time.Time, err error)
	RevokeSession(ctx context.Context, rawRefreshToken string) error
	RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error
}

type tokenService struct {
	cfg         *config.Config
	sessionRepo repositories.RefreshSessionRepository
	userRepo    repositories.UserRepository
}

func NewTokenService(cfg *config.Config, sessionRepo repositories.RefreshSessionRepository, userRepo repositories.UserRepository) TokenService {
	return &tokenService{
		cfg:         cfg,
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
	}
}

const PatientAndClinicSessionTTL = 30 * 24 * time.Hour

func (s *tokenService) GenerateTokenPair(ctx context.Context, userID uuid.UUID, role string) (string, string, time.Time, error) {
	now := time.Now().UTC()

	var sessionTTL time.Duration
	var accessTTL time.Duration
	if role == string(models.RolePatient) || role == string(models.RoleClinicOwner) || role == string(models.RoleClinicAdmin) || role == string(models.RoleDoctor) {
		sessionTTL = PatientAndClinicSessionTTL // Bemor, klinika va shifokor uchun 30 kunlik sessiya
		accessTTL = PatientAndClinicSessionTTL  // 30 kunlik token
	} else {
		// SuperAdmin uchun standart konfiguratsiyadagi TTL (7 kun refresh, 15 min access) saqlanadi
		sessionTTL = s.cfg.RefreshTokenTTL
		accessTTL = s.cfg.AccessTokenTTL
	}

	// Access Token yaratish
	accessToken, err := utils.GenerateAccessToken(userID, role, accessTTL, s.cfg.JWTSecret, s.cfg.JWTIssuer, s.cfg.JWTAudience)
	if err != nil {
		return "", "", time.Time{}, err
	}

	// Kriptografik xavfsiz Refresh Token yaratish
	rawRefreshToken, err := utils.GenerateRandomToken(32)
	if err != nil {
		return "", "", time.Time{}, err
	}

	tokenHash := utils.HashToken(rawRefreshToken)

	maxExpiresAt := now.Add(sessionTTL)
	expiresAt := maxExpiresAt

	session := &models.RefreshSession{
		UserID:       userID,
		TokenHash:    tokenHash,
		ExpiresAt:    expiresAt,
		MaxExpiresAt: maxExpiresAt,
		CreatedAt:    now,
		LastUsedAt:   now,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return "", "", time.Time{}, err
	}

	return accessToken, rawRefreshToken, expiresAt, nil
}

func (s *tokenService) RotateRefreshToken(ctx context.Context, rawRefreshToken string) (string, string, time.Time, error) {
	tokenHash := utils.HashToken(rawRefreshToken)
	session, err := s.sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return "", "", time.Time{}, repositories.ErrSessionNotFound
	}

	// Agar token allaqachon bekor qilingan bo'lsa (re-use attack!)
	if session.RevokedAt != nil {
		// Ushbu foydalanuvchining barcha aktiv sessiyalarini bekor qilish
		_ = s.sessionRepo.RevokeAllByUserID(ctx, session.UserID)
		return "", "", time.Time{}, ErrTokenReused
	}

	now := time.Now().UTC()
	maxExpiresAt := session.MaxExpiresAt
	if maxExpiresAt.IsZero() {
		maxExpiresAt = session.ExpiresAt
	}

	// 1. Dastlabki login vaqtidan hisoblangan mutlaq (absolute) maksimal 30 kunlik muddat tekshiriladi
	// 2. Oddiy expires_at muddati tekshiriladi
	if now.After(session.ExpiresAt) || now.After(maxExpiresAt) {
		return "", "", time.Time{}, ErrSessionExpired
	}

	// Foydalanuvchini olish va statusini tekshirish
	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return "", "", time.Time{}, err
	}
	if !user.IsActive() {
		return "", "", time.Time{}, errors.New("foydalanuvchi akkaunti faol emas")
	}

	// Eski sessiyani bekor qilish (Revoke)
	if err := s.sessionRepo.Revoke(ctx, session.ID); err != nil {
		return "", "", time.Time{}, err
	}

	// Yangi Refresh Token yaratish
	newRawRefreshToken, err := utils.GenerateRandomToken(32)
	if err != nil {
		return "", "", time.Time{}, err
	}

	newTokenHash := utils.HashToken(newRawRefreshToken)

	// Yangi sessiyaning amal qilish muddati:
	// Bemor, klinika va shifokor akkauntlari uchun dastlabki logindan hisoblangan mutlaq 30 kunlik chegara (maxExpiresAt) saqlanadi
	// 30 kundan ortiq uzaytirilmaydi!
	// SuperAdmin uchun standart RefreshTokenTTL saqlanadi
	var accessTTL time.Duration
	var newExpiresAt time.Time
	var newMaxExpiresAt time.Time

	if user.Role == models.RolePatient || user.Role == models.RoleClinicOwner || user.Role == models.RoleClinicAdmin || user.Role == models.RoleDoctor {
		accessTTL = PatientAndClinicSessionTTL
		newMaxExpiresAt = maxExpiresAt // Dastlabki logindagi mutlaq 30 kunlik sana uzaytirilmaydi
		newExpiresAt = maxExpiresAt
	} else {
		// Superadmin
		accessTTL = s.cfg.AccessTokenTTL
		newExpiresAt = now.Add(s.cfg.RefreshTokenTTL)
		newMaxExpiresAt = newExpiresAt
	}

	// Yangi Access Token yaratish
	accessToken, err := utils.GenerateAccessToken(user.ID, string(user.Role), accessTTL, s.cfg.JWTSecret, s.cfg.JWTIssuer, s.cfg.JWTAudience)
	if err != nil {
		return "", "", time.Time{}, err
	}

	newSession := &models.RefreshSession{
		UserID:       user.ID,
		TokenHash:    newTokenHash,
		ExpiresAt:    newExpiresAt,
		MaxExpiresAt: newMaxExpiresAt,
		CreatedAt:    now,
		LastUsedAt:   now,
	}

	if err := s.sessionRepo.Create(ctx, newSession); err != nil {
		return "", "", time.Time{}, err
	}

	return accessToken, newRawRefreshToken, newExpiresAt, nil
}

func (s *tokenService) RevokeSession(ctx context.Context, rawRefreshToken string) error {
	tokenHash := utils.HashToken(rawRefreshToken)
	session, err := s.sessionRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		// Agar sessiya topilmasa, allaqachon yo'q deb hisoblaymiz
		return nil
	}
	return s.sessionRepo.Revoke(ctx, session.ID)
}

func (s *tokenService) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	return s.sessionRepo.RevokeAllByUserID(ctx, userID)
}
