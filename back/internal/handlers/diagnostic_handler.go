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

type DiagnosticHandler struct {
	diagnosticService services.DiagnosticService
}

func NewDiagnosticHandler(diagnosticService services.DiagnosticService) *DiagnosticHandler {
	return &DiagnosticHandler{diagnosticService: diagnosticService}
}

// ==========================================
// PUBLIC ENDPOINTS (Mehmonlar va Bemorlar)
// ==========================================

// ListCategories godoc
// @Summary Diagnostika kategoriyalari ro‘yxati (MRT, MSKT, UZI, va h.k.)
// @Tags Diagnostics
// @Produce json
// @Success 200 {object} dto.SuccessResponse{data=dto.ListDiagnosticCategoriesResponse}
// @Router /api/v1/diagnostics/categories [get]
func (h *DiagnosticHandler) ListCategories(c *gin.Context) {
	res, err := h.diagnosticService.ListCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Kategoriyalarni yuklashda xatolik yuz berdi", nil))
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResponse("Diagnostika kategoriyalari ro‘yxati", res))
}

// ListServices godoc
// @Summary Diagnostika xizmatlarini qidirish va filtrlash
// @Tags Diagnostics
// @Produce json
// @Param clinic_id query string false "Klinika ID"
// @Param category_id query string false "Kategoriya ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.ListDiagnosticServicesResponse}
// @Router /api/v1/diagnostics/services [get]
func (h *DiagnosticHandler) ListServices(c *gin.Context) {
	clinicIDStr := c.Query("clinic_id")
	categoryIDStr := c.Query("category_id")

	var clinicID *uuid.UUID
	if clinicIDStr != "" {
		parsed, err := uuid.Parse(clinicIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto'g'ri klinika ID formati", nil))
			return
		}
		clinicID = &parsed
	}

	var categoryID *uuid.UUID
	if categoryIDStr != "" {
		parsedCatID, err := uuid.Parse(categoryIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto'g'ri kategoriya ID formati", nil))
			return
		}
		categoryID = &parsedCatID
	}

	res, err := h.diagnosticService.ListPublicServices(c.Request.Context(), clinicID, categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Diagnostika xizmatlarini yuklashda xatolik", nil))
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResponse("Diagnostika xizmatlari ro'yxati", res))
}

// GetServiceByID godoc
// @Summary Diagnostika xizmati tafsilotlari (narxi va tayyorgarlik qoidalari bilan)
// @Tags Diagnostics
// @Produce json
// @Param serviceId path string true "Xizmat ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.DiagnosticServiceDTO}
// @Router /api/v1/diagnostics/services/{serviceId} [get]
func (h *DiagnosticHandler) GetServiceByID(c *gin.Context) {
	serviceID, err := uuid.Parse(c.Param("serviceId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri xizmat ID formati", nil))
		return
	}

	svc, err := h.diagnosticService.GetService(c.Request.Context(), serviceID)
	if err != nil {
		if errors.Is(err, repositories.ErrDiagnosticServiceNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("Diagnostika xizmati topilmadi", nil))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Xizmat ma’lumotlarini olishda xatolik", nil))
		return
	}
	c.JSON(http.StatusOK, dto.NewSuccessResponse("Diagnostika xizmati tafsilotlari", svc))
}

// GetEquipmentAvailability godoc
// @Summary Diagnostika uskunasi uchun bo‘sh slotlarni olish
// @Tags Diagnostics
// @Produce json
// @Param equipmentId path string true "Uskuna ID"
// @Param date query string true "Sana (YYYY-MM-DD)"
// @Success 200 {object} dto.SuccessResponse{data=dto.DiagnosticAvailabilityResponse}
// @Router /api/v1/diagnostics/equipments/{equipmentId}/availability [get]
func (h *DiagnosticHandler) GetEquipmentAvailability(c *gin.Context) {
	equipmentID, err := uuid.Parse(c.Param("equipmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri uskuna ID formati", nil))
		return
	}

	dateStr := strings.TrimSpace(c.Query("date"))
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("date parametri majburiy (format: YYYY-MM-DD)", nil))
		return
	}

	avail, err := h.diagnosticService.GetDiagnosticEquipmentAvailability(c.Request.Context(), equipmentID, dateStr)
	if err != nil {
		if errors.Is(err, services.ErrInvalidDateFormat) {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, repositories.ErrDiagnosticEquipmentNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("Diagnostika uskunasi topilmadi", nil))
			return
		}
		if errors.Is(err, services.ErrClinicNotApproved) || errors.Is(err, services.ErrEquipmentInactive) {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Bo‘sh vaqtlarni hisoblashda xatolik", nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Uskunaning bo‘sh vaqtlari", avail))
}

