package models

import (
	"time"

	"github.com/google/uuid"
)

// DiagnosticAppointmentStatus represents the current status of a diagnostic booking
type DiagnosticAppointmentStatus string

const (
	DiagAppointmentStatusConfirmed DiagnosticAppointmentStatus = "confirmed"
	DiagAppointmentStatusCancelled DiagnosticAppointmentStatus = "cancelled"
	DiagAppointmentStatusCompleted DiagnosticAppointmentStatus = "completed"
	DiagAppointmentStatusNoShow    DiagnosticAppointmentStatus = "no_show"
)

// DiagnosticCategory represents a medical diagnostic category (MRT, MSKT, UZI, etc.)
type DiagnosticCategory struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null;unique" json:"name"`
	Slug        string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"slug"`
	IconURL     string    `gorm:"type:varchar(500)" json:"icon_url"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (DiagnosticCategory) TableName() string {
	return "diagnostic_categories"
}

// DiagnosticEquipment represents a medical machine or diagnostic room in a clinic
type DiagnosticEquipment struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID   uuid.UUID `gorm:"type:uuid;not null;index" json:"clinic_id"`
	CategoryID uuid.UUID `gorm:"type:uuid;not null;index" json:"category_id"`
	Name       string    `gorm:"type:varchar(255);not null" json:"name"`
	RoomNumber string    `gorm:"type:varchar(50)" json:"room_number"`
	IsActive   bool      `gorm:"type:boolean;not null;default:true" json:"is_active"`
	CreatedAt  time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relationships
	Clinic      *Clinic                     `gorm:"foreignKey:ClinicID;constraint:OnDelete:CASCADE" json:"clinic,omitempty"`
	Category    *DiagnosticCategory         `gorm:"foreignKey:CategoryID;constraint:OnDelete:RESTRICT" json:"category,omitempty"`
	Services    []DiagnosticService         `gorm:"foreignKey:EquipmentID;constraint:OnDelete:CASCADE" json:"services,omitempty"`
	WeeklySlots []DiagnosticWeeklySlot      `gorm:"foreignKey:EquipmentID;constraint:OnDelete:CASCADE" json:"weekly_slots,omitempty"`
	Overrides   []DiagnosticDateSlotOverride `gorm:"foreignKey:EquipmentID;constraint:OnDelete:CASCADE" json:"overrides,omitempty"`
}

func (DiagnosticEquipment) TableName() string {
	return "diagnostic_equipments"
}

