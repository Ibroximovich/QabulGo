package handlers

import (
	"errors"
	"net/http"

	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/middleware"
	"qabulgo-backend/internal/repositories"
	"qabulgo-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ScheduleHandler struct {
	scheduleService services.ScheduleService
}

func NewScheduleHandler(scheduleService services.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{scheduleService: scheduleService}
}

// GenerateSchedulePreview godoc
// @Summary Shifokor qabul vaqtlarini avtomatik hisoblash (Preview)
// @Description Parametrlar asosida slotlar ro‘yxatini hisoblaydi, lekin bazaga saqlamaydi
// @Tags Doctor Schedule
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Klinika UUID"
// @Param doctorId path string true "Shifokor yoki ClinicDoctor UUID"
// @Param request body dto.GenerateScheduleRequest true "Parametrlar"
// @Success 200 {object} dto.SuccessResponse{data=dto.GenerateScheduleResponse} "Hisoblangan slotlar"
// @Failure 400 {object} dto.ErrorResponse "Noto‘g‘ri ma’lumotlar"
// @Failure 403 {object} dto.ErrorResponse "Ruxsat berilmagan"
// @Router /api/v1/clinics/{id}/doctors/{doctorId}/schedule/generate [post]
func (h *ScheduleHandler) GenerateSchedulePreview(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	doctorID, err := uuid.Parse(c.Param("doctorId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Shifokor ID formati noto‘g‘ri", nil))
		return
	}

	var req dto.GenerateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Kiritilgan ma’lumotlar formati noto‘g‘ri", map[string]string{"body": err.Error()}))
		return
	}

	preview, err := h.scheduleService.GenerateSlotsPreview(c.Request.Context(), clinicID, userID, doctorID, &req)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Qabul vaqtlari avtomatik hisoblandi (Preview)", preview))
}

// GetWeeklySchedule godoc
// @Summary Shifokorning haftalik jadvalini olish
// @Description Haftaning har bir kuni uchun saqlangan qabul vaqtlari
// @Tags Doctor Schedule
// @Produce json
// @Security BearerAuth
// @Param id path string true "Klinika UUID"
// @Param doctorId path string true "Shifokor yoki ClinicDoctor UUID"
// @Success 200 {object} dto.SuccessResponse{data=dto.DoctorWeeklyScheduleDTO} "Haftalik jadval"
// @Failure 404 {object} dto.ErrorResponse "Shifokor topilmadi"
// @Router /api/v1/clinics/{id}/doctors/{doctorId}/schedule [get]
func (h *ScheduleHandler) GetWeeklySchedule(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	doctorID, err := uuid.Parse(c.Param("doctorId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Shifokor ID formati noto‘g‘ri", nil))
		return
	}

	schedule, err := h.scheduleService.GetWeeklySchedule(c.Request.Context(), clinicID, doctorID)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicDoctorNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, services.ErrScheduleBookingConflict) {
			c.JSON(http.StatusConflict, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Jadvalni yuklashda xatolik", nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Haftalik qabul jadvali", schedule))
}

