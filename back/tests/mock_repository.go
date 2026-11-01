package tests

import (
	"context"
	"errors"
	"math"
	"sort"
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

func (m *MockSpecialtyRepository) Create(ctx context.Context, s *models.Specialty) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	sCopy := *s
	m.specialties[s.ID] = &sCopy
	return nil
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

// MockPublicSearchRepository
type MockPublicSearchRepository struct {
	clinicRepo    *MockClinicRepository
	doctorRepo    *MockDoctorRepository
	specialtyRepo *MockSpecialtyRepository
}

func NewMockPublicSearchRepository(clinicRepo *MockClinicRepository, doctorRepo *MockDoctorRepository, specialtyRepo *MockSpecialtyRepository) *MockPublicSearchRepository {
	return &MockPublicSearchRepository{
		clinicRepo:    clinicRepo,
		doctorRepo:    doctorRepo,
		specialtyRepo: specialtyRepo,
	}
}

func (m *MockPublicSearchRepository) SearchClinics(ctx context.Context, filter *dto.PublicClinicFilter) ([]models.Clinic, int64, error) {
	m.clinicRepo.mu.Lock()
	defer m.clinicRepo.mu.Unlock()

	var matched []models.Clinic
	for _, c := range m.clinicRepo.clinics {
		if c.Status != models.ClinicStatusApproved {
			continue
		}
		if filter != nil {
			if filter.Search != "" {
				term := strings.ToLower(strings.TrimSpace(filter.Search))
				if !strings.Contains(strings.ToLower(c.Name), term) &&
					!strings.Contains(strings.ToLower(c.Description), term) &&
					!strings.Contains(strings.ToLower(c.Address), term) {
					continue
				}
			}
			if filter.Region != "" && !strings.EqualFold(c.Region, strings.TrimSpace(filter.Region)) {
				continue
			}
			if filter.District != "" && !strings.EqualFold(c.District, strings.TrimSpace(filter.District)) {
				continue
			}
			if filter.SpecialtyID != nil && strings.TrimSpace(*filter.SpecialtyID) != "" {
				specUUID, _ := uuid.Parse(strings.TrimSpace(*filter.SpecialtyID))
				hasSpec := false
				m.doctorRepo.mu.Lock()
				for _, cd := range m.doctorRepo.clinicDoctors {
					if cd.ClinicID == c.ID && cd.SpecialtyID == specUUID && cd.IsActive && !cd.DeletedAt.Valid {
						hasSpec = true
						break
					}
				}
				m.doctorRepo.mu.Unlock()
				if !hasSpec {
					continue
				}
			}
		}

		cCopy := *c
		cCopy.WorkingHours = m.clinicRepo.workingHours[c.ID]
		matched = append(matched, cCopy)
	}

	total := int64(len(matched))

	// Stable sort by name
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].Name < matched[j].Name
	})

	limit := 12
	page := 1
	if filter != nil {
		if filter.Limit > 0 && filter.Limit <= 50 {
			limit = filter.Limit
		} else if filter.Limit > 50 {
			limit = 50
		}
		if filter.Page > 0 {
			page = filter.Page
		}
	}
	offset := (page - 1) * limit
	if offset >= len(matched) {
		return []models.Clinic{}, total, nil
	}
	end := offset + limit
	if end > len(matched) {
		end = len(matched)
	}

	return matched[offset:end], total, nil
}

func (m *MockPublicSearchRepository) GetApprovedClinicByID(ctx context.Context, id uuid.UUID) (*models.Clinic, error) {
	m.clinicRepo.mu.Lock()
	defer m.clinicRepo.mu.Unlock()

	c, ok := m.clinicRepo.clinics[id]
	if !ok || c.Status != models.ClinicStatusApproved {
		return nil, repositories.ErrClinicNotFound
	}
	cCopy := *c
	cCopy.WorkingHours = m.clinicRepo.workingHours[c.ID]
	return &cCopy, nil
}

