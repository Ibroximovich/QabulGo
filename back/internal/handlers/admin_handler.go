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

type AdminHandler struct {
	adminService services.AdminService
}

func NewAdminHandler(adminService services.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

// ListClinics godoc
// @Summary Superadmin uchun klinikalar ro‘yxati
// @Description Status bo‘yicha filtrlash (draft, pending, approved, rejected, suspended), qidiruv va sahifalash
// @Tags Superadmin
// @Produce json
// @Security BearerAuth
// @Param status query string false "Status (masalan: pending)"
// @Param search query string false "Qidiruv"
// @Param page query int false "Sahifa"
// @Param limit query int false "Miqdor"
// @Success 200 {object} dto.SuccessResponse{data=[]dto.ClinicSummaryDTO} "Klinikalar ro‘yxati"
// @Router /api/v1/admin/clinics [get]
func (h *AdminHandler) ListClinics(c *gin.Context) {
	var filter dto.AdminClinicFilter
	_ = c.ShouldBindQuery(&filter)

	clinics, total, err := h.adminService.ListClinics(c.Request.Context(), &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Ma’lumotlarni olishda xatolik", nil))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Klinikalar ro‘yxati",
		"data":    clinics,
		"meta": gin.H{
			"total": total,
			"page":  filter.Page,
			"limit": filter.Limit,
		},
	})
}

// GetClinicDetails godoc
// @Summary Superadmin uchun klinika to‘liq ma’lumotlari va tekshiruv tarixi
// @Description Klinika ma’lumotlari, ish vaqtlari, shifokorlar soni va audit tarixi
// @Tags Superadmin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Klinika UUID"
// @Success 200 {object} dto.SuccessResponse{data=dto.ClinicDetailDTO} "Klinika tafsilotlari"
// @Router /api/v1/admin/clinics/{id} [get]
func (h *AdminHandler) GetClinicDetails(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	detail, reviews, err := h.adminService.GetClinicDetails(c.Request.Context(), clinicID)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("Klinika topilmadi", nil))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Serverda xatolik", nil))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Klinika tafsilotlari",
		"data": gin.H{
			"clinic":  detail,
			"reviews": reviews,
		},
	})
}

