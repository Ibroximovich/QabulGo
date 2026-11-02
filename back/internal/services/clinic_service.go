package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/models"
	"qabulgo-backend/internal/repositories"
	"qabulgo-backend/internal/utils"

	"github.com/google/uuid"
)

var (
	ErrClinicNameRequired        = errors.New("klinika nomi kiritilishi shart (kamida 2 ta belgi)")
	ErrClinicPhoneRequired       = errors.New("klinika telefon raqami kiritilishi shart")
	ErrClinicRegionRequired      = errors.New("viloyat/shahar kiritilishi shart")
	ErrClinicDistrictRequired    = errors.New("tuman/shahar kiritilishi shart")
	ErrClinicAddressRequired     = errors.New("manzil kiritilishi shart")
	ErrClinicSuspended           = errors.New("ushbu klinika bloklangan holatda, amallar bajarish taqiqlanadi")
	ErrClinicAlreadySubmitted    = errors.New("klinika allaqachon tasdiqlash uchun yuborilgan (pending holatida)")
	ErrClinicAlreadyApproved     = errors.New("klinika allaqachon tasdiqlangan")
	ErrClinicWorkingHourRequired = errors.New("tasdiqlashga yuborish uchun kamida 1 ta ish kuni soatlari ko‘rsatilishi shart")
	ErrInvalidStatusTransition   = errors.New("statusni bu holatga o‘tkazish taqiqlanadi")
)

type ClinicService interface {
	CreateClinic(ctx context.Context, ownerID uuid.UUID, req *dto.CreateClinicRequest) (*dto.ClinicDetailDTO, error)
	GetMyClinic(ctx context.Context, userID uuid.UUID) (*dto.ClinicDetailDTO, error)
	GetClinicByID(ctx context.Context, clinicID uuid.UUID, requestingUserID uuid.UUID, isSuperAdmin bool) (*dto.ClinicDetailDTO, error)
	UpdateClinic(ctx context.Context, clinicID, userID uuid.UUID, req *dto.UpdateClinicRequest) (*dto.ClinicDetailDTO, error)
	SubmitClinic(ctx context.Context, clinicID, userID uuid.UUID) (*dto.ClinicDetailDTO, error)
}

type clinicService struct {
	clinicRepo repositories.ClinicRepository
	doctorRepo repositories.DoctorRepository
	userRepo   repositories.UserRepository
	auditRepo  repositories.AuditRepository
	outboxRepo repositories.OutboxRepository
}

func NewClinicService(
	clinicRepo repositories.ClinicRepository,
	doctorRepo repositories.DoctorRepository,
	userRepo repositories.UserRepository,
	auditRepo repositories.AuditRepository,
	outboxRepo ...repositories.OutboxRepository,
) ClinicService {
	svc := &clinicService{
		clinicRepo: clinicRepo,
		doctorRepo: doctorRepo,
		userRepo:   userRepo,
		auditRepo:  auditRepo,
	}
	if len(outboxRepo) > 0 {
		svc.outboxRepo = outboxRepo[0]
	}
	return svc
}

