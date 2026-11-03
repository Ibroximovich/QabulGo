package dto

import (
	"time"

	"github.com/google/uuid"
)

// AdminDashboardSummaryDTO represents high-level KPI metrics across the platform
type AdminDashboardSummaryDTO struct {
	TotalClinics          int64   `json:"total_clinics"`
	ApprovedClinics       int64   `json:"approved_clinics"`
	PendingClinics        int64   `json:"pending_clinics"`
	RejectedClinics       int64   `json:"rejected_clinics"`
	SuspendedClinics      int64   `json:"suspended_clinics"`
	TotalPatients         int64   `json:"total_patients"`
	TotalDoctors          int64   `json:"total_doctors"`
	TotalAppointments     int64   `json:"total_appointments"`
	TodayAppointments     int64   `json:"today_appointments"`
	ThisWeekAppointments int64   `json:"this_week_appointments"`
	ThisMonthAppointments int64   `json:"this_month_appointments"`
	ConfirmedAppointments int64   `json:"confirmed_appointments"`
	CancelledAppointments int64   `json:"cancelled_appointments"`
	CompletedAppointments int64   `json:"completed_appointments"`
	NoShowAppointments    int64   `json:"no_show_appointments"`
	EstimatedTotalRevenue float64 `json:"estimated_total_revenue"` // "Bronlarning taxminiy umumiy qiymati"
}

