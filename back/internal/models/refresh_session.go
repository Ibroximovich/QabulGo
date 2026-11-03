package models

import (
	"time"

	"github.com/google/uuid"
)

type RefreshSession struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	TokenHash    string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"-"`
	ExpiresAt    time.Time  `gorm:"type:timestamptz;not null" json:"expires_at"`
	MaxExpiresAt time.Time  `gorm:"type:timestamptz;not null" json:"max_expires_at"`
	RevokedAt    *time.Time `gorm:"type:timestamptz" json:"revoked_at"`
	CreatedAt    time.Time  `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	LastUsedAt   time.Time  `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"last_used_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (RefreshSession) TableName() string {
	return "refresh_sessions"
}

func (s *RefreshSession) IsActive() bool {
	now := time.Now().UTC()
	return s.RevokedAt == nil && now.Before(s.ExpiresAt) && (s.MaxExpiresAt.IsZero() || now.Before(s.MaxExpiresAt))
}
