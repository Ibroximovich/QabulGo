package dto

import (
	"time"

	"github.com/google/uuid"
)

// TelegramLinkTokenResponse returns generated deep link for clinic manager
type TelegramLinkTokenResponse struct {
	Token       string    `json:"token"`
	DeepLink    string    `json:"deep_link"`
	BotUsername string    `json:"bot_username"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// ClinicTelegramRecipientDTO represents recipient details for frontend
type ClinicTelegramRecipientDTO struct {
	ID                   uuid.UUID `json:"id"`
	TelegramUserID       int64     `json:"telegram_user_id"`
	TelegramUsername     string    `json:"telegram_username"`
	TelegramFirstName    string    `json:"telegram_first_name"`
	IsActive             bool      `json:"is_active"`
	NotificationsEnabled bool      `json:"notifications_enabled"`
	CreatedAt            time.Time `json:"created_at"`
}

// ClinicTelegramStatusResponse represents clinic telegram connection status
type ClinicTelegramStatusResponse struct {
	IsConnected bool                         `json:"is_connected"`
	BotUsername string                       `json:"bot_username"`
	Recipients  []ClinicTelegramRecipientDTO `json:"recipients"`
}

// AdminTelegramStatusResponse represents superadmin bot status
type AdminTelegramStatusResponse struct {
	IsConfigured    bool    `json:"is_configured"`
	BotUsername     string  `json:"bot_username"`
	BotName         string  `json:"bot_name"`
	DeepLink        string  `json:"deep_link"`
	AllowedAdminIDs []int64 `json:"allowed_admin_ids"`
}

