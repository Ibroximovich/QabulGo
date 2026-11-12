package dto

import (
	"time"

	"github.com/google/uuid"
)

// --- Category DTOs ---

type DiagnosticCategoryDTO struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	IconURL     string    `json:"icon_url"`
	Description string    `json:"description"`
}

type ListDiagnosticCategoriesResponse struct {
	Categories []DiagnosticCategoryDTO `json:"categories"`
	Total      int                     `json:"total"`
}

// --- Equipment DTOs ---

type CreateDiagnosticEquipmentRequest struct {
	CategoryID uuid.UUID `json:"category_id" binding:"required"`
	Name       string    `json:"name" binding:"required"`
	RoomNumber string    `json:"room_number"`
	IsActive   *bool     `json:"is_active"`
}

type UpdateDiagnosticEquipmentRequest struct {
	CategoryID *uuid.UUID `json:"category_id"`
	Name       *string    `json:"name"`
	RoomNumber *string    `json:"room_number"`
	IsActive   *bool      `json:"is_active"`
}

type DiagnosticEquipmentDTO struct {
	ID         uuid.UUID              `json:"id"`
	ClinicID   uuid.UUID              `json:"clinic_id"`
	CategoryID uuid.UUID              `json:"category_id"`
	Category   *DiagnosticCategoryDTO `json:"category,omitempty"`
	Name       string                 `json:"name"`
	RoomNumber string                 `json:"room_number"`
	IsActive   bool                   `json:"is_active"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

type ListDiagnosticEquipmentsResponse struct {
	Equipments []DiagnosticEquipmentDTO `json:"equipments"`
	Total      int                      `json:"total"`
}

// --- Service DTOs ---

type CreateDiagnosticServiceRequest struct {
	CategoryID       uuid.UUID `json:"category_id" binding:"required"`
	EquipmentID      uuid.UUID `json:"equipment_id" binding:"required"`
	Name             string    `json:"name" binding:"required"`
	Price            float64   `json:"price" binding:"required,min=0"`
	DurationMinutes  int       `json:"duration_minutes" binding:"required,min=5,max=240"`
	PreparationRules string    `json:"preparation_rules"`
	IsActive         *bool     `json:"is_active"`
}

type UpdateDiagnosticServiceRequest struct {
	CategoryID       *uuid.UUID `json:"category_id"`
	EquipmentID      *uuid.UUID `json:"equipment_id"`
	Name             *string    `json:"name"`
	Price            *float64   `json:"price"`
	DurationMinutes  *int       `json:"duration_minutes"`
	PreparationRules *string    `json:"preparation_rules"`
	IsActive         *bool      `json:"is_active"`
}

type DiagnosticServiceDTO struct {
	ID               uuid.UUID               `json:"id"`
	ClinicID         uuid.UUID               `json:"clinic_id"`
	CategoryID       uuid.UUID               `json:"category_id"`
	EquipmentID      uuid.UUID               `json:"equipment_id"`
	Category         *DiagnosticCategoryDTO  `json:"category,omitempty"`
	Equipment        *DiagnosticEquipmentDTO `json:"equipment,omitempty"`
	Name             string                  `json:"name"`
	Price            float64                 `json:"price"`
	DurationMinutes  int                     `json:"duration_minutes"`
	PreparationRules string                  `json:"preparation_rules"`
	IsActive         bool                    `json:"is_active"`
	CreatedAt        time.Time               `json:"created_at"`
	UpdatedAt        time.Time               `json:"updated_at"`
}

type ListDiagnosticServicesResponse struct {
	Services []DiagnosticServiceDTO `json:"services"`
	Total    int                    `json:"total"`
}

// --- Schedule & Slots DTOs ---

type DiagnosticWeeklySlotDTO struct {
	ID          uuid.UUID `json:"id"`
	EquipmentID uuid.UUID `json:"equipment_id"`
	Weekday     int       `json:"weekday"`
	StartTime   string    `json:"start_time"`
	EndTime     string    `json:"end_time"`
	Status      string    `json:"status"`
}

type SaveDiagnosticWeeklySlotItem struct {
	Weekday   int    `json:"weekday" binding:"required,min=1,max=7"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
	Status    string `json:"status"`
}

type SaveDiagnosticWeeklySlotsRequest struct {
	Slots []SaveDiagnosticWeeklySlotItem `json:"slots" binding:"required"`
}