// ApproveClinic godoc
// @Summary Klinikani tasdiqlash
// @Description Superadmin klinikani tasdiqlaydi (status: approved) va tizimda faol qiladi
// @Tags Superadmin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Klinika UUID"
// @Success 200 {object} dto.SuccessResponse "Klinika muvaffaqiyatli tasdiqlandi"
// @Router /api/v1/admin/clinics/{id}/approve [post]
func (h *AdminHandler) ApproveClinic(c *gin.Context) {
	adminID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	if err := h.adminService.ApproveClinic(c.Request.Context(), adminID, clinicID, ipAddress, userAgent); err != nil {
		if errors.Is(err, repositories.ErrClinicNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("Klinika topilmadi", nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Klinika muvaffaqiyatli tasdiqlandi va platformada faollashtirildi", nil))
}

// RejectClinic godoc
// @Summary Klinikani rad etish
// @Description Superadmin arizani rad etadi va sababini qayd qiladi (status: rejected)
// @Tags Superadmin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Klinika UUID"
// @Param request body dto.AdminReviewClinicRequest true "Rad etish sababi"
// @Success 200 {object} dto.SuccessResponse "Klinika rad etildi"
// @Router /api/v1/admin/clinics/{id}/reject [post]
func (h *AdminHandler) RejectClinic(c *gin.Context) {
	adminID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	var req dto.AdminReviewClinicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Rad etish sababi yozilishi shart", nil))
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	if err := h.adminService.RejectClinic(c.Request.Context(), adminID, clinicID, req.Reason, ipAddress, userAgent); err != nil {
		if errors.Is(err, repositories.ErrClinicNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("Klinika topilmadi", nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Klinika arizasi rad etildi va sababi qayd etildi", nil))
}

// SuspendClinic godoc
// @Summary Klinikani bloklash / to‘xtatish
// @Description Superadmin qoidabuzarlik yoki tekshiruv tufayli klinikani bloklaydi (status: suspended)
// @Tags Superadmin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Klinika UUID"
// @Param request body dto.AdminReviewClinicRequest true "Bloklash sababi"
// @Success 200 {object} dto.SuccessResponse "Klinika bloklandi"
// @Router /api/v1/admin/clinics/{id}/suspend [post]
func (h *AdminHandler) SuspendClinic(c *gin.Context) {
	adminID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	var req dto.AdminReviewClinicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Bloklash sababi yozilishi shart", nil))
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	if err := h.adminService.SuspendClinic(c.Request.Context(), adminID, clinicID, req.Reason, ipAddress, userAgent); err != nil {
		if errors.Is(err, repositories.ErrClinicNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("Klinika topilmadi", nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Klinika faoliyati to‘xtatildi (bloklandi)", nil))
}

// GetDashboardSummary godoc
// @Summary Superadmin boshqaruv paneli umumiy statistikasi
// @Description Real PostgreSQL aggregate ma'lumotlari bo'yicha 13 ta KPI ko'rsatkichlari
// @Tags Superadmin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.SuccessResponse{data=dto.AdminDashboardSummaryDTO} "Umumiy statistika"
// @Router /api/v1/admin/dashboard/summary [get]
func (h *AdminHandler) GetDashboardSummary(c *gin.Context) {
	summary, err := h.adminService.GetDashboardSummary(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Statistikani yuklashda xatolik yuz berdi", nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Boshqaruv paneli ko‘rsatkichlari", summary))
}

// GetDashboardAnalytics godoc
// @Summary Superadmin grafik tahlillari
// @Description Bronlar dinamikasi, klinikalar va bemorlar o'sishi, status taqsimoti
// @Tags Superadmin
// @Produce json
// @Security BearerAuth
// @Param range query string false "Vaqt oralig'i (today, 7d, 30d, 90d, custom)"
// @Param start_date query string false "Boshlanish sanasi (YYYY-MM-DD)"
// @Param end_date query string false "Tugash sanasi (YYYY-MM-DD)"
// @Success 200 {object} dto.SuccessResponse{data=dto.AdminDashboardAnalyticsDTO} "Analitika ma'lumotlari"
// @Router /api/v1/admin/dashboard/analytics [get]
func (h *AdminHandler) GetDashboardAnalytics(c *gin.Context) {
	var filter dto.AdminDashboardAnalyticsFilter
	_ = c.ShouldBindQuery(&filter)

	analytics, err := h.adminService.GetDashboardAnalytics(c.Request.Context(), &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Analitikani yuklashda xatolik yuz berdi", nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Analitika ma'lumotlari", analytics))
}

// GetClinicDoctors godoc
// @Summary Klinikaga biriktirilgan shifokorlar ro'yxati
// @Tags Superadmin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Klinika UUID"
// @Success 200 {object} dto.SuccessResponse{data=[]dto.AdminClinicDoctorItemDTO} "Klinika shifokorlari"
// @Router /api/v1/admin/clinics/{id}/doctors [get]
func (h *AdminHandler) GetClinicDoctors(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	doctors, err := h.adminService.GetClinicDoctors(c.Request.Context(), clinicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Shifokorlarni yuklashda xatolik", nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Klinika shifokorlari", doctors))
}

// GetClinicAppointments godoc
// @Summary Klinikaning barcha bronlari
// @Tags Superadmin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Klinika UUID"
// @Param page query int false "Sahifa"
// @Param limit query int false "Limit"
// @Success 200 {object} dto.SuccessResponse{data=[]dto.AdminAppointmentListItemDTO} "Klinika bronlari"
// @Router /api/v1/admin/clinics/{id}/appointments [get]
func (h *AdminHandler) GetClinicAppointments(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	var filter dto.AdminAppointmentFilter
	_ = c.ShouldBindQuery(&filter)

	appointments, total, err := h.adminService.GetClinicAppointments(c.Request.Context(), clinicID, &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Bronlarni yuklashda xatolik", nil))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Klinika bronlari",
		"data":    appointments,
		"meta": gin.H{
			"total": total,
			"page":  filter.Page,
			"limit": filter.Limit,
		},
	})
}

// GetClinicAnalytics godoc
// @Summary Klinikaning xususiy tahlillari va statistikasi
// @Tags Superadmin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Klinika UUID"
// @Success 200 {object} dto.SuccessResponse{data=dto.AdminClinicAnalyticsDTO} "Klinika analitikasi"
// @Router /api/v1/admin/clinics/{id}/analytics [get]
func (h *AdminHandler) GetClinicAnalytics(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	analytics, err := h.adminService.GetClinicAnalytics(c.Request.Context(), clinicID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Analitikani yuklashda xatolik", nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Klinika analitikasi", analytics))
}

// ListPatients godoc
// @Summary Superadmin uchun barcha bemorlar ro'yxati
// @Tags Superadmin
// @Produce json
// @Security BearerAuth
// @Param search query string false "Qidiruv (ism, familiya, email, telefon)"
// @Param page query int false "Sahifa"
// @Param limit query int false "Limit"
// @Success 200 {object} dto.SuccessResponse{data=[]dto.AdminPatientListItemDTO} "Bemorlar ro‘yxati"
// @Router /api/v1/admin/patients [get]
func (h *AdminHandler) ListPatients(c *gin.Context) {
	var filter dto.AdminPatientFilter
	_ = c.ShouldBindQuery(&filter)

	patients, total, err := h.adminService.ListPatients(c.Request.Context(), &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Bemorlarni yuklashda xatolik", nil))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Bemorlar ro‘yxati",
		"data":    patients,
		"meta": gin.H{
			"total": total,
			"page":  filter.Page,
			"limit": filter.Limit,
		},
	})
}

// GetPatientDetails godoc
// @Summary Bemorning to'liq profili va bronlar tarixi
// @Tags Superadmin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Bemor UUID"
// @Success 200 {object} dto.SuccessResponse{data=dto.AdminPatientDetailDTO} "Bemor tafsilotlari"
// @Router /api/v1/admin/patients/{id} [get]
func (h *AdminHandler) GetPatientDetails(c *gin.Context) {
	patientID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Bemor ID formati noto‘g‘ri", nil))
		return
	}

	patient, err := h.adminService.GetPatientDetails(c.Request.Context(), patientID)
	if err != nil {
		if errors.Is(err, repositories.ErrPatientNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("Bemor topilmadi", nil))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Serverda xatolik", nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Bemor ma'lumotlari", patient))
}

// ListDoctors godoc
// @Summary Superadmin uchun barcha shifokorlar ro'yxati
// @Tags Superadmin
// @Produce json
// @Security BearerAuth
// @Param search query string false "Ism bo'yicha qidiruv"
// @Param specialty_id query string false "Mutaxassislik ID"
// @Param clinic_id query string false "Klinika ID"
// @Param page query int false "Sahifa"
// @Param limit query int false "Limit"
// @Success 200 {object} dto.SuccessResponse{data=[]dto.AdminDoctorListItemDTO} "Shifokorlar ro‘yxati"
// @Router /api/v1/admin/doctors [get]
func (h *AdminHandler) ListDoctors(c *gin.Context) {
	var filter dto.AdminDoctorFilter
	_ = c.ShouldBindQuery(&filter)

	doctors, total, err := h.adminService.ListDoctors(c.Request.Context(), &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Shifokorlarni yuklashda xatolik", nil))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Shifokorlar ro‘yxati",
		"data":    doctors,
		"meta": gin.H{
			"total": total,
			"page":  filter.Page,
			"limit": filter.Limit,
		},
	})
}

