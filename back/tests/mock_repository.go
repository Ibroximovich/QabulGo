package tests

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/models"
	"qabulgo-backend/internal/repositories"

	"github.com/google/uuid"
)

// MockUserRepository implements in-memory thread-safe UserRepository
type MockUserRepository struct {
	mu    sync.RWMutex
	users map[uuid.UUID]*models.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[uuid.UUID]*models.User),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, u := range m.users {
		if strings.EqualFold(u.Email, user.Email) {
			return repositories.ErrEmailAlreadyExist
		}
		if u.Phone == user.Phone {
			return repositories.ErrPhoneAlreadyExist
		}
	}

	userCopy := *user
	m.users[user.ID] = &userCopy
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[id]
	if !exists {
		return nil, repositories.ErrUserNotFound
	}
	userCopy := *user
	return &userCopy, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, u := range m.users {
		if strings.EqualFold(u.Email, email) {
			userCopy := *u
			return &userCopy, nil
		}
	}
	return nil, repositories.ErrUserNotFound
}

func (m *MockUserRepository) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, u := range m.users {
		if u.Phone == phone {
			userCopy := *u
			return &userCopy, nil
		}
	}
	return nil, repositories.ErrUserNotFound
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[user.ID]; !exists {
		return repositories.ErrUserNotFound
	}
	userCopy := *user
	m.users[user.ID] = &userCopy
	return nil
}

func (m *MockUserRepository) UpdateEmailVerified(ctx context.Context, id uuid.UUID, verifiedAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, exists := m.users[id]
	if !exists {
		return repositories.ErrUserNotFound
	}
	user.EmailVerifiedAt = &verifiedAt
	return nil
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, exists := m.users[id]
	if !exists {
		return repositories.ErrUserNotFound
	}
	user.PasswordHash = passwordHash
	user.UpdatedAt = time.Now().UTC()
	return nil
}

// MockRefreshSessionRepository implements in-memory RefreshSessionRepository
type MockRefreshSessionRepository struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]*models.RefreshSession
}

func NewMockRefreshSessionRepository() *MockRefreshSessionRepository {
	return &MockRefreshSessionRepository{
		sessions: make(map[uuid.UUID]*models.RefreshSession),
	}
}

func (m *MockRefreshSessionRepository) Create(ctx context.Context, session *models.RefreshSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	sessionCopy := *session
	m.sessions[session.ID] = &sessionCopy
	return nil
}

func (m *MockRefreshSessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshSession, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, s := range m.sessions {
		if s.TokenHash == tokenHash {
			sessionCopy := *s
			return &sessionCopy, nil
		}
	}
	return nil, repositories.ErrSessionNotFound
}

func (m *MockRefreshSessionRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, exists := m.sessions[id]
	if !exists {
		return repositories.ErrSessionNotFound
	}
	now := time.Now().UTC()
	s.RevokedAt = &now
	return nil
}

func (m *MockRefreshSessionRepository) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	for _, s := range m.sessions {
		if s.UserID == userID && s.RevokedAt == nil {
			s.RevokedAt = &now
		}
	}
	return nil
}

func (m *MockRefreshSessionRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, exists := m.sessions[id]
	if !exists {
		return repositories.ErrSessionNotFound
	}
	s.LastUsedAt = time.Now().UTC()
	return nil
}

// MockVerificationTokenRepository implements in-memory VerificationTokenRepository
type MockVerificationTokenRepository struct {
	mu     sync.RWMutex
	tokens map[uuid.UUID]*models.VerificationToken
}

func NewMockVerificationTokenRepository() *MockVerificationTokenRepository {
	return &MockVerificationTokenRepository{
		tokens: make(map[uuid.UUID]*models.VerificationToken),
	}
}

func (m *MockVerificationTokenRepository) Create(ctx context.Context, token *models.VerificationToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	tokenCopy := *token
	m.tokens[token.ID] = &tokenCopy
	return nil
}

