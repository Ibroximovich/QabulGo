package models

import (
	"time"

	"github.com/google/uuid"
)

type AppointmentStatus string

const (
	AppointmentStatusConfirmed AppointmentStatus = "confirmed"
	AppointmentStatusCancelled AppointmentStatus = "cancelled"
	AppointmentStatusCompleted AppointmentStatus = "completed"
	AppointmentStatusNoShow    AppointmentStatus = "no_show"
)

// Appointment represents a booked medical appointment
type Appointment struct {
	ID                 uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PatientID          uuid.UUID         `gorm:"type:uuid;not null;index" json:"patient_id"`
	ClinicID           uuid.UUID         `gorm:"type:uuid;not null;index" json:"clinic_id"`
	DoctorID           uuid.UUID         `gorm:"type:uuid;not null;index" json:"doctor_id"`
	ClinicDoctorID     uuid.UUID         `gorm:"type:uuid;not null;index" json:"clinic_doctor_id"`
	AppointmentDate    string            `gorm:"type:date;not null" json:"appointment_date"` // "YYYY-MM-DD"
	StartTime          string            `gorm:"type:varchar(5);not null" json:"start_time"`       // "09:00"
	EndTime            string            `gorm:"type:varchar(5);not null" json:"end_time"`         // "10:00"
	Status             AppointmentStatus `gorm:"type:varchar(20);not null;default:'confirmed'" json:"status"`
	ConsultationPrice  float64           `gorm:"type:numeric(12,2);not null;default:0" json:"consultation_price"`
	IdempotencyKey     *string           `gorm:"type:varchar(255);index" json:"idempotency_key,omitempty"`
	CancellationReason *string           `gorm:"type:text" json:"cancellation_reason,omitempty"`
	CancelledAt        *time.Time        `gorm:"type:timestamptz" json:"cancelled_at,omitempty"`
	CompletedAt        *time.Time        `gorm:"type:timestamptz" json:"completed_at,omitempty"`
	NoShowAt           *time.Time        `gorm:"type:timestamptz" json:"no_show_at,omitempty"`
	CreatedAt          time.Time         `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt          time.Time         `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relationships
	Patient      *User         `gorm:"foreignKey:PatientID;constraint:OnDelete:RESTRICT" json:"patient,omitempty"`
	Clinic       *Clinic       `gorm:"foreignKey:ClinicID;constraint:OnDelete:RESTRICT" json:"clinic,omitempty"`
	Doctor       *Doctor       `gorm:"foreignKey:DoctorID;constraint:OnDelete:RESTRICT" json:"doctor,omitempty"`
	ClinicDoctor *ClinicDoctor `gorm:"foreignKey:ClinicDoctorID;constraint:OnDelete:RESTRICT" json:"clinic_doctor,omitempty"`
}

func (Appointment) TableName() string {
	return "appointments"
}

