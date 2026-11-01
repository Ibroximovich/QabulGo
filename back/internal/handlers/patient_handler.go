package handlers

import (
	"errors"
	"net/http"

	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/middleware"
	"qabulgo-backend/internal/repositories"
	"qabulgo-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type PatientHandler struct {
	patientService services.PatientService
}

func NewPatientHandler(patientService services.PatientService) *PatientHandler {
	return &PatientHandler{patientService: patientService}
}

// GetProfile godoc
// @Summary Bemor profilini ko‘rish
// @Description Avtorizatsiyadan o‘tgan bemorning shaxsiy ma’lumotlarini qaytaradi
// @Tags Patient
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.SuccessResponse{data=dto.PatientProfileResponse} "Bemor profili"
// @Failure 401 {object} dto.ErrorResponse "Avtorizatsiyadan o‘tilmagan"
// @Failure 403 {object} dto.ErrorResponse "Ruxsat berilmagan rol"
// @Failure 404 {object} dto.ErrorResponse "Foydalanuvchi topilmadi"
// @Router /api/v1/patient/me [get]
func (h *PatientHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Foydalanuvchi identifikatori topilmadi", nil))
		return
	}

	profile, err := h.patientService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Profilni yuklashda xatolik yuz berdi", nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Bemor profili muvaffaqiyatli olindi", profile))
}

// UpdateProfile godoc
// @Summary Bemor profilini tahrirlash
// @Description Bemor ism, familiya va telefon raqamini o‘zgartira oladi (email o‘zgarmaydi)
// @Tags Patient
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdatePatientProfileRequest true "Yangi ma’lumotlar"
// @Success 200 {object} dto.SuccessResponse{data=dto.PatientProfileResponse} "Profil muvaffaqiyatli yangilandi"
// @Failure 400 {object} dto.ErrorResponse "Validatsiya xatoligi"
// @Failure 401 {object} dto.ErrorResponse "Avtorizatsiyadan o‘tilmagan"
// @Failure 409 {object} dto.ErrorResponse "Telefon raqami allaqachon mavjud"
// @Router /api/v1/patient/me [patch]
func (h *PatientHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Foydalanuvchi identifikatori topilmadi", nil))
		return
	}

	var req dto.UpdatePatientProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.NewErrorResponse("Kiritilgan ma'lumotlar formati noto‘g‘ri", nil))
		return
	}

	updatedProfile, err := h.patientService.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, repositories.ErrPhoneAlreadyExist) {
			c.JSON(http.StatusConflict, dto.NewErrorResponse(err.Error(), map[string]string{"phone": err.Error()}))
			return
		}
		if errors.Is(err, services.ErrNameLength) || errors.Is(err, services.ErrInvalidPhoneFormat) {
			c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Profilni yangilashda xatolik yuz berdi", nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Profil ma’lumotlari muvaffaqiyatli yangilandi", updatedProfile))
}
