package services

import (
	"context"
	"testing"
	"time"

	"qabulgo-backend/internal/config"
	"qabulgo-backend/internal/models"
	"qabulgo-backend/internal/repositories"
	"qabulgo-backend/internal/utils"

	"github.com/google/uuid"
)

type mockSessionRepo struct {
	sessions    map[string]*models.RefreshSession
	revokedIDs  map[uuid.UUID]bool
	revokedUser map[uuid.UUID]bool
}

func newMockSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{
		sessions:    make(map[string]*models.RefreshSession),
		revokedIDs:  make(map[uuid.UUID]bool),
		revokedUser: make(map[uuid.UUID]bool),
	}
}

func (m *mockSessionRepo) Create(ctx context.Context, session *models.RefreshSession) error {
	m.sessions[session.TokenHash] = session
	return nil
}

func (m *mockSessionRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshSession, error) {
	s, ok := m.sessions[tokenHash]
	if !ok {
		return nil, repositories.ErrSessionNotFound
	}
	return s, nil
}

func (m *mockSessionRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	m.revokedIDs[id] = true
	now := time.Now().UTC()
	for _, s := range m.sessions {
		if s.ID == id {
			s.RevokedAt = &now
		}
	}
	return nil
}

func (m *mockSessionRepo) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	m.revokedUser[userID] = true
	now := time.Now().UTC()
	for _, s := range m.sessions {
		if s.UserID == userID {
			s.RevokedAt = &now
		}
	}
	return nil
}

func (m *mockSessionRepo) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	return nil
}

type mockUserRepo struct {
	users map[uuid.UUID]*models.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[uuid.UUID]*models.User)}
}

func (m *mockUserRepo) Create(ctx context.Context, user *models.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, repositories.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, repositories.ErrUserNotFound
}

func (m *mockUserRepo) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	return nil, repositories.ErrUserNotFound
}

func (m *mockUserRepo) Update(ctx context.Context, user *models.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) UpdateEmailVerified(ctx context.Context, id uuid.UUID, verifiedAt time.Time) error {
	if u, ok := m.users[id]; ok {
		u.EmailVerifiedAt = &verifiedAt
	}
	return nil
}

func (m *mockUserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return nil
}

func TestPatientSession30DaysTTL(t *testing.T) {
	cfg := &config.Config{
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		JWTSecret:       "test-secret-at-least-32-chars-long-123456",
		JWTIssuer:       "qabulgo-test",
		JWTAudience:     "qabulgo-test",
	}

	sessionRepo := newMockSessionRepo()
	userRepo := newMockUserRepo()
	svc := NewTokenService(cfg, sessionRepo, userRepo)

	userID := uuid.New()
	_, rawRefresh, expiresAt, err := svc.GenerateTokenPair(context.Background(), userID, string(models.RolePatient))
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}

	// Bemor sessiyasi 30 kun bo'lishi kerak
	expectedMinExpiry := time.Now().UTC().Add(29 * 24 * time.Hour)
	expectedMaxExpiry := time.Now().UTC().Add(31 * 24 * time.Hour)

	if expiresAt.Before(expectedMinExpiry) || expiresAt.After(expectedMaxExpiry) {
		t.Errorf("Expected expiresAt ~30 days, got: %v", expiresAt)
	}

	hash := utils.HashToken(rawRefresh)
	savedSession := sessionRepo.sessions[hash]
	if savedSession == nil {
		t.Fatal("Session not saved in repository")
	}

	if savedSession.MaxExpiresAt.Before(expectedMinExpiry) {
		t.Errorf("Expected MaxExpiresAt to be ~30 days, got: %v", savedSession.MaxExpiresAt)
	}
}

func TestSuperAdminSessionKeepsStandardTTL(t *testing.T) {
	cfg := &config.Config{
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour, // Superadmin 7 kun
		JWTSecret:       "test-secret-at-least-32-chars-long-123456",
		JWTIssuer:       "qabulgo-test",
		JWTAudience:     "qabulgo-test",
	}

	sessionRepo := newMockSessionRepo()
	userRepo := newMockUserRepo()
	svc := NewTokenService(cfg, sessionRepo, userRepo)

	adminID := uuid.New()
	_, _, expiresAt, err := svc.GenerateTokenPair(context.Background(), adminID, string(models.RoleSuperAdmin))
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}

	// Superadmin sessiyasi 7 kun bo'lishi kerak
	expectedMinExpiry := time.Now().UTC().Add(6 * 24 * time.Hour)
	expectedMaxExpiry := time.Now().UTC().Add(8 * 24 * time.Hour)

	if expiresAt.Before(expectedMinExpiry) || expiresAt.After(expectedMaxExpiry) {
		t.Errorf("SuperAdmin expiresAt should remain 7 days, got: %v", expiresAt)
	}
}