func (s *clinicService) CreateClinic(ctx context.Context, ownerID uuid.UUID, req *dto.CreateClinicRequest) (*dto.ClinicDetailDTO, error) {
	// 1. Foydalanuvchini tekshirish
	user, err := s.userRepo.GetByID(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive() {
		return nil, ErrAccountBlocked
	}
	if !user.IsEmailVerified() {
		return nil, ErrEmailNotVerified
	}

	// 2. Foydalanuvchi allaqachon klinika ochganmi?
	existing, err := s.clinicRepo.GetByOwnerID(ctx, ownerID)
	if err == nil && existing != nil {
		return nil, repositories.ErrClinicAlreadyExists
	}

	// 3. Validatsiya
	name := utils.SanitizeString(req.Name)
	if len(name) < 2 || len(name) > 255 {
		return nil, ErrClinicNameRequired
	}
	phone := utils.SanitizeString(req.Phone)
	if phone == "" {
		return nil, ErrClinicPhoneRequired
	}
	region := utils.SanitizeString(req.Region)
	if region == "" {
		return nil, ErrClinicRegionRequired
	}
	district := utils.SanitizeString(req.District)
	if district == "" {
		return nil, ErrClinicDistrictRequired
	}
	address := utils.SanitizeString(req.Address)
	if address == "" {
		return nil, ErrClinicAddressRequired
	}

	// 4. Model yaratish
	// 4. Koordinatalarni tekshirish (agar kiritilgan bo'lsa)
	var lat, lon *float64
	if req.Latitude != nil || req.Longitude != nil {
		if req.Latitude == nil || req.Longitude == nil {
			return nil, errors.New("kenglik (latitude) va uzunlik (longitude) ikkalasi birga kiritilishi shart")
		}
		if *req.Latitude < -90 || *req.Latitude > 90 {
			return nil, errors.New("kenglik (latitude) -90 va 90 oralig'ida bo'lishi kerak")
		}
		if *req.Longitude < -180 || *req.Longitude > 180 {
			return nil, errors.New("uzunlik (longitude) -180 va 180 oralig'ida bo'lishi kerak")
		}
		lat = req.Latitude
		lon = req.Longitude
	}

	clinic := &models.Clinic{
		ID:            uuid.New(),
		Name:          name,
		Description:   utils.SanitizeString(req.Description),
		Phone:         phone,
		Email:         strings.ToLower(strings.TrimSpace(req.Email)),
		Region:        region,
		District:      district,
		Address:       address,
		LocationURL:   strings.TrimSpace(req.LocationURL),
		Latitude:      lat,
		Longitude:     lon,
		LogoURL:       req.LogoURL,
		CoverImageURL: req.CoverImageURL,
		Status:        models.ClinicStatusDraft,
		OwnerID:       ownerID,
	}

	// 5. Ish vaqtlari
	var workingHours []models.ClinicWorkingHour
	for _, wh := range req.WorkingHours {
		if wh.Weekday >= 1 && wh.Weekday <= 7 {
			workingHours = append(workingHours, models.ClinicWorkingHour{
				Weekday:     wh.Weekday,
				OpeningTime: wh.OpeningTime,
				ClosingTime: wh.ClosingTime,
				IsClosed:    wh.IsClosed,
			})
		}
	}

	// 6. Bazaga saqlash
	if err := s.clinicRepo.Create(ctx, clinic, workingHours); err != nil {
		return nil, err
	}

	return s.buildClinicDetailDTO(ctx, clinic)
}

func (s *clinicService) GetMyClinic(ctx context.Context, userID uuid.UUID) (*dto.ClinicDetailDTO, error) {
	clinic, err := s.clinicRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.buildClinicDetailDTO(ctx, clinic)
}

func (s *clinicService) GetClinicByID(ctx context.Context, clinicID uuid.UUID, requestingUserID uuid.UUID, isSuperAdmin bool) (*dto.ClinicDetailDTO, error) {
	clinic, err := s.clinicRepo.GetByID(ctx, clinicID)
	if err != nil {
		return nil, err
	}

	// Agar klinika tasdiqlanmagan (draft, pending, rejected, suspended) bo'lsa:
	// Uni faqat superadmin yoki shu klinikaning o'z admin/egasi ko'ra oladi!
	if clinic.Status != models.ClinicStatusApproved && !isSuperAdmin {
		if requestingUserID == uuid.Nil {
			return nil, repositories.ErrClinicNotFound
		}
		isAdmin, err := s.clinicRepo.IsUserClinicAdmin(ctx, clinicID, requestingUserID)
		if err != nil || !isAdmin {
			return nil, repositories.ErrClinicNotFound
		}
	}

	return s.buildClinicDetailDTO(ctx, clinic)
}

func (s *clinicService) UpdateClinic(ctx context.Context, clinicID, userID uuid.UUID, req *dto.UpdateClinicRequest) (*dto.ClinicDetailDTO, error) {
	// 1. Huquq tekshiruvi (IDOR himoyasi)
	isAdmin, err := s.clinicRepo.IsUserClinicAdmin(ctx, clinicID, userID)
	if err != nil || !isAdmin {
		return nil, repositories.ErrClinicAccessDenied
	}

	// 2. Klinikani olish
	clinic, err := s.clinicRepo.GetByID(ctx, clinicID)
	if err != nil {
		return nil, err
	}

	if clinic.Status == models.ClinicStatusSuspended {
		return nil, ErrClinicSuspended
	}

	// 3. Maydonlarni yangilash
	if req.Name != nil {
		trimmed := utils.SanitizeString(*req.Name)
		if len(trimmed) >= 2 {
			clinic.Name = trimmed
		}
	}
	if req.Description != nil {
		clinic.Description = utils.SanitizeString(*req.Description)
	}
	if req.Phone != nil {
		clinic.Phone = utils.SanitizeString(*req.Phone)
	}
	if req.Email != nil {
		clinic.Email = strings.ToLower(strings.TrimSpace(*req.Email))
	}
	if req.Region != nil {
		clinic.Region = utils.SanitizeString(*req.Region)
	}
	if req.District != nil {
		clinic.District = utils.SanitizeString(*req.District)
	}
	if req.Address != nil {
		clinic.Address = utils.SanitizeString(*req.Address)
	}
	if req.LocationURL != nil {
		clinic.LocationURL = strings.TrimSpace(*req.LocationURL)
	}
	if req.Latitude != nil || req.Longitude != nil {
		if req.Latitude == nil || req.Longitude == nil {
			return nil, errors.New("kenglik (latitude) va uzunlik (longitude) ikkalasi birga kiritilishi shart")
		}
		if *req.Latitude < -90 || *req.Latitude > 90 {
			return nil, errors.New("kenglik (latitude) -90 va 90 oralig'ida bo'lishi kerak")
		}
		if *req.Longitude < -180 || *req.Longitude > 180 {
			return nil, errors.New("uzunlik (longitude) -180 va 180 oralig'ida bo'lishi kerak")
		}
		clinic.Latitude = req.Latitude
		clinic.Longitude = req.Longitude
	}
	if req.LogoURL != nil {
		clinic.LogoURL = *req.LogoURL
	}
	if req.CoverImageURL != nil {
		clinic.CoverImageURL = *req.CoverImageURL
	}

	if err := s.clinicRepo.Update(ctx, clinic); err != nil {
		return nil, err
	}

	// 4. Ish soatlarini yangilash
	if req.WorkingHours != nil {
		var newHours []models.ClinicWorkingHour
		for _, wh := range *req.WorkingHours {
			if wh.Weekday >= 1 && wh.Weekday <= 7 {
				newHours = append(newHours, models.ClinicWorkingHour{
					ClinicID:    clinicID,
					Weekday:     wh.Weekday,
					OpeningTime: wh.OpeningTime,
					ClosingTime: wh.ClosingTime,
					IsClosed:    wh.IsClosed,
				})
			}
		}
		if err := s.clinicRepo.SetWorkingHours(ctx, clinicID, newHours); err != nil {
			return nil, err
		}
	}

	// Qayta yuklash
	updatedClinic, err := s.clinicRepo.GetByID(ctx, clinicID)
	if err != nil {
		return nil, err
	}

	return s.buildClinicDetailDTO(ctx, updatedClinic)
}

func (s *clinicService) SubmitClinic(ctx context.Context, clinicID, userID uuid.UUID) (*dto.ClinicDetailDTO, error) {
	// 1. Huquq tekshiruvi (IDOR)
	isAdmin, err := s.clinicRepo.IsUserClinicAdmin(ctx, clinicID, userID)
	if err != nil || !isAdmin {
		return nil, repositories.ErrClinicAccessDenied
	}

	clinic, err := s.clinicRepo.GetByID(ctx, clinicID)
	if err != nil {
		return nil, err
	}

	// 2. Status tekshiruvi
	if clinic.Status == models.ClinicStatusPending {
		return nil, ErrClinicAlreadySubmitted
	}
	if clinic.Status == models.ClinicStatusApproved {
		return nil, ErrClinicAlreadyApproved
	}
	if clinic.Status == models.ClinicStatusSuspended {
		return nil, ErrClinicSuspended
	}

	// 3. To'liqlik tekshiruvi
	if clinic.Name == "" || clinic.Phone == "" || clinic.Region == "" || clinic.District == "" || clinic.Address == "" {
		return nil, errors.New("klinikani tekshiruvga yuborish uchun barcha asosiy ma’lumotlar (nomi, telefon, viloyat, tuman, manzil) to‘ldirilgan bo‘lishi shart")
	}

	// Ish soatlari mavjudligi
	hours, err := s.clinicRepo.GetWorkingHours(ctx, clinicID)
	if err != nil || len(hours) == 0 {
		return nil, ErrClinicWorkingHourRequired
	}

	// 4. Statusni pending ga o'tkazish
	if err := s.clinicRepo.UpdateStatus(ctx, clinicID, models.ClinicStatusPending, ""); err != nil {
		return nil, err
	}

	// 5. Audit log
	_ = s.auditRepo.CreateAuditLog(ctx, &models.AuditLog{
		UserID:     &userID,
		Action:     "CLINIC_SUBMITTED_FOR_REVIEW",
		EntityType: "clinic",
		EntityID:   &clinicID,
		Details:    `{"status": "pending"}`,
	})

	clinic.Status = models.ClinicStatusPending
	clinic.RejectionReason = ""

	// 6. Enqueue Outbox Notification for Superadmin Telegram Bot
	if s.outboxRepo != nil {
		ownerName := ""
		ownerPhone := clinic.Phone
		ownerEmail := clinic.Email
		if clinic.Owner != nil {
			ownerName = fmt.Sprintf("%s %s", clinic.Owner.FirstName, clinic.Owner.LastName)
			if clinic.Owner.Phone != "" {
				ownerPhone = clinic.Owner.Phone
			}
			if clinic.Owner.Email != "" {
				ownerEmail = clinic.Owner.Email
			}
		}

		payloadBytes, _ := json.Marshal(map[string]any{
			"clinic_id":    clinic.ID.String(),
			"clinic_name":  clinic.Name,
			"owner_name":   ownerName,
			"phone":        ownerPhone,
			"email":        ownerEmail,
			"region":       clinic.Region,
			"district":     clinic.District,
			"submitted_at": time.Now().Format("02.01.2006 15:04"),
		})

		_ = s.outboxRepo.Enqueue(ctx, &models.NotificationOutbox{
			ID:         uuid.New(),
			EventType:  "clinic_submitted",
			EntityType: "clinic",
			EntityID:   clinic.ID,
			TargetRole: "admin",
			Payload:    string(payloadBytes),
			Status:     "pending",
		})
	}

	return s.buildClinicDetailDTO(ctx, clinic)
}

func (s *clinicService) buildClinicDetailDTO(ctx context.Context, clinic *models.Clinic) (*dto.ClinicDetailDTO, error) {
	hours, err := s.clinicRepo.GetWorkingHours(ctx, clinic.ID)
	if err != nil {
		hours = clinic.WorkingHours
	}

	whItems := make([]dto.WorkingHourItem, 0, len(hours))
	for _, h := range hours {
		whItems = append(whItems, dto.WorkingHourItem{
			Weekday:     h.Weekday,
			OpeningTime: h.OpeningTime,
			ClosingTime: h.ClosingTime,
			IsClosed:    h.IsClosed,
		})
	}

	doctorCount, _ := s.doctorRepo.CountDoctorsByClinic(ctx, clinic.ID)

	return &dto.ClinicDetailDTO{
		ID:              clinic.ID,
		Name:            clinic.Name,
		Description:     clinic.Description,
		Phone:           clinic.Phone,
		Email:           clinic.Email,
		Region:          clinic.Region,
		District:        clinic.District,
		Address:         clinic.Address,
		LocationURL:     clinic.LocationURL,
		Latitude:        clinic.Latitude,
		Longitude:       clinic.Longitude,
		LogoURL:         clinic.LogoURL,
		CoverImageURL:   clinic.CoverImageURL,
		Status:          string(clinic.Status),
		RejectionReason: clinic.RejectionReason,
		OwnerID:         clinic.OwnerID,
		WorkingHours:    whItems,
		DoctorsCount:    doctorCount,
		CreatedAt:       clinic.CreatedAt,
		UpdatedAt:       clinic.UpdatedAt,
	}, nil
}

