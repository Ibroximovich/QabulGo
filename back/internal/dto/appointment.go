package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateAppointmentRequest payload for booking an appointment
type CreateAppointmentRequest struct {
	ClinicID  uuid.UUID `json:"clinic_id" binding:"required" example:"e202c1aa-9f7b-6492-88b4-75ce67d7808b"`
	DoctorID  uuid.UUID `json:"doctor_id" binding:"required" example:"23f77e06-d956-4789-8c8a-2ccbec8c4a2d"`
	Date      string    `json:"date" binding:"required" example:"2026-11-09"`       // YYYY-MM-DD
	StartTime string    `json:"start_time" binding:"required" example:"10:00"` // HH:mm
	EndTime   string    `json:"end_time" binding:"required" example:"11:00"`   // HH:mm
}

// CancelAppointmentRequest optional reason when cancelling an appointment
type CancelAppointmentRequest struct {
	Reason string `json:"reason" example:"Rejalarim o‘zgarib qoldi"`
}

// AppointmentDoctorDTO sanitized doctor info for appointment responses
type AppointmentDoctorDTO struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Specialty string    `json:"specialty,omitempty"`
	PhotoURL  string    `json:"photo_url,omitempty"`
}

// AppointmentClinicDTO sanitized clinic info for appointment responses
type AppointmentClinicDTO struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Address  string    `json:"address"`
	Region   string    `json:"region"`
	District string    `json:"district"`
	Phone    string    `json:"phone,omitempty"`
}

// AppointmentPatientDTO sanitized patient info for clinic view
type AppointmentPatientDTO struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"`
}

// AppointmentResponseDTO unified appointment detail response
type AppointmentResponseDTO struct {
	ID                 uuid.UUID              `json:"id"`
	PatientID          uuid.UUID              `json:"patient_id"`
	ClinicID           uuid.UUID              `json:"clinic_id"`
	DoctorID           uuid.UUID              `json:"doctor_id"`
	ClinicDoctorID     uuid.UUID              `json:"clinic_doctor_id"`
	AppointmentDate    string                 `json:"appointment_date"`
	StartTime          string                 `json:"start_time"`
	EndTime            string                 `json:"end_time"`
	Status             string                 `json:"status"`
	ConsultationPrice  float64                `json:"consultation_price"`
	CancellationReason *string                `json:"cancellation_reason,omitempty"`
	CancelledAt        *time.Time             `json:"cancelled_at,omitempty"`
	CompletedAt        *time.Time             `json:"completed_at,omitempty"`
	NoShowAt           *time.Time             `json:"no_show_at,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	Doctor             *AppointmentDoctorDTO  `json:"doctor,omitempty"`
	Clinic             *AppointmentClinicDTO  `json:"clinic,omitempty"`
	Patient            *AppointmentPatientDTO `json:"patient,omitempty"`
}

// PatientAppointmentFilter filter criteria for patient's appointments
type PatientAppointmentFilter struct {
	Status    string `form:"status"`    // confirmed, cancelled, completed, no_show
	Timeframe string `form:"timeframe"` // upcoming, past, cancelled
	Page      int    `form:"page,default=1"`
	Limit     int    `form:"limit,default=10"`
}

// ClinicAppointmentFilter filter criteria for clinic's appointments
type ClinicAppointmentFilter struct {
	Date     string     `form:"date"`      // YYYY-MM-DD
	DateFrom string     `form:"date_from"` // YYYY-MM-DD
	DateTo   string     `form:"date_to"`   // YYYY-MM-DD
	DoctorID *uuid.UUID `form:"doctor_id"`
	Status   string     `form:"status"` // confirmed, cancelled, completed, no_show
	Page     int        `form:"page,default=1"`
	Limit    int        `form:"limit,default=20"`
}

// AppointmentListResponseDTO paginated list response
type AppointmentListResponseDTO struct {
	Appointments []AppointmentResponseDTO `json:"appointments"`
	Total        int64                    `json:"total"`
	Page         int                      `json:"page"`
	Limit        int                      `json:"limit"`
}

