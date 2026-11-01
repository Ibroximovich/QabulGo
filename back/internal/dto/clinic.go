package dto

import (
	"time"

	"github.com/google/uuid"
)

// ClinicRegisterRequest payload for clinic owner registration
type ClinicRegisterRequest struct {
	FirstName       string `json:"first_name" binding:"required" example:"Jamshid"`
	LastName        string `json:"last_name" binding:"required" example:"Karimov"`
	Email           string `json:"email" binding:"required" example:"jamshid@clinic.uz"`
	Phone           string `json:"phone" binding:"required" example:"+998901234567"`
	Password        string `json:"password" binding:"required" example:"ClinicPass123!"`
	ConfirmPassword string `json:"confirm_password" binding:"required" example:"ClinicPass123!"`
	AcceptTerms     bool   `json:"accept_terms" example:"true"`
}

// ClinicLoginRequest payload for clinic owner or admin login
type ClinicLoginRequest struct {
	Email    string `json:"email" binding:"required" example:"jamshid@clinic.uz"`
	Password string `json:"password" binding:"required" example:"ClinicPass123!"`
}

// SuperAdminLoginRequest payload for superadmin login
type SuperAdminLoginRequest struct {
	Email    string `json:"email" binding:"required" example:"admin@qabulgo.uz"`
	Password string `json:"password" binding:"required" example:"SuperAdminPass123!"`
}

// WorkingHourItem represents working hour for a day
type WorkingHourItem struct {
	Weekday     int    `json:"weekday" binding:"required,min=1,max=7" example:"1"` // 1=Dushanba, 7=Yakshanba
	OpeningTime string `json:"opening_time" example:"08:00"`
	ClosingTime string `json:"closing_time" example:"18:00"`
	IsClosed    bool   `json:"is_closed" example:"false"`
}

// CreateClinicRequest represents request to create a new clinic
type CreateClinicRequest struct {
	Name          string            `json:"name" binding:"required" example:"Shifo Nur Med"`
	Description   string            `json:"description" example:"Zamonaviy ko'p tarmoqli xususiy klinika"`
	Phone         string            `json:"phone" binding:"required" example:"+998712001122"`
	Email         string            `json:"email" example:"info@shifonur.uz"`
	Region        string            `json:"region" binding:"required" example:"Toshkent shahri"`
	District      string            `json:"district" binding:"required" example:"Yunusobod tumani"`
	Address       string            `json:"address" binding:"required" example:"Amir Temur ko'chasi, 45-uy"`
	LocationURL   string            `json:"location_url,omitempty" example:"https://maps.app.goo.gl/xyz"`
	LogoURL       string            `json:"logo_url" example:"/uploads/clinics/logo-123.jpg"`
	CoverImageURL string            `json:"cover_image_url" example:"/uploads/clinics/cover-123.jpg"`
	WorkingHours  []WorkingHourItem `json:"working_hours,omitempty"`
}

// UpdateClinicRequest represents request to update clinic profile
type UpdateClinicRequest struct {
	Name          *string            `json:"name,omitempty" example:"Shifo Nur Med Clinic"`
	Description   *string            `json:"description,omitempty"`
	Phone         *string            `json:"phone,omitempty" example:"+998712001122"`
	Email         *string            `json:"email,omitempty"`
	Region        *string            `json:"region,omitempty" example:"Toshkent shahri"`
	District      *string            `json:"district,omitempty" example:"Yunusobod tumani"`
	Address       *string            `json:"address,omitempty"`
	LocationURL   *string            `json:"location_url,omitempty" example:"https://maps.app.goo.gl/xyz"`
	LogoURL       *string            `json:"logo_url,omitempty"`
	CoverImageURL *string            `json:"cover_image_url,omitempty"`
	WorkingHours  *[]WorkingHourItem `json:"working_hours,omitempty"`
}

// ClinicSummaryDTO represents public/summary view of clinic
type ClinicSummaryDTO struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Phone         string    `json:"phone"`
	Email         string    `json:"email"`
	Region        string    `json:"region"`
	District      string    `json:"district"`
	Address       string    `json:"address"`
	LocationURL   string    `json:"location_url,omitempty"`
	LogoURL       string    `json:"logo_url"`
	CoverImageURL string    `json:"cover_image_url"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

// ClinicDetailDTO represents full clinic detail with working hours and doctors
type ClinicDetailDTO struct {
	ID              uuid.UUID             `json:"id"`
	Name            string                `json:"name"`
	Description     string                `json:"description"`
	Phone           string                `json:"phone"`
	Email           string                `json:"email"`
	Region          string                `json:"region"`
	District        string                `json:"district"`
	Address         string                `json:"address"`
	LocationURL     string                `json:"location_url,omitempty"`
	LogoURL         string                `json:"logo_url"`
	CoverImageURL   string                `json:"cover_image_url"`
	Status          string                `json:"status"`
	RejectionReason string                `json:"rejection_reason,omitempty"`
	OwnerID         uuid.UUID             `json:"owner_id"`
	WorkingHours    []WorkingHourItem     `json:"working_hours"`
	DoctorsCount    int                   `json:"doctors_count"`
	Doctors         []ClinicDoctorItemDTO `json:"doctors,omitempty"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
}

