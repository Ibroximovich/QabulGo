package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/models"
	"qabulgo-backend/internal/repositories"

	"github.com/google/uuid"
)

var (
	timeRegex = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)
	dateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

	ErrInvalidTimeFormat      = errors.New("vaqt formati noto‘g‘ri, HH:MM ko‘rinishida bo‘lishi shart (masalan, 09:00)")
	ErrInvalidDateFormat      = errors.New("sana formati noto‘g‘ri, YYYY-MM-DD ko‘rinishida bo‘lishi shart")
	ErrStartMustBeBeforeEnd   = errors.New("boshlanish vaqti tugash vaqtidan kichik bo‘lishi shart")
	ErrInvalidWeekday         = errors.New("hafta kuni 1 (Dushanba) dan 7 (Yakshanba) gacha bo‘lishi shart")
	ErrInvalidSlotDuration    = errors.New("qabul davomiyligi 5 dan 240 daqiqagacha bo‘lishi shart")
	ErrBreakOutsideHours      = errors.New("tanaffus ish vaqti oralig‘ida bo‘lishi shart")
	ErrOverlappingBreaks      = errors.New("tanaffuslar bir-birini qoplamasligi shart")
	ErrInternalSlotOverlap    = errors.New("kiritilgan qabul vaqtlari bir-birini qoplamasligi shart")
	ErrClinicNotApproved      = errors.New("klinika hali tasdiqlanmagan")
	ErrDoctorInactive         = errors.New("shifokor ayni paytda faol emas")
)

type ScheduleService interface {
	GenerateSlotsPreview(ctx context.Context, clinicID, userID, doctorIDOrCDID uuid.UUID, req *dto.GenerateScheduleRequest) (*dto.GenerateScheduleResponse, error)
	GetWeeklySchedule(ctx context.Context, clinicID, doctorIDOrCDID uuid.UUID) (*dto.DoctorWeeklyScheduleDTO, error)
	SaveWeeklySchedule(ctx context.Context, clinicID, userID, doctorIDOrCDID uuid.UUID, req *dto.SaveWeeklyScheduleRequest) (*dto.DoctorWeeklyScheduleDTO, error)
	GetDateExceptions(ctx context.Context, clinicID, doctorIDOrCDID uuid.UUID, date, startDate, endDate string) ([]dto.DoctorDateOverrideDTO, error)
	CreateDateException(ctx context.Context, clinicID, userID, doctorIDOrCDID uuid.UUID, req *dto.CreateDateOverrideRequest) (*dto.DoctorDateOverrideDTO, error)
	UpdateDateException(ctx context.Context, clinicID, userID, doctorIDOrCDID, exceptionID uuid.UUID, req *dto.UpdateDateOverrideRequest) (*dto.DoctorDateOverrideDTO, error)
	DeleteDateException(ctx context.Context, clinicID, userID, doctorIDOrCDID, exceptionID uuid.UUID) error
	GetPublicAvailability(ctx context.Context, clinicID, doctorIDOrCDID uuid.UUID, dateStr string) (*dto.PublicAvailabilityResponse, error)
}

type scheduleService struct {
	scheduleRepo repositories.ScheduleRepository
	clinicRepo   repositories.ClinicRepository
	auditRepo    repositories.AuditRepository
}

func NewScheduleService(
	scheduleRepo repositories.ScheduleRepository,
	clinicRepo repositories.ClinicRepository,
	auditRepo repositories.AuditRepository,
) ScheduleService {
	return &scheduleService{
		scheduleRepo: scheduleRepo,
		clinicRepo:   clinicRepo,
		auditRepo:    auditRepo,
	}
}

// Tashkent timezone (+05:00) helper
func getTashkentLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		return time.FixedZone("Asia/Tashkent", 5*3600)
	}
	return loc
}

func parseTimeToMinutes(t string) (int, error) {
	if !timeRegex.MatchString(t) {
		return 0, ErrInvalidTimeFormat
	}
	parts := strings.Split(t, ":")
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	return h*60 + m, nil
}

func minutesToTime(mins int) string {
	h := mins / 60
	m := mins % 60
	return fmt.Sprintf("%02d:%02d", h, m)
}

