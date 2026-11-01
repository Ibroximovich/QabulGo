package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RolePatient     UserRole = "patient"
	RoleClinicOwner UserRole = "clinic_owner"
	RoleClinicAdmin UserRole = "clinic_admin"
	RoleDoctor      UserRole = "doctor"
	RoleSuperAdmin  UserRole = "super_admin"
)

type UserStatus string

const (
	StatusActive  UserStatus = "active"
	StatusBlocked UserStatus = "blocked"
)

type User struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FirstName       string     `gorm:"type:varchar(50);not null" json:"first_name"`
	LastName        string     `gorm:"type:varchar(50);not null" json:"last_name"`
	Email           string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Phone           string     `gorm:"type:varchar(20);uniqueIndex;not null" json:"phone"`
	PasswordHash    string     `gorm:"type:varchar(255);not null" json:"-"`
	Role            UserRole   `gorm:"type:varchar(20);not null;default:'patient'" json:"role"`
	Status          UserStatus `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	EmailVerifiedAt *time.Time `gorm:"type:timestamptz" json:"email_verified_at"`
	TermsAcceptedAt time.Time  `gorm:"type:timestamptz;not null" json:"terms_accepted_at"`
	TermsVersion    string     `gorm:"type:varchar(20);not null" json:"terms_version"`
	CreatedAt       time.Time  `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Aloqalar
	RefreshSessions    []RefreshSession    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	VerificationTokens []VerificationToken `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}

func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

func (u *User) IsClinicStaff() bool {
	return u.Role == RoleClinicOwner || u.Role == RoleClinicAdmin
}

func (u *User) IsSuperAdmin() bool {
	return u.Role == RoleSuperAdmin
}
