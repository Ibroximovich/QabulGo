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

type DoctorHandler struct {
	doctorService services.DoctorService
}

func NewDoctorHandler(doctorService services.DoctorService) *DoctorHandler {
	return &DoctorHandler{doctorService: doctorService}
}

// AddDoctor godoc
// @Summary Klinikaga yangi shifokor qo‘shish
// @Description Klinika ma’muri yangi shifokor yaratib, o‘z klinikasiga bog‘laydi
// @Tags Doctor Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Klinika UUID"
// @Param request body dto.CreateDoctorRequest true "Shifokor ma’lumotlari"
// @Success 201 {object} dto.SuccessResponse{data=dto.ClinicDoctorItemDTO} "Shifokor qo‘shildi"
// @Failure 400 {object} dto.ErrorResponse "Validatsiya xatoligi"
// @Failure 403 {object} dto.ErrorResponse "Ruxsat yo‘q"
// @Router /api/v1/clinics/{id}/doctors [post]
func (h *DoctorHandler) AddDoctor(c *gin.Context) {
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

	var req dto.CreateDoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Kiritilgan ma’lumotlar formati noto‘g‘ri", map[string]string{"body": err.Error()}))
		return
	}

	doc, err := h.doctorService.AddDoctor(c.Request.Context(), clinicID, userID, &req)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, repositories.ErrSpecialtyNotFound) {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusCreated, dto.NewSuccessResponse("Shifokor muvaffaqiyatli qo‘shildi", doc))
}

// ListDoctors godoc
// @Summary Klinika shifokorlarini ro‘yxatini olish
// @Description Filtrlash, qidirish va sahifalash imkoniyati bilan
// @Tags Doctor Management
// @Produce json
// @Param id path string true "Klinika UUID"
// @Param specialty_id query string false "Mutaxassislik ID"
// @Param is_active query bool false "Faollik holati"
// @Param search query string false "Qidiruv (ism/familiya)"
// @Param page query int false "Sahifa (default 1)"
// @Param limit query int false "Miqdor (default 20)"
// @Success 200 {object} dto.SuccessResponse{data=[]dto.ClinicDoctorItemDTO} "Shifokorlar ro‘yxati"
// @Router /api/v1/clinics/{id}/doctors [get]
func (h *DoctorHandler) ListDoctors(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Klinika ID formati noto‘g‘ri", nil))
		return
	}

	var filter dto.DoctorFilter
	_ = c.ShouldBindQuery(&filter)

	items, total, err := h.doctorService.ListClinicDoctors(c.Request.Context(), clinicID, &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Ma’lumotlarni olishda xatolik", nil))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Klinika shifokorlari ro‘yxati",
		"data":    items,
		"meta": gin.H{
			"total": total,
			"page":  filter.Page,
			"limit": filter.Limit,
		},
	})
}

// GetDoctor godoc
// @Summary Shifokor batafsil ma’lumotlarini olish
// @Description Klinika va shifokor bog‘lanmasi ma’lumotlari
// @Tags Doctor Management
// @Produce json
// @Param id path string true "Klinika UUID"
// @Param doctorId path string true "Shifokor (clinic_doctor) UUID"
// @Success 200 {object} dto.SuccessResponse{data=dto.ClinicDoctorItemDTO} "Shifokor ma’lumotlari"
// @Failure 404 {object} dto.ErrorResponse "Shifokor topilmadi"
// @Router /api/v1/clinics/{id}/doctors/{doctorId} [get]
func (h *DoctorHandler) GetDoctor(c *gin.Context) {
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

	doc, err := h.doctorService.GetClinicDoctor(c.Request.Context(), clinicID, doctorID)
	if err != nil {
		if errors.Is(err, repositories.ErrDoctorNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("Shifokor topilmadi", nil))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Xatolik", nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Shifokor ma’lumotlari", doc))
}

// UpdateDoctor godoc
// @Summary Shifokor ma’lumotlarini tahrirlash
// @Description Narx, qabul vaqti, tavsif, faollik va shaxsiy ma’lumotlarni yangilash
// @Tags Doctor Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Klinika UUID"
// @Param doctorId path string true "Shifokor UUID"
// @Param request body dto.UpdateDoctorRequest true "Tahrirlash ma’lumotlari"
// @Success 200 {object} dto.SuccessResponse{data=dto.ClinicDoctorItemDTO} "Shifokor yangilandi"
// @Failure 403 {object} dto.ErrorResponse "Ruxsat yo‘q"
// @Router /api/v1/clinics/{id}/doctors/{doctorId} [patch]
func (h *DoctorHandler) UpdateDoctor(c *gin.Context) {
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

	var req dto.UpdateDoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Kiritilgan ma’lumotlar formati noto‘g‘ri", map[string]string{"body": err.Error()}))
		return
	}

	updated, err := h.doctorService.UpdateClinicDoctor(c.Request.Context(), clinicID, userID, doctorID, &req)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, repositories.ErrDoctorNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Shifokor ma’lumotlari yangilandi", updated))
}

// DeactivateDoctor godoc
// @Summary Shifokorni o‘chirish / nofaol qilish
// @Description Soft-delete mexanizmi orqali shifokorni klinikadan o‘chirish
// @Tags Doctor Management
// @Produce json
// @Security BearerAuth
// @Param id path string true "Klinika UUID"
// @Param doctorId path string true "Shifokor UUID"
// @Success 200 {object} dto.SuccessResponse "Shifokor o‘chirildi"
// @Failure 403 {object} dto.ErrorResponse "Ruxsat yo‘q"
// @Router /api/v1/clinics/{id}/doctors/{doctorId} [delete]
func (h *DoctorHandler) DeactivateDoctor(c *gin.Context) {
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

	if err := h.doctorService.DeactivateClinicDoctor(c.Request.Context(), clinicID, userID, doctorID); err != nil {
		if errors.Is(err, repositories.ErrClinicAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		if errors.Is(err, repositories.ErrDoctorNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Shifokor muvaffaqiyatli o‘chirildi", nil))
}

