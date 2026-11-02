package models

import (
	"time"

	"github.com/google/uuid"
)

// ClinicTelegramRecipient represents a connected Telegram chat recipient for a clinic
type ClinicTelegramRecipient struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID             uuid.UUID `gorm:"type:uuid;not null;index" json:"clinic_id"`
	UserID               *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	TelegramUserID       int64     `gorm:"type:bigint;not null;index" json:"telegram_user_id"`
	TelegramChatID       int64     `gorm:"type:bigint;not null" json:"telegram_chat_id"`
	TelegramUsername     string    `gorm:"type:varchar(255)" json:"telegram_username"`
	TelegramFirstName    string    `gorm:"type:varchar(255)" json:"telegram_first_name"`
	IsActive             bool      `gorm:"type:boolean;not null;default:true" json:"is_active"`
	NotificationsEnabled bool      `gorm:"type:boolean;not null;default:true" json:"notifications_enabled"`
	CreatedAt            time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt            time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relationships
	Clinic *Clinic `gorm:"foreignKey:ClinicID;constraint:OnDelete:CASCADE" json:"clinic,omitempty"`
	User   *User   `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL" json:"user,omitempty"`
}

func (ClinicTelegramRecipient) TableName() string {
	return "clinic_telegram_recipients"
}

// TelegramLinkingToken represents a one-time cryptographic token to link a clinic to Telegram
type TelegramLinkingToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"clinic_id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	TokenHash string     `gorm:"type:varchar(64);not null;uniqueIndex" json:"token_hash"`
	IsUsed    bool       `gorm:"type:boolean;not null;default:false" json:"is_used"`
	ExpiresAt time.Time  `gorm:"type:timestamptz;not null" json:"expires_at"`
	UsedAt    *time.Time `gorm:"type:timestamptz" json:"used_at,omitempty"`
	CreatedAt time.Time  `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	Clinic *Clinic `gorm:"foreignKey:ClinicID;constraint:OnDelete:CASCADE" json:"clinic,omitempty"`
	User   *User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

func (TelegramLinkingToken) TableName() string {
	return "telegram_linking_tokens"
}

// NotificationOutbox represents a durable event to be sent asynchronously to Telegram
type NotificationOutbox struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EventType      string     `gorm:"type:varchar(50);not null" json:"event_type"` // appointment_created, appointment_cancelled, clinic_submitted
	EntityType     string     `gorm:"type:varchar(50);not null" json:"entity_type"` // appointment, clinic
	EntityID       uuid.UUID  `gorm:"type:uuid;not null" json:"entity_id"`
	TargetClinicID *uuid.UUID `gorm:"type:uuid;index" json:"target_clinic_id,omitempty"`
	TargetRole     string     `gorm:"type:varchar(50);not null" json:"target_role"` // clinic, admin
	Payload        string     `gorm:"type:jsonb;not null" json:"payload"`
	Status         string     `gorm:"type:varchar(20);not null;default:'pending'" json:"status"` // pending, delivered, failed, skipped
	RetryCount     int        `gorm:"type:int;not null;default:0" json:"retry_count"`
	MaxRetries     int        `gorm:"type:int;not null;default:5" json:"max_retries"`
	NextRetryAt    time.Time  `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"next_retry_at"`
	DeliveredAt    *time.Time `gorm:"type:timestamptz" json:"delivered_at,omitempty"`
	LastError      *string    `gorm:"type:text" json:"last_error,omitempty"`
	CreatedAt      time.Time  `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (NotificationOutbox) TableName() string {
	return "notification_outbox"
}

// TelegramUpdateDedup stores processed Telegram update IDs to ensure idempotency
type TelegramUpdateDedup struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BotType     string    `gorm:"type:varchar(20);not null" json:"bot_type"` // clinic, admin
	UpdateID    int64     `gorm:"type:bigint;not null" json:"update_id"`
	ProcessedAt time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"processed_at"`
}

func (TelegramUpdateDedup) TableName() string {
	return "telegram_update_dedup"
}

// TelegramAdminSession maintains conversational state for Superadmins (e.g. entering rejection reason)
type TelegramAdminSession struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TelegramUserID int64      `gorm:"type:bigint;not null;uniqueIndex" json:"telegram_user_id"`
	AwaitingAction string     `gorm:"type:varchar(50)" json:"awaiting_action"` // reject_clinic
	TargetClinicID *uuid.UUID `gorm:"type:uuid" json:"target_clinic_id,omitempty"`
	UpdatedAt      time.Time  `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (TelegramAdminSession) TableName() string {
	return "telegram_admin_sessions"
}

