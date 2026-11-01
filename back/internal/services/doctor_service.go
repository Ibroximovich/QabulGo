package services

import (
	"context"
	"errors"

	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/models"
	"qabulgo-backend/internal/repositories"
	"qabulgo-backend/internal/utils"

	"github.com/google/uuid"
)

var (
	ErrDoctorFirstNameRequired = errors.New("shifokor ismi kiritilishi shart")
	ErrDoctorLastNameRequired  = errors.New("shifokor familiyasi kiritilishi shart")
	ErrInvalidConsultationPrice = errors.New("konsultatsiya narxi manfiy bo‘lishi mumkin emas")
	ErrInvalidAppointmentDuration = errors.New("qabul davomiyligi 5 dan 240 daqiqagacha bo‘lishi kerak")
	ErrInvalidExperienceYears  = errors.New("ish tajribasi yillari 0 dan 70 gacha bo‘lishi kerak")
)

type DoctorService interface {
	AddDoctor(ctx context.Context, clinicID, userID uuid.UUID, req *dto.CreateDoctorRequest) (*dto.ClinicDoctorItemDTO, error)
	GetClinicDoctor(ctx context.Context, clinicID, clinicDoctorID uuid.UUID) (*dto.ClinicDoctorItemDTO, error)
	ListClinicDoctors(ctx context.Context, clinicID uuid.UUID, filter *dto.DoctorFilter) ([]dto.ClinicDoctorItemDTO, int64, error)
	UpdateClinicDoctor(ctx context.Context, clinicID, userID, clinicDoctorID uuid.UUID, req *dto.UpdateDoctorRequest) (*dto.ClinicDoctorItemDTO, error)
	DeactivateClinicDoctor(ctx context.Context, clinicID, userID, clinicDoctorID uuid.UUID) error
}

type doctorService struct {
	doctorRepo    repositories.DoctorRepository
	clinicRepo    repositories.ClinicRepository
	specialtyRepo repositories.SpecialtyRepository
	auditRepo     repositories.AuditRepository
}

func NewDoctorService(
	doctorRepo repositories.DoctorRepository,
	clinicRepo repositories.ClinicRepository,
	specialtyRepo repositories.SpecialtyRepository,
	auditRepo repositories.AuditRepository,
) DoctorService {
	return &doctorService{
		doctorRepo:    doctorRepo,
		clinicRepo:    clinicRepo,
		specialtyRepo: specialtyRepo,
		auditRepo:     auditRepo,
	}
}

func (s *doctorService) AddDoctor(ctx context.Context, clinicID, userID uuid.UUID, req *dto.CreateDoctorRequest) (*dto.ClinicDoctorItemDTO, error) {
	// 1. Huquq tekshiruvi (IDOR)
	isAdmin, err := s.clinicRepo.IsUserClinicAdmin(ctx, clinicID, userID)
	if err != nil || !isAdmin {
		return nil, repositories.ErrClinicAccessDenied
	}

	clinic, err := s.clinicRepo.GetByID(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	if clinic.Status == models.ClinicStatusSuspended {
		return nil, ErrClinicSuspended
	}

	// 2. Validatsiya
	firstName := utils.SanitizeString(req.FirstName)
	if len(firstName) < 2 || len(firstName) > 100 {
		return nil, ErrDoctorFirstNameRequired
	}
	lastName := utils.SanitizeString(req.LastName)
	if len(lastName) < 2 || len(lastName) > 100 {
		return nil, ErrDoctorLastNameRequired
	}

	if req.ConsultationPrice < 0 {
		return nil, ErrInvalidConsultationPrice
	}
	if req.AppointmentDurationMinutes < 5 || req.AppointmentDurationMinutes > 240 {
		return nil, ErrInvalidAppointmentDuration
	}
	if req.ExperienceYears < 0 || req.ExperienceYears > 70 {
		return nil, ErrInvalidExperienceYears
	}

	// 3. Mutaxassislik mavjudligini tekshirish
	specialty, err := s.specialtyRepo.GetByID(ctx, req.SpecialtyID)
	if err != nil {
		return nil, repositories.ErrSpecialtyNotFound
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	doc := &models.Doctor{
		ID:              uuid.New(),
		FirstName:       firstName,
		LastName:        lastName,
		Bio:             utils.SanitizeString(req.Bio),
		ExperienceYears: req.ExperienceYears,
		PhotoURL:        req.PhotoURL,
	}

	clinicDoctor, err := s.doctorRepo.CreateDoctorWithClinic(
		ctx,
		clinicID,
		doc,
		req.SpecialtyID,
		req.ConsultationPrice,
		req.AppointmentDurationMinutes,
		isActive,
	)
	if err != nil {
		return nil, err
	}

	// Audit log
	_ = s.auditRepo.CreateAuditLog(ctx, &models.AuditLog{
		UserID:     &userID,
		Action:     "DOCTOR_CREATED",
		EntityType: "doctor",
		EntityID:   &doc.ID,
		Details:    `{"clinic_id": "` + clinicID.String() + `"}`,
	})

	return s.toClinicDoctorDTO(clinicDoctor, specialty), nil
}

func (s *doctorService) GetClinicDoctor(ctx context.Context, clinicID, clinicDoctorID uuid.UUID) (*dto.ClinicDoctorItemDTO, error) {
	cd, err := s.doctorRepo.GetClinicDoctor(ctx, clinicID, clinicDoctorID)
	if err != nil {
		return nil, err
	}
	return s.toClinicDoctorDTO(cd, cd.Specialty), nil
}

func (s *doctorService) ListClinicDoctors(ctx context.Context, clinicID uuid.UUID, filter *dto.DoctorFilter) ([]dto.ClinicDoctorItemDTO, int64, error) {
	items, total, err := s.doctorRepo.ListClinicDoctors(ctx, clinicID, filter)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]dto.ClinicDoctorItemDTO, 0, len(items))
	for _, it := range items {
		dtos = append(dtos, *s.toClinicDoctorDTO(&it, it.Specialty))
	}
	return dtos, total, nil
}

