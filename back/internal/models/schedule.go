package models

import (
	"time"

	"github.com/google/uuid"
)

type SlotStatus string

const (
	SlotStatusActive   SlotStatus = "active"
	SlotStatusInactive SlotStatus = "inactive"
)

type OverrideStatus string

const (
	OverrideStatusClosed    OverrideStatus = "closed"    // Individual slot closed
	OverrideStatusAvailable OverrideStatus = "available" // Extra slot added
	OverrideStatusDayOff    OverrideStatus = "day_off"   // Entire day closed
)

// DoctorWeeklySlot defines a doctor's repeating weekly working slot in a specific clinic
type DoctorWeeklySlot struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicDoctorID uuid.UUID  `gorm:"type:uuid;not null;index" json:"clinic_doctor_id"`
	Weekday        int        `gorm:"type:int;not null" json:"weekday"` // 1=Dushanba .. 7=Yakshanba
	StartTime      string     `gorm:"type:varchar(5);not null" json:"start_time"` // "09:00"
	EndTime        string     `gorm:"type:varchar(5);not null" json:"end_time"`   // "10:00"
	Status         SlotStatus `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	CreatedAt      time.Time  `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	ClinicDoctor *ClinicDoctor `gorm:"foreignKey:ClinicDoctorID;constraint:OnDelete:CASCADE" json:"clinic_doctor,omitempty"`
}

func (DoctorWeeklySlot) TableName() string {
	return "doctor_weekly_slots"
}

// DoctorDateSlotOverride defines a specific date override (closed slot, extra slot, or whole day off)
type DoctorDateSlotOverride struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicDoctorID uuid.UUID      `gorm:"type:uuid;not null;index" json:"clinic_doctor_id"`
	Date           string         `gorm:"type:date;not null" json:"date"` // "YYYY-MM-DD"
	StartTime      *string        `gorm:"type:varchar(5)" json:"start_time,omitempty"`
	EndTime        *string        `gorm:"type:varchar(5)" json:"end_time,omitempty"`
	Status         OverrideStatus `gorm:"type:varchar(20);not null;default:'closed'" json:"status"`
	Reason         string         `gorm:"type:varchar(255)" json:"reason,omitempty"`
	CreatedAt      time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	ClinicDoctor *ClinicDoctor `gorm:"foreignKey:ClinicDoctorID;constraint:OnDelete:CASCADE" json:"clinic_doctor,omitempty"`
}

func (DoctorDateSlotOverride) TableName() string {
	return "doctor_date_slot_overrides"
}

