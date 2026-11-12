package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/models"
	"qabulgo-backend/internal/repositories"

	"github.com/google/uuid"
)

var (
	ErrPreparationNotAcknowledged    = errors.New("tekshiruvga tayyorgarlik qoidalari bilan tanishish va tasdiqlash shart")
	ErrEquipmentNotAssignedToClinic  = errors.New("uskuna ushbu klinikaga tegishli emas yoki nofaol")
	ErrServiceNotAssignedToEquipment = errors.New("xizmat ushbu uskunaga biriktirilmagan")
	ErrEquipmentInactive             = errors.New("uskuna hozirda faol emas")
	ErrSlotNotInEquipmentSchedule    = errors.New("tanlangan vaqt uskunaning ish jadvalida mavjud emas")
	ErrDiagPrematureCompletion       = errors.New("qabul vaqti hali kelmagan, qabulni faqat boshlanish vaqtidan so‘ng yakunlash mumkin")
	ErrDiagPrematureNoShow           = errors.New("qabul vaqti hali kelmagan, bemor kelmaganligini faqat qabul vaqtidan so‘ng belgilash mumkin")
)

var (
	diagDateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	diagTimeRegex = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)
)

type DiagnosticService interface {
	// Categories
	ListCategories(ctx context.Context) (*dto.ListDiagnosticCategoriesResponse, error)

	// Equipment Management
	CreateEquipment(ctx context.Context, clinicID, userID uuid.UUID, req *dto.CreateDiagnosticEquipmentRequest) (*dto.DiagnosticEquipmentDTO, error)
	UpdateEquipment(ctx context.Context, clinicID, userID, equipmentID uuid.UUID, req *dto.UpdateDiagnosticEquipmentRequest) (*dto.DiagnosticEquipmentDTO, error)
	GetEquipment(ctx context.Context, clinicID, equipmentID uuid.UUID) (*dto.DiagnosticEquipmentDTO, error)
	ListEquipments(ctx context.Context, clinicID uuid.UUID) (*dto.ListDiagnosticEquipmentsResponse, error)
	DeleteEquipment(ctx context.Context, clinicID, userID, equipmentID uuid.UUID) error

	// Service Management
	CreateService(ctx context.Context, clinicID, userID uuid.UUID, req *dto.CreateDiagnosticServiceRequest) (*dto.DiagnosticServiceDTO, error)
	UpdateService(ctx context.Context, clinicID, userID, serviceID uuid.UUID, req *dto.UpdateDiagnosticServiceRequest) (*dto.DiagnosticServiceDTO, error)
	GetService(ctx context.Context, serviceID uuid.UUID) (*dto.DiagnosticServiceDTO, error)
	ListServicesByClinic(ctx context.Context, clinicID uuid.UUID, categoryID *uuid.UUID) (*dto.ListDiagnosticServicesResponse, error)
	DeleteService(ctx context.Context, clinicID, userID, serviceID uuid.UUID) error

	// Equipment Schedule & Slots
	GetEquipmentWeeklySlots(ctx context.Context, clinicID, equipmentID uuid.UUID) ([]dto.DiagnosticWeeklySlotDTO, error)
	SaveEquipmentWeeklySlots(ctx context.Context, clinicID, userID, equipmentID uuid.UUID, req *dto.SaveDiagnosticWeeklySlotsRequest) ([]dto.DiagnosticWeeklySlotDTO, error)
	GetEquipmentDateOverrides(ctx context.Context, clinicID, equipmentID uuid.UUID, date, startDate, endDate string) ([]dto.DiagnosticDateOverrideDTO, error)
	CreateEquipmentDateOverride(ctx context.Context, clinicID, userID, equipmentID uuid.UUID, req *dto.CreateDiagnosticDateOverrideRequest) (*dto.DiagnosticDateOverrideDTO, error)
	DeleteEquipmentDateOverride(ctx context.Context, clinicID, userID, equipmentID, overrideID uuid.UUID) error
	GetDiagnosticEquipmentAvailability(ctx context.Context, equipmentID uuid.UUID, dateStr string) (*dto.DiagnosticAvailabilityResponse, error)

	// Appointments
	BookDiagnosticAppointment(ctx context.Context, patientID uuid.UUID, idempotencyKey string, req *dto.CreateDiagnosticAppointmentRequest) (*dto.DiagnosticAppointmentResponseDTO, error)
	GetPatientDiagnosticAppointments(ctx context.Context, patientID uuid.UUID, filter *dto.PatientDiagnosticAppointmentFilter) (*dto.DiagnosticAppointmentListResponseDTO, error)
	GetPatientDiagnosticAppointmentByID(ctx context.Context, patientID, appointmentID uuid.UUID) (*dto.DiagnosticAppointmentResponseDTO, error)
	CancelPatientDiagnosticAppointment(ctx context.Context, patientID, appointmentID uuid.UUID, req *dto.CancelDiagnosticAppointmentRequest) (*dto.DiagnosticAppointmentResponseDTO, error)
	GetClinicDiagnosticAppointments(ctx context.Context, clinicID, userID uuid.UUID, filter *dto.ClinicDiagnosticAppointmentFilter) (*dto.DiagnosticAppointmentListResponseDTO, error)
	CancelClinicDiagnosticAppointment(ctx context.Context, clinicID, userID, appointmentID uuid.UUID, req *dto.CancelDiagnosticAppointmentRequest) (*dto.DiagnosticAppointmentResponseDTO, error)
	CompleteClinicDiagnosticAppointment(ctx context.Context, clinicID, userID, appointmentID uuid.UUID) (*dto.DiagnosticAppointmentResponseDTO, error)
	NoShowClinicDiagnosticAppointment(ctx context.Context, clinicID, userID, appointmentID uuid.UUID) (*dto.DiagnosticAppointmentResponseDTO, error)
}

type diagnosticService struct {
	diagRepo   repositories.DiagnosticRepository
	clinicRepo repositories.ClinicRepository
	auditRepo  repositories.AuditRepository
	outboxRepo repositories.OutboxRepository
}