// ==========================================
// PATIENT APPOINTMENT ENDPOINTS (Bemorlar)
// ==========================================

// BookAppointment godoc
// @Summary Diagnostika tekshiruviga bron qilish
// @Tags Diagnostics Appointments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Idempotency-Key header string false "Takroriy so‘rov kaliti"
// @Param request body dto.CreateDiagnosticAppointmentRequest true "Bron ma’lumotlari"
// @Success 201 {object} dto.SuccessResponse{data=dto.DiagnosticAppointmentResponseDTO}
// @Router /api/v1/diagnostics/appointments [post]
func (h *DiagnosticHandler) BookAppointment(c *gin.Context) {
	patientID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	role, ok := middleware.GetCurrentUserRole(c)
	if !ok || role != models.RolePatient {
		c.JSON(http.StatusForbidden, dto.NewErrorResponse("Faqat bemorlar diagnostikaga yozilishi mumkin", nil))
		return
	}

	var req dto.CreateDiagnosticAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Kiritilgan ma’lumotlar formati noto‘g‘ri", map[string]string{"body": err.Error()}))
		return
	}

	idempotencyKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))

	appt, err := h.diagnosticService.BookDiagnosticAppointment(c.Request.Context(), patientID, idempotencyKey, &req)
	if err != nil {
		if errors.Is(err, services.ErrPreparationNotAcknowledged) {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, repositories.ErrDoubleBookingConflict) ||
			errors.Is(err, repositories.ErrIdempotencyConflict) {
			c.JSON(http.StatusConflict, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, repositories.ErrDiagnosticEquipmentNotFound) ||
			errors.Is(err, repositories.ErrDiagnosticServiceNotFound) ||
			errors.Is(err, repositories.ErrClinicNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, services.ErrSlotNotInEquipmentSchedule) ||
			errors.Is(err, services.ErrSlotDateOverrideClosed) ||
			errors.Is(err, services.ErrPastTimeBooking) ||
			errors.Is(err, services.ErrEquipmentInactive) ||
			errors.Is(err, services.ErrClinicNotApproved) {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusCreated, dto.NewSuccessResponse("Diagnostika tekshiruviga muvaffaqiyatli yozildingiz", appt))
}

// GetPatientAppointments godoc
// @Summary Bemorning o‘z diagnostika qabullari tarixi
// @Tags Diagnostics Appointments
// @Produce json
// @Security BearerAuth
// @Router /api/v1/diagnostics/appointments/my [get]
func (h *DiagnosticHandler) GetPatientAppointments(c *gin.Context) {
	patientID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	var filter dto.PatientDiagnosticAppointmentFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri filter parametrlari", nil))
		return
	}

	res, err := h.diagnosticService.GetPatientDiagnosticAppointments(c.Request.Context(), patientID, &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Qabullarni yuklashda xatolik", nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Mening diagnostika qabullarim", res))
}

// GetPatientAppointmentByID godoc
// @Summary Bemorning aniq bir diagnostika qabuli tafsiloti
// @Tags Diagnostics Appointments
// @Produce json
// @Security BearerAuth
// @Router /api/v1/diagnostics/appointments/{appointmentId} [get]
func (h *DiagnosticHandler) GetPatientAppointmentByID(c *gin.Context) {
	patientID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	appointmentID, err := uuid.Parse(c.Param("appointmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri appointment ID formati", nil))
		return
	}

	appt, err := h.diagnosticService.GetPatientDiagnosticAppointmentByID(c.Request.Context(), patientID, appointmentID)
	if err != nil {
		if errors.Is(err, repositories.ErrDiagnosticAppointmentNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("Diagnostika qabuli topilmadi", nil))
			return
		}
		if errors.Is(err, services.ErrUnauthorizedAppointmentAccess) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Qabul ma’lumotlarini olishda xatolik", nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Diagnostika qabuli tafsilotlari", appt))
}

