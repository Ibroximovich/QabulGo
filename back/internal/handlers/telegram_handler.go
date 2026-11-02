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

type TelegramHandler struct {
	telegramService services.TelegramService
}

func NewTelegramHandler(telegramService services.TelegramService) *TelegramHandler {
	return &TelegramHandler{telegramService: telegramService}
}

// GenerateLinkToken generates a one-time token and deep link to connect the clinic to Telegram
func (h *TelegramHandler) GenerateLinkToken(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri klinika identifikatori", nil))
		return
	}

	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Autentifikatsiyadan o‘tilmagan", nil))
		return
	}

	resp, err := h.telegramService.GenerateClinicLinkingToken(c.Request.Context(), clinicID, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse("Sizda ushbu klinika sozlamalarini boshqarish huquqi yo‘q", nil))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Telegram ulanish havolasini yaratishda xatolik: "+err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Ulanish havolasi yaratildi", resp))
}

// GetStatus returns the Telegram connection status and list of recipients for the clinic
func (h *TelegramHandler) GetStatus(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri klinika identifikatori", nil))
		return
	}

	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Autentifikatsiyadan o‘tilmagan", nil))
		return
	}

	resp, err := h.telegramService.GetClinicTelegramStatus(c.Request.Context(), clinicID, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse("Sizda ushbu klinika sozlamalarini ko‘rish huquqi yo‘q", nil))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Statusni olishda xatolik: "+err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Telegram holati muvaffaqiyatli olindi", resp))
}

// DisconnectRecipient disconnects a linked Telegram user
func (h *TelegramHandler) DisconnectRecipient(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri klinika identifikatori", nil))
		return
	}

	recipientID, err := uuid.Parse(c.Param("recipientId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri qabul qiluvchi identifikatori", nil))
		return
	}

	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Autentifikatsiyadan o‘tilmagan", nil))
		return
	}

	if err := h.telegramService.DisconnectClinicRecipient(c.Request.Context(), clinicID, userID, recipientID); err != nil {
		if errors.Is(err, repositories.ErrClinicAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse("Huquq yetarli emas", nil))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Ulanishni uzishda xatolik: "+err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Telegram qabul qiluvchi muvaffaqiyatli uzildi", nil))
}

// ToggleNotifications toggles whether a linked Telegram user receives appointment notifications
func (h *TelegramHandler) ToggleNotifications(c *gin.Context) {
	clinicID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri klinika identifikatori", nil))
		return
	}

	recipientID, err := uuid.Parse(c.Param("recipientId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Noto‘g‘ri qabul qiluvchi identifikatori", nil))
		return
	}

	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Autentifikatsiyadan o‘tilmagan", nil))
		return
	}

	updated, err := h.telegramService.ToggleClinicNotifications(c.Request.Context(), clinicID, userID, recipientID)
	if err != nil {
		if errors.Is(err, repositories.ErrClinicAccessDenied) {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse("Huquq yetarli emas", nil))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Xabarnomani o‘zgartirishda xatolik: "+err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Xabarnoma holati yangilandi", updated))
}

// GetAdminStatus returns the Telegram Superadmin Bot configuration status
func (h *TelegramHandler) GetAdminStatus(c *gin.Context) {
	resp, err := h.telegramService.GetAdminTelegramStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Admin bot statusini olishda xatolik: "+err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, dto.NewSuccessResponse("Admin Telegram holati olindi", resp))
}
