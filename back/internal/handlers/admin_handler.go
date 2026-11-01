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