// checkPermission verifies if user is an owner or admin of the clinic
func (s *scheduleService) checkPermission(ctx context.Context, clinicID, userID uuid.UUID) error {
	isAdmin, err := s.clinicRepo.IsUserClinicAdmin(ctx, clinicID, userID)
	if err != nil || !isAdmin {
		return repositories.ErrClinicAccessDenied
	}
	return nil
}

// GenerateSlotsPreview calculates slots for selected weekdays without saving
func (s *scheduleService) GenerateSlotsPreview(
	ctx context.Context,
	clinicID, userID, doctorIDOrCDID uuid.UUID,
	req *dto.GenerateScheduleRequest,
) (*dto.GenerateScheduleResponse, error) {
	if err := s.checkPermission(ctx, clinicID, userID); err != nil {
		return nil, err
	}

	startMins, err := parseTimeToMinutes(req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("boshlanish vaqti xatosi: %w", err)
	}
	endMins, err := parseTimeToMinutes(req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("tugash vaqti xatosi: %w", err)
	}
	if startMins >= endMins {
		return nil, ErrStartMustBeBeforeEnd
	}

	if req.SlotDurationMinutes < 5 || req.SlotDurationMinutes > 240 {
		return nil, ErrInvalidSlotDuration
	}

	// Validate breaks
	type parsedBreak struct {
		start int
		end   int
	}
	var breaks []parsedBreak
	for i, b := range req.Breaks {
		bStart, err := parseTimeToMinutes(b.StartTime)
		if err != nil {
			return nil, fmt.Errorf("tanaffus #%d boshlanish vaqti xatosi: %w", i+1, err)
		}
		bEnd, err := parseTimeToMinutes(b.EndTime)
		if err != nil {
			return nil, fmt.Errorf("tanaffus #%d tugash vaqti xatosi: %w", i+1, err)
		}
		if bStart >= bEnd {
			return nil, fmt.Errorf("tanaffus #%d: boshlanish vaqti tugashidan kichik bo‘lishi shart", i+1)
		}
		if bStart < startMins || bEnd > endMins {
			return nil, ErrBreakOutsideHours
		}
		breaks = append(breaks, parsedBreak{start: bStart, end: bEnd})
	}

	// Check overlapping breaks
	sort.Slice(breaks, func(i, j int) bool {
		return breaks[i].start < breaks[j].start
	})
	for i := 1; i < len(breaks); i++ {
		if breaks[i].start < breaks[i-1].end {
			return nil, ErrOverlappingBreaks
		}
	}

	// Sort & validate weekdays
	weekdayMap := make(map[int]bool)
	var sortedWeekdays []int
	for _, wd := range req.Weekdays {
		if wd < 1 || wd > 7 {
			return nil, ErrInvalidWeekday
		}
		if !weekdayMap[wd] {
			weekdayMap[wd] = true
			sortedWeekdays = append(sortedWeekdays, wd)
		}
	}
	sort.Ints(sortedWeekdays)

	var generatedSlots []dto.GeneratedSlotItem

	// Generate interval slots
	for _, wd := range sortedWeekdays {
		curr := startMins
		for curr+req.SlotDurationMinutes <= endMins {
			slotStart := curr
			slotEnd := curr + req.SlotDurationMinutes

			// Check if slot overlaps with any break
			overlapWithBreak := false
			for _, brk := range breaks {
				if slotStart < brk.end && brk.start < slotEnd {
					overlapWithBreak = true
					break
				}
			}

			if !overlapWithBreak {
				generatedSlots = append(generatedSlots, dto.GeneratedSlotItem{
					Weekday:   wd,
					StartTime: minutesToTime(slotStart),
					EndTime:   minutesToTime(slotEnd),
					Status:    "active",
				})
			}

			curr += req.SlotDurationMinutes
		}
	}

	return &dto.GenerateScheduleResponse{
		TotalSlots: len(generatedSlots),
		Slots:      generatedSlots,
	}, nil
}

