package handlers

import (
	"errors"
	"net/http"
	"strings"

	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/middleware"
	"qabulgo-backend/internal/models"
	"qabulgo-backend/internal/repositories"
	"qabulgo-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AppointmentHandler struct {
	appointmentService services.AppointmentService
}

func NewAppointmentHandler(appointmentService services.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{appointmentService: appointmentService}
}

// BookAppointment godoc
// @Summary Shifokor qabuliga bron qilish (Avtomatik tasdiqlanadi)
// @Description Bemor tanlagan shifokor va slot bo‘yicha appointment yaratadi (darhol confirmed bo‘ladi)
// @Tags Appointments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Idempotency-Key header string false "Takroriy yuborishning oldini olish kaliti"
// @Param request body dto.CreateAppointmentRequest true "Bron ma’lumotlari"
// @Success 201 {object} dto.SuccessResponse{data=dto.AppointmentResponseDTO} "Qabulga muvaffaqiyatli yozildingiz"
// @Failure 400 {object} dto.ErrorResponse "Noto‘g‘ri ma’lumot yoki o‘tib ketgan vaqt"
// @Failure 401 {object} dto.ErrorResponse "Avtorizatsiya talab qilinadi"
// @Failure 403 {object} dto.ErrorResponse "Faqat bemorlar bron qila oladi"
// @Failure 404 {object} dto.ErrorResponse "Klinika yoki shifokor topilmadi"
// @Failure 409 {object} dto.ErrorResponse "Double booking to‘qnashuvi"
// @Router /api/v1/appointments [post]
func (h *AppointmentHandler) BookAppointment(c *gin.Context) {
	patientID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	role, ok := middleware.GetCurrentUserRole(c)
	if !ok || role != models.RolePatient {
		c.JSON(http.StatusForbidden, dto.NewErrorResponse("Faqat bemorlar qabulga yozilishi mumkin", nil))
		return
	}

	var req dto.CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Kiritilgan ma’lumotlar formati noto‘g‘ri", map[string]string{"body": err.Error()}))
		return
	}

	idempotencyKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))

	appt, err := h.appointmentService.BookAppointment(c.Request.Context(), patientID, idempotencyKey, &req)
	if err != nil {
		if errors.Is(err, repositories.ErrDoubleBookingConflict) ||
			errors.Is(err, repositories.ErrIdempotencyConflict) ||
			errors.Is(err, services.ErrScheduleBookingConflict) {
			c.JSON(http.StatusConflict, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, services.ErrPatientOnly) || errors.Is(err, repositories.ErrClinicAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, repositories.ErrClinicNotFound) || errors.Is(err, repositories.ErrClinicDoctorNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusCreated, dto.NewSuccessResponse("Qabulga muvaffaqiyatli yozildingiz", appt))
}

// GetMyAppointments godoc
// @Summary Bemorning o‘z qabullari ro‘yxatini olish
// @Description Bemor o‘ziga tegishli barcha qabullarni ko‘radi
// @Tags Appointments
// @Produce json
// @Security BearerAuth
// @Param timeframe query string false "upcoming, past, cancelled"
// @Param status query string false "confirmed, cancelled, completed, no_show"
// @Param page query int false "Sahifa raqami"
// @Param limit query int false "Sahifadagi yozuvlar soni"
// @Success 200 {object} dto.SuccessResponse{data=dto.AppointmentListResponseDTO} "Qabullar ro‘yxati"
// @Router /api/v1/appointments/my [get]
func (h *AppointmentHandler) GetMyAppointments(c *gin.Context) {
	patientID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	var filter dto.PatientAppointmentFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Filtr parametrlari noto‘g‘ri", nil))
		return
	}

	res, err := h.appointmentService.GetPatientAppointments(c.Request.Context(), patientID, &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Bemor qabullari ro‘yxati", res))
}

// GetMyAppointmentByID godoc
// @Summary Bemorning bitta qabul ma’lumotini olish
// @Tags Appointments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Appointment UUID"
// @Success 200 {object} dto.SuccessResponse{data=dto.AppointmentResponseDTO} "Qabul ma’lumotlari"
// @Failure 403 {object} dto.ErrorResponse "Boshqa bemorning bron ma'lumotlariga kirish taqiqlanadi"
// @Failure 404 {object} dto.ErrorResponse "Qabul topilmadi"
// @Router /api/v1/appointments/{id} [get]
func (h *AppointmentHandler) GetMyAppointmentByID(c *gin.Context) {
	patientID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	apptID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Appointment ID formati noto‘g‘ri", nil))
		return
	}

	appt, err := h.appointmentService.GetPatientAppointmentByID(c.Request.Context(), patientID, apptID)
	if err != nil {
		if errors.Is(err, repositories.ErrAppointmentNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, services.ErrUnauthorizedAppointmentAccess) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Qabul ma’lumotlari", appt))
}

// CancelMyAppointment godoc
// @Summary Bemor tomonidan qabulni bekor qilish
// @Tags Appointments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Appointment UUID"
// @Param request body dto.CancelAppointmentRequest false "Bekor qilish sababi"
// @Success 200 {object} dto.SuccessResponse{data=dto.AppointmentResponseDTO} "Qabul bekor qilindi"
// @Router /api/v1/appointments/{id}/cancel [post]
func (h *AppointmentHandler) CancelMyAppointment(c *gin.Context) {
	patientID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	apptID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Appointment ID formati noto‘g‘ri", nil))
		return
	}

	var req dto.CancelAppointmentRequest
	_ = c.ShouldBindJSON(&req)

	updated, err := h.appointmentService.CancelPatientAppointment(c.Request.Context(), patientID, apptID, &req)
	if err != nil {
		if errors.Is(err, repositories.ErrAppointmentNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, services.ErrUnauthorizedAppointmentAccess) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Qabul muvaffaqiyatli bekor qilindi", updated))
}

// GetClinicAppointments godoc
// @Summary Klinikaning barcha shifokorlari bo‘yicha qabullarni ko‘rish
// @Tags Clinic Appointments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Klinika UUID"
// @Param date query string false "Aniq sana (YYYY-MM-DD)"
// @Param date_from query string false "Sanadan boshlab"
// @Param date_to query string false "Sanagacha"
// @Param doctor_id query string false "Shifokor UUID"
// @Param status query string false "confirmed, cancelled, completed, no_show"
// @Param page query int false "Sahifa"
// @Param limit query int false "Limit"
// @Success 200 {object} dto.SuccessResponse{data=dto.AppointmentListResponseDTO} "Klinika qabullari"
// @Router /api/v1/clinics/{id}/appointments [get]
func (h *AppointmentHandler) GetClinicAppointments(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	clinicIDStr := c.Param("id")
	if clinicIDStr == "" {
		clinicIDStr = c.Param("clinicId")
	}
	clinicID, err := uuid.Parse(clinicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	var filter dto.ClinicAppointmentFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Filtr parametrlari noto‘g‘ri", nil))
		return
	}

	res, err := h.appointmentService.GetClinicAppointments(c.Request.Context(), clinicID, userID, &filter)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Klinika qabullari ro‘yxati", res))
}

// GetClinicAppointmentByID godoc
// @Summary Klinika uchun bitta qabulning to‘liq ma’lumotini olish
// @Tags Clinic Appointments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Klinika UUID"
// @Param appointmentId path string true "Appointment UUID"
// @Success 200 {object} dto.SuccessResponse{data=dto.AppointmentResponseDTO} "Qabul ma’lumotlari"
// @Router /api/v1/clinics/{id}/appointments/{appointmentId} [get]
func (h *AppointmentHandler) GetClinicAppointmentByID(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	clinicIDStr := c.Param("id")
	if clinicIDStr == "" {
		clinicIDStr = c.Param("clinicId")
	}
	clinicID, err := uuid.Parse(clinicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	apptID, err := uuid.Parse(c.Param("appointmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Appointment ID formati noto‘g‘ri", nil))
		return
	}

	appt, err := h.appointmentService.GetClinicAppointmentByID(c.Request.Context(), clinicID, userID, apptID)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicAccessDenied) || errors.Is(err, services.ErrUnauthorizedAppointmentAccess) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, repositories.ErrAppointmentNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Qabul ma’lumotlari", appt))
}

