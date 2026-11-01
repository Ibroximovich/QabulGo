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

func (s *tokenService) GenerateTokenPair(ctx context.Context, userID uuid.UUID, role string) (string, string, time.Time, error) {
	// Access Token yaratish
	accessToken, err := utils.GenerateAccessToken(userID, role, s.cfg.AccessTokenTTL, s.cfg.JWTSecret, s.cfg.JWTIssuer, s.cfg.JWTAudience)
	if err != nil {
		return "", "", time.Time{}, err
	}

	// Kriptografik xavfsiz Refresh Token yaratish
	rawRefreshToken, err := utils.GenerateRandomToken(32)
	if err != nil {
		return "", "", time.Time{}, err
	}

	tokenHash := utils.HashToken(rawRefreshToken)
	expiresAt := time.Now().UTC().Add(s.cfg.RefreshTokenTTL)

	session := &models.RefreshSession{
		UserID:     userID,
		TokenHash:  tokenHash,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now().UTC(),
		LastUsedAt: time.Now().UTC(),
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

	// Muddati tekshiriladi
	if time.Now().UTC().After(session.ExpiresAt) {
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

	// Yangi juftlik yaratish
	return s.GenerateTokenPair(ctx, user.ID, string(user.Role))
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