// DiagnosticService represents a specific diagnostic test offered by a clinic with price and preparation rules
type DiagnosticService struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID         uuid.UUID `gorm:"type:uuid;not null;index" json:"clinic_id"`
	CategoryID       uuid.UUID `gorm:"type:uuid;not null;index" json:"category_id"`
	EquipmentID      uuid.UUID `gorm:"type:uuid;not null;index" json:"equipment_id"`
	Name             string    `gorm:"type:varchar(255);not null" json:"name"`
	Price            float64   `gorm:"type:numeric(12,2);not null;default:0" json:"price"`
	DurationMinutes  int       `gorm:"type:int;not null;default:30" json:"duration_minutes"`
	PreparationRules string    `gorm:"type:text" json:"preparation_rules"`
	IsActive         bool      `gorm:"type:boolean;not null;default:true" json:"is_active"`
	CreatedAt        time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        time.Time `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relationships
	Clinic    *Clinic              `gorm:"foreignKey:ClinicID;constraint:OnDelete:CASCADE" json:"clinic,omitempty"`
	Category  *DiagnosticCategory  `gorm:"foreignKey:CategoryID;constraint:OnDelete:RESTRICT" json:"category,omitempty"`
	Equipment *DiagnosticEquipment `gorm:"foreignKey:EquipmentID;constraint:OnDelete:CASCADE" json:"equipment,omitempty"`
}

func (DiagnosticService) TableName() string {
	return "diagnostic_services"
}

// DiagnosticWeeklySlot defines an equipment's repeating weekly working slot
type DiagnosticWeeklySlot struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EquipmentID uuid.UUID  `gorm:"type:uuid;not null;index" json:"equipment_id"`
	Weekday     int        `gorm:"type:int;not null" json:"weekday"` // 1=Dushanba .. 7=Yakshanba
	StartTime   string     `gorm:"type:varchar(5);not null" json:"start_time"` // "09:00"
	EndTime     string     `gorm:"type:varchar(5);not null" json:"end_time"`   // "09:30"
	Status      SlotStatus `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	CreatedAt   time.Time  `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Equipment *DiagnosticEquipment `gorm:"foreignKey:EquipmentID;constraint:OnDelete:CASCADE" json:"equipment,omitempty"`
}

func (DiagnosticWeeklySlot) TableName() string {
	return "diagnostic_weekly_slots"
}

// DiagnosticDateSlotOverride defines a specific date override for an equipment (maintenance, holidays, extra hours)
type DiagnosticDateSlotOverride struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	EquipmentID uuid.UUID      `gorm:"type:uuid;not null;index" json:"equipment_id"`
	Date        string         `gorm:"type:date;not null" json:"date"` // "YYYY-MM-DD"
	StartTime   *string        `gorm:"type:varchar(5)" json:"start_time,omitempty"`
	EndTime     *string        `gorm:"type:varchar(5)" json:"end_time,omitempty"`
	Status      OverrideStatus `gorm:"type:varchar(20);not null;default:'closed'" json:"status"`
	Reason      string         `gorm:"type:varchar(255)" json:"reason,omitempty"`
	CreatedAt   time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	Equipment *DiagnosticEquipment `gorm:"foreignKey:EquipmentID;constraint:OnDelete:CASCADE" json:"equipment,omitempty"`
}

func (DiagnosticDateSlotOverride) TableName() string {
	return "diagnostic_date_slot_overrides"
}

// DiagnosticAppointment represents a booked diagnostic test appointment
type DiagnosticAppointment struct {
	ID                      uuid.UUID                   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PatientID               uuid.UUID                   `gorm:"type:uuid;not null;index" json:"patient_id"`
	ClinicID                uuid.UUID                   `gorm:"type:uuid;not null;index" json:"clinic_id"`
	ServiceID               uuid.UUID                   `gorm:"type:uuid;not null;index" json:"service_id"`
	EquipmentID             uuid.UUID                   `gorm:"type:uuid;not null;index" json:"equipment_id"`
	AppointmentDate         string                      `gorm:"type:date;not null" json:"appointment_date"` // "YYYY-MM-DD"
	StartTime               string                      `gorm:"type:varchar(5);not null" json:"start_time"`       // "09:00"
	EndTime                 string                      `gorm:"type:varchar(5);not null" json:"end_time"`         // "09:30"
	Price                   float64                     `gorm:"type:numeric(12,2);not null;default:0" json:"price"`
	Status                  DiagnosticAppointmentStatus `gorm:"type:varchar(20);not null;default:'confirmed'" json:"status"`
	PreparationAcknowledged bool                        `gorm:"type:boolean;not null;default:false" json:"preparation_acknowledged"`
	IdempotencyKey          *string                     `gorm:"type:varchar(255);index" json:"idempotency_key,omitempty"`
	CancellationReason      *string                     `gorm:"type:text" json:"cancellation_reason,omitempty"`
	CancelledAt             *time.Time                  `gorm:"type:timestamptz" json:"cancelled_at,omitempty"`
	CompletedAt             *time.Time                  `gorm:"type:timestamptz" json:"completed_at,omitempty"`
	NoShowAt                *time.Time                  `gorm:"type:timestamptz" json:"no_show_at,omitempty"`
	CreatedAt               time.Time                   `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt               time.Time                   `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relationships
	Patient   *User                `gorm:"foreignKey:PatientID;constraint:OnDelete:RESTRICT" json:"patient,omitempty"`
	Clinic    *Clinic              `gorm:"foreignKey:ClinicID;constraint:OnDelete:RESTRICT" json:"clinic,omitempty"`
	Service   *DiagnosticService   `gorm:"foreignKey:ServiceID;constraint:OnDelete:RESTRICT" json:"service,omitempty"`
	Equipment *DiagnosticEquipment `gorm:"foreignKey:EquipmentID;constraint:OnDelete:RESTRICT" json:"equipment,omitempty"`
}

func (DiagnosticAppointment) TableName() string {
	return "diagnostic_appointments"
}