// CancelClinicAppointment godoc
// @Summary Klinika xodimi tomonidan qabulni bekor qilish
// @Tags Clinic Appointments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Klinika UUID"
// @Param appointmentId path string true "Appointment UUID"
// @Param request body dto.CancelAppointmentRequest false "Bekor qilish sababi"
// @Success 200 {object} dto.SuccessResponse{data=dto.AppointmentResponseDTO} "Qabul bekor qilindi"
// @Router /api/v1/clinics/{id}/appointments/{appointmentId}/cancel [post]
func (h *AppointmentHandler) CancelClinicAppointment(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	clinicIDStr := c.Param("id")
	if clinicIDStr == "" {
		clinicIDStr = c.Param("clinicId")
	}
	clinicID, err := uuid.Parse(clinicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	apptID, err := uuid.Parse(c.Param("appointmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Appointment ID formati noto‘g‘ri", nil))
		return
	}

	var req dto.CancelAppointmentRequest
	_ = c.ShouldBindJSON(&req)

	updated, err := h.appointmentService.CancelClinicAppointment(c.Request.Context(), clinicID, userID, apptID, &req)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicAccessDenied) || errors.Is(err, services.ErrUnauthorizedAppointmentAccess) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, repositories.ErrAppointmentNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Qabul muvaffaqiyatli bekor qilindi", updated))
}

