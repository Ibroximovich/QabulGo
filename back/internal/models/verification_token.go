package models

import (
	"time"

	"github.com/google/uuid"
)

type TokenPurpose string

const (
	PurposeEmailVerification TokenPurpose = "email_verification"
	PurposePasswordReset     TokenPurpose = "password_reset"
)

type VerificationToken struct {
	ID         uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     uuid.UUID    `gorm:"type:uuid;not null;index" json:"user_id"`
	Purpose    TokenPurpose `gorm:"type:varchar(50);not null" json:"purpose"`
	TokenHash  string       `gorm:"type:varchar(255);not null;index" json:"-"`
	ExpiresAt  time.Time    `gorm:"type:timestamptz;not null" json:"expires_at"`
	Attempts   int          `gorm:"type:int;not null;default:0" json:"attempts"`
	ConsumedAt *time.Time   `gorm:"type:timestamptz" json:"consumed_at"`
	CreatedAt  time.Time    `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (VerificationToken) TableName() string {
	return "verification_tokens"
}

func (t *VerificationToken) IsValid() bool {
	return t.ConsumedAt == nil && time.Now().Before(t.ExpiresAt) && t.Attempts < 5
}
