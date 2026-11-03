package services

import (
	"context"
	"errors"
	"fmt"

	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/models"
	"qabulgo-backend/internal/repositories"
	"qabulgo-backend/internal/utils"

	"github.com/google/uuid"
)

var (
	ErrRejectionReasonRequired = errors.New("rad etish sababi ko‘rsatilishi shart")
	ErrSuspendReasonRequired   = errors.New("bloklash sababi ko‘rsatilishi shart")
)

type AdminService interface {
	GetDashboardSummary(ctx context.Context) (*dto.AdminDashboardSummaryDTO, error)
	GetDashboardAnalytics(ctx context.Context, filter *dto.AdminDashboardAnalyticsFilter) (*dto.AdminDashboardAnalyticsDTO, error)
	ListClinics(ctx context.Context, filter *dto.AdminClinicFilter) ([]dto.ClinicSummaryDTO, int64, error)
	GetClinicDetails(ctx context.Context, clinicID uuid.UUID) (*dto.ClinicDetailDTO, []models.ClinicVerificationReview, error)
	GetClinicDoctors(ctx context.Context, clinicID uuid.UUID) ([]dto.AdminClinicDoctorItemDTO, error)
	GetClinicAppointments(ctx context.Context, clinicID uuid.UUID, filter *dto.AdminAppointmentFilter) ([]dto.AdminAppointmentListItemDTO, int64, error)
	GetClinicAnalytics(ctx context.Context, clinicID uuid.UUID) (*dto.AdminClinicAnalyticsDTO, error)
	ApproveClinic(ctx context.Context, adminID, clinicID uuid.UUID, ipAddress, userAgent string) error
	RejectClinic(ctx context.Context, adminID, clinicID uuid.UUID, reason string, ipAddress, userAgent string) error
	SuspendClinic(ctx context.Context, adminID, clinicID uuid.UUID, reason string, ipAddress, userAgent string) error
	ListPatients(ctx context.Context, filter *dto.AdminPatientFilter) ([]dto.AdminPatientListItemDTO, int64, error)
	GetPatientDetails(ctx context.Context, patientID uuid.UUID) (*dto.AdminPatientDetailDTO, error)
	ListDoctors(ctx context.Context, filter *dto.AdminDoctorFilter) ([]dto.AdminDoctorListItemDTO, int64, error)
	GetDoctorDetails(ctx context.Context, doctorID uuid.UUID) (*dto.AdminDoctorDetailDTO, error)
	ListAppointments(ctx context.Context, filter *dto.AdminAppointmentFilter) ([]dto.AdminAppointmentListItemDTO, int64, error)
	GetAppointmentDetails(ctx context.Context, appointmentID uuid.UUID) (*dto.AdminAppointmentListItemDTO, error)
}

type adminService struct {
	clinicRepo repositories.ClinicRepository
	doctorRepo repositories.DoctorRepository
	auditRepo  repositories.AuditRepository
	adminRepo  repositories.AdminRepository
}

func NewAdminService(
	clinicRepo repositories.ClinicRepository,
	doctorRepo repositories.DoctorRepository,
	auditRepo repositories.AuditRepository,
	adminRepo repositories.AdminRepository,
) AdminService {
	return &adminService{
		clinicRepo: clinicRepo,
		doctorRepo: doctorRepo,
		auditRepo:  auditRepo,
		adminRepo:  adminRepo,
	}
}

func (s *adminService) GetDashboardSummary(ctx context.Context) (*dto.AdminDashboardSummaryDTO, error) {
	if s.adminRepo == nil {
		return nil, errors.New("admin repository mavjud emas")
	}
	return s.adminRepo.GetDashboardSummary(ctx)
}

func (s *adminService) GetDashboardAnalytics(ctx context.Context, filter *dto.AdminDashboardAnalyticsFilter) (*dto.AdminDashboardAnalyticsDTO, error) {
	if s.adminRepo == nil {
		return nil, errors.New("admin repository mavjud emas")
	}
	return s.adminRepo.GetDashboardAnalytics(ctx, filter)
}

func (s *adminService) GetClinicDoctors(ctx context.Context, clinicID uuid.UUID) ([]dto.AdminClinicDoctorItemDTO, error) {
	if s.adminRepo == nil {
		return nil, errors.New("admin repository mavjud emas")
	}
	return s.adminRepo.GetClinicDoctors(ctx, clinicID)
}

