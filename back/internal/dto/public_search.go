package dto

import (
	"time"

	"github.com/google/uuid"
)

// PublicClinicFilter for searching approved clinics
type PublicClinicFilter struct {
	Search      string  `form:"search"`
	Region      string  `form:"region"`
	District    string  `form:"district"`
	SpecialtyID *string `form:"specialty_id"`
	Page        int     `form:"page,default=1"`
	Limit       int     `form:"limit,default=12"`
}

// PublicClinicItemDTO represents clinic summary in public listings
type PublicClinicItemDTO struct {
	ID            uuid.UUID         `json:"id"`
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	Phone         string            `json:"phone"`
	Email         string            `json:"email,omitempty"`
	Region        string            `json:"region"`
	District      string            `json:"district"`
	Address       string            `json:"address"`
	LocationURL   string            `json:"location_url,omitempty"`
	LogoURL       string            `json:"logo_url"`
	CoverImageURL string            `json:"cover_image_url"`
	Latitude      *float64          `json:"latitude,omitempty"`
	Longitude     *float64          `json:"longitude,omitempty"`
	WorkingHours  []WorkingHourItem `json:"working_hours,omitempty"`
	DoctorsCount  int               `json:"doctors_count"`
	Specialties   []SpecialtyDTO    `json:"specialties,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
}

// NearbyClinicFilter query parameters for searching clinics near a location
type NearbyClinicFilter struct {
	Latitude    float64 `form:"latitude" binding:"required"`
	Longitude   float64 `form:"longitude" binding:"required"`
	RadiusKM    float64 `form:"radius_km,default=10"`
	SpecialtyID *string `form:"specialty_id"`
	Page        int     `form:"page,default=1"`
	Limit       int     `form:"limit,default=12"`
}

// NearbyClinicItemDTO represents a clinic with calculated distance from user
type NearbyClinicItemDTO struct {
	ID            uuid.UUID      `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Phone         string         `json:"phone"`
	Email         string         `json:"email,omitempty"`
	Region        string         `json:"region"`
	District      string         `json:"district"`
	Address       string         `json:"address"`
	LocationURL   string         `json:"location_url,omitempty"`
	LogoURL       string         `json:"logo_url"`
	CoverImageURL string         `json:"cover_image_url"`
	Latitude      float64        `json:"latitude"`
	Longitude     float64        `json:"longitude"`
	DistanceKM    float64        `json:"distance_km"` // To'g'ri chiziq bo'yicha masofa (km)
	DoctorsCount  int            `json:"doctors_count"`
	Specialties   []SpecialtyDTO `json:"specialties,omitempty"`
}

// PublicDoctorFilter for searching doctors across approved clinics
type PublicDoctorFilter struct {
	Search      string   `form:"search"`
	Region      string   `form:"region"`
	District    string   `form:"district"`
	SpecialtyID *string  `form:"specialty_id"`
	ClinicID    *string  `form:"clinic_id"`
	MinPrice    *float64 `form:"min_price"`
	MaxPrice    *float64 `form:"max_price"`
	Sort        string   `form:"sort,default=name_asc"` // price_asc, price_desc, experience_desc, name_asc
	Page        int      `form:"page,default=1"`
	Limit       int      `form:"limit,default=12"`
}

// PublicDoctorClinicDTO clinic summary embedded inside public doctor search result
type PublicDoctorClinicDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Region      string    `json:"region"`
	District    string    `json:"district"`
	Address     string    `json:"address"`
	Phone       string    `json:"phone"`
	LocationURL string    `json:"location_url,omitempty"`
	LogoURL     string    `json:"logo_url"`
	Latitude    *float64  `json:"latitude,omitempty"`
	Longitude   *float64  `json:"longitude,omitempty"`
}

// PublicDoctorItemDTO represents a doctor in a specific clinic
type PublicDoctorItemDTO struct {
	ClinicDoctorID             uuid.UUID             `json:"clinic_doctor_id"` // clinic_doctor row ID
	DoctorID                   uuid.UUID             `json:"doctor_id"`
	FirstName                  string                `json:"first_name"`
	LastName                   string                `json:"last_name"`
	PhotoURL                   string                `json:"photo_url"`
	Bio                        string                `json:"bio"`
	ExperienceYears            int                   `json:"experience_years"`
	Specialty                  *SpecialtyDTO         `json:"specialty"`
	ConsultationPrice          float64               `json:"consultation_price"`
	AppointmentDurationMinutes int                   `json:"appointment_duration_minutes"`
	Clinic                     PublicDoctorClinicDTO `json:"clinic"`
}

// DoctorClinicWorkplaceDTO workplace of a doctor
type DoctorClinicWorkplaceDTO struct {
	ClinicDoctorID             uuid.UUID      `json:"clinic_doctor_id"`
	ClinicID                   uuid.UUID      `json:"clinic_id"`
	ClinicName                 string         `json:"clinic_name"`
	Region                     string         `json:"region"`
	District                   string         `json:"district"`
	Address                    string         `json:"address"`
	Phone                      string         `json:"phone"`
	LocationURL                string         `json:"location_url,omitempty"`
	LogoURL                    string         `json:"logo_url"`
	Latitude                   *float64       `json:"latitude,omitempty"`
	Longitude                  *float64       `json:"longitude,omitempty"`
	Specialty                  *SpecialtyDTO  `json:"specialty"`
	ConsultationPrice          float64        `json:"consultation_price"`
	AppointmentDurationMinutes int            `json:"appointment_duration_minutes"`
}

// PublicDoctorDetailDTO full detail of a doctor across all their approved workplaces
type PublicDoctorDetailDTO struct {
	DoctorID        uuid.UUID                  `json:"doctor_id"`
	FirstName       string                     `json:"first_name"`
	LastName        string                     `json:"last_name"`
	PhotoURL        string                     `json:"photo_url"`
	Bio             string                     `json:"bio"`
	ExperienceYears int                        `json:"experience_years"`
	PrimarySpecialty *SpecialtyDTO             `json:"primary_specialty,omitempty"`
	Workplaces      []DoctorClinicWorkplaceDTO `json:"workplaces"`
}

// RegionItemDTO catalog item for regions and districts
type RegionItemDTO struct {
	Region    string   `json:"region"`
	Districts []string `json:"districts"`
}