// CancelPatientAppointment godoc
// @Summary Bemor tomonidan diagnostika qabulini bekor qilish
// @Tags Diagnostics Appointments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Router /api/v1/diagnostics/appointments/{appointmentId}/cancel [post]
func (h *DiagnosticHandler) CancelPatientAppointment(c *gin.Context) {
	patientID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	appointmentID, err := uuid.Parse(c.Param("appointmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri appointment ID formati", nil))
		return
	}

	var req dto.CancelDiagnosticAppointmentRequest
	_ = c.ShouldBindJSON(&req)

	appt, err := h.diagnosticService.CancelPatientDiagnosticAppointment(c.Request.Context(), patientID, appointmentID, &req)
	if err != nil {
		if errors.Is(err, repositories.ErrDiagnosticAppointmentNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("Diagnostika qabuli topilmadi", nil))
			return
		}
		if errors.Is(err, services.ErrUnauthorizedAppointmentAccess) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, services.ErrOnlyConfirmedCancellable) || errors.Is(err, services.ErrPastAppointmentCancellation) {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Qabulni bekor qilishda xatolik", nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Diagnostika qabuli bekor qilindi", appt))
}

// ==========================================
// CLINIC MANAGEMENT ENDPOINTS (Klinikalar)
// ==========================================

// CreateEquipment godoc
// @Summary Klinika uchun yangi diagnostika uskunasi qo‘shish
// @Tags Clinic Diagnostics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Router /api/v1/clinics/{id}/diagnostic-equipments [post]
func (h *DiagnosticHandler) CreateEquipment(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri klinika ID formati", nil))
		return
	}

	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	var req dto.CreateDiagnosticEquipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Noto‘g‘ri ma’lumot formati", map[string]string{"body": err.Error()}))
		return
	}

	eq, err := h.diagnosticService.CreateEquipment(c.Request.Context(), clinicID, userID, &req)
	if err != nil {
		if errors.Is(err, repositories.ErrDiagnosticCategoryNotFound) {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Kategoriya topilmadi", nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusCreated, dto.NewSuccessResponse("Uskuna muvaffaqiyatli qo‘shildi", eq))
}

// ListEquipments godoc
// @Summary Klinikaning diagnostika uskunalari ro‘yxati
// @Tags Clinic Diagnostics
// @Produce json
// @Security BearerAuth
// @Router /api/v1/clinics/{id}/diagnostic-equipments [get]
func (h *DiagnosticHandler) ListEquipments(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri klinika ID formati", nil))
		return
	}

	res, err := h.diagnosticService.ListEquipments(c.Request.Context(), clinicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Uskunalar ro‘yxatini olishda xatolik", nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Diagnostika uskunalari", res))
}