// SaveWeeklySchedule godoc
// @Summary Shifokorning haftalik jadvalini saqlash yoki yangilash
// @Description Avtomatik yoki qo‘lda kiritilgan slotlarni atomik tarzda saqlaydi
// @Tags Doctor Schedule
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Klinika UUID"
// @Param doctorId path string true "Shifokor yoki ClinicDoctor UUID"
// @Param request body dto.SaveWeeklyScheduleRequest true "Haftalik slotlar ro‘yxati"
// @Success 200 {object} dto.SuccessResponse{data=dto.DoctorWeeklyScheduleDTO} "Jadval muvaffaqiyatli saqlandi"
// @Failure 400 {object} dto.ErrorResponse "To‘qnashuv yoki validatsiya xatoligi"
// @Failure 403 {object} dto.ErrorResponse "Ruxsat berilmagan"
// @Router /api/v1/clinics/{id}/doctors/{doctorId}/schedule [put]
func (h *ScheduleHandler) SaveWeeklySchedule(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	doctorID, err := uuid.Parse(c.Param("doctorId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Shifokor ID formati noto‘g‘ri", nil))
		return
	}

	var req dto.SaveWeeklyScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Kiritilgan ma’lumotlar formati noto‘g‘ri", map[string]string{"body": err.Error()}))
		return
	}

	updated, err := h.scheduleService.SaveWeeklySchedule(c.Request.Context(), clinicID, userID, doctorID, &req)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, repositories.ErrClinicDoctorNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, services.ErrScheduleBookingConflict) {
			c.JSON(http.StatusConflict, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, repositories.ErrCrossClinicConflict) || errors.Is(err, services.ErrScheduleBookingConflict) {
			c.JSON(http.StatusConflict, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Haftalik qabul jadvali muvaffaqiyatli saqlandi", updated))
}

// GetExceptions godoc
// @Summary Shifokorning istisno sanalari va override'larini olish
// @Description Dam olish kunlari, yopilgan slotlar yoki qo‘shilgan maxsus vaqtlar
// @Tags Doctor Schedule
// @Produce json
// @Security BearerAuth
// @Param id path string true "Klinika UUID"
// @Param doctorId path string true "Shifokor yoki ClinicDoctor UUID"
// @Param date query string false "Aniq sana (YYYY-MM-DD)"
// @Param start_date query string false "Oraliq boshi (YYYY-MM-DD)"
// @Param end_date query string false "Oraliq oxiri (YYYY-MM-DD)"
// @Success 200 {object} dto.SuccessResponse{data=[]dto.DoctorDateOverrideDTO} "Istisnolar ro‘yxati"
// @Router /api/v1/clinics/{id}/doctors/{doctorId}/schedule/exceptions [get]
func (h *ScheduleHandler) GetExceptions(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	doctorID, err := uuid.Parse(c.Param("doctorId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Shifokor ID formati noto‘g‘ri", nil))
		return
	}

	date := c.Query("date")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	exceptions, err := h.scheduleService.GetDateExceptions(c.Request.Context(), clinicID, doctorID, date, startDate, endDate)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicDoctorNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, services.ErrScheduleBookingConflict) {
			c.JSON(http.StatusConflict, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Istisnolarni yuklashda xatolik", nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Jadval istisnolari ro‘yxati", exceptions))
}

// CreateException godoc
// @Summary Jadvalga yangi istisno sana yoki maxsus slot qo‘shish
// @Description Butun kunni yopish (day_off), bitta vaqtni yopish (closed) yoki yangi vaqt qo‘shish (available)
// @Tags Doctor Schedule
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Klinika UUID"
// @Param doctorId path string true "Shifokor yoki ClinicDoctor UUID"
// @Param request body dto.CreateDateOverrideRequest true "Istisno ma’lumotlari"
// @Success 201 {object} dto.SuccessResponse{data=dto.DoctorDateOverrideDTO} "Istisno yaratildi"
// @Failure 400 {object} dto.ErrorResponse "Validatsiya xatoligi"
// @Router /api/v1/clinics/{id}/doctors/{doctorId}/schedule/exceptions [post]
func (h *ScheduleHandler) CreateException(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	doctorID, err := uuid.Parse(c.Param("doctorId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Shifokor ID formati noto‘g‘ri", nil))
		return
	}

	var req dto.CreateDateOverrideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Kiritilgan ma’lumotlar formati noto‘g‘ri", map[string]string{"body": err.Error()}))
		return
	}

	res, err := h.scheduleService.CreateDateException(c.Request.Context(), clinicID, userID, doctorID, &req)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, repositories.ErrClinicDoctorNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, services.ErrScheduleBookingConflict) {
			c.JSON(http.StatusConflict, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusCreated, dto.NewSuccessResponse("Jadval istisnosi muvaffaqiyatli saqlandi", res))
}