func (m *MockVerificationTokenRepository) GetLatestActive(ctx context.Context, userID uuid.UUID, purpose models.TokenPurpose) (*models.VerificationToken, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var latest *models.VerificationToken
	now := time.Now().UTC()

	for _, t := range m.tokens {
		if t.UserID == userID && t.Purpose == purpose && t.ConsumedAt == nil && t.ExpiresAt.After(now) {
			if latest == nil || t.CreatedAt.After(latest.CreatedAt) {
				tokenCopy := *t
				latest = &tokenCopy
			}
		}
	}

	if latest == nil {
		return nil, repositories.ErrTokenNotFound
	}
	return latest, nil
}

func (m *MockVerificationTokenRepository) IncrementAttempts(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, exists := m.tokens[id]
	if !exists {
		return repositories.ErrTokenNotFound
	}
	t.Attempts++
	return nil
}

func (m *MockVerificationTokenRepository) MarkConsumed(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, exists := m.tokens[id]
	if !exists {
		return repositories.ErrTokenNotFound
	}
	now := time.Now().UTC()
	t.ConsumedAt = &now
	return nil
}

func (m *MockVerificationTokenRepository) RevokeAllActiveByUserID(ctx context.Context, userID uuid.UUID, purpose models.TokenPurpose) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	for _, t := range m.tokens {
		if t.UserID == userID && t.Purpose == purpose && t.ConsumedAt == nil {
			t.ConsumedAt = &now
		}
	}
	return nil
}

// MockEmailService for tests
type MockEmailService struct {
	SentVerificationEmails map[string]string // email -> code
	SentResetEmails        map[string]string // email -> code
	mu                     sync.Mutex
}

func NewMockEmailService() *MockEmailService {
	return &MockEmailService{
		SentVerificationEmails: make(map[string]string),
		SentResetEmails:        make(map[string]string),
	}
}

func (m *MockEmailService) SendVerificationEmail(toEmail, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SentVerificationEmails[toEmail] = code
	return nil
}

func (m *MockEmailService) SendPasswordResetEmail(toEmail, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SentResetEmails[toEmail] = code
	return nil
}

func (m *MockEmailService) GetVerificationCode(email string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	code, exists := m.SentVerificationEmails[email]
	if !exists {
		return "", errors.New("no code sent")
	}
	return code, nil
}

func (m *MockEmailService) GetResetCode(email string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	code, exists := m.SentResetEmails[email]
	if !exists {
		return "", errors.New("no code sent")
	}
	return code, nil
}


// MockClinicRepository
type MockClinicRepository struct {
	mu           sync.RWMutex
	clinics      map[uuid.UUID]*models.Clinic
	workingHours map[uuid.UUID][]models.ClinicWorkingHour
	members      map[uuid.UUID][]models.ClinicMember
}

func NewMockClinicRepository() *MockClinicRepository {
	return &MockClinicRepository{
		clinics:      make(map[uuid.UUID]*models.Clinic),
		workingHours: make(map[uuid.UUID][]models.ClinicWorkingHour),
		members:      make(map[uuid.UUID][]models.ClinicMember),
	}
}

func (m *MockClinicRepository) Create(ctx context.Context, clinic *models.Clinic, workingHours []models.ClinicWorkingHour) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cCopy := *clinic
	m.clinics[clinic.ID] = &cCopy
	m.workingHours[clinic.ID] = workingHours
	m.members[clinic.ID] = append(m.members[clinic.ID], models.ClinicMember{
		ID:       uuid.New(),
		ClinicID: clinic.ID,
		UserID:   clinic.OwnerID,
		Role:     models.MemberRoleOwner,
	})
	return nil
}

func (m *MockClinicRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Clinic, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	c, exists := m.clinics[id]
	if !exists {
		return nil, repositories.ErrClinicNotFound
	}
	cCopy := *c
	cCopy.WorkingHours = m.workingHours[id]
	return &cCopy, nil
}

func (m *MockClinicRepository) GetByOwnerID(ctx context.Context, ownerID uuid.UUID) (*models.Clinic, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, c := range m.clinics {
		if c.OwnerID == ownerID {
			cCopy := *c
			cCopy.WorkingHours = m.workingHours[c.ID]
			return &cCopy, nil
		}
	}
	return nil, repositories.ErrClinicNotFound
}