// GetEquipment godoc
// @Summary Aniq bir uskuna tafsiloti
// @Tags Clinic Diagnostics
// @Produce json
// @Security BearerAuth
// @Router /api/v1/clinics/{id}/diagnostic-equipments/{equipmentId} [get]
func (h *DiagnosticHandler) GetEquipment(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri klinika ID formati", nil))
		return
	}

	equipmentID, err := uuid.Parse(c.Param("equipmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri uskuna ID formati", nil))
		return
	}

	eq, err := h.diagnosticService.GetEquipment(c.Request.Context(), clinicID, equipmentID)
	if err != nil {
		if errors.Is(err, repositories.ErrDiagnosticEquipmentNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("Uskuna topilmadi", nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Uskuna tafsiloti", eq))
}

// UpdateEquipment godoc
// @Summary Uskuna ma’lumotlarini yangilash
// @Tags Clinic Diagnostics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Router /api/v1/clinics/{id}/diagnostic-equipments/{equipmentId} [patch]
func (h *DiagnosticHandler) UpdateEquipment(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri klinika ID formati", nil))
		return
	}

	equipmentID, err := uuid.Parse(c.Param("equipmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri uskuna ID formati", nil))
		return
	}

	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	var req dto.UpdateDiagnosticEquipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Noto‘g‘ri ma’lumot formati", map[string]string{"body": err.Error()}))
		return
	}

	eq, err := h.diagnosticService.UpdateEquipment(c.Request.Context(), clinicID, userID, equipmentID, &req)
	if err != nil {
		if errors.Is(err, repositories.ErrDiagnosticEquipmentNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("Uskuna topilmadi", nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Uskuna muvaffaqiyatli yangilandi", eq))
}

// DeleteEquipment godoc
// @Summary Uskunani o‘chirish
// @Tags Clinic Diagnostics
// @Produce json
// @Security BearerAuth
// @Router /api/v1/clinics/{id}/diagnostic-equipments/{equipmentId} [delete]
func (h *DiagnosticHandler) DeleteEquipment(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri klinika ID formati", nil))
		return
	}

	equipmentID, err := uuid.Parse(c.Param("equipmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri uskuna ID formati", nil))
		return
	}

	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	if err := h.diagnosticService.DeleteEquipment(c.Request.Context(), clinicID, userID, equipmentID); err != nil {
		if errors.Is(err, services.ErrUnauthorizedEquipmentAccess) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse("Ushbu klinikani boshqarish huquqiga ega emassiz", nil))
			return
		}
		if errors.Is(err, repositories.ErrDiagnosticEquipmentNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("Uskuna topilmadi", nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Uskuna muvaffaqiyatli o‘chirildi", nil))
}

// --- Clinic Services ---

// CreateService godoc
// @Summary Klinika diagnostika xizmatini qo‘shish
// @Tags Clinic Diagnostics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Router /api/v1/clinics/{id}/diagnostic-services [post]
func (h *DiagnosticHandler) CreateService(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri klinika ID formati", nil))
		return
	}

	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	var req dto.CreateDiagnosticServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Noto‘g‘ri ma’lumot formati", map[string]string{"body": err.Error()}))
		return
	}

	svc, err := h.diagnosticService.CreateService(c.Request.Context(), clinicID, userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusCreated, dto.NewSuccessResponse("Diagnostika xizmati muvaffaqiyatli yaratildi", svc))
}

// ListClinicServices godoc
// @Summary Klinikaning o‘z diagnostika xizmatlari ro‘yxati
// @Tags Clinic Diagnostics
// @Produce json
// @Security BearerAuth
// @Router /api/v1/clinics/{id}/diagnostic-services [get]
func (h *DiagnosticHandler) ListClinicServices(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri klinika ID formati", nil))
		return
	}

	res, err := h.diagnosticService.ListServicesByClinic(c.Request.Context(), clinicID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Xizmatlarni yuklashda xatolik", nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Klinika diagnostika xizmatlari", res))
}

// UpdateService godoc
// @Summary Diagnostika xizmatini tahrirlash
// @Tags Clinic Diagnostics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Router /api/v1/clinics/{id}/diagnostic-services/{serviceId} [patch]
func (h *DiagnosticHandler) UpdateService(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri klinika ID formati", nil))
		return
	}

	serviceID, err := uuid.Parse(c.Param("serviceId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri xizmat ID formati", nil))
		return
	}

	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	var req dto.UpdateDiagnosticServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Noto‘g‘ri ma’lumot formati", map[string]string{"body": err.Error()}))
		return
	}

	svc, err := h.diagnosticService.UpdateService(c.Request.Context(), clinicID, userID, serviceID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Xizmat muvaffaqiyatli yangilandi", svc))
}