// GetWeeklySchedule retrieves persisted weekly slots for a doctor
func (s *scheduleService) GetWeeklySchedule(ctx context.Context, clinicID, doctorIDOrCDID uuid.UUID) (*dto.DoctorWeeklyScheduleDTO, error) {
	cd, err := s.scheduleRepo.FindClinicDoctor(ctx, clinicID, doctorIDOrCDID)
	if err != nil {
		return nil, err
	}

	slots, err := s.scheduleRepo.GetWeeklySlots(ctx, cd.ID)
	if err != nil {
		return nil, err
	}

	var dtoList []dto.DoctorWeeklySlotDTO
	for _, slot := range slots {
		dtoList = append(dtoList, dto.DoctorWeeklySlotDTO{
			ID:             slot.ID,
			ClinicDoctorID: slot.ClinicDoctorID,
			Weekday:        slot.Weekday,
			StartTime:      slot.StartTime,
			EndTime:        slot.EndTime,
			Status:         string(slot.Status),
			CreatedAt:      slot.CreatedAt,
			UpdatedAt:      slot.UpdatedAt,
		})
	}

	return &dto.DoctorWeeklyScheduleDTO{
		ClinicDoctorID: cd.ID,
		DoctorID:       cd.DoctorID,
		ClinicID:       cd.ClinicID,
		Slots:          dtoList,
	}, nil
}

// SaveWeeklySchedule atomically saves the full weekly schedule after thorough validation
func (s *scheduleService) SaveWeeklySchedule(
	ctx context.Context,
	clinicID, userID, doctorIDOrCDID uuid.UUID,
	req *dto.SaveWeeklyScheduleRequest,
) (*dto.DoctorWeeklyScheduleDTO, error) {
	if err := s.checkPermission(ctx, clinicID, userID); err != nil {
		return nil, err
	}

	cd, err := s.scheduleRepo.FindClinicDoctor(ctx, clinicID, doctorIDOrCDID)
	if err != nil {
		return nil, err
	}

	type parsedSlot struct {
		weekday   int
		startMins int
		endMins   int
		item      dto.WeeklySlotItemRequest
	}

	// 1. Validate individual slot times and ranges
	var parsedList []parsedSlot
	weekdayGroups := make(map[int][]parsedSlot)

	for i, slotReq := range req.Slots {
		if slotReq.Weekday < 1 || slotReq.Weekday > 7 {
			return nil, fmt.Errorf("slot #%d: %w", i+1, ErrInvalidWeekday)
		}
		sMins, err := parseTimeToMinutes(slotReq.StartTime)
		if err != nil {
			return nil, fmt.Errorf("slot #%d: boshlanish vaqti xatosi: %w", i+1, err)
		}
		eMins, err := parseTimeToMinutes(slotReq.EndTime)
		if err != nil {
			return nil, fmt.Errorf("slot #%d: tugash vaqti xatosi: %w", i+1, err)
		}
		if sMins >= eMins {
			return nil, fmt.Errorf("slot #%d: boshlanish vaqti (%s) tugash vaqtidan (%s) kichik bo‘lishi shart", i+1, slotReq.StartTime, slotReq.EndTime)
		}

		ps := parsedSlot{
			weekday:   slotReq.Weekday,
			startMins: sMins,
			endMins:   eMins,
			item:      slotReq,
		}
		parsedList = append(parsedList, ps)
		weekdayGroups[slotReq.Weekday] = append(weekdayGroups[slotReq.Weekday], ps)
	}

	// 2. Validate internal overlaps within the same weekday
	for wd, group := range weekdayGroups {
		sort.Slice(group, func(i, j int) bool {
			return group[i].startMins < group[j].startMins
		})
		for i := 1; i < len(group); i++ {
			if group[i].startMins < group[i-1].endMins {
				return nil, fmt.Errorf("kun %d da slotlar to‘qnashuvi: %s-%s va %s-%s", wd,
					group[i-1].item.StartTime, group[i-1].item.EndTime,
					group[i].item.StartTime, group[i].item.EndTime)
			}
		}
	}

	// 3. Validate cross-clinic conflicts for the doctor
	for _, ps := range parsedList {
		conflictSlot, otherClinic, err := s.scheduleRepo.CheckDoctorCrossClinicConflict(
			ctx,
			cd.DoctorID,
			cd.ID,
			ps.weekday,
			ps.item.StartTime,
			ps.item.EndTime,
		)
		if err != nil {
			if errors.Is(err, repositories.ErrCrossClinicConflict) {
				clinicName := "Boshqa klinika"
				if otherClinic != nil && otherClinic.Name != "" {
					clinicName = otherClinic.Name
				}
				weekdayName := getWeekdayName(ps.weekday)
				return nil, fmt.Errorf("%w: %s-%s (%s, %s)",
					repositories.ErrCrossClinicConflict,
					conflictSlot.StartTime, conflictSlot.EndTime, weekdayName, clinicName)
			}
			return nil, fmt.Errorf("boshqa klinikalar jadvalini tekshirishda xatolik: %w", err)
		}
	}

	// 4. Convert to models and replace atomically
	var modelSlots []models.DoctorWeeklySlot
	for _, ps := range parsedList {
		status := models.SlotStatusActive
		if ps.item.Status == string(models.SlotStatusInactive) {
			status = models.SlotStatusInactive
		}
		modelSlots = append(modelSlots, models.DoctorWeeklySlot{
			ClinicDoctorID: cd.ID,
			Weekday:        ps.weekday,
			StartTime:      ps.item.StartTime,
			EndTime:        ps.item.EndTime,
			Status:         status,
		})
	}

	if err := s.scheduleRepo.ReplaceWeeklySlots(ctx, cd.ID, modelSlots); err != nil {
		return nil, err
	}

	// Audit log
	_ = s.auditRepo.CreateAuditLog(ctx, &models.AuditLog{
		UserID:     &userID,
		Action:     "UPDATE_DOCTOR_WEEKLY_SCHEDULE",
		EntityType: "clinic_doctors",
		EntityID:   &cd.ID,
		Details:    fmt.Sprintf(`{"slots_count": %d}`, len(modelSlots)),
	})

	return s.GetWeeklySchedule(ctx, clinicID, cd.ID)
}