// AdminDashboardAnalyticsFilter for querying chart analytics
type AdminDashboardAnalyticsFilter struct {
	Range     string `form:"range"` // today, 7d, 30d, 90d, custom
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

// DailyMetricDTO represents aggregated metrics for a specific date
type DailyMetricDTO struct {
	Date   string  `json:"date"`
	Count  int64   `json:"count"`
	Amount float64 `json:"amount,omitempty"`
}

// StatusCountDTO represents appointment counts grouped by status
type StatusCountDTO struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

// TopClinicMetricDTO represents top-performing clinics by bookings
type TopClinicMetricDTO struct {
	ClinicID      uuid.UUID `json:"clinic_id"`
	ClinicName    string    `json:"clinic_name"`
	BookingsCount int64     `json:"bookings_count"`
	TotalRevenue  float64   `json:"total_revenue"`
}

// AdminDashboardAnalyticsDTO contains all chart series for the dashboard
type AdminDashboardAnalyticsDTO struct {
	DailyBookings      []DailyMetricDTO     `json:"daily_bookings"`
	WeeklyBookings     []DailyMetricDTO     `json:"weekly_bookings"`
	MonthlyBookings    []DailyMetricDTO     `json:"monthly_bookings"`
	ClinicGrowth       []DailyMetricDTO     `json:"clinic_growth"`
	PatientGrowth      []DailyMetricDTO     `json:"patient_growth"`
	StatusDistribution []StatusCountDTO     `json:"status_distribution"`
	TopClinics         []TopClinicMetricDTO `json:"top_clinics"`
}

// AdminPatientFilter for searching and paginating patients
type AdminPatientFilter struct {
	Search string `form:"search"`
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
}

func (f *AdminPatientFilter) Normalize() {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
}

// AdminPatientListItemDTO represents a patient in the table list
type AdminPatientListItemDTO struct {
	ID                  uuid.UUID `json:"id"`
	FirstName           string    `json:"first_name"`
	LastName            string    `json:"last_name"`
	Email               string    `json:"email"`
	Phone               string    `json:"phone"`
	Status              string    `json:"status"`
	CreatedAt           time.Time `json:"created_at"`
	AppointmentsCount   int64     `json:"appointments_count"`
	LastAppointmentDate *string   `json:"last_appointment_date,omitempty"`
}

// AdminPatientDetailDTO represents a single patient's profile and booking history
type AdminPatientDetailDTO struct {
	ID                    uuid.UUID                     `json:"id"`
	FirstName             string                        `json:"first_name"`
	LastName              string                        `json:"last_name"`
	Email                 string                        `json:"email"`
	Phone                 string                        `json:"phone"`
	Status                string                        `json:"status"`
	CreatedAt             time.Time                     `json:"created_at"`
	TotalAppointments     int64                         `json:"total_appointments"`
	UpcomingAppointments  int64                         `json:"upcoming_appointments"`
	CompletedAppointments int64                         `json:"completed_appointments"`
	CancelledAppointments int64                         `json:"cancelled_appointments"`
	RecentAppointments    []AdminAppointmentListItemDTO `json:"recent_appointments"`
}

// AdminDoctorFilter for searching and paginating doctors
type AdminDoctorFilter struct {
	Search      string `form:"search"`
	SpecialtyID string `form:"specialty_id"`
	ClinicID    string `form:"clinic_id"`
	Page        int    `form:"page"`
	Limit       int    `form:"limit"`
}

func (f *AdminDoctorFilter) Normalize() {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
}

// AdminDoctorListItemDTO represents unique doctor with aggregated clinic/specialty data
type AdminDoctorListItemDTO struct {
	ID                uuid.UUID `json:"id"`
	FirstName         string    `json:"first_name"`
	LastName          string    `json:"last_name"`
	Bio               string    `json:"bio"`
	ExperienceYears   int       `json:"experience_years"`
	PhotoURL          string    `json:"photo_url"`
	Specialties       []string  `json:"specialties"`
	Clinics           []string  `json:"clinics"`
	ClinicsCount      int       `json:"clinics_count"`
	TotalAppointments int64     `json:"total_appointments"`
	CreatedAt         time.Time `json:"created_at"`
}

// AdminDoctorClinicInfoDTO represents clinic-specific attachment of a doctor
type AdminDoctorClinicInfoDTO struct {
	ClinicID          uuid.UUID `json:"clinic_id"`
	ClinicName        string    `json:"clinic_name"`
	ClinicDoctorID    uuid.UUID `json:"clinic_doctor_id"`
	SpecialtyName     string    `json:"specialty_name"`
	ConsultationPrice float64   `json:"consultation_price"`
	IsActive          bool      `json:"is_active"`
}

// AdminDoctorDetailDTO represents full profile of a doctor for superadmin
type AdminDoctorDetailDTO struct {
	ID                    uuid.UUID                     `json:"id"`
	FirstName             string                        `json:"first_name"`
	LastName              string                        `json:"last_name"`
	Bio                   string                        `json:"bio"`
	ExperienceYears       int                           `json:"experience_years"`
	PhotoURL              string                        `json:"photo_url"`
	CreatedAt             time.Time                     `json:"created_at"`
	Clinics               []AdminDoctorClinicInfoDTO    `json:"clinics"`
	TotalAppointments     int64                         `json:"total_appointments"`
	UpcomingAppointments  int64                         `json:"upcoming_appointments"`
	CompletedAppointments int64                         `json:"completed_appointments"`
	CancelledAppointments int64                         `json:"cancelled_appointments"`
	RecentAppointments    []AdminAppointmentListItemDTO `json:"recent_appointments"`
}

// AdminAppointmentFilter for querying and paginating all appointments
type AdminAppointmentFilter struct {
	ClinicID      string `form:"clinic_id"`
	DoctorID      string `form:"doctor_id"`
	PatientSearch string `form:"patient_search"`
	Status        string `form:"status"`
	StartDate     string `form:"start_date"`
	EndDate       string `form:"end_date"`
	Page          int    `form:"page"`
	Limit         int    `form:"limit"`
}

func (f *AdminAppointmentFilter) Normalize() {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
}

// AdminAppointmentListItemDTO represents an appointment in the platform-wide table
type AdminAppointmentListItemDTO struct {
	ID                 uuid.UUID `json:"id"`
	PatientID          uuid.UUID `json:"patient_id"`
	PatientName        string    `json:"patient_name"`
	PatientPhone       string    `json:"patient_phone"`
	ClinicID           uuid.UUID `json:"clinic_id"`
	ClinicName         string    `json:"clinic_name"`
	DoctorID           uuid.UUID `json:"doctor_id"`
	DoctorName         string    `json:"doctor_name"`
	SpecialtyName      string    `json:"specialty_name"`
	AppointmentDate    string    `json:"appointment_date"`
	StartTime          string    `json:"start_time"`
	EndTime            string    `json:"end_time"`
	Status             string    `json:"status"`
	ConsultationPrice  float64   `json:"consultation_price"`
	CancellationReason *string   `json:"cancellation_reason,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}

// AdminClinicTopDoctorDTO for clinic-specific analytics
type AdminClinicTopDoctorDTO struct {
	DoctorID      uuid.UUID `json:"doctor_id"`
	DoctorName    string    `json:"doctor_name"`
	SpecialtyName string    `json:"specialty_name"`
	BookingsCount int64     `json:"bookings_count"`
}

// AdminClinicAnalyticsDTO represents clinic-specific performance statistics
type AdminClinicAnalyticsDTO struct {
	TotalAppointments     int64                     `json:"total_appointments"`
	TodayAppointments     int64                     `json:"today_appointments"`
	UpcomingAppointments  int64                     `json:"upcoming_appointments"`
	CompletedAppointments int64                     `json:"completed_appointments"`
	CancelledAppointments int64                     `json:"cancelled_appointments"`
	NoShowAppointments    int64                     `json:"no_show_appointments"`
	EstimatedTotalRevenue float64                   `json:"estimated_total_revenue"`
	StatusDistribution    []StatusCountDTO          `json:"status_distribution"`
	TopDoctors            []AdminClinicTopDoctorDTO `json:"top_doctors"`
}

// AdminClinicDoctorItemDTO represents a doctor inside clinic details
type AdminClinicDoctorItemDTO struct {
	ClinicDoctorID             uuid.UUID `json:"clinic_doctor_id"`
	DoctorID                   uuid.UUID `json:"doctor_id"`
	FirstName                  string    `json:"first_name"`
	LastName                   string    `json:"last_name"`
	SpecialtyName              string    `json:"specialty_name"`
	ConsultationPrice          float64   `json:"consultation_price"`
	AppointmentDurationMinutes int       `json:"appointment_duration_minutes"`
	ExperienceYears            int       `json:"experience_years"`
	IsActive                   bool      `json:"is_active"`
	PhotoURL                   string    `json:"photo_url"`
}