// DeleteService godoc
// @Summary Diagnostika xizmatini o‘chirish
// @Tags Clinic Diagnostics
// @Produce json
// @Security BearerAuth
// @Router /api/v1/clinics/{id}/diagnostic-services/{serviceId} [delete]
func (h *DiagnosticHandler) DeleteService(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri klinika ID formati", nil))
		return
	}

	serviceID, err := uuid.Parse(c.Param("serviceId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri xizmat ID formati", nil))
		return
	}

	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	if err := h.diagnosticService.DeleteService(c.Request.Context(), clinicID, userID, serviceID); err != nil {
		if errors.Is(err, services.ErrUnauthorizedEquipmentAccess) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse("Ushbu klinikani boshqarish huquqiga ega emassiz", nil))
			return
		}
		if errors.Is(err, repositories.ErrDiagnosticServiceNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("Diagnostika xizmati topilmadi", nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Xizmat muvaffaqiyatli o‘chirildi", nil))
}

// --- Equipment Schedule & Overrides ---

// GetEquipmentWeeklySlots godoc
// @Summary Uskunaning haftalik ish jadvallari
// @Tags Clinic Diagnostics Schedule
// @Produce json
// @Security BearerAuth
// @Router /api/v1/clinics/{id}/diagnostic-equipments/{equipmentId}/schedule [get]
func (h *DiagnosticHandler) GetEquipmentWeeklySlots(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri klinika ID formati", nil))
		return
	}

	equipmentID, err := uuid.Parse(c.Param("equipmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri uskuna ID formati", nil))
		return
	}

	slots, err := h.diagnosticService.GetEquipmentWeeklySlots(c.Request.Context(), clinicID, equipmentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Uskuna haftalik ish vaqtlari", slots))
}

// SaveEquipmentWeeklySlots godoc
// @Summary Uskunaning haftalik ish jadvalini saqlash
// @Tags Clinic Diagnostics Schedule
// @Accept json
// @Produce json
// @Security BearerAuth
// @Router /api/v1/clinics/{id}/diagnostic-equipments/{equipmentId}/schedule [put]
func (h *DiagnosticHandler) SaveEquipmentWeeklySlots(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri klinika ID formati", nil))
		return
	}

	equipmentID, err := uuid.Parse(c.Param("equipmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri uskuna ID formati", nil))
		return
	}

	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	var req dto.SaveDiagnosticWeeklySlotsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Noto‘g‘ri ma’lumot formati", map[string]string{"body": err.Error()}))
		return
	}

	slots, err := h.diagnosticService.SaveEquipmentWeeklySlots(c.Request.Context(), clinicID, userID, equipmentID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Haftalik ish vaqtlari saqlandi", slots))
}

// GetEquipmentDateOverrides godoc
// @Summary Uskunaning istisno va maxsus sanalari
// @Tags Clinic Diagnostics Schedule
// @Produce json
// @Security BearerAuth
// @Router /api/v1/clinics/{id}/diagnostic-equipments/{equipmentId}/schedule/exceptions [get]
func (h *DiagnosticHandler) GetEquipmentDateOverrides(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri klinika ID formati", nil))
		return
	}

	equipmentID, err := uuid.Parse(c.Param("equipmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri uskuna ID formati", nil))
		return
	}

	date := c.Query("date")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	overrides, err := h.diagnosticService.GetEquipmentDateOverrides(c.Request.Context(), clinicID, equipmentID, date, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Uskuna istisno sanalari", overrides))
}

// CreateEquipmentDateOverride godoc
// @Summary Uskunaga istisno yoki profilaktika sanasi qo‘shish
// @Tags Clinic Diagnostics Schedule
// @Accept json
// @Produce json
// @Security BearerAuth
// @Router /api/v1/clinics/{id}/diagnostic-equipments/{equipmentId}/schedule/exceptions [post]
func (h *DiagnosticHandler) CreateEquipmentDateOverride(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri klinika ID formati", nil))
		return
	}

	equipmentID, err := uuid.Parse(c.Param("equipmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri uskuna ID formati", nil))
		return
	}

	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	var req dto.CreateDiagnosticDateOverrideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Noto‘g‘ri ma’lumot formati", map[string]string{"body": err.Error()}))
		return
	}

	override, err := h.diagnosticService.CreateEquipmentDateOverride(c.Request.Context(), clinicID, userID, equipmentID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusCreated, dto.NewSuccessResponse("Istisno muvaffaqiyatli qo‘shildi", override))
}