func (m *MockPublicSearchRepository) GetClinicSpecialties(ctx context.Context, clinicID uuid.UUID) ([]models.Specialty, error) {
	m.doctorRepo.mu.Lock()
	defer m.doctorRepo.mu.Unlock()
	m.specialtyRepo.mu.Lock()
	defer m.specialtyRepo.mu.Unlock()

	specMap := make(map[uuid.UUID]models.Specialty)
	for _, cd := range m.doctorRepo.clinicDoctors {
		if cd.ClinicID == clinicID && cd.IsActive && !cd.DeletedAt.Valid {
			if sp, ok := m.specialtyRepo.specialties[cd.SpecialtyID]; ok {
				specMap[sp.ID] = *sp
			}
		}
	}

	var list []models.Specialty
	for _, sp := range specMap {
		list = append(list, sp)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})
	return list, nil
}

func (m *MockPublicSearchRepository) GetClinicActiveDoctors(ctx context.Context, clinicID uuid.UUID) ([]models.ClinicDoctor, error) {
	m.doctorRepo.mu.Lock()
	defer m.doctorRepo.mu.Unlock()
	m.specialtyRepo.mu.Lock()
	defer m.specialtyRepo.mu.Unlock()

	var list []models.ClinicDoctor
	for _, cd := range m.doctorRepo.clinicDoctors {
		if cd.ClinicID == clinicID && cd.IsActive && !cd.DeletedAt.Valid {
			cdCopy := *cd
			if doc, ok := m.doctorRepo.doctors[cd.DoctorID]; ok {
				docCopy := *doc
				cdCopy.Doctor = &docCopy
			}
			if sp, ok := m.specialtyRepo.specialties[cd.SpecialtyID]; ok {
				spCopy := *sp
				cdCopy.Specialty = &spCopy
			}
			list = append(list, cdCopy)
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].ConsultationPrice < list[j].ConsultationPrice
	})
	return list, nil
}