func (m *MockClinicRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Clinic, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for clinicID, memberList := range m.members {
		for _, mem := range memberList {
			if mem.UserID == userID {
				c := m.clinics[clinicID]
				cCopy := *c
				cCopy.WorkingHours = m.workingHours[clinicID]
				return &cCopy, nil
			}
		}
	}

	return m.GetByOwnerID(ctx, userID)
}

func (m *MockClinicRepository) Update(ctx context.Context, clinic *models.Clinic) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cCopy := *clinic
	m.clinics[clinic.ID] = &cCopy
	return nil
}

func (m *MockClinicRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.ClinicStatus, rejectionReason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, exists := m.clinics[id]
	if !exists {
		return repositories.ErrClinicNotFound
	}
	c.Status = status
	c.RejectionReason = rejectionReason
	return nil
}

func (m *MockClinicRepository) SetWorkingHours(ctx context.Context, clinicID uuid.UUID, workingHours []models.ClinicWorkingHour) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.workingHours[clinicID] = workingHours
	return nil
}

func (m *MockClinicRepository) GetWorkingHours(ctx context.Context, clinicID uuid.UUID) ([]models.ClinicWorkingHour, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.workingHours[clinicID], nil
}

func (m *MockClinicRepository) IsUserClinicAdmin(ctx context.Context, clinicID, userID uuid.UUID) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	c, exists := m.clinics[clinicID]
	if exists && c.OwnerID == userID {
		return true, nil
	}

	for _, mem := range m.members[clinicID] {
		if mem.UserID == userID {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockClinicRepository) ListForAdmin(ctx context.Context, filter *dto.AdminClinicFilter) ([]models.Clinic, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []models.Clinic
	for _, c := range m.clinics {
		if filter.Status != "" && string(c.Status) != filter.Status {
			continue
		}
		cCopy := *c
		result = append(result, cCopy)
	}
	return result, int64(len(result)), nil
}

// MockDoctorRepository
type MockDoctorRepository struct {
	mu            sync.RWMutex
	doctors       map[uuid.UUID]*models.Doctor
	clinicDoctors map[uuid.UUID]*models.ClinicDoctor
}

func NewMockDoctorRepository() *MockDoctorRepository {
	return &MockDoctorRepository{
		doctors:       make(map[uuid.UUID]*models.Doctor),
		clinicDoctors: make(map[uuid.UUID]*models.ClinicDoctor),
	}
}

func (m *MockDoctorRepository) CreateDoctorWithClinic(
	ctx context.Context,
	clinicID uuid.UUID,
	doc *models.Doctor,
	specialtyID uuid.UUID,
	price float64,
	duration int,
	isActive bool,
) (*models.ClinicDoctor, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	docCopy := *doc
	m.doctors[doc.ID] = &docCopy

	cd := &models.ClinicDoctor{
		ID:                         uuid.New(),
		ClinicID:                   clinicID,
		DoctorID:                   doc.ID,
		SpecialtyID:                specialtyID,
		ConsultationPrice:          price,
		AppointmentDurationMinutes: duration,
		IsActive:                   isActive,
		Doctor:                     &docCopy,
	}
	m.clinicDoctors[cd.ID] = cd
	return cd, nil
}

func (m *MockDoctorRepository) GetClinicDoctor(ctx context.Context, clinicID, clinicDoctorID uuid.UUID) (*models.ClinicDoctor, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cd, exists := m.clinicDoctors[clinicDoctorID]
	if !exists || cd.ClinicID != clinicID || cd.DeletedAt.Valid {
		return nil, repositories.ErrDoctorNotFound
	}
	return cd, nil
}

func (m *MockDoctorRepository) UpdateClinicDoctor(ctx context.Context, cd *models.ClinicDoctor, doc *models.Doctor) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if doc != nil {
		m.doctors[doc.ID] = doc
	}
	m.clinicDoctors[cd.ID] = cd
	return nil
}