// DeleteEquipmentDateOverride godoc
// @Summary Uskunaning istisno sanasini o‘chirish
// @Tags Clinic Diagnostics Schedule
// @Produce json
// @Security BearerAuth
// @Router /api/v1/clinics/{id}/diagnostic-equipments/{equipmentId}/schedule/exceptions/{overrideId} [delete]
func (h *DiagnosticHandler) DeleteEquipmentDateOverride(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri klinika ID formati", nil))
		return
	}

	equipmentID, err := uuid.Parse(c.Param("equipmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri uskuna ID formati", nil))
		return
	}

	overrideID, err := uuid.Parse(c.Param("overrideId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri istisno ID formati", nil))
		return
	}

	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	if err := h.diagnosticService.DeleteEquipmentDateOverride(c.Request.Context(), clinicID, userID, equipmentID, overrideID); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Istisno muvaffaqiyatli o‘chirildi", nil))
}

// --- Clinic Appointments ---

// GetClinicAppointments godoc
// @Summary Klinikaning diagnostika qabullari ro‘yxati
// @Tags Clinic Diagnostics Appointments
// @Produce json
// @Security BearerAuth
// @Router /api/v1/clinics/{id}/diagnostic-appointments [get]
func (h *DiagnosticHandler) GetClinicAppointments(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri klinika ID formati", nil))
		return
	}

	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	var filter dto.ClinicDiagnosticAppointmentFilter
	_ = c.ShouldBindQuery(&filter)

	res, err := h.diagnosticService.GetClinicDiagnosticAppointments(c.Request.Context(), clinicID, userID, &filter)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Klinika diagnostika qabullari", res))
}

// CancelClinicAppointment godoc
// @Summary Klinika tomonidan diagnostika qabulini bekor qilish
// @Tags Clinic Diagnostics Appointments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Router /api/v1/clinics/{id}/diagnostic-appointments/{appointmentId}/cancel [post]
func (h *DiagnosticHandler) CancelClinicAppointment(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri klinika ID formati", nil))
		return
	}

	appointmentID, err := uuid.Parse(c.Param("appointmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri qabul ID formati", nil))
		return
	}

	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	var req dto.CancelDiagnosticAppointmentRequest
	_ = c.ShouldBindJSON(&req)

	appt, err := h.diagnosticService.CancelClinicDiagnosticAppointment(c.Request.Context(), clinicID, userID, appointmentID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Qabul bekor qilindi", appt))
}

// CompleteClinicAppointment godoc
// @Summary Diagnostika tekshiruvi yakunlanganligini belgilash
// @Tags Clinic Diagnostics Appointments
// @Produce json
// @Security BearerAuth
// @Router /api/v1/clinics/{id}/diagnostic-appointments/{appointmentId}/complete [post]
func (h *DiagnosticHandler) CompleteClinicAppointment(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri klinika ID formati", nil))
		return
	}

	appointmentID, err := uuid.Parse(c.Param("appointmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri qabul ID formati", nil))
		return
	}

	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	appt, err := h.diagnosticService.CompleteClinicDiagnosticAppointment(c.Request.Context(), clinicID, userID, appointmentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Qabul muvaffaqiyatli yakunlandi", appt))
}

// NoShowClinicAppointment godoc
// @Summary Bemor kelmaganligini belgilash
// @Tags Clinic Diagnostics Appointments
// @Produce json
// @Security BearerAuth
// @Router /api/v1/clinics/{id}/diagnostic-appointments/{appointmentId}/no-show [post]
func (h *DiagnosticHandler) NoShowClinicAppointment(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri klinika ID formati", nil))
		return
	}

	appointmentID, err := uuid.Parse(c.Param("appointmentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri qabul ID formati", nil))
		return
	}

	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	appt, err := h.diagnosticService.NoShowClinicDiagnosticAppointment(c.Request.Context(), clinicID, userID, appointmentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Bemor kelmadi deb belgilandi", appt))
}

