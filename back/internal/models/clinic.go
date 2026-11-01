package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClinicStatus string

const (
	ClinicStatusDraft     ClinicStatus = "draft"
	ClinicStatusPending   ClinicStatus = "pending"
	ClinicStatusApproved  ClinicStatus = "approved"
	ClinicStatusRejected  ClinicStatus = "rejected"
	ClinicStatusSuspended ClinicStatus = "suspended"
)

type MemberRole string

const (
	MemberRoleOwner MemberRole = "owner"
	MemberRoleAdmin MemberRole = "admin"
)

// Clinic model
type Clinic struct {
	ID              uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name            string       `gorm:"type:varchar(255);not null" json:"name"`
	Description     string       `gorm:"type:text" json:"description"`
	Phone           string       `gorm:"type:varchar(50);not null" json:"phone"`
	Email           string       `gorm:"type:varchar(255)" json:"email"`
	Region          string       `gorm:"type:varchar(100);not null" json:"region"`
	District        string       `gorm:"type:varchar(100);not null" json:"district"`
	Address         string       `gorm:"type:varchar(255);not null" json:"address"`
	LocationURL     string       `gorm:"type:varchar(500)" json:"location_url"`
	Latitude        *float64     `gorm:"type:double precision" json:"latitude,omitempty"`
	Longitude       *float64     `gorm:"type:double precision" json:"longitude,omitempty"`
	LogoURL         string       `gorm:"type:varchar(500)" json:"logo_url"`
	CoverImageURL   string       `gorm:"type:varchar(500)" json:"cover_image_url"`
	Status          ClinicStatus `gorm:"type:varchar(20);not null;default:'draft'" json:"status"`
	RejectionReason string       `gorm:"type:text" json:"rejection_reason,omitempty"`
	OwnerID         uuid.UUID    `gorm:"type:uuid;not null" json:"owner_id"`
	CreatedAt       time.Time    `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time    `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Aloqalar
	Owner        *User               `gorm:"foreignKey:OwnerID;constraint:OnDelete:RESTRICT" json:"owner,omitempty"`
	Members      []ClinicMember      `gorm:"foreignKey:ClinicID;constraint:OnDelete:CASCADE" json:"members,omitempty"`
	Doctors      []ClinicDoctor      `gorm:"foreignKey:ClinicID;constraint:OnDelete:CASCADE" json:"doctors,omitempty"`
	WorkingHours []ClinicWorkingHour `gorm:"foreignKey:ClinicID;constraint:OnDelete:CASCADE" json:"working_hours,omitempty"`
}

func (Clinic) TableName() string {
	return "clinics"
}

// ClinicMember represents a user belonging to a clinic administration
type ClinicMember struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"clinic_id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Role      MemberRole `gorm:"type:varchar(20);not null;default:'owner'" json:"role"`
	CreatedAt time.Time  `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	User   *User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Clinic *Clinic `gorm:"foreignKey:ClinicID;constraint:OnDelete:CASCADE" json:"clinic,omitempty"`
}

func (ClinicMember) TableName() string {
	return "clinic_members"
}

// Specialty represents a medical specialty
type Specialty struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null;unique" json:"name"`
	Slug      string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"slug"`
	IconURL   string    `gorm:"type:varchar(500)" json:"icon_url"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (Specialty) TableName() string {
	return "specialties"
}

// Doctor represents a doctor profile in the platform
type Doctor struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FirstName       string    `gorm:"type:varchar(100);not null" json:"first_name"`
	LastName        string    `gorm:"type:varchar(100);not null" json:"last_name"`
	Bio             string    `gorm:"type:text" json:"bio"`
	ExperienceYears int       `gorm:"type:int;not null;default:0" json:"experience_years"`
	PhotoURL        string    `gorm:"type:varchar(500)" json:"photo_url"`
	CreatedAt       time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Doctor) TableName() string {
	return "doctors"
}

// ClinicDoctor connects a clinic, doctor, and specialty with pricing
type ClinicDoctor struct {
	ID                         uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID                   uuid.UUID      `gorm:"type:uuid;not null;index" json:"clinic_id"`
	DoctorID                   uuid.UUID      `gorm:"type:uuid;not null;index" json:"doctor_id"`
	SpecialtyID                uuid.UUID      `gorm:"type:uuid;not null;index" json:"specialty_id"`
	ConsultationPrice          float64        `gorm:"type:numeric(12,2);not null;default:0" json:"consultation_price"`
	AppointmentDurationMinutes int            `gorm:"type:int;not null;default:30" json:"appointment_duration_minutes"`
	IsActive                   bool           `gorm:"type:boolean;not null;default:true" json:"is_active"`
	DeletedAt                  gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt                  time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt                  time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Clinic    *Clinic    `gorm:"foreignKey:ClinicID;constraint:OnDelete:CASCADE" json:"clinic,omitempty"`
	Doctor    *Doctor    `gorm:"foreignKey:DoctorID;constraint:OnDelete:CASCADE" json:"doctor,omitempty"`
	Specialty *Specialty `gorm:"foreignKey:SpecialtyID;constraint:OnDelete:RESTRICT" json:"specialty,omitempty"`
}

func (ClinicDoctor) TableName() string {
	return "clinic_doctors"
}

// ClinicWorkingHour defines weekly operational hours
type ClinicWorkingHour struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID    uuid.UUID `gorm:"type:uuid;not null;index" json:"clinic_id"`
	Weekday     int       `gorm:"type:int;not null" json:"weekday"` // 1=Dushanba .. 7=Yakshanba
	OpeningTime string    `gorm:"type:varchar(5);not null;default:'08:00'" json:"opening_time"`
	ClosingTime string    `gorm:"type:varchar(5);not null;default:'18:00'" json:"closing_time"`
	IsClosed    bool      `gorm:"type:boolean;not null;default:false" json:"is_closed"`
}

func (ClinicWorkingHour) TableName() string {
	return "clinic_working_hours"
}

// ClinicVerificationReview tracks superadmin review history
type ClinicVerificationReview struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID   uuid.UUID `gorm:"type:uuid;not null;index" json:"clinic_id"`
	ReviewedBy uuid.UUID `gorm:"type:uuid;not null" json:"reviewed_by"`
	Decision   string    `gorm:"type:varchar(20);not null" json:"decision"`
	Reason     string    `gorm:"type:text" json:"reason,omitempty"`
	ReviewedAt time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"reviewed_at"`

	Reviewer *User   `gorm:"foreignKey:ReviewedBy;constraint:OnDelete:RESTRICT" json:"reviewer,omitempty"`
	Clinic   *Clinic `gorm:"foreignKey:ClinicID;constraint:OnDelete:CASCADE" json:"clinic,omitempty"`
}

func (ClinicVerificationReview) TableName() string {
	return "clinic_verification_reviews"
}

// AuditLog captures administrative and sensitive actions
type AuditLog struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Action     string     `gorm:"type:varchar(100);not null;index" json:"action"`
	EntityType string     `gorm:"type:varchar(50);not null;index" json:"entity_type"`
	EntityID   *uuid.UUID `gorm:"type:uuid;index" json:"entity_id,omitempty"`
	Details    string     `gorm:"type:jsonb" json:"details,omitempty"`
	IPAddress  string     `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	UserAgent  string     `gorm:"type:text" json:"user_agent,omitempty"`
	CreatedAt  time.Time  `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL" json:"user,omitempty"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