func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0
	rLat1 := lat1 * math.Pi / 180.0
	rLat2 := lat2 * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(rLat1)*math.Cos(rLat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func (m *MockPublicSearchRepository) SearchNearbyClinics(ctx context.Context, filter *dto.NearbyClinicFilter) ([]repositories.NearbyClinicRaw, int64, error) {
	m.clinicRepo.mu.Lock()
	defer m.clinicRepo.mu.Unlock()

	radius := filter.RadiusKM
	if radius <= 0 {
		radius = 10
	} else if radius > 100 {
		radius = 100
	}

	var matched []repositories.NearbyClinicRaw
	for _, c := range m.clinicRepo.clinics {
		if c.Status != models.ClinicStatusApproved || c.Latitude == nil || c.Longitude == nil {
			continue
		}

		dist := haversineKm(filter.Latitude, filter.Longitude, *c.Latitude, *c.Longitude)
		if dist > radius {
			continue
		}

		if filter.SpecialtyID != nil && strings.TrimSpace(*filter.SpecialtyID) != "" {
			specUUID, _ := uuid.Parse(strings.TrimSpace(*filter.SpecialtyID))
			hasSpec := false
			m.doctorRepo.mu.Lock()
			for _, cd := range m.doctorRepo.clinicDoctors {
				if cd.ClinicID == c.ID && cd.SpecialtyID == specUUID && cd.IsActive && !cd.DeletedAt.Valid {
					hasSpec = true
					break
				}
			}
			m.doctorRepo.mu.Unlock()
			if !hasSpec {
				continue
			}
		}

		matched = append(matched, repositories.NearbyClinicRaw{
			Clinic:     *c,
			DistanceKM: dist,
		})
	}

	total := int64(len(matched))

	// Sort by distance ASC, then Name ASC
	sort.Slice(matched, func(i, j int) bool {
		if matched[i].DistanceKM == matched[j].DistanceKM {
			return matched[i].Name < matched[j].Name
		}
		return matched[i].DistanceKM < matched[j].DistanceKM
	})

	limit := 12
	page := 1
	if filter.Limit > 0 && filter.Limit <= 50 {
		limit = filter.Limit
	} else if filter.Limit > 50 {
		limit = 50
	}
	if filter.Page > 0 {
		page = filter.Page
	}
	offset := (page - 1) * limit
	if offset >= len(matched) {
		return []repositories.NearbyClinicRaw{}, total, nil
	}
	end := offset + limit
	if end > len(matched) {
		end = len(matched)
	}

	return matched[offset:end], total, nil
}

func (m *MockPublicSearchRepository) SearchDoctors(ctx context.Context, filter *dto.PublicDoctorFilter) ([]models.ClinicDoctor, int64, error) {
	m.doctorRepo.mu.Lock()
	defer m.doctorRepo.mu.Unlock()
	m.clinicRepo.mu.Lock()
	defer m.clinicRepo.mu.Unlock()
	m.specialtyRepo.mu.Lock()
	defer m.specialtyRepo.mu.Unlock()

	var matched []models.ClinicDoctor
	for _, cd := range m.doctorRepo.clinicDoctors {
		if !cd.IsActive || cd.DeletedAt.Valid {
			continue
		}
		c, ok := m.clinicRepo.clinics[cd.ClinicID]
		if !ok || c.Status != models.ClinicStatusApproved {
			continue
		}
		doc, ok := m.doctorRepo.doctors[cd.DoctorID]
		if !ok {
			continue
		}

		if filter != nil {
			if filter.Search != "" {
				term := strings.ToLower(strings.TrimSpace(filter.Search))
				fullName := strings.ToLower(doc.FirstName + " " + doc.LastName)
				revName := strings.ToLower(doc.LastName + " " + doc.FirstName)
				if !strings.Contains(strings.ToLower(doc.FirstName), term) &&
					!strings.Contains(strings.ToLower(doc.LastName), term) &&
					!strings.Contains(fullName, term) &&
					!strings.Contains(revName, term) &&
					!strings.Contains(strings.ToLower(doc.Bio), term) {
					continue
				}
			}
			if filter.Region != "" && !strings.EqualFold(c.Region, strings.TrimSpace(filter.Region)) {
				continue
			}
			if filter.District != "" && !strings.EqualFold(c.District, strings.TrimSpace(filter.District)) {
				continue
			}
			if filter.SpecialtyID != nil && strings.TrimSpace(*filter.SpecialtyID) != "" {
				specUUID, _ := uuid.Parse(strings.TrimSpace(*filter.SpecialtyID))
				if cd.SpecialtyID != specUUID {
					continue
				}
			}
			if filter.ClinicID != nil && strings.TrimSpace(*filter.ClinicID) != "" {
				clinicUUID, _ := uuid.Parse(strings.TrimSpace(*filter.ClinicID))
				if cd.ClinicID != clinicUUID {
					continue
				}
			}
			if filter.MinPrice != nil && cd.ConsultationPrice < *filter.MinPrice {
				continue
			}
			if filter.MaxPrice != nil && cd.ConsultationPrice > *filter.MaxPrice {
				continue
			}
		}

		cdCopy := *cd
		docCopy := *doc
		cCopy := *c
		cdCopy.Doctor = &docCopy
		cdCopy.Clinic = &cCopy
		if sp, spOk := m.specialtyRepo.specialties[cd.SpecialtyID]; spOk {
			spCopy := *sp
			cdCopy.Specialty = &spCopy
		}
		matched = append(matched, cdCopy)
	}

	total := int64(len(matched))

	sortOrder := "name_asc"
	if filter != nil && filter.Sort != "" {
		sortOrder = filter.Sort
	}

	switch sortOrder {
	case "price_asc":
		sort.Slice(matched, func(i, j int) bool {
			return matched[i].ConsultationPrice < matched[j].ConsultationPrice
		})
	case "price_desc":
		sort.Slice(matched, func(i, j int) bool {
			return matched[i].ConsultationPrice > matched[j].ConsultationPrice
		})
	case "experience_desc":
		sort.Slice(matched, func(i, j int) bool {
			return matched[i].Doctor.ExperienceYears > matched[j].Doctor.ExperienceYears
		})
	default:
		sort.Slice(matched, func(i, j int) bool {
			if matched[i].Doctor.LastName == matched[j].Doctor.LastName {
				return matched[i].Doctor.FirstName < matched[j].Doctor.FirstName
			}
			return matched[i].Doctor.LastName < matched[j].Doctor.LastName
		})
	}

	limit := 12
	page := 1
	if filter != nil {
		if filter.Limit > 0 && filter.Limit <= 50 {
			limit = filter.Limit
		} else if filter.Limit > 50 {
			limit = 50
		}
		if filter.Page > 0 {
			page = filter.Page
		}
	}
	offset := (page - 1) * limit
	if offset >= len(matched) {
		return []models.ClinicDoctor{}, total, nil
	}
	end := offset + limit
	if end > len(matched) {
		end = len(matched)
	}

	return matched[offset:end], total, nil
}

func (m *MockPublicSearchRepository) GetDoctorByIDOrClinicDoctorID(ctx context.Context, targetID uuid.UUID) (*models.Doctor, []models.ClinicDoctor, error) {
	m.doctorRepo.mu.Lock()
	defer m.doctorRepo.mu.Unlock()
	m.clinicRepo.mu.Lock()
	defer m.clinicRepo.mu.Unlock()
	m.specialtyRepo.mu.Lock()
	defer m.specialtyRepo.mu.Unlock()

	var foundDoctor *models.Doctor
	if doc, ok := m.doctorRepo.doctors[targetID]; ok {
		foundDoctor = doc
	} else {
		for _, cd := range m.doctorRepo.clinicDoctors {
			if cd.ID == targetID {
				if d, dOk := m.doctorRepo.doctors[cd.DoctorID]; dOk {
					foundDoctor = d
					break
				}
			}
		}
	}

	if foundDoctor == nil {
		return nil, nil, repositories.ErrDoctorNotFound
	}

	var workplaces []models.ClinicDoctor
	for _, cd := range m.doctorRepo.clinicDoctors {
		if cd.DoctorID == foundDoctor.ID && cd.IsActive && !cd.DeletedAt.Valid {
			c, cOk := m.clinicRepo.clinics[cd.ClinicID]
			if cOk && c.Status == models.ClinicStatusApproved {
				cdCopy := *cd
				cCopy := *c
				cdCopy.Clinic = &cCopy
				if sp, spOk := m.specialtyRepo.specialties[cd.SpecialtyID]; spOk {
					spCopy := *sp
					cdCopy.Specialty = &spCopy
				}
				workplaces = append(workplaces, cdCopy)
			}
		}
	}

	if len(workplaces) == 0 {
		return nil, nil, repositories.ErrDoctorNotFound
	}

	docCopy := *foundDoctor
	return &docCopy, workplaces, nil
}

func (m *MockPublicSearchRepository) GetRegionsCatalog(ctx context.Context) ([]dto.RegionItemDTO, error) {
	m.clinicRepo.mu.Lock()
	defer m.clinicRepo.mu.Unlock()

	regMap := make(map[string]map[string]bool)
	for _, c := range m.clinicRepo.clinics {
		if c.Status == models.ClinicStatusApproved && c.Region != "" {
			if _, ok := regMap[c.Region]; !ok {
				regMap[c.Region] = make(map[string]bool)
			}
			if c.District != "" {
				regMap[c.Region][c.District] = true
			}
		}
	}

	var res []dto.RegionItemDTO
	for reg, distMap := range regMap {
		var dists []string
		for dist := range distMap {
			dists = append(dists, dist)
		}
		sort.Strings(dists)
		res = append(res, dto.RegionItemDTO{
			Region:    reg,
			Districts: dists,
		})
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Region < res[j].Region
	})
	return res, nil
}