// GetDateExceptions retrieves date overrides for a doctor
func (s *scheduleService) GetDateExceptions(ctx context.Context, clinicID, doctorIDOrCDID uuid.UUID, date, startDate, endDate string) ([]dto.DoctorDateOverrideDTO, error) {
	cd, err := s.scheduleRepo.FindClinicDoctor(ctx, clinicID, doctorIDOrCDID)
	if err != nil {
		return nil, err
	}

	overrides, err := s.scheduleRepo.GetDateOverrides(ctx, cd.ID, date, startDate, endDate)
	if err != nil {
		return nil, err
	}

	var dtoList []dto.DoctorDateOverrideDTO
	for _, o := range overrides {
		dtoList = append(dtoList, dto.DoctorDateOverrideDTO{
			ID:             o.ID,
			ClinicDoctorID: o.ClinicDoctorID,
			Date:           o.Date,
			StartTime:      o.StartTime,
			EndTime:        o.EndTime,
			Status:         string(o.Status),
			Reason:         o.Reason,
			CreatedAt:      o.CreatedAt,
			UpdatedAt:      o.UpdatedAt,
		})
	}
	return dtoList, nil
}

// CreateDateException creates a single date exception
func (s *scheduleService) CreateDateException(
	ctx context.Context,
	clinicID, userID, doctorIDOrCDID uuid.UUID,
	req *dto.CreateDateOverrideRequest,
) (*dto.DoctorDateOverrideDTO, error) {
	if err := s.checkPermission(ctx, clinicID, userID); err != nil {
		return nil, err
	}

	cd, err := s.scheduleRepo.FindClinicDoctor(ctx, clinicID, doctorIDOrCDID)
	if err != nil {
		return nil, err
	}

	if !dateRegex.MatchString(req.Date) {
		return nil, ErrInvalidDateFormat
	}
	if _, err := time.Parse("2006-01-02", req.Date); err != nil {
		return nil, ErrInvalidDateFormat
	}

	status := models.OverrideStatus(req.Status)
	var startTime, endTime *string

	if status == models.OverrideStatusDayOff {
		// whole day off
		startTime = nil
		endTime = nil
	} else {
		if req.StartTime == nil || req.EndTime == nil {
			return nil, errors.New("start_time va end_time ko‘rsatilishi shart")
		}
		sMins, err := parseTimeToMinutes(*req.StartTime)
		if err != nil {
			return nil, err
		}
		eMins, err := parseTimeToMinutes(*req.EndTime)
		if err != nil {
			return nil, err
		}
		if sMins >= eMins {
			return nil, ErrStartMustBeBeforeEnd
		}
		startTime = req.StartTime
		endTime = req.EndTime
	}

	override := &models.DoctorDateSlotOverride{
		ClinicDoctorID: cd.ID,
		Date:           req.Date,
		StartTime:      startTime,
		EndTime:        endTime,
		Status:         status,
		Reason:         req.Reason,
	}

	if err := s.scheduleRepo.CreateDateOverride(ctx, override); err != nil {
		return nil, err
	}

	_ = s.auditRepo.CreateAuditLog(ctx, &models.AuditLog{
		UserID:     &userID,
		Action:     "CREATE_DOCTOR_SCHEDULE_EXCEPTION",
		EntityType: "doctor_date_slot_overrides",
		EntityID:   &override.ID,
		Details:    fmt.Sprintf(`{"date": "%s", "status": "%s"}`, req.Date, req.Status),
	})

	return &dto.DoctorDateOverrideDTO{
		ID:             override.ID,
		ClinicDoctorID: override.ClinicDoctorID,
		Date:           override.Date,
		StartTime:      override.StartTime,
		EndTime:        override.EndTime,
		Status:         string(override.Status),
		Reason:         override.Reason,
		CreatedAt:      override.CreatedAt,
		UpdatedAt:      override.UpdatedAt,
	}, nil
}