func (s *doctorService) UpdateClinicDoctor(ctx context.Context, clinicID, userID, clinicDoctorID uuid.UUID, req *dto.UpdateDoctorRequest) (*dto.ClinicDoctorItemDTO, error) {
	// 1. Huquq tekshiruvi (IDOR)
	isAdmin, err := s.clinicRepo.IsUserClinicAdmin(ctx, clinicID, userID)
	if err != nil || !isAdmin {
		return nil, repositories.ErrClinicAccessDenied
	}

	cd, err := s.doctorRepo.GetClinicDoctor(ctx, clinicID, clinicDoctorID)
	if err != nil {
		return nil, err
	}

	doc := cd.Doctor

	// 2. Maydonlarni yangilash
	if req.FirstName != nil {
		name := utils.SanitizeString(*req.FirstName)
		if len(name) >= 2 {
			doc.FirstName = name
		}
	}
	if req.LastName != nil {
		name := utils.SanitizeString(*req.LastName)
		if len(name) >= 2 {
			doc.LastName = name
		}
	}
	if req.Bio != nil {
		doc.Bio = utils.SanitizeString(*req.Bio)
	}
	if req.ExperienceYears != nil {
		if *req.ExperienceYears >= 0 && *req.ExperienceYears <= 70 {
			doc.ExperienceYears = *req.ExperienceYears
		}
	}
	if req.PhotoURL != nil {
		doc.PhotoURL = *req.PhotoURL
	}

	if req.SpecialtyID != nil {
		specialty, err := s.specialtyRepo.GetByID(ctx, *req.SpecialtyID)
		if err != nil {
			return nil, repositories.ErrSpecialtyNotFound
		}
		cd.SpecialtyID = specialty.ID
		cd.Specialty = specialty
	}

	if req.ConsultationPrice != nil {
		if *req.ConsultationPrice >= 0 {
			cd.ConsultationPrice = *req.ConsultationPrice
		}
	}

	if req.AppointmentDurationMinutes != nil {
		if *req.AppointmentDurationMinutes >= 5 && *req.AppointmentDurationMinutes <= 240 {
			cd.AppointmentDurationMinutes = *req.AppointmentDurationMinutes
		}
	}

	if req.IsActive != nil {
		cd.IsActive = *req.IsActive
	}

	if err := s.doctorRepo.UpdateClinicDoctor(ctx, cd, doc); err != nil {
		return nil, err
	}

	return s.toClinicDoctorDTO(cd, cd.Specialty), nil
}

func (s *doctorService) DeactivateClinicDoctor(ctx context.Context, clinicID, userID, clinicDoctorID uuid.UUID) error {
	isAdmin, err := s.clinicRepo.IsUserClinicAdmin(ctx, clinicID, userID)
	if err != nil || !isAdmin {
		return repositories.ErrClinicAccessDenied
	}

	return s.doctorRepo.SoftDeleteClinicDoctor(ctx, clinicID, clinicDoctorID)
}

func (s *doctorService) toClinicDoctorDTO(cd *models.ClinicDoctor, spec *models.Specialty) *dto.ClinicDoctorItemDTO {
	var specDTO *dto.SpecialtyDTO
	if spec != nil {
		specDTO = &dto.SpecialtyDTO{
			ID:      spec.ID,
			Name:    spec.Name,
			Slug:    spec.Slug,
			IconURL: spec.IconURL,
		}
	}

	doc := cd.Doctor
	firstName := ""
	lastName := ""
	bio := ""
	exp := 0
	photo := ""
	if doc != nil {
		firstName = doc.FirstName
		lastName = doc.LastName
		bio = doc.Bio
		exp = doc.ExperienceYears
		photo = doc.PhotoURL
	}

	return &dto.ClinicDoctorItemDTO{
		ID:                         cd.ID,
		DoctorID:                   cd.DoctorID,
		ClinicID:                   cd.ClinicID,
		FirstName:                  firstName,
		LastName:                   lastName,
		Bio:                        bio,
		ExperienceYears:            exp,
		PhotoURL:                   photo,
		Specialty:                  specDTO,
		ConsultationPrice:          cd.ConsultationPrice,
		AppointmentDurationMinutes: cd.AppointmentDurationMinutes,
		IsActive:                   cd.IsActive,
		CreatedAt:                  cd.CreatedAt,
		UpdatedAt:                  cd.UpdatedAt,
	}
}