// MockScheduleRepository
type MockScheduleRepository struct {
	mu            sync.RWMutex
	clinicRepo    *MockClinicRepository
	doctorRepo    *MockDoctorRepository
	weeklySlots   map[uuid.UUID][]models.DoctorWeeklySlot // keyed by clinicDoctorID
	dateOverrides map[uuid.UUID]models.DoctorDateSlotOverride // keyed by overrideID
}

func NewMockScheduleRepository(clinicRepo *MockClinicRepository, doctorRepo *MockDoctorRepository) *MockScheduleRepository {
	return &MockScheduleRepository{
		clinicRepo:    clinicRepo,
		doctorRepo:    doctorRepo,
		weeklySlots:   make(map[uuid.UUID][]models.DoctorWeeklySlot),
		dateOverrides: make(map[uuid.UUID]models.DoctorDateSlotOverride),
	}
}

func (m *MockScheduleRepository) FindClinicDoctor(ctx context.Context, clinicID, doctorIDOrCDID uuid.UUID) (*models.ClinicDoctor, error) {
	m.doctorRepo.mu.RLock()
	defer m.doctorRepo.mu.RUnlock()

	// 1. By ID
	if cd, ok := m.doctorRepo.clinicDoctors[doctorIDOrCDID]; ok {
		if cd.ClinicID == clinicID && !cd.DeletedAt.Valid {
			cdCopy := *cd
			if doc, docOk := m.doctorRepo.doctors[cd.DoctorID]; docOk {
				docCopy := *doc
				cdCopy.Doctor = &docCopy
			}
			return &cdCopy, nil
		}
	}

	// 2. By DoctorID
	for _, cd := range m.doctorRepo.clinicDoctors {
		if cd.ClinicID == clinicID && cd.DoctorID == doctorIDOrCDID && !cd.DeletedAt.Valid {
			cdCopy := *cd
			if doc, docOk := m.doctorRepo.doctors[cd.DoctorID]; docOk {
				docCopy := *doc
				cdCopy.Doctor = &docCopy
			}
			return &cdCopy, nil
		}
	}

	return nil, repositories.ErrClinicDoctorNotFound
}