func (s *adminService) GetClinicAppointments(ctx context.Context, clinicID uuid.UUID, filter *dto.AdminAppointmentFilter) ([]dto.AdminAppointmentListItemDTO, int64, error) {
	if s.adminRepo == nil {
		return nil, 0, errors.New("admin repository mavjud emas")
	}
	return s.adminRepo.GetClinicAppointments(ctx, clinicID, filter)
}

func (s *adminService) GetClinicAnalytics(ctx context.Context, clinicID uuid.UUID) (*dto.AdminClinicAnalyticsDTO, error) {
	if s.adminRepo == nil {
		return nil, errors.New("admin repository mavjud emas")
	}
	return s.adminRepo.GetClinicAnalytics(ctx, clinicID)
}

func (s *adminService) ListPatients(ctx context.Context, filter *dto.AdminPatientFilter) ([]dto.AdminPatientListItemDTO, int64, error) {
	if s.adminRepo == nil {
		return nil, 0, errors.New("admin repository mavjud emas")
	}
	return s.adminRepo.ListPatients(ctx, filter)
}

func (s *adminService) GetPatientDetails(ctx context.Context, patientID uuid.UUID) (*dto.AdminPatientDetailDTO, error) {
	if s.adminRepo == nil {
		return nil, errors.New("admin repository mavjud emas")
	}
	return s.adminRepo.GetPatientDetails(ctx, patientID)
}

func (s *adminService) ListDoctors(ctx context.Context, filter *dto.AdminDoctorFilter) ([]dto.AdminDoctorListItemDTO, int64, error) {
	if s.adminRepo == nil {
		return nil, 0, errors.New("admin repository mavjud emas")
	}
	return s.adminRepo.ListDoctors(ctx, filter)
}

func (s *adminService) GetDoctorDetails(ctx context.Context, doctorID uuid.UUID) (*dto.AdminDoctorDetailDTO, error) {
	if s.adminRepo == nil {
		return nil, errors.New("admin repository mavjud emas")
	}
	return s.adminRepo.GetDoctorDetails(ctx, doctorID)
}

func (s *adminService) ListAppointments(ctx context.Context, filter *dto.AdminAppointmentFilter) ([]dto.AdminAppointmentListItemDTO, int64, error) {
	if s.adminRepo == nil {
		return nil, 0, errors.New("admin repository mavjud emas")
	}
	return s.adminRepo.ListAppointments(ctx, filter)
}

func (s *adminService) GetAppointmentDetails(ctx context.Context, appointmentID uuid.UUID) (*dto.AdminAppointmentListItemDTO, error) {
	if s.adminRepo == nil {
		return nil, errors.New("admin repository mavjud emas")
	}
	return s.adminRepo.GetAppointmentDetails(ctx, appointmentID)
}

