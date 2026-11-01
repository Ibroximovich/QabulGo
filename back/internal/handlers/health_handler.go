package handlers

import (
	"net/http"

	"qabulgo-backend/internal/dto"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck godoc
// @Summary Server holatini tekshirish
// @Description Server faol ekanligini bildiradi
// @Tags System
// @Produce json
// @Success 200 {object} dto.SuccessResponse
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, dto.NewSuccessResponse("Server muvaffaqiyatli ishlamoqda", gin.H{
		"status": "UP",
	}))
}