func (m *MockScheduleRepository) GetWeeklySlots(ctx context.Context, clinicDoctorID uuid.UUID) ([]models.DoctorWeeklySlot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	slots := m.weeklySlots[clinicDoctorID]
	var res []models.DoctorWeeklySlot
	for _, s := range slots {
		res = append(res, s)
	}
	sort.Slice(res, func(i, j int) bool {
		if res[i].Weekday != res[j].Weekday {
			return res[i].Weekday < res[j].Weekday
		}
		return res[i].StartTime < res[j].StartTime
	})
	return res, nil
}

func (m *MockScheduleRepository) ReplaceWeeklySlots(ctx context.Context, clinicDoctorID uuid.UUID, slots []models.DoctorWeeklySlot) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range slots {
		if slots[i].ID == uuid.Nil {
			slots[i].ID = uuid.New()
		}
		slots[i].ClinicDoctorID = clinicDoctorID
		slots[i].CreatedAt = time.Now()
		slots[i].UpdatedAt = time.Now()
	}
	m.weeklySlots[clinicDoctorID] = slots
	return nil
}

func (m *MockScheduleRepository) CheckDoctorCrossClinicConflict(
	ctx context.Context,
	doctorID, currentClinicDoctorID uuid.UUID,
	weekday int,
	startTime, endTime string,
) (*models.DoctorWeeklySlot, *models.Clinic, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.doctorRepo.mu.RLock()
	var otherCDIDs []uuid.UUID
	clinicByCDID := make(map[uuid.UUID]*models.Clinic)
	for cdID, cd := range m.doctorRepo.clinicDoctors {
		if cd.DoctorID == doctorID && cd.ID != currentClinicDoctorID && !cd.DeletedAt.Valid {
			otherCDIDs = append(otherCDIDs, cdID)
			m.clinicRepo.mu.RLock()
			if c, ok := m.clinicRepo.clinics[cd.ClinicID]; ok {
				cCopy := *c
				clinicByCDID[cdID] = &cCopy
			}
			m.clinicRepo.mu.RUnlock()
		}
	}
	m.doctorRepo.mu.RUnlock()

	for _, cdID := range otherCDIDs {
		for _, slot := range m.weeklySlots[cdID] {
			if slot.Weekday == weekday && slot.Status == models.SlotStatusActive {
				// overlap: slot.StartTime < endTime && slot.EndTime > startTime
				if slot.StartTime < endTime && slot.EndTime > startTime {
					slotCopy := slot
					return &slotCopy, clinicByCDID[cdID], repositories.ErrCrossClinicConflict
				}
			}
		}
	}

	return nil, nil, nil
}