func NewDiagnosticService(
	diagRepo repositories.DiagnosticRepository,
	clinicRepo repositories.ClinicRepository,
	auditRepo repositories.AuditRepository,
	outboxRepo ...repositories.OutboxRepository,
) DiagnosticService {
	svc := &diagnosticService{
		diagRepo:   diagRepo,
		clinicRepo: clinicRepo,
		auditRepo:  auditRepo,
	}
	if len(outboxRepo) > 0 {
		svc.outboxRepo = outboxRepo[0]
	}
	return svc
}

func (s *diagnosticService) checkClinicPermission(ctx context.Context, clinicID, userID uuid.UUID) error {
	isAdmin, err := s.clinicRepo.IsUserClinicAdmin(ctx, clinicID, userID)
	if err != nil || !isAdmin {
		return errors.New("ushbu klinikani boshqarish huquqiga ega emassiz")
	}
	return nil
}

// --- Categories ---

func (s *diagnosticService) ListCategories(ctx context.Context) (*dto.ListDiagnosticCategoriesResponse, error) {
	cats, err := s.diagRepo.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]dto.DiagnosticCategoryDTO, len(cats))
	for i, c := range cats {
		res[i] = dto.DiagnosticCategoryDTO{
			ID:          c.ID,
			Name:        c.Name,
			Slug:        c.Slug,
			IconURL:     c.IconURL,
			Description: c.Description,
		}
	}
	return &dto.ListDiagnosticCategoriesResponse{
		Categories: res,
		Total:      len(res),
	}, nil
}

// --- Equipment Management ---