func TestRotationPreservesAbsolute30DayLimit(t *testing.T) {
	cfg := &config.Config{
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		JWTSecret:       "test-secret-at-least-32-chars-long-123456",
		JWTIssuer:       "qabulgo-test",
		JWTAudience:     "qabulgo-test",
	}

	sessionRepo := newMockSessionRepo()
	userRepo := newMockUserRepo()

	user := &models.User{
		ID:     uuid.New(),
		Role:   models.RolePatient,
		Status: models.StatusActive,
	}
	_ = userRepo.Create(context.Background(), user)

	svc := NewTokenService(cfg, sessionRepo, userRepo)

	// 1. Initial login (MaxExpiresAt = 30 days from now)
	_, rawRefresh1, initialExpiry, err := svc.GenerateTokenPair(context.Background(), user.ID, string(models.RolePatient))
	if err != nil {
		t.Fatalf("Failed initial token pair: %v", err)
	}

	// 2. Rotate after some time
	_, rawRefresh2, rotatedExpiry1, err := svc.RotateRefreshToken(context.Background(), rawRefresh1)
	if err != nil {
		t.Fatalf("First rotation failed: %v", err)
	}

	// Rotated token must have the EXACT SAME MaxExpiresAt as the initial session!
	hash2 := utils.HashToken(rawRefresh2)
	session2 := sessionRepo.sessions[hash2]
	if !session2.MaxExpiresAt.Equal(initialExpiry) {
		t.Errorf("Expected MaxExpiresAt to remain %v, got %v", initialExpiry, session2.MaxExpiresAt)
	}
	if !rotatedExpiry1.Equal(initialExpiry) {
		t.Errorf("Expected rotatedExpiry to equal initialExpiry %v, got %v", initialExpiry, rotatedExpiry1)
	}

	// 3. Simulate session that has exceeded 30 days (expired)
	session2.MaxExpiresAt = time.Now().UTC().Add(-1 * time.Minute)
	session2.ExpiresAt = time.Now().UTC().Add(-1 * time.Minute)

	_, _, _, err = svc.RotateRefreshToken(context.Background(), rawRefresh2)
	if err != ErrSessionExpired {
		t.Errorf("Expected ErrSessionExpired for session past 30 days, got: %v", err)
	}
}

func TestRevokedTokenReuseTriggersAllRevoked(t *testing.T) {
	cfg := &config.Config{
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		JWTSecret:       "test-secret-at-least-32-chars-long-123456",
		JWTIssuer:       "qabulgo-test",
		JWTAudience:     "qabulgo-test",
	}

	sessionRepo := newMockSessionRepo()
	userRepo := newMockUserRepo()

	user := &models.User{
		ID:     uuid.New(),
		Role:   models.RoleClinicOwner,
		Status: models.StatusActive,
	}
	_ = userRepo.Create(context.Background(), user)

	svc := NewTokenService(cfg, sessionRepo, userRepo)

	_, rawRefresh1, _, _ := svc.GenerateTokenPair(context.Background(), user.ID, string(models.RoleClinicOwner))

	// Rotate token once
	_, _, _, err := svc.RotateRefreshToken(context.Background(), rawRefresh1)
	if err != nil {
		t.Fatalf("First rotation failed: %v", err)
	}

	// Attacker tries to reuse rawRefresh1 which is now revoked!
	_, _, _, err = svc.RotateRefreshToken(context.Background(), rawRefresh1)
	if err != ErrTokenReused {
		t.Errorf("Expected ErrTokenReused, got: %v", err)
	}

	if !sessionRepo.revokedUser[user.ID] {
		t.Errorf("Expected all user sessions to be revoked on token reuse attempt")
	}
}