func (m *MockScheduleRepository) GetDateOverrides(ctx context.Context, clinicDoctorID uuid.UUID, date string, startDate, endDate string) ([]models.DoctorDateSlotOverride, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var res []models.DoctorDateSlotOverride
	for _, ov := range m.dateOverrides {
		if ov.ClinicDoctorID == clinicDoctorID {
			if date != "" && ov.Date != date {
				continue
			}
			if startDate != "" && ov.Date < startDate {
				continue
			}
			if endDate != "" && ov.Date > endDate {
				continue
			}
			res = append(res, ov)
		}
	}

	sort.Slice(res, func(i, j int) bool {
		if res[i].Date != res[j].Date {
			return res[i].Date < res[j].Date
		}
		stI := ""
		if res[i].StartTime != nil {
			stI = *res[i].StartTime
		}
		stJ := ""
		if res[j].StartTime != nil {
			stJ = *res[j].StartTime
		}
		return stI < stJ
	})

	return res, nil
}

func (m *MockScheduleRepository) GetDateOverrideByID(ctx context.Context, overrideID uuid.UUID) (*models.DoctorDateSlotOverride, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ov, ok := m.dateOverrides[overrideID]
	if !ok {
		return nil, repositories.ErrDateOverrideNotFound
	}
	return &ov, nil
}

func (m *MockScheduleRepository) CreateDateOverride(ctx context.Context, override *models.DoctorDateSlotOverride) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if override.ID == uuid.Nil {
		override.ID = uuid.New()
	}
	override.CreatedAt = time.Now()
	override.UpdatedAt = time.Now()
	m.dateOverrides[override.ID] = *override
	return nil
}

func (m *MockScheduleRepository) UpdateDateOverride(ctx context.Context, override *models.DoctorDateSlotOverride) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.dateOverrides[override.ID]; !ok {
		return repositories.ErrDateOverrideNotFound
	}
	override.UpdatedAt = time.Now()
	m.dateOverrides[override.ID] = *override
	return nil
}

func (m *MockScheduleRepository) DeleteDateOverride(ctx context.Context, overrideID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.dateOverrides[overrideID]; !ok {
		return repositories.ErrDateOverrideNotFound
	}
	delete(m.dateOverrides, overrideID)
	return nil
}

func (m *MockScheduleRepository) GetAvailabilityData(
	ctx context.Context,
	clinicID, doctorIDOrCDID uuid.UUID,
	dateStr string,
	weekday int,
) (*models.Clinic, *models.ClinicDoctor, []models.DoctorWeeklySlot, []models.DoctorDateSlotOverride, error) {
	m.clinicRepo.mu.RLock()
	clinic, cExists := m.clinicRepo.clinics[clinicID]
	var clinicCopy models.Clinic
	if cExists {
		clinicCopy = *clinic
	}
	m.clinicRepo.mu.RUnlock()
	if !cExists {
		return nil, nil, nil, nil, repositories.ErrClinicNotFound
	}

	cd, err := m.FindClinicDoctor(ctx, clinicID, doctorIDOrCDID)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	slots, err := m.GetWeeklySlots(ctx, cd.ID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	var weekdaySlots []models.DoctorWeeklySlot
	for _, s := range slots {
		if s.Weekday == weekday {
			weekdaySlots = append(weekdaySlots, s)
		}
	}

	overrides, err := m.GetDateOverrides(ctx, cd.ID, dateStr, "", "")
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return &clinicCopy, cd, weekdaySlots, overrides, nil
}

