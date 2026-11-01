package handlers

import (
	"errors"
	"net/http"

	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/middleware"
	"qabulgo-backend/internal/models"
	"qabulgo-backend/internal/repositories"
	"qabulgo-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ClinicHandler struct {
	clinicService services.ClinicService
}

func NewClinicHandler(clinicService services.ClinicService) *ClinicHandler {
	return &ClinicHandler{clinicService: clinicService}
}

// CreateClinic godoc
// @Summary Yangi klinika yaratish
// @Description Klinika egasi yangi klinika profilini yaratadi (status: draft)
// @Tags Clinic Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateClinicRequest true "Klinika ma’lumotlari"
// @Success 201 {object} dto.SuccessResponse{data=dto.ClinicDetailDTO} "Klinika muvaffaqiyatli yaratildi"
// @Failure 400 {object} dto.ErrorResponse "Validatsiya xatoligi"
// @Failure 401 {object} dto.ErrorResponse "Avtorizatsiya talab qilinadi"
// @Failure 409 {object} dto.ErrorResponse "Siz allaqachon klinika yaratgansiz"
// @Router /api/v1/clinics [post]
func (h *ClinicHandler) CreateClinic(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	var req dto.CreateClinicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Kiritilgan ma’lumotlar formati noto‘g‘ri", map[string]string{"body": err.Error()}))
		return
	}

	clinic, err := h.clinicService.CreateClinic(c.Request.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicAlreadyExists) {
			c.JSON(http.StatusConflict, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusCreated, dto.NewSuccessResponse("Klinika muvaffaqiyatli yaratildi (qoralama holatida)", clinic))
}

// GetMyClinic godoc
// @Summary O‘z klinikasini ko‘rish
// @Description Avtorizatsiyadan o‘tgan klinika xodimi/egasining klinika ma’lumotlari
// @Tags Clinic Profile
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.SuccessResponse{data=dto.ClinicDetailDTO} "Klinika ma’lumotlari"
// @Failure 401 {object} dto.ErrorResponse "Avtorizatsiya talab qilinadi"
// @Failure 404 {object} dto.ErrorResponse "Klinika topilmadi"
// @Router /api/v1/clinics/mine [get]
func (h *ClinicHandler) GetMyClinic(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	clinic, err := h.clinicService.GetMyClinic(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("Sizga tegishli klinika topilmadi", nil))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Serverda xatolik yuz berdi", nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Klinika ma’lumotlari", clinic))
}

// GetClinicByID godoc
// @Summary Klinika ma’lumotlarini ID orqali ko‘rish
// @Description Klinika ma’lumotlarini olish (tasdiqlangan klinikalar ochiq, boshqalari faqat egasi yoki superadmin uchun)
// @Tags Clinic Profile
// @Produce json
// @Param id path string true "Klinika UUID"
// @Success 200 {object} dto.SuccessResponse{data=dto.ClinicDetailDTO} "Klinika ma’lumotlari"
// @Failure 404 {object} dto.ErrorResponse "Klinika topilmadi"
// @Router /api/v1/clinics/{id} [get]
func (h *ClinicHandler) GetClinicByID(c *gin.Context) {
	idStr := c.Param("id")
	clinicID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	requestingUserID, _ := middleware.GetCurrentUserID(c)
	roleVal, _ := c.Get(middleware.RoleContextKey)
	isSuperAdmin := false
	if r, ok := roleVal.(models.UserRole); ok && r == models.RoleSuperAdmin {
		isSuperAdmin = true
	}

	clinic, err := h.clinicService.GetClinicByID(c.Request.Context(), clinicID, requestingUserID, isSuperAdmin)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("Klinika topilmadi", nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Klinika ma’lumotlari", clinic))
}

// UpdateClinic godoc
// @Summary Klinika ma’lumotlarini tahrirlash
// @Description Klinika profilini va ish soatlarini yangilash (faqat klinika egasi/admini)
// @Tags Clinic Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Klinika UUID"
// @Param request body dto.UpdateClinicRequest true "Tahrirlash ma’lumotlari"
// @Success 200 {object} dto.SuccessResponse{data=dto.ClinicDetailDTO} "Klinika yangilandi"
// @Failure 400 {object} dto.ErrorResponse "Xatolik"
// @Failure 403 {object} dto.ErrorResponse "Ruxsat yo‘q"
// @Failure 404 {object} dto.ErrorResponse "Klinika topilmadi"
// @Router /api/v1/clinics/{id} [patch]
func (h *ClinicHandler) UpdateClinic(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	idStr := c.Param("id")
	clinicID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	var req dto.UpdateClinicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Kiritilgan ma’lumotlar formati noto‘g‘ri", map[string]string{"body": err.Error()}))
		return
	}

	updated, err := h.clinicService.UpdateClinic(c.Request.Context(), clinicID, userID, &req)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, repositories.ErrClinicNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Klinika ma’lumotlari yangilandi", updated))
}

// SubmitClinic godoc
// @Summary Klinikani tekshiruvga yuborish
// @Description Klinika ma’lumotlari to‘liq kiritilgach, superadmin tekshiruvi uchun yuboradi (status: pending)
// @Tags Clinic Profile
// @Produce json
// @Security BearerAuth
// @Param id path string true "Klinika UUID"
// @Success 200 {object} dto.SuccessResponse{data=dto.ClinicDetailDTO} "Klinika tekshiruvga yuborildi"
// @Failure 400 {object} dto.ErrorResponse "Yetarli ma’lumot to‘ldirilmagan"
// @Failure 403 {object} dto.ErrorResponse "Ruxsat yo‘q"
// @Router /api/v1/clinics/{id}/submit [post]
func (h *ClinicHandler) SubmitClinic(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Avtorizatsiya talab qilinadi", nil))
		return
	}

	idStr := c.Param("id")
	clinicID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	submitted, err := h.clinicService.SubmitClinic(c.Request.Context(), clinicID, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Klinika superadmin tekshiruviga yuborildi", submitted))
}