func (m *MockDoctorRepository) SoftDeleteClinicDoctor(ctx context.Context, clinicID, clinicDoctorID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cd, exists := m.clinicDoctors[clinicDoctorID]
	if !exists || cd.ClinicID != clinicID || cd.DeletedAt.Valid {
		return repositories.ErrDoctorNotFound
	}
	cd.DeletedAt.Time = time.Now()
	cd.DeletedAt.Valid = true
	return nil
}

func (m *MockDoctorRepository) ListClinicDoctors(ctx context.Context, clinicID uuid.UUID, filter *dto.DoctorFilter) ([]models.ClinicDoctor, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []models.ClinicDoctor
	for _, cd := range m.clinicDoctors {
		if cd.ClinicID != clinicID || cd.DeletedAt.Valid {
			continue
		}
		if filter != nil && filter.SpecialtyID != nil && cd.SpecialtyID != *filter.SpecialtyID {
			continue
		}
		if filter != nil && filter.IsActive != nil && cd.IsActive != *filter.IsActive {
			continue
		}
		result = append(result, *cd)
	}
	return result, int64(len(result)), nil
}

func (m *MockDoctorRepository) CountDoctorsByClinic(ctx context.Context, clinicID uuid.UUID) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, cd := range m.clinicDoctors {
		if cd.ClinicID == clinicID && !cd.DeletedAt.Valid {
			count++
		}
	}
	return count, nil
}

// MockSpecialtyRepository
type MockSpecialtyRepository struct {
	mu          sync.RWMutex
	specialties map[uuid.UUID]*models.Specialty
}

func NewMockSpecialtyRepository() *MockSpecialtyRepository {
	m := &MockSpecialtyRepository{
		specialties: make(map[uuid.UUID]*models.Specialty),
	}
	_ = m.SeedIfEmpty(context.Background())
	return m
}

func (m *MockSpecialtyRepository) ListAll(ctx context.Context) ([]models.Specialty, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []models.Specialty
	for _, s := range m.specialties {
		list = append(list, *s)
	}
	return list, nil
}

func (m *MockSpecialtyRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Specialty, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	s, exists := m.specialties[id]
	if !exists {
		return nil, repositories.ErrSpecialtyNotFound
	}
	return s, nil
}

func (m *MockSpecialtyRepository) GetBySlug(ctx context.Context, slug string) (*models.Specialty, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, s := range m.specialties {
		if s.Slug == slug {
			return s, nil
		}
	}
	return nil, repositories.ErrSpecialtyNotFound
}

func (m *MockSpecialtyRepository) SeedIfEmpty(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.specialties) > 0 {
		return nil
	}

	seeds := []models.Specialty{
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000001"), Name: "Terapevt", Slug: "terapevt"},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000002"), Name: "Kardiolog", Slug: "kardiolog"},
		{ID: uuid.MustParse("00000000-0000-0000-0000-000000000003"), Name: "Stomatolog", Slug: "stomatolog"},
	}
	for _, s := range seeds {
		sCopy := s
		m.specialties[s.ID] = &sCopy
	}
	return nil
}

// MockAuditRepository
type MockAuditRepository struct {
	mu      sync.Mutex
	logs    []models.AuditLog
	reviews []models.ClinicVerificationReview
}

func NewMockAuditRepository() *MockAuditRepository {
	return &MockAuditRepository{}
}

func (m *MockAuditRepository) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = append(m.logs, *log)
	return nil
}

func (m *MockAuditRepository) CreateVerificationReview(ctx context.Context, review *models.ClinicVerificationReview) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reviews = append(m.reviews, *review)
	return nil
}

func (m *MockAuditRepository) GetReviewsByClinic(ctx context.Context, clinicID uuid.UUID) ([]models.ClinicVerificationReview, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var list []models.ClinicVerificationReview
	for _, r := range m.reviews {
		if r.ClinicID == clinicID {
			list = append(list, r)
		}
	}
	return list, nil
}
