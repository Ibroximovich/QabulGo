package dto

import (
	"time"

	"github.com/google/uuid"
)

// TimeBreakItem defines an excluded break interval during auto generation
type TimeBreakItem struct {
	StartTime string `json:"start_time" binding:"required" example:"13:00"`
	EndTime   string `json:"end_time" binding:"required" example:"14:00"`
}

// GenerateScheduleRequest payload for calculating slot previews
type GenerateScheduleRequest struct {
	Weekdays            []int           `json:"weekdays" binding:"required,min=1" example:"[1,2,3,4,5]"`
	StartTime           string          `json:"start_time" binding:"required" example:"09:00"`
	EndTime             string          `json:"end_time" binding:"required" example:"17:00"`
	SlotDurationMinutes int             `json:"slot_duration_minutes" binding:"required,min=5,max=240" example:"60"`
	Breaks              []TimeBreakItem `json:"breaks"`
}

// GeneratedSlotItem represents a single generated slot in preview
type GeneratedSlotItem struct {
	Weekday   int    `json:"weekday" example:"1"`
	StartTime string `json:"start_time" example:"09:00"`
	EndTime   string `json:"end_time" example:"10:00"`
	Status    string `json:"status" example:"active"`
}

// GenerateScheduleResponse response for previewing calculated slots
type GenerateScheduleResponse struct {
	TotalSlots int                 `json:"total_slots"`
	Slots      []GeneratedSlotItem `json:"slots"`
}

// WeeklySlotItemRequest a single slot item when saving weekly schedule
type WeeklySlotItemRequest struct {
	Weekday   int    `json:"weekday" binding:"required,min=1,max=7" example:"1"`
	StartTime string `json:"start_time" binding:"required" example:"09:00"`
	EndTime   string `json:"end_time" binding:"required" example:"10:00"`
	Status    string `json:"status,omitempty" example:"active"`
}

// SaveWeeklyScheduleRequest atomic replacement of doctor's weekly slots
type SaveWeeklyScheduleRequest struct {
	Slots []WeeklySlotItemRequest `json:"slots" binding:"required"`
}

// DoctorWeeklySlotDTO represents a persisted weekly slot
type DoctorWeeklySlotDTO struct {
	ID             uuid.UUID `json:"id"`
	ClinicDoctorID uuid.UUID `json:"clinic_doctor_id"`
	Weekday        int       `json:"weekday"`
	StartTime      string    `json:"start_time"`
	EndTime        string    `json:"end_time"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// DoctorWeeklyScheduleDTO full weekly schedule response
type DoctorWeeklyScheduleDTO struct {
	ClinicDoctorID uuid.UUID             `json:"clinic_doctor_id"`
	DoctorID       uuid.UUID             `json:"doctor_id"`
	ClinicID       uuid.UUID             `json:"clinic_id"`
	Slots          []DoctorWeeklySlotDTO `json:"slots"`
}

// CreateDateOverrideRequest creates a single date exception/override
type CreateDateOverrideRequest struct {
	Date      string  `json:"date" binding:"required" example:"2026-11-05"` // YYYY-MM-DD
	StartTime *string `json:"start_time,omitempty" example:"10:00"`
	EndTime   *string `json:"end_time,omitempty" example:"11:00"`
	Status    string  `json:"status" binding:"required,oneof=closed available day_off" example:"closed"`
	Reason    string  `json:"reason,omitempty" example:"Ta'til"`
}

// UpdateDateOverrideRequest updates an existing date exception
type UpdateDateOverrideRequest struct {
	Date      *string `json:"date,omitempty" example:"2026-11-05"`
	StartTime *string `json:"start_time,omitempty" example:"11:00"`
	EndTime   *string `json:"end_time,omitempty" example:"12:00"`
	Status    *string `json:"status,omitempty" binding:"omitempty,oneof=closed available day_off" example:"closed"`
	Reason    *string `json:"reason,omitempty" example:"Shaxsiy sabab"`
}

// DoctorDateOverrideDTO persisted date override
type DoctorDateOverrideDTO struct {
	ID             uuid.UUID `json:"id"`
	ClinicDoctorID uuid.UUID `json:"clinic_doctor_id"`
	Date           string    `json:"date"`
	StartTime      *string   `json:"start_time,omitempty"`
	EndTime        *string   `json:"end_time,omitempty"`
	Status         string    `json:"status"`
	Reason         string    `json:"reason,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// AvailabilitySlotDTO individual public availability time slot
type AvailabilitySlotDTO struct {
	StartTime string `json:"start_time" example:"09:00"`
	EndTime   string `json:"end_time" example:"10:00"`
}

// PublicAvailabilityResponse public endpoint response for doctor availability on a given date
type PublicAvailabilityResponse struct {
	Date           string                `json:"date" example:"2026-11-05"`
	Timezone       string                `json:"timezone" example:"Asia/Tashkent"`
	ClinicID       uuid.UUID             `json:"clinic_id"`
	DoctorID       uuid.UUID             `json:"doctor_id"`
	ClinicDoctorID uuid.UUID             `json:"clinic_doctor_id"`
	Slots          []AvailabilitySlotDTO `json:"slots"`
}