type DiagnosticDateOverrideDTO struct {
	ID          uuid.UUID `json:"id"`
	EquipmentID uuid.UUID `json:"equipment_id"`
	Date        string    `json:"date"`
	StartTime   *string   `json:"start_time,omitempty"`
	EndTime     *string   `json:"end_time,omitempty"`
	Status      string    `json:"status"`
	Reason      string    `json:"reason,omitempty"`
}

type CreateDiagnosticDateOverrideRequest struct {
	Date      string  `json:"date" binding:"required"`
	StartTime *string `json:"start_time"`
	EndTime   *string `json:"end_time"`
	Status    string  `json:"status" binding:"required"`
	Reason    string  `json:"reason"`
}

type DiagnosticAvailabilitySlotDTO struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type DiagnosticAvailabilityResponse struct {
	Date        string                          `json:"date"`
	Timezone    string                          `json:"timezone"`
	ClinicID    uuid.UUID                       `json:"clinic_id"`
	EquipmentID uuid.UUID                       `json:"equipment_id"`
	Slots       []DiagnosticAvailabilitySlotDTO `json:"slots"`
}

// --- Appointment DTOs ---

type CreateDiagnosticAppointmentRequest struct {
	ClinicID                uuid.UUID `json:"clinic_id" binding:"required"`
	ServiceID               uuid.UUID `json:"service_id" binding:"required"`
	EquipmentID             uuid.UUID `json:"equipment_id" binding:"required"`
	Date                    string    `json:"date" binding:"required"`       // YYYY-MM-DD
	StartTime               string    `json:"start_time" binding:"required"` // HH:mm
	EndTime                 string    `json:"end_time" binding:"required"`   // HH:mm
	PreparationAcknowledged bool      `json:"preparation_acknowledged" binding:"required"`
}

type CancelDiagnosticAppointmentRequest struct {
	Reason string `json:"reason"`
}

type DiagnosticAppointmentResponseDTO struct {
	ID                      uuid.UUID               `json:"id"`
	PatientID               uuid.UUID               `json:"patient_id"`
	ClinicID                uuid.UUID               `json:"clinic_id"`
	ServiceID               uuid.UUID               `json:"service_id"`
	EquipmentID             uuid.UUID               `json:"equipment_id"`
	AppointmentDate         string                  `json:"appointment_date"`
	StartTime               string                  `json:"start_time"`
	EndTime                 string                  `json:"end_time"`
	Price                   float64                 `json:"price"`
	Status                  string                  `json:"status"`
	PreparationAcknowledged bool                    `json:"preparation_acknowledged"`
	PreparationRules        string                  `json:"preparation_rules,omitempty"`
	CancellationReason      *string                 `json:"cancellation_reason,omitempty"`
	CancelledAt             *time.Time              `json:"cancelled_at,omitempty"`
	CompletedAt             *time.Time              `json:"completed_at,omitempty"`
	NoShowAt                *time.Time              `json:"no_show_at,omitempty"`
	CreatedAt               time.Time               `json:"created_at"`
	UpdatedAt               time.Time               `json:"updated_at"`
	Service                 *DiagnosticServiceDTO   `json:"service,omitempty"`
	Equipment               *DiagnosticEquipmentDTO `json:"equipment,omitempty"`
	Clinic                  *AppointmentClinicDTO   `json:"clinic,omitempty"`
	Patient                 *AppointmentPatientDTO  `json:"patient,omitempty"`
}

type PatientDiagnosticAppointmentFilter struct {
	Timeframe string `form:"timeframe"` // upcoming, past, cancelled
	Status    string `form:"status"`
	Page      int    `form:"page,default=1"`
	Limit     int    `form:"limit,default=10"`
}

type ClinicDiagnosticAppointmentFilter struct {
	EquipmentID *uuid.UUID `form:"equipment_id"`
	ServiceID   *uuid.UUID `form:"service_id"`
	Date        string     `form:"date"`
	DateFrom    string     `form:"date_from"`
	DateTo      string     `form:"date_to"`
	Status      string     `form:"status"`
	Page        int        `form:"page,default=1"`
	Limit       int        `form:"limit,default=10"`
}

type DiagnosticAppointmentListResponseDTO struct {
	Appointments []DiagnosticAppointmentResponseDTO `json:"appointments"`
	Total        int64                              `json:"total"`
	Page         int                                `json:"page"`
	Limit        int                                `json:"limit"`
}