// SpecialtyDTO represents a medical specialty
type SpecialtyDTO struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Slug    string    `json:"slug"`
	IconURL string    `json:"icon_url"`
}

// CreateDoctorRequest represents request to add a doctor to a clinic
type CreateDoctorRequest struct {
	FirstName                  string    `json:"first_name" binding:"required" example:"Anvar"`
	LastName                   string    `json:"last_name" binding:"required" example:"Saidov"`
	SpecialtyID                uuid.UUID `json:"specialty_id" binding:"required" example:"a1b2c3d4-e5f6-7890-abcd-ef1234567890"`
	Bio                        string    `json:"bio" example:"15 yillik tajribaga ega oliy toifali kardiolog"`
	ExperienceYears            int       `json:"experience_years" binding:"min=0,max=70" example:"15"`
	PhotoURL                   string    `json:"photo_url" example:"/uploads/doctors/doctor-123.jpg"`
	ConsultationPrice          float64   `json:"consultation_price" binding:"required,min=0" example:"150000"`
	AppointmentDurationMinutes int       `json:"appointment_duration_minutes" binding:"required,min=5,max=240" example:"30"`
	IsActive                   *bool     `json:"is_active,omitempty" example:"true"`
}

// UpdateDoctorRequest represents request to update doctor in a clinic
type UpdateDoctorRequest struct {
	FirstName                  *string    `json:"first_name,omitempty" example:"Anvar"`
	LastName                   *string    `json:"last_name,omitempty" example:"Saidov"`
	SpecialtyID                *uuid.UUID `json:"specialty_id,omitempty"`
	Bio                        *string    `json:"bio,omitempty"`
	ExperienceYears            *int       `json:"experience_years,omitempty"`
	PhotoURL                   *string    `json:"photo_url,omitempty"`
	ConsultationPrice          *float64   `json:"consultation_price,omitempty"`
	AppointmentDurationMinutes *int       `json:"appointment_duration_minutes,omitempty"`
	IsActive                   *bool      `json:"is_active,omitempty"`
}

// ClinicDoctorItemDTO represents doctor details in a clinic
type ClinicDoctorItemDTO struct {
	ID                         uuid.UUID     `json:"id"` // clinic_doctor id
	DoctorID                   uuid.UUID     `json:"doctor_id"`
	ClinicID                   uuid.UUID     `json:"clinic_id"`
	FirstName                  string        `json:"first_name"`
	LastName                   string        `json:"last_name"`
	Bio                        string        `json:"bio"`
	ExperienceYears            int           `json:"experience_years"`
	PhotoURL                   string        `json:"photo_url"`
	Specialty                  *SpecialtyDTO `json:"specialty,omitempty"`
	ConsultationPrice          float64       `json:"consultation_price"`
	AppointmentDurationMinutes int           `json:"appointment_duration_minutes"`
	IsActive                   bool          `json:"is_active"`
	CreatedAt                  time.Time     `json:"created_at"`
	UpdatedAt                  time.Time     `json:"updated_at"`
}

// DoctorFilter query params for doctor list
type DoctorFilter struct {
	SpecialtyID *uuid.UUID `form:"specialty_id"`
	IsActive    *bool      `form:"is_active"`
	Search      string     `form:"search"`
	Page        int        `form:"page,default=1"`
	Limit       int        `form:"limit,default=20"`
}

// AdminClinicFilter query params for superadmin clinic list
type AdminClinicFilter struct {
	Status string `form:"status"`
	Search string `form:"search"`
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=20"`
}

// AdminReviewClinicRequest represents rejection or suspend reason
type AdminReviewClinicRequest struct {
	Reason string `json:"reason" example:"Tibbiy litsenziya ma'lumotlari kiritilmagan"`
}

// UploadResponse represents file upload result
type UploadResponse struct {
	URL      string `json:"url" example:"/uploads/clinics/uuid.jpg"`
	Filename string `json:"filename" example:"uuid.jpg"`
	Size     int64  `json:"size" example:"102400"`
	MimeType string `json:"mime_type" example:"image/jpeg"`
}