func (s *diagnosticService) CreateEquipment(
	ctx context.Context,
	clinicID, userID uuid.UUID,
	req *dto.CreateDiagnosticEquipmentRequest,
) (*dto.DiagnosticEquipmentDTO, error) {
	if err := s.checkClinicPermission(ctx, clinicID, userID); err != nil {
		return nil, err
	}

	// Verify category exists
	_, err := s.diagRepo.GetCategoryByID(ctx, req.CategoryID)
	if err != nil {
		return nil, err
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	eq := &models.DiagnosticEquipment{
		ClinicID:   clinicID,
		CategoryID: req.CategoryID,
		Name:       strings.TrimSpace(req.Name),
		RoomNumber: strings.TrimSpace(req.RoomNumber),
		IsActive:   isActive,
	}

	if err := s.diagRepo.CreateEquipment(ctx, eq); err != nil {
		return nil, err
	}

	return s.GetEquipment(ctx, clinicID, eq.ID)
}

func (s *diagnosticService) UpdateEquipment(
	ctx context.Context,
	clinicID, userID, equipmentID uuid.UUID,
	req *dto.UpdateDiagnosticEquipmentRequest,
) (*dto.DiagnosticEquipmentDTO, error) {
	if err := s.checkClinicPermission(ctx, clinicID, userID); err != nil {
		return nil, err
	}

	eq, err := s.diagRepo.GetEquipmentByID(ctx, equipmentID)
	if err != nil {
		return nil, err
	}
	if eq.ClinicID != clinicID {
		return nil, ErrEquipmentNotAssignedToClinic
	}

	if req.CategoryID != nil {
		if _, err := s.diagRepo.GetCategoryByID(ctx, *req.CategoryID); err != nil {
			return nil, err
		}
		eq.CategoryID = *req.CategoryID
	}
	if req.Name != nil {
		eq.Name = strings.TrimSpace(*req.Name)
	}
	if req.RoomNumber != nil {
		eq.RoomNumber = strings.TrimSpace(*req.RoomNumber)
	}
	if req.IsActive != nil {
		eq.IsActive = *req.IsActive
	}

	if err := s.diagRepo.UpdateEquipment(ctx, eq); err != nil {
		return nil, err
	}

	return s.GetEquipment(ctx, clinicID, equipmentID)
}

func (s *diagnosticService) GetEquipment(ctx context.Context, clinicID, equipmentID uuid.UUID) (*dto.DiagnosticEquipmentDTO, error) {
	eq, err := s.diagRepo.GetEquipmentByID(ctx, equipmentID)
	if err != nil {
		return nil, err
	}
	if clinicID != uuid.Nil && eq.ClinicID != clinicID {
		return nil, ErrEquipmentNotAssignedToClinic
	}

	res := &dto.DiagnosticEquipmentDTO{
		ID:         eq.ID,
		ClinicID:   eq.ClinicID,
		CategoryID: eq.CategoryID,
		Name:       eq.Name,
		RoomNumber: eq.RoomNumber,
		IsActive:   eq.IsActive,
		CreatedAt:  eq.CreatedAt,
		UpdatedAt:  eq.UpdatedAt,
	}
	if eq.Category != nil {
		res.Category = &dto.DiagnosticCategoryDTO{
			ID:          eq.Category.ID,
			Name:        eq.Category.Name,
			Slug:        eq.Category.Slug,
			IconURL:     eq.Category.IconURL,
			Description: eq.Category.Description,
		}
	}
	return res, nil
}

func (s *diagnosticService) ListEquipments(ctx context.Context, clinicID uuid.UUID) (*dto.ListDiagnosticEquipmentsResponse, error) {
	list, err := s.diagRepo.ListEquipmentsByClinic(ctx, clinicID)
	if err != nil {
		return nil, err
	}
	res := make([]dto.DiagnosticEquipmentDTO, len(list))
	for i, eq := range list {
		res[i] = dto.DiagnosticEquipmentDTO{
			ID:         eq.ID,
			ClinicID:   eq.ClinicID,
			CategoryID: eq.CategoryID,
			Name:       eq.Name,
			RoomNumber: eq.RoomNumber,
			IsActive:   eq.IsActive,
			CreatedAt:  eq.CreatedAt,
			UpdatedAt:  eq.UpdatedAt,
		}
		if eq.Category != nil {
			res[i].Category = &dto.DiagnosticCategoryDTO{
				ID:          eq.Category.ID,
				Name:        eq.Category.Name,
				Slug:        eq.Category.Slug,
				IconURL:     eq.Category.IconURL,
				Description: eq.Category.Description,
			}
		}
	}
	return &dto.ListDiagnosticEquipmentsResponse{
		Equipments: res,
		Total:      len(res),
	}, nil
}

func (s *diagnosticService) DeleteEquipment(ctx context.Context, clinicID, userID, equipmentID uuid.UUID) error {
	if err := s.checkClinicPermission(ctx, clinicID, userID); err != nil {
		return err
	}
	eq, err := s.diagRepo.GetEquipmentByID(ctx, equipmentID)
	if err != nil {
		return err
	}
	if eq.ClinicID != clinicID {
		return ErrEquipmentNotAssignedToClinic
	}
	return s.diagRepo.DeleteEquipment(ctx, equipmentID)
}

// --- Service Management ---

func (s *diagnosticService) CreateService(
	ctx context.Context,
	clinicID, userID uuid.UUID,
	req *dto.CreateDiagnosticServiceRequest,
) (*dto.DiagnosticServiceDTO, error) {
	if err := s.checkClinicPermission(ctx, clinicID, userID); err != nil {
		return nil, err
	}

	// Verify equipment belongs to clinic
	eq, err := s.diagRepo.GetEquipmentByID(ctx, req.EquipmentID)
	if err != nil {
		return nil, err
	}
	if eq.ClinicID != clinicID {
		return nil, ErrEquipmentNotAssignedToClinic
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	svc := &models.DiagnosticService{
		ClinicID:         clinicID,
		CategoryID:       req.CategoryID,
		EquipmentID:      req.EquipmentID,
		Name:             strings.TrimSpace(req.Name),
		Price:            req.Price,
		DurationMinutes:  req.DurationMinutes,
		PreparationRules: strings.TrimSpace(req.PreparationRules),
		IsActive:         isActive,
	}

	if err := s.diagRepo.CreateService(ctx, svc); err != nil {
		return nil, err
	}

	return s.GetService(ctx, svc.ID)
}

func (s *diagnosticService) UpdateService(
	ctx context.Context,
	clinicID, userID, serviceID uuid.UUID,
	req *dto.UpdateDiagnosticServiceRequest,
) (*dto.DiagnosticServiceDTO, error) {
	if err := s.checkClinicPermission(ctx, clinicID, userID); err != nil {
		return nil, err
	}

	svc, err := s.diagRepo.GetServiceByID(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	if svc.ClinicID != clinicID {
		return nil, errors.New("ushbu xizmat klinikangizga tegishli emas")
	}

	if req.EquipmentID != nil {
		eq, err := s.diagRepo.GetEquipmentByID(ctx, *req.EquipmentID)
		if err != nil {
			return nil, err
		}
		if eq.ClinicID != clinicID {
			return nil, ErrEquipmentNotAssignedToClinic
		}
		svc.EquipmentID = *req.EquipmentID
	}
	if req.CategoryID != nil {
		svc.CategoryID = *req.CategoryID
	}
	if req.Name != nil {
		svc.Name = strings.TrimSpace(*req.Name)
	}
	if req.Price != nil {
		svc.Price = *req.Price
	}
	if req.DurationMinutes != nil {
		svc.DurationMinutes = *req.DurationMinutes
	}
	if req.PreparationRules != nil {
		svc.PreparationRules = strings.TrimSpace(*req.PreparationRules)
	}
	if req.IsActive != nil {
		svc.IsActive = *req.IsActive
	}

	if err := s.diagRepo.UpdateService(ctx, svc); err != nil {
		return nil, err
	}

	return s.GetService(ctx, serviceID)
}

func (s *diagnosticService) GetService(ctx context.Context, serviceID uuid.UUID) (*dto.DiagnosticServiceDTO, error) {
	svc, err := s.diagRepo.GetServiceByID(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	res := &dto.DiagnosticServiceDTO{
		ID:               svc.ID,
		ClinicID:         svc.ClinicID,
		CategoryID:       svc.CategoryID,
		EquipmentID:      svc.EquipmentID,
		Name:             svc.Name,
		Price:            svc.Price,
		DurationMinutes:  svc.DurationMinutes,
		PreparationRules: svc.PreparationRules,
		IsActive:         svc.IsActive,
		CreatedAt:        svc.CreatedAt,
		UpdatedAt:        svc.UpdatedAt,
	}
	if svc.Category != nil {
		res.Category = &dto.DiagnosticCategoryDTO{
			ID:          svc.Category.ID,
			Name:        svc.Category.Name,
			Slug:        svc.Category.Slug,
			IconURL:     svc.Category.IconURL,
			Description: svc.Category.Description,
		}
	}
	if svc.Equipment != nil {
		res.Equipment = &dto.DiagnosticEquipmentDTO{
			ID:         svc.Equipment.ID,
			ClinicID:   svc.Equipment.ClinicID,
			CategoryID: svc.Equipment.CategoryID,
			Name:       svc.Equipment.Name,
			RoomNumber: svc.Equipment.RoomNumber,
			IsActive:   svc.Equipment.IsActive,
			CreatedAt:  svc.Equipment.CreatedAt,
			UpdatedAt:  svc.Equipment.UpdatedAt,
		}
	}
	return res, nil
}

func (s *diagnosticService) ListServicesByClinic(ctx context.Context, clinicID uuid.UUID, categoryID *uuid.UUID) (*dto.ListDiagnosticServicesResponse, error) {
	list, err := s.diagRepo.ListServicesByClinic(ctx, clinicID, categoryID)
	if err != nil {
		return nil, err
	}
	res := make([]dto.DiagnosticServiceDTO, len(list))
	for i, svc := range list {
		res[i] = dto.DiagnosticServiceDTO{
			ID:               svc.ID,
			ClinicID:         svc.ClinicID,
			CategoryID:       svc.CategoryID,
			EquipmentID:      svc.EquipmentID,
			Name:             svc.Name,
			Price:            svc.Price,
			DurationMinutes:  svc.DurationMinutes,
			PreparationRules: svc.PreparationRules,
			IsActive:         svc.IsActive,
			CreatedAt:        svc.CreatedAt,
			UpdatedAt:        svc.UpdatedAt,
		}
		if svc.Category != nil {
			res[i].Category = &dto.DiagnosticCategoryDTO{
				ID:          svc.Category.ID,
				Name:        svc.Category.Name,
				Slug:        svc.Category.Slug,
				IconURL:     svc.Category.IconURL,
				Description: svc.Category.Description,
			}
		}
		if svc.Equipment != nil {
			res[i].Equipment = &dto.DiagnosticEquipmentDTO{
				ID:         svc.Equipment.ID,
				ClinicID:   svc.Equipment.ClinicID,
				CategoryID: svc.Equipment.CategoryID,
				Name:       svc.Equipment.Name,
				RoomNumber: svc.Equipment.RoomNumber,
				IsActive:   svc.Equipment.IsActive,
				CreatedAt:  svc.Equipment.CreatedAt,
				UpdatedAt:  svc.Equipment.UpdatedAt,
			}
		}
	}
	return &dto.ListDiagnosticServicesResponse{
		Services: res,
		Total:    len(res),
	}, nil
}

func (s *diagnosticService) DeleteService(ctx context.Context, clinicID, userID, serviceID uuid.UUID) error {
	if err := s.checkClinicPermission(ctx, clinicID, userID); err != nil {
		return err
	}
	svc, err := s.diagRepo.GetServiceByID(ctx, serviceID)
	if err != nil {
		return err
	}
	if svc.ClinicID != clinicID {
		return errors.New("ushbu xizmat klinikangizga tegishli emas")
	}
	return s.diagRepo.DeleteService(ctx, serviceID)
}

// --- Schedule & Slots ---

func (s *diagnosticService) GetEquipmentWeeklySlots(ctx context.Context, clinicID, equipmentID uuid.UUID) ([]dto.DiagnosticWeeklySlotDTO, error) {
	eq, err := s.diagRepo.GetEquipmentByID(ctx, equipmentID)
	if err != nil {
		return nil, err
	}
	if clinicID != uuid.Nil && eq.ClinicID != clinicID {
		return nil, ErrEquipmentNotAssignedToClinic
	}

	slots, err := s.diagRepo.GetWeeklySlots(ctx, equipmentID)
	if err != nil {
		return nil, err
	}
	res := make([]dto.DiagnosticWeeklySlotDTO, len(slots))
	for i, sl := range slots {
		res[i] = dto.DiagnosticWeeklySlotDTO{
			ID:          sl.ID,
			EquipmentID: sl.EquipmentID,
			Weekday:     sl.Weekday,
			StartTime:   sl.StartTime,
			EndTime:     sl.EndTime,
			Status:      string(sl.Status),
		}
	}
	return res, nil
}

func (s *diagnosticService) SaveEquipmentWeeklySlots(
	ctx context.Context,
	clinicID, userID, equipmentID uuid.UUID,
	req *dto.SaveDiagnosticWeeklySlotsRequest,
) ([]dto.DiagnosticWeeklySlotDTO, error) {
	if err := s.checkClinicPermission(ctx, clinicID, userID); err != nil {
		return nil, err
	}
	eq, err := s.diagRepo.GetEquipmentByID(ctx, equipmentID)
	if err != nil {
		return nil, err
	}
	if eq.ClinicID != clinicID {
		return nil, ErrEquipmentNotAssignedToClinic
	}

	// Validate slots
	for _, sl := range req.Slots {
		if sl.Weekday < 1 || sl.Weekday > 7 {
			return nil, ErrInvalidWeekday
		}
		if !diagTimeRegex.MatchString(sl.StartTime) || !diagTimeRegex.MatchString(sl.EndTime) {
			return nil, ErrInvalidTimeFormat
		}
		if sl.StartTime >= sl.EndTime {
			return nil, ErrStartMustBeBeforeEnd
		}
	}

	modelSlots := make([]models.DiagnosticWeeklySlot, len(req.Slots))
	for i, sl := range req.Slots {
		st := models.SlotStatusActive
		if sl.Status == "inactive" {
			st = models.SlotStatusInactive
		}
		modelSlots[i] = models.DiagnosticWeeklySlot{
			EquipmentID: equipmentID,
			Weekday:     sl.Weekday,
			StartTime:   sl.StartTime,
			EndTime:     sl.EndTime,
			Status:      st,
		}
	}

	if err := s.diagRepo.SaveWeeklySlots(ctx, equipmentID, modelSlots); err != nil {
		return nil, err
	}

	return s.GetEquipmentWeeklySlots(ctx, clinicID, equipmentID)
}

func (s *diagnosticService) GetEquipmentDateOverrides(
	ctx context.Context,
	clinicID, equipmentID uuid.UUID,
	date, startDate, endDate string,
) ([]dto.DiagnosticDateOverrideDTO, error) {
	eq, err := s.diagRepo.GetEquipmentByID(ctx, equipmentID)
	if err != nil {
		return nil, err
	}
	if clinicID != uuid.Nil && eq.ClinicID != clinicID {
		return nil, ErrEquipmentNotAssignedToClinic
	}

	overrides, err := s.diagRepo.GetDateOverrides(ctx, equipmentID, date, startDate, endDate)
	if err != nil {
		return nil, err
	}
	res := make([]dto.DiagnosticDateOverrideDTO, len(overrides))
	for i, o := range overrides {
		res[i] = dto.DiagnosticDateOverrideDTO{
			ID:          o.ID,
			EquipmentID: o.EquipmentID,
			Date:        o.Date,
			StartTime:   o.StartTime,
			EndTime:     o.EndTime,
			Status:      string(o.Status),
			Reason:      o.Reason,
		}
	}
	return res, nil
}

func (s *diagnosticService) CreateEquipmentDateOverride(
	ctx context.Context,
	clinicID, userID, equipmentID uuid.UUID,
	req *dto.CreateDiagnosticDateOverrideRequest,
) (*dto.DiagnosticDateOverrideDTO, error) {
	if err := s.checkClinicPermission(ctx, clinicID, userID); err != nil {
		return nil, err
	}
	eq, err := s.diagRepo.GetEquipmentByID(ctx, equipmentID)
	if err != nil {
		return nil, err
	}
	if eq.ClinicID != clinicID {
		return nil, ErrEquipmentNotAssignedToClinic
	}

	override := &models.DiagnosticDateSlotOverride{
		EquipmentID: equipmentID,
		Date:        req.Date,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Status:      models.OverrideStatus(req.Status),
		Reason:      req.Reason,
	}

	if err := s.diagRepo.CreateDateOverride(ctx, override); err != nil {
		return nil, err
	}

	return &dto.DiagnosticDateOverrideDTO{
		ID:          override.ID,
		EquipmentID: override.EquipmentID,
		Date:        override.Date,
		StartTime:   override.StartTime,
		EndTime:     override.EndTime,
		Status:      string(override.Status),
		Reason:      override.Reason,
	}, nil
}

func (s *diagnosticService) DeleteEquipmentDateOverride(ctx context.Context, clinicID, userID, equipmentID, overrideID uuid.UUID) error {
	if err := s.checkClinicPermission(ctx, clinicID, userID); err != nil {
		return err
	}
	eq, err := s.diagRepo.GetEquipmentByID(ctx, equipmentID)
	if err != nil {
		return err
	}
	if eq.ClinicID != clinicID {
		return ErrEquipmentNotAssignedToClinic
	}
	return s.diagRepo.DeleteDateOverride(ctx, equipmentID, overrideID)
}

// GetDiagnosticEquipmentAvailability calculates free available slots for an equipment on a target date
func (s *diagnosticService) GetDiagnosticEquipmentAvailability(
	ctx context.Context,
	equipmentID uuid.UUID,
	dateStr string,
) (*dto.DiagnosticAvailabilityResponse, error) {
	if !diagDateRegex.MatchString(dateStr) {
		return nil, ErrInvalidDateFormat
	}
	targetDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, ErrInvalidDateFormat
	}

	wd := int(targetDate.Weekday())
	if wd == 0 {
		wd = 7
	}

	eq, weeklySlots, dateOverrides, err := s.diagRepo.GetEquipmentAvailabilityData(ctx, equipmentID, dateStr, wd)
	if err != nil {
		return nil, err
	}

	if !eq.IsActive {
		return nil, ErrEquipmentInactive
	}
	if eq.Clinic == nil || eq.Clinic.Status != models.ClinicStatusApproved {
		return nil, ErrClinicNotApproved
	}

	tashkentLoc := getTashkentLocation()
	nowTashkent := time.Now().In(tashkentLoc)
	todayTashkentStr := nowTashkent.Format("2006-01-02")
	currentTimeStr := nowTashkent.Format("15:04")

	emptyRes := &dto.DiagnosticAvailabilityResponse{
		Date:        dateStr,
		Timezone:    "Asia/Tashkent",
		ClinicID:    eq.ClinicID,
		EquipmentID: eq.ID,
		Slots:       []dto.DiagnosticAvailabilitySlotDTO{},
	}

	// Past dates return empty slots
	if dateStr < todayTashkentStr {
		return emptyRes, nil
	}

	// Check day_off override
	for _, o := range dateOverrides {
		if o.Status == models.OverrideStatusDayOff {
			return emptyRes, nil
		}
	}

	// Candidate slots from weekly schedule
	type slotEntry struct {
		start string
		end   string
	}
	var candidateSlots []slotEntry

	for _, ws := range weeklySlots {
		if ws.Status != models.SlotStatusActive {
			continue
		}
		isClosed := false
		for _, o := range dateOverrides {
			if o.Status == models.OverrideStatusClosed && o.StartTime != nil && o.EndTime != nil {
				if ws.StartTime < *o.EndTime && *o.StartTime < ws.EndTime {
					isClosed = true
					break
				}
			}
		}
		if !isClosed {
			candidateSlots = append(candidateSlots, slotEntry{
				start: ws.StartTime,
				end:   ws.EndTime,
			})
		}
	}

	// Add available overrides
	for _, o := range dateOverrides {
		if o.Status == models.OverrideStatusAvailable && o.StartTime != nil && o.EndTime != nil {
			candidateSlots = append(candidateSlots, slotEntry{
				start: *o.StartTime,
				end:   *o.EndTime,
			})
		}
	}

	// Filter booked appointments
	bookedAppts, err := s.diagRepo.GetBookedSlotsForEquipment(ctx, equipmentID, dateStr)
	if err != nil {
		return nil, err
	}

	var availableSlots []dto.DiagnosticAvailabilitySlotDTO
	for _, cand := range candidateSlots {
		// Past time slots on today's date
		if dateStr == todayTashkentStr && cand.start <= currentTimeStr {
			continue
		}

		isBooked := false
		for _, booked := range bookedAppts {
			if cand.start < booked.EndTime && booked.StartTime < cand.end {
				isBooked = true
				break
			}
		}

		if !isBooked {
			availableSlots = append(availableSlots, dto.DiagnosticAvailabilitySlotDTO{
				StartTime: cand.start,
				EndTime:   cand.end,
			})
		}
	}

	// Sort slots
	sort.Slice(availableSlots, func(i, j int) bool {
		return availableSlots[i].StartTime < availableSlots[j].StartTime
	})

	return &dto.DiagnosticAvailabilityResponse{
		Date:        dateStr,
		Timezone:    "Asia/Tashkent",
		ClinicID:    eq.ClinicID,
		EquipmentID: eq.ID,
		Slots:       availableSlots,
	}, nil
}

// --- Appointments ---

func (s *diagnosticService) BookDiagnosticAppointment(
	ctx context.Context,
	patientID uuid.UUID,
	idempotencyKey string,
	req *dto.CreateDiagnosticAppointmentRequest,
) (*dto.DiagnosticAppointmentResponseDTO, error) {
	if patientID == uuid.Nil {
		return nil, ErrPatientOnly
	}

	// 1. Preparation Rules acknowledgment validation
	if !req.PreparationAcknowledged {
		return nil, ErrPreparationNotAcknowledged
	}

	// 2. Validate request payload formats
	if !diagDateRegex.MatchString(req.Date) {
		return nil, ErrInvalidDateFormat
	}
	if !diagTimeRegex.MatchString(req.StartTime) || !diagTimeRegex.MatchString(req.EndTime) {
		return nil, ErrInvalidTimeFormat
	}
	if req.StartTime >= req.EndTime {
		return nil, ErrStartMustBeBeforeEnd
	}

	// 3. Idempotency Key check
	cleanKey := strings.TrimSpace(idempotencyKey)
	var keyPtr *string
	if cleanKey != "" {
		keyPtr = &cleanKey
		existingAppt, err := s.diagRepo.FindByIdempotencyKey(ctx, patientID, cleanKey)
		if err != nil {
			return nil, fmt.Errorf("idempotency tekshiruvida xatolik: %w", err)
		}
		if existingAppt != nil {
			if existingAppt.ClinicID == req.ClinicID &&
				existingAppt.ServiceID == req.ServiceID &&
				existingAppt.EquipmentID == req.EquipmentID &&
				existingAppt.AppointmentDate == req.Date &&
				existingAppt.StartTime == req.StartTime &&
				existingAppt.EndTime == req.EndTime {
				return mapDiagAppointmentToDTO(existingAppt), nil
			}
			return nil, repositories.ErrIdempotencyConflict
		}
	}

	// 4. Check Clinic exists and is Approved
	clinic, err := s.clinicRepo.GetByID(ctx, req.ClinicID)
	if err != nil {
		return nil, err
	}
	if clinic.Status != models.ClinicStatusApproved {
		return nil, ErrClinicNotApproved
	}

	// 5. Check Service exists, belongs to clinic and equipment, and is Active
	svc, err := s.diagRepo.GetServiceByID(ctx, req.ServiceID)
	if err != nil {
		return nil, err
	}
	if svc.ClinicID != req.ClinicID {
		return nil, errors.New("xizmat ushbu klinikaga tegishli emas")
	}
	if svc.EquipmentID != req.EquipmentID {
		return nil, ErrServiceNotAssignedToEquipment
	}
	if !svc.IsActive {
		return nil, errors.New("ushbu diagnostika xizmati ayni paytda nofaol")
	}

	// 6. Check Equipment active
	eq, err := s.diagRepo.GetEquipmentByID(ctx, req.EquipmentID)
	if err != nil {
		return nil, err
	}
	if !eq.IsActive {
		return nil, ErrEquipmentInactive
	}

	// 7. Timezone validation (Asia/Tashkent): check past date/time
	loc := getTashkentLocation()
	now := time.Now().In(loc)
	slotDateTime, err := time.ParseInLocation("2006-01-02 15:04", req.Date+" "+req.StartTime, loc)
	if err != nil {
		return nil, ErrInvalidDateFormat
	}
	if slotDateTime.Before(now) {
		return nil, ErrPastTimeBooking
	}

	// 8. Check Equipment Weekly Schedule / Overrides
	weekday := int(slotDateTime.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	weeklySlots, err := s.diagRepo.GetWeeklySlots(ctx, req.EquipmentID)
	if err != nil {
		return nil, err
	}
	dateOverrides, err := s.diagRepo.GetDateOverrides(ctx, req.EquipmentID, req.Date, "", "")
	if err != nil {
		return nil, err
	}

	// Day off override
	for _, o := range dateOverrides {
		if o.Status == models.OverrideStatusDayOff {
			return nil, ErrSlotDateOverrideClosed
		}
	}

	// Closed override
	for _, o := range dateOverrides {
		if o.Status == models.OverrideStatusClosed && o.StartTime != nil && o.EndTime != nil {
			if req.StartTime < *o.EndTime && *o.StartTime < req.EndTime {
				return nil, ErrSlotDateOverrideClosed
			}
		}
	}

	// Weekly slot fit
	slotInWeekly := false
	for _, ws := range weeklySlots {
		if ws.Status == models.SlotStatusActive && ws.Weekday == weekday {
			if ws.StartTime <= req.StartTime && ws.EndTime >= req.EndTime {
				slotInWeekly = true
				break
			}
		}
	}

	slotInOverride := false
	for _, o := range dateOverrides {
		if o.Status == models.OverrideStatusAvailable && o.StartTime != nil && o.EndTime != nil {
			if *o.StartTime <= req.StartTime && *o.EndTime >= req.EndTime {
				slotInOverride = true
				break
			}
		}
	}

	if !slotInWeekly && !slotInOverride {
		return nil, ErrSlotNotInEquipmentSchedule
	}

	// 9. Check Equipment Overlap
	conflictAppt, err := s.diagRepo.CheckEquipmentBookingOverlap(ctx, req.EquipmentID, req.Date, req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}
	if conflictAppt != nil {
		return nil, repositories.ErrDoubleBookingConflict
	}

	// 10. Create Appointment inside Transaction with Row Locking
	appt := &models.DiagnosticAppointment{
		PatientID:               patientID,
		ClinicID:                req.ClinicID,
		ServiceID:               req.ServiceID,
		EquipmentID:             req.EquipmentID,
		AppointmentDate:         req.Date,
		StartTime:               req.StartTime,
		EndTime:                 req.EndTime,
		Price:                   svc.Price, // Authoritative price from DB
		Status:                  models.DiagAppointmentStatusConfirmed,
		PreparationAcknowledged: true,
		IdempotencyKey:          keyPtr,
	}

	if err := s.diagRepo.CreateAppointment(ctx, appt); err != nil {
		return nil, err
	}

	fullAppt, err := s.diagRepo.GetAppointmentByID(ctx, appt.ID)
	if err != nil {
		return mapDiagAppointmentToDTO(appt), nil
	}
	return mapDiagAppointmentToDTO(fullAppt), nil
}

func (s *diagnosticService) GetPatientDiagnosticAppointments(
	ctx context.Context,
	patientID uuid.UUID,
	filter *dto.PatientDiagnosticAppointmentFilter,
) (*dto.DiagnosticAppointmentListResponseDTO, error) {
	loc := getTashkentLocation()
	now := time.Now().In(loc)
	todayStr := now.Format("2006-01-02")
	nowTimeStr := now.Format("15:04")

	list, total, err := s.diagRepo.ListForPatient(ctx, patientID, filter, todayStr, nowTimeStr)
	if err != nil {
		return nil, err
	}

	dtos := make([]dto.DiagnosticAppointmentResponseDTO, len(list))
	for i, appt := range list {
		dtos[i] = *mapDiagAppointmentToDTO(&appt)
	}

	page := 1
	limit := 10
	if filter != nil {
		if filter.Page > 0 {
			page = filter.Page
		}
		if filter.Limit > 0 {
			limit = filter.Limit
		}
	}

	return &dto.DiagnosticAppointmentListResponseDTO{
		Appointments: dtos,
		Total:        total,
		Page:         page,
		Limit:        limit,
	}, nil
}

func (s *diagnosticService) GetPatientDiagnosticAppointmentByID(
	ctx context.Context,
	patientID, appointmentID uuid.UUID,
) (*dto.DiagnosticAppointmentResponseDTO, error) {
	appt, err := s.diagRepo.GetAppointmentByID(ctx, appointmentID)
	if err != nil {
		return nil, err
	}
	if appt.PatientID != patientID {
		return nil, ErrUnauthorizedAppointmentAccess
	}
	return mapDiagAppointmentToDTO(appt), nil
}

func (s *diagnosticService) CancelPatientDiagnosticAppointment(
	ctx context.Context,
	patientID, appointmentID uuid.UUID,
	req *dto.CancelDiagnosticAppointmentRequest,
) (*dto.DiagnosticAppointmentResponseDTO, error) {
	appt, err := s.diagRepo.GetAppointmentByID(ctx, appointmentID)
	if err != nil {
		return nil, err
	}
	if appt.PatientID != patientID {
		return nil, ErrUnauthorizedAppointmentAccess
	}
	if appt.Status != models.DiagAppointmentStatusConfirmed {
		return nil, ErrOnlyConfirmedCancellable
	}

	loc := getTashkentLocation()
	now := time.Now().In(loc)
	slotStart, err := time.ParseInLocation("2006-01-02 15:04", appt.AppointmentDate+" "+appt.StartTime, loc)
	if err == nil && slotStart.Before(now) {
		return nil, ErrPastAppointmentCancellation
	}

	var reasonPtr *string
	if req != nil && strings.TrimSpace(req.Reason) != "" {
		trimmed := strings.TrimSpace(req.Reason)
		reasonPtr = &trimmed
	}

	if err := s.diagRepo.UpdateAppointmentStatus(ctx, appointmentID, models.DiagAppointmentStatusConfirmed, models.DiagAppointmentStatusCancelled, reasonPtr, "cancelled_at", now); err != nil {
		return nil, err
	}

	updated, err := s.diagRepo.GetAppointmentByID(ctx, appointmentID)
	if err != nil {
		return nil, err
	}
	return mapDiagAppointmentToDTO(updated), nil
}

func (s *diagnosticService) GetClinicDiagnosticAppointments(
	ctx context.Context,
	clinicID, userID uuid.UUID,
	filter *dto.ClinicDiagnosticAppointmentFilter,
) (*dto.DiagnosticAppointmentListResponseDTO, error) {
	if err := s.checkClinicPermission(ctx, clinicID, userID); err != nil {
		return nil, err
	}

	list, total, err := s.diagRepo.ListForClinic(ctx, clinicID, filter)
	if err != nil {
		return nil, err
	}

	dtos := make([]dto.DiagnosticAppointmentResponseDTO, len(list))
	for i, appt := range list {
		dtos[i] = *mapDiagAppointmentToDTO(&appt)
	}

	page := 1
	limit := 10
	if filter != nil {
		if filter.Page > 0 {
			page = filter.Page
		}
		if filter.Limit > 0 {
			limit = filter.Limit
		}
	}

	return &dto.DiagnosticAppointmentListResponseDTO{
		Appointments: dtos,
		Total:        total,
		Page:         page,
		Limit:        limit,
	}, nil
}

func (s *diagnosticService) CancelClinicDiagnosticAppointment(
	ctx context.Context,
	clinicID, userID, appointmentID uuid.UUID,
	req *dto.CancelDiagnosticAppointmentRequest,
) (*dto.DiagnosticAppointmentResponseDTO, error) {
	if err := s.checkClinicPermission(ctx, clinicID, userID); err != nil {
		return nil, err
	}
	appt, err := s.diagRepo.GetAppointmentByID(ctx, appointmentID)
	if err != nil {
		return nil, err
	}
	if appt.ClinicID != clinicID {
		return nil, ErrUnauthorizedAppointmentAccess
	}
	if appt.Status != models.DiagAppointmentStatusConfirmed {
		return nil, ErrOnlyConfirmedCancellable
	}

	loc := getTashkentLocation()
	now := time.Now().In(loc)

	var reasonPtr *string
	if req != nil && strings.TrimSpace(req.Reason) != "" {
		trimmed := strings.TrimSpace(req.Reason)
		reasonPtr = &trimmed
	}

	if err := s.diagRepo.UpdateAppointmentStatus(ctx, appointmentID, models.DiagAppointmentStatusConfirmed, models.DiagAppointmentStatusCancelled, reasonPtr, "cancelled_at", now); err != nil {
		return nil, err
	}

	updated, err := s.diagRepo.GetAppointmentByID(ctx, appointmentID)
	if err != nil {
		return nil, err
	}
	return mapDiagAppointmentToDTO(updated), nil
}

func (s *diagnosticService) CompleteClinicDiagnosticAppointment(
	ctx context.Context,
	clinicID, userID, appointmentID uuid.UUID,
) (*dto.DiagnosticAppointmentResponseDTO, error) {
	if err := s.checkClinicPermission(ctx, clinicID, userID); err != nil {
		return nil, err
	}
	appt, err := s.diagRepo.GetAppointmentByID(ctx, appointmentID)
	if err != nil {
		return nil, err
	}
	if appt.ClinicID != clinicID {
		return nil, ErrUnauthorizedAppointmentAccess
	}
	if appt.Status != models.DiagAppointmentStatusConfirmed {
		return nil, repositories.ErrAppointmentNotEditable
	}

	loc := getTashkentLocation()
	now := time.Now().In(loc)
	slotStart, err := time.ParseInLocation("2006-01-02 15:04", appt.AppointmentDate+" "+appt.StartTime, loc)
	if err == nil && now.Before(slotStart) {
		return nil, ErrDiagPrematureCompletion
	}

	if err := s.diagRepo.UpdateAppointmentStatus(ctx, appointmentID, models.DiagAppointmentStatusConfirmed, models.DiagAppointmentStatusCompleted, nil, "completed_at", now); err != nil {
		return nil, err
	}

	updated, err := s.diagRepo.GetAppointmentByID(ctx, appointmentID)
	if err != nil {
		return nil, err
	}
	return mapDiagAppointmentToDTO(updated), nil
}

func (s *diagnosticService) NoShowClinicDiagnosticAppointment(
	ctx context.Context,
	clinicID, userID, appointmentID uuid.UUID,
) (*dto.DiagnosticAppointmentResponseDTO, error) {
	if err := s.checkClinicPermission(ctx, clinicID, userID); err != nil {
		return nil, err
	}
	appt, err := s.diagRepo.GetAppointmentByID(ctx, appointmentID)
	if err != nil {
		return nil, err
	}
	if appt.ClinicID != clinicID {
		return nil, ErrUnauthorizedAppointmentAccess
	}
	if appt.Status != models.DiagAppointmentStatusConfirmed {
		return nil, repositories.ErrAppointmentNotEditable
	}

	loc := getTashkentLocation()
	now := time.Now().In(loc)
	slotStart, err := time.ParseInLocation("2006-01-02 15:04", appt.AppointmentDate+" "+appt.StartTime, loc)
	if err == nil && now.Before(slotStart) {
		return nil, ErrDiagPrematureNoShow
	}

	if err := s.diagRepo.UpdateAppointmentStatus(ctx, appointmentID, models.DiagAppointmentStatusConfirmed, models.DiagAppointmentStatusNoShow, nil, "no_show_at", now); err != nil {
		return nil, err
	}

	updated, err := s.diagRepo.GetAppointmentByID(ctx, appointmentID)
	if err != nil {
		return nil, err
	}
	return mapDiagAppointmentToDTO(updated), nil
}

func mapDiagAppointmentToDTO(appt *models.DiagnosticAppointment) *dto.DiagnosticAppointmentResponseDTO {
	res := &dto.DiagnosticAppointmentResponseDTO{
		ID:                      appt.ID,
		PatientID:               appt.PatientID,
		ClinicID:                appt.ClinicID,
		ServiceID:               appt.ServiceID,
		EquipmentID:             appt.EquipmentID,
		AppointmentDate:         appt.AppointmentDate,
		StartTime:               appt.StartTime,
		EndTime:                 appt.EndTime,
		Price:                   appt.Price,
		Status:                  string(appt.Status),
		PreparationAcknowledged: appt.PreparationAcknowledged,
		CancellationReason:      appt.CancellationReason,
		CancelledAt:             appt.CancelledAt,
		CompletedAt:             appt.CompletedAt,
		NoShowAt:                appt.NoShowAt,
		CreatedAt:               appt.CreatedAt,
		UpdatedAt:               appt.UpdatedAt,
	}

	if appt.Service != nil {
		res.PreparationRules = appt.Service.PreparationRules
		res.Service = &dto.DiagnosticServiceDTO{
			ID:               appt.Service.ID,
			ClinicID:         appt.Service.ClinicID,
			CategoryID:       appt.Service.CategoryID,
			EquipmentID:      appt.Service.EquipmentID,
			Name:             appt.Service.Name,
			Price:            appt.Service.Price,
			DurationMinutes:  appt.Service.DurationMinutes,
			PreparationRules: appt.Service.PreparationRules,
			IsActive:         appt.Service.IsActive,
		}
	}

	if appt.Equipment != nil {
		res.Equipment = &dto.DiagnosticEquipmentDTO{
			ID:         appt.Equipment.ID,
			ClinicID:   appt.Equipment.ClinicID,
			CategoryID: appt.Equipment.CategoryID,
			Name:       appt.Equipment.Name,
			RoomNumber: appt.Equipment.RoomNumber,
			IsActive:   appt.Equipment.IsActive,
		}
	}

	if appt.Clinic != nil {
		res.Clinic = &dto.AppointmentClinicDTO{
			ID:       appt.Clinic.ID,
			Name:     appt.Clinic.Name,
			Address:  appt.Clinic.Address,
			Region:   appt.Clinic.Region,
			District: appt.Clinic.District,
			Phone:    appt.Clinic.Phone,
		}
	}

	if appt.Patient != nil {
		res.Patient = &dto.AppointmentPatientDTO{
			ID:        appt.Patient.ID,
			FirstName: appt.Patient.FirstName,
			LastName:  appt.Patient.LastName,
			Phone:     appt.Patient.Phone,
		}
	}

	return res
}