// UpdateDateException updates an existing exception
func (s *scheduleService) UpdateDateException(
	ctx context.Context,
	clinicID, userID, doctorIDOrCDID, exceptionID uuid.UUID,
	req *dto.UpdateDateOverrideRequest,
) (*dto.DoctorDateOverrideDTO, error) {
	if err := s.checkPermission(ctx, clinicID, userID); err != nil {
		return nil, err
	}

	cd, err := s.scheduleRepo.FindClinicDoctor(ctx, clinicID, doctorIDOrCDID)
	if err != nil {
		return nil, err
	}

	override, err := s.scheduleRepo.GetDateOverrideByID(ctx, exceptionID)
	if err != nil {
		return nil, err
	}
	if override.ClinicDoctorID != cd.ID {
		return nil, repositories.ErrClinicAccessDenied
	}

	if req.Date != nil {
		if !dateRegex.MatchString(*req.Date) {
			return nil, ErrInvalidDateFormat
		}
		override.Date = *req.Date
	}
	if req.Status != nil {
		override.Status = models.OverrideStatus(*req.Status)
	}
	if req.Reason != nil {
		override.Reason = *req.Reason
	}

	if override.Status == models.OverrideStatusDayOff {
		override.StartTime = nil
		override.EndTime = nil
	} else {
		if req.StartTime != nil {
			override.StartTime = req.StartTime
		}
		if req.EndTime != nil {
			override.EndTime = req.EndTime
		}
		if override.StartTime == nil || override.EndTime == nil {
			return nil, errors.New("start_time va end_time ko‘rsatilishi shart")
		}
		sMins, err := parseTimeToMinutes(*override.StartTime)
		if err != nil {
			return nil, err
		}
		eMins, err := parseTimeToMinutes(*override.EndTime)
		if err != nil {
			return nil, err
		}
		if sMins >= eMins {
			return nil, ErrStartMustBeBeforeEnd
		}
	}

	if err := s.scheduleRepo.UpdateDateOverride(ctx, override); err != nil {
		return nil, err
	}

	return &dto.DoctorDateOverrideDTO{
		ID:             override.ID,
		ClinicDoctorID: override.ClinicDoctorID,
		Date:           override.Date,
		StartTime:      override.StartTime,
		EndTime:        override.EndTime,
		Status:         string(override.Status),
		Reason:         override.Reason,
		CreatedAt:      override.CreatedAt,
		UpdatedAt:      override.UpdatedAt,
	}, nil
}

// DeleteDateException deletes an exception
func (s *scheduleService) DeleteDateException(ctx context.Context, clinicID, userID, doctorIDOrCDID, exceptionID uuid.UUID) error {
	if err := s.checkPermission(ctx, clinicID, userID); err != nil {
		return err
	}

	cd, err := s.scheduleRepo.FindClinicDoctor(ctx, clinicID, doctorIDOrCDID)
	if err != nil {
		return err
	}

	override, err := s.scheduleRepo.GetDateOverrideByID(ctx, exceptionID)
	if err != nil {
		return err
	}
	if override.ClinicDoctorID != cd.ID {
		return repositories.ErrClinicAccessDenied
	}

	return s.scheduleRepo.DeleteDateOverride(ctx, exceptionID)
}