// UpdateException godoc
// @Summary Mavjud istisno sanani tahrirlash
// @Tags Doctor Schedule
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Klinika UUID"
// @Param doctorId path string true "Shifokor yoki ClinicDoctor UUID"
// @Param exceptionId path string true "Istisno UUID"
// @Param request body dto.UpdateDateOverrideRequest true "O‘zgartirishlar"
// @Success 200 {object} dto.SuccessResponse{data=dto.DoctorDateOverrideDTO} "Istisno yangilandi"
// @Router /api/v1/clinics/{id}/doctors/{doctorId}/schedule/exceptions/{exceptionId} [patch]
func (h *ScheduleHandler) UpdateException(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	doctorID, err := uuid.Parse(c.Param("doctorId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Shifokor ID formati noto‘g‘ri", nil))
		return
	}

	exceptionID, err := uuid.Parse(c.Param("exceptionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Istisno ID formati noto‘g‘ri", nil))
		return
	}

	var req dto.UpdateDateOverrideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Kiritilgan ma’lumotlar formati noto‘g‘ri", map[string]string{"body": err.Error()}))
		return
	}

	res, err := h.scheduleService.UpdateDateException(c.Request.Context(), clinicID, userID, doctorID, exceptionID, &req)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, repositories.ErrDateOverrideNotFound) || errors.Is(err, repositories.ErrClinicDoctorNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Jadval istisnosi yangilandi", res))
}

// DeleteException godoc
// @Summary Mavjud istisno sanani o‘chirish
// @Tags Doctor Schedule
// @Produce json
// @Security BearerAuth
// @Param id path string true "Klinika UUID"
// @Param doctorId path string true "Shifokor yoki ClinicDoctor UUID"
// @Param exceptionId path string true "Istisno UUID"
// @Success 200 {object} dto.SuccessResponse "Istisno o‘chirildi"
// @Router /api/v1/clinics/{id}/doctors/{doctorId}/schedule/exceptions/{exceptionId} [delete]
func (h *ScheduleHandler) DeleteException(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	doctorID, err := uuid.Parse(c.Param("doctorId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Shifokor ID formati noto‘g‘ri", nil))
		return
	}

	exceptionID, err := uuid.Parse(c.Param("exceptionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Istisno ID formati noto‘g‘ri", nil))
		return
	}

	if err := h.scheduleService.DeleteDateException(c.Request.Context(), clinicID, userID, doctorID, exceptionID); err != nil {
		if errors.Is(err, repositories.ErrClinicAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, repositories.ErrDateOverrideNotFound) || errors.Is(err, repositories.ErrClinicDoctorNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Jadval istisnosi muvaffaqiyatli o‘chirildi", nil))
}

// GetPublicAvailability godoc
// @Summary Bemorlar uchun shifokorning ma’lum sanadagi ochiq qabul vaqtlarini olish
// @Description Faqat tasdiqlangan klinika va faol shifokor uchun, o‘tib ketgan vaqtlar chiqarilmaydi
// @Tags Public Search
// @Produce json
// @Param clinicId path string true "Klinika UUID"
// @Param doctorId path string true "Shifokor yoki ClinicDoctor UUID"
// @Param date query string true "Sana (YYYY-MM-DD)"
// @Success 200 {object} dto.SuccessResponse{data=dto.PublicAvailabilityResponse} "Ochiq qabul vaqtlari"
// @Failure 400 {object} dto.ErrorResponse "Noto‘g‘ri sana formati"
// @Failure 404 {object} dto.ErrorResponse "Topilmadi"
// @Router /api/v1/public/clinics/{clinicId}/doctors/{doctorId}/availability [get]
func (h *ScheduleHandler) GetPublicAvailability(c *gin.Context) {
	clinicIDStr := c.Param("clinicId")
	if clinicIDStr == "" {
		clinicIDStr = c.Param("id")
	}
	clinicID, err := uuid.Parse(clinicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	doctorID, err := uuid.Parse(c.Param("doctorId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Shifokor ID formati noto‘g‘ri", nil))
		return
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("date so‘rov parametri majburiy (YYYY-MM-DD)", nil))
		return
	}

	res, err := h.scheduleService.GetPublicAvailability(c.Request.Context(), clinicID, doctorID, dateStr)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicNotFound) || errors.Is(err, repositories.ErrClinicDoctorNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, services.ErrClinicNotApproved) || errors.Is(err, services.ErrDoctorInactive) {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Shifokorning ochiq qabul vaqtlari", res))
}