// CompleteClinicAppointment godoc
// @Summary Qabul yakunlanganini belgilash (faqat vaqti kelgandan so‘ng)
// @Tags Clinic Appointments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Klinika UUID"
// @Param appointmentId path string true "Appointment UUID"
// @Success 200 {object} dto.SuccessResponse{data=dto.AppointmentResponseDTO} "Qabul yakunlandi"
// @Router /api/v1/clinics/{id}/appointments/{appointmentId}/complete [post]
func (h *AppointmentHandler) CompleteClinicAppointment(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	clinicIDStr := c.Param("id")
	if clinicIDStr == "" {
		clinicIDStr = c.Param("clinicId")
	}
	clinicID, err := uuid.Parse(clinicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	apptID, err := uuid.Parse(c.Param("appointmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Appointment ID formati noto‘g‘ri", nil))
		return
	}

	updated, err := h.appointmentService.CompleteClinicAppointment(c.Request.Context(), clinicID, userID, apptID)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicAccessDenied) || errors.Is(err, services.ErrUnauthorizedAppointmentAccess) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, repositories.ErrAppointmentNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Qabul muvaffaqiyatli yakunlandi", updated))
}

// NoShowClinicAppointment godoc
// @Summary Bemor qabulga kelmaganini belgilash (faqat vaqti kelgandan so‘ng)
// @Tags Clinic Appointments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Klinika UUID"
// @Param appointmentId path string true "Appointment UUID"
// @Success 200 {object} dto.SuccessResponse{data=dto.AppointmentResponseDTO} "Bemor kelmagan deb qayd etildi"
// @Router /api/v1/clinics/{id}/appointments/{appointmentId}/no-show [post]
func (h *AppointmentHandler) NoShowClinicAppointment(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	clinicIDStr := c.Param("id")
	if clinicIDStr == "" {
		clinicIDStr = c.Param("clinicId")
	}
	clinicID, err := uuid.Parse(clinicIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	apptID, err := uuid.Parse(c.Param("appointmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Appointment ID formati noto‘g‘ri", nil))
		return
	}

	updated, err := h.appointmentService.NoShowClinicAppointment(c.Request.Context(), clinicID, userID, apptID)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicAccessDenied) || errors.Is(err, services.ErrUnauthorizedAppointmentAccess) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, repositories.ErrAppointmentNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Bemor kelmagan deb qayd etildi", updated))
}