// GetPublicAvailability computes available slots for a specific date in Asia/Tashkent timezone
func (s *scheduleService) GetPublicAvailability(
	ctx context.Context,
	clinicID, doctorIDOrCDID uuid.UUID,
	dateStr string,
) (*dto.PublicAvailabilityResponse, error) {
	if !dateRegex.MatchString(dateStr) {
		return nil, ErrInvalidDateFormat
	}
	targetDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, ErrInvalidDateFormat
	}

	// Calculate weekday in 1=Monday .. 7=Sunday
	wd := int(targetDate.Weekday())
	if wd == 0 {
		wd = 7
	}

	clinic, cd, weeklySlots, dateOverrides, err := s.scheduleRepo.GetAvailabilityData(ctx, clinicID, doctorIDOrCDID, dateStr, wd)
	if err != nil {
		return nil, err
	}

	// Only approved clinics and active doctors
	if clinic.Status != models.ClinicStatusApproved {
		return nil, ErrClinicNotApproved
	}
	if !cd.IsActive {
		return nil, ErrDoctorInactive
	}

	tashkentLoc := getTashkentLocation()
	nowTashkent := time.Now().In(tashkentLoc)
	todayTashkentStr := nowTashkent.Format("2006-01-02")
	currentTimeStr := nowTashkent.Format("15:04")

	// 1. If date is in the past, return empty slots
	if dateStr < todayTashkentStr {
		return &dto.PublicAvailabilityResponse{
			Date:           dateStr,
			Timezone:       "Asia/Tashkent",
			ClinicID:       clinic.ID,
			DoctorID:       cd.DoctorID,
			ClinicDoctorID: cd.ID,
			Slots:          []dto.AvailabilitySlotDTO{},
		}, nil
	}

	// 2. Check for day_off override
	for _, o := range dateOverrides {
		if o.Status == models.OverrideStatusDayOff {
			return &dto.PublicAvailabilityResponse{
				Date:           dateStr,
				Timezone:       "Asia/Tashkent",
				ClinicID:       clinic.ID,
				DoctorID:       cd.DoctorID,
				ClinicDoctorID: cd.ID,
				Slots:          []dto.AvailabilitySlotDTO{},
			}, nil
		}
	}

	// 3. Process base weekly slots
	type slotEntry struct {
		start string
		end   string
	}
	var candidateSlots []slotEntry

	for _, ws := range weeklySlots {
		if ws.Status != models.SlotStatusActive {
			continue
		}

		// Check if this slot is explicitly closed by an override
		isClosed := false
		for _, o := range dateOverrides {
			if o.Status == models.OverrideStatusClosed && o.StartTime != nil && o.EndTime != nil {
				// Exact match or overlap
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

	// 4. Add extra available slots from overrides
	for _, o := range dateOverrides {
		if o.Status == models.OverrideStatusAvailable && o.StartTime != nil && o.EndTime != nil {
			candidateSlots = append(candidateSlots, slotEntry{
				start: *o.StartTime,
				end:   *o.EndTime,
			})
		}
	}

	// 5. If date is TODAY, filter out past time slots
	var finalSlots []dto.AvailabilitySlotDTO
	for _, s := range candidateSlots {
		if dateStr == todayTashkentStr {
			if s.start <= currentTimeStr {
				continue // past slot
			}
		}
		finalSlots = append(finalSlots, dto.AvailabilitySlotDTO{
			StartTime: s.start,
			EndTime:   s.end,
		})
	}

	// 6. Sort chronologically by start time
	sort.Slice(finalSlots, func(i, j int) bool {
		return finalSlots[i].StartTime < finalSlots[j].StartTime
	})

	return &dto.PublicAvailabilityResponse{
		Date:           dateStr,
		Timezone:       "Asia/Tashkent",
		ClinicID:       clinic.ID,
		DoctorID:       cd.DoctorID,
		ClinicDoctorID: cd.ID,
		Slots:          finalSlots,
	}, nil
}

func getWeekdayName(wd int) string {
	names := map[int]string{
		1: "Dushanba",
		2: "Seshanba",
		3: "Chorshanba",
		4: "Payshanba",
		5: "Juma",
		6: "Shanba",
		7: "Yakshanba",
	}
	return names[wd]
}
