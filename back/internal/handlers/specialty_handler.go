package handlers

import (
	"net/http"

	"qabulgo-backend/internal/dto"
	"qabulgo-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type SpecialtyHandler struct {
	specialtyService services.SpecialtyService
}

func NewSpecialtyHandler(specialtyService services.SpecialtyService) *SpecialtyHandler {
	return &SpecialtyHandler{specialtyService: specialtyService}
}

// ListSpecialties godoc
// @Summary Barcha tibbiy mutaxassisliklar ro‘yxati
// @Description Klinikaga shifokor biriktirish va qidiruv uchun mutaxassisliklar
// @Tags Specialties
// @Produce json
// @Success 200 {object} dto.SuccessResponse{data=[]dto.SpecialtyDTO} "Mutaxassisliklar ro‘yxati"
// @Router /api/v1/specialties [get]
func (h *SpecialtyHandler) ListSpecialties(c *gin.Context) {
	items, err := h.specialtyService.ListSpecialties(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Mutaxassisliklarni olishda xatolik", nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Mutaxassisliklar ro‘yxati", items))
}