func (s *adminService) ListClinics(ctx context.Context, filter *dto.AdminClinicFilter) ([]dto.ClinicSummaryDTO, int64, error) {
	clinics, total, err := s.clinicRepo.ListForAdmin(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	summaries := make([]dto.ClinicSummaryDTO, 0, len(clinics))
	for _, c := range clinics {
		summaries = append(summaries, dto.ClinicSummaryDTO{
			ID:            c.ID,
			Name:          c.Name,
			Description:   c.Description,
			Phone:         c.Phone,
			Email:         c.Email,
			Region:        c.Region,
			District:      c.District,
			Address:       c.Address,
			LocationURL:   c.LocationURL,
			Latitude:      c.Latitude,
			Longitude:     c.Longitude,
			LogoURL:       c.LogoURL,
			CoverImageURL: c.CoverImageURL,
			Status:        string(c.Status),
			CreatedAt:     c.CreatedAt,
		})
	}
	return summaries, total, nil
}

func (s *adminService) GetClinicDetails(ctx context.Context, clinicID uuid.UUID) (*dto.ClinicDetailDTO, []models.ClinicVerificationReview, error) {
	clinic, err := s.clinicRepo.GetByID(ctx, clinicID)
	if err != nil {
		return nil, nil, err
	}

	hours, _ := s.clinicRepo.GetWorkingHours(ctx, clinicID)
	whItems := make([]dto.WorkingHourItem, 0, len(hours))
	for _, h := range hours {
		whItems = append(whItems, dto.WorkingHourItem{
			Weekday:     h.Weekday,
			OpeningTime: h.OpeningTime,
			ClosingTime: h.ClosingTime,
			IsClosed:    h.IsClosed,
		})
	}

	doctorCount, _ := s.doctorRepo.CountDoctorsByClinic(ctx, clinicID)
	reviews, _ := s.auditRepo.GetReviewsByClinic(ctx, clinicID)

	detail := &dto.ClinicDetailDTO{
		ID:              clinic.ID,
		Name:            clinic.Name,
		Description:     clinic.Description,
		Phone:           clinic.Phone,
		Email:           clinic.Email,
		Region:          clinic.Region,
		District:        clinic.District,
		Address:         clinic.Address,
		LogoURL:         clinic.LogoURL,
		CoverImageURL:   clinic.CoverImageURL,
		Status:          string(clinic.Status),
		RejectionReason: clinic.RejectionReason,
		OwnerID:         clinic.OwnerID,
		WorkingHours:    whItems,
		DoctorsCount:    doctorCount,
		CreatedAt:       clinic.CreatedAt,
		UpdatedAt:       clinic.UpdatedAt,
	}

	return detail, reviews, nil
}

func (s *adminService) ApproveClinic(ctx context.Context, adminID, clinicID uuid.UUID, ipAddress, userAgent string) error {
	clinic, err := s.clinicRepo.GetByID(ctx, clinicID)
	if err != nil {
		return err
	}

	if clinic.Status == models.ClinicStatusApproved {
		return ErrClinicAlreadyApproved
	}

	if err := s.clinicRepo.UpdateStatus(ctx, clinicID, models.ClinicStatusApproved, ""); err != nil {
		return err
	}

	// Yozuvlarni saqlash
	_ = s.auditRepo.CreateVerificationReview(ctx, &models.ClinicVerificationReview{
		ClinicID:   clinicID,
		ReviewedBy: adminID,
		Decision:   "approved",
		Reason:     "Klinika barcha talablarga javob beradi va tasdiqlandi",
	})

	_ = s.auditRepo.CreateAuditLog(ctx, &models.AuditLog{
		UserID:     &adminID,
		Action:     "CLINIC_APPROVED",
		EntityType: "clinic",
		EntityID:   &clinicID,
		Details:    `{"previous_status": "` + string(clinic.Status) + `", "new_status": "approved"}`,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	})

	return nil
}

func (s *adminService) RejectClinic(ctx context.Context, adminID, clinicID uuid.UUID, reason string, ipAddress, userAgent string) error {
	cleanReason := utils.SanitizeString(reason)
	if cleanReason == "" {
		return ErrRejectionReasonRequired
	}

	clinic, err := s.clinicRepo.GetByID(ctx, clinicID)
	if err != nil {
		return err
	}

	if err := s.clinicRepo.UpdateStatus(ctx, clinicID, models.ClinicStatusRejected, cleanReason); err != nil {
		return err
	}

	_ = s.auditRepo.CreateVerificationReview(ctx, &models.ClinicVerificationReview{
		ClinicID:   clinicID,
		ReviewedBy: adminID,
		Decision:   "rejected",
		Reason:     cleanReason,
	})

	_ = s.auditRepo.CreateAuditLog(ctx, &models.AuditLog{
		UserID:     &adminID,
		Action:     "CLINIC_REJECTED",
		EntityType: "clinic",
		EntityID:   &clinicID,
		Details:    fmt.Sprintf(`{"reason": %q, "previous_status": %q}`, cleanReason, string(clinic.Status)),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	})

	return nil
}

func (s *adminService) SuspendClinic(ctx context.Context, adminID, clinicID uuid.UUID, reason string, ipAddress, userAgent string) error {
	cleanReason := utils.SanitizeString(reason)
	if cleanReason == "" {
		return ErrSuspendReasonRequired
	}

	clinic, err := s.clinicRepo.GetByID(ctx, clinicID)
	if err != nil {
		return err
	}

	if err := s.clinicRepo.UpdateStatus(ctx, clinicID, models.ClinicStatusSuspended, cleanReason); err != nil {
		return err
	}

	_ = s.auditRepo.CreateVerificationReview(ctx, &models.ClinicVerificationReview{
		ClinicID:   clinicID,
		ReviewedBy: adminID,
		Decision:   "suspended",
		Reason:     cleanReason,
	})

	_ = s.auditRepo.CreateAuditLog(ctx, &models.AuditLog{
		UserID:     &adminID,
		Action:     "CLINIC_SUSPENDED",
		EntityType: "clinic",
		EntityID:   &clinicID,
		Details:    fmt.Sprintf(`{"reason": %q, "previous_status": %q}`, cleanReason, string(clinic.Status)),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	})

	return nil
}