// GetDoctorDetails godoc
// @Summary Shifokor to'liq profili, ish joylari va jadvallari
// @Tags Superadmin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Shifokor UUID"
// @Success 200 {object} dto.SuccessResponse{data=dto.AdminDoctorDetailDTO} "Shifokor tafsilotlari"
// @Router /api/v1/admin/doctors/{id} [get]
func (h *AdminHandler) GetDoctorDetails(c *gin.Context) {
	doctorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Shifokor ID formati noto‘g‘ri", nil))
		return
	}

	doctor, err := h.adminService.GetDoctorDetails(c.Request.Context(), doctorID)
	if err != nil {
		if errors.Is(err, repositories.ErrDoctorNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("Shifokor topilmadi", nil))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Serverda xatolik", nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Shifokor ma'lumotlari", doctor))
}

// ListAppointments godoc
// @Summary Superadmin uchun barcha platforma bronlari
// @Tags Superadmin
// @Produce json
// @Security BearerAuth
// @Param clinic_id query string false "Klinika ID"
// @Param doctor_id query string false "Shifokor ID"
// @Param patient_search query string false "Bemor qidiruv"
// @Param status query string false "Status (confirmed, completed, cancelled, no_show)"
// @Param start_date query string false "Boshlanish sanasi (YYYY-MM-DD)"
// @Param end_date query string false "Tugash sanasi (YYYY-MM-DD)"
// @Param page query int false "Sahifa"
// @Param limit query int false "Limit"
// @Success 200 {object} dto.SuccessResponse{data=[]dto.AdminAppointmentListItemDTO} "Bronlar ro‘yxati"
// @Router /api/v1/admin/appointments [get]
func (h *AdminHandler) ListAppointments(c *gin.Context) {
	var filter dto.AdminAppointmentFilter
	_ = c.ShouldBindQuery(&filter)

	appointments, total, err := h.adminService.ListAppointments(c.Request.Context(), &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Bronlarni yuklashda xatolik", nil))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Platforma bronlari ro‘yxati",
		"data":    appointments,
		"meta": gin.H{
			"total": total,
			"page":  filter.Page,
			"limit": filter.Limit,
		},
	})
}

// GetAppointmentDetails godoc
// @Summary Bron to'liq tafsilotlari
// @Tags Superadmin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Bron UUID"
// @Success 200 {object} dto.SuccessResponse{data=dto.AdminAppointmentListItemDTO} "Bron tafsilotlari"
// @Router /api/v1/admin/appointments/{id} [get]
func (h *AdminHandler) GetAppointmentDetails(c *gin.Context) {
	appointmentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Bron ID formati noto‘g‘ri", nil))
		return
	}

	appointment, err := h.adminService.GetAppointmentDetails(c.Request.Context(), appointmentID)
	if err != nil {
		if errors.Is(err, repositories.ErrAppointmentNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("Bron topilmadi", nil))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Serverda xatolik", nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Bron tafsilotlari", appointment))
}

